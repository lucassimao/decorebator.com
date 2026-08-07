package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenReplay  = errors.New("refresh token replay")
	ErrSessionExpired      = errors.New("session expired")
)

type AuthSessionRepository struct {
	db *pgxpool.Pool
}

func NewAuthSessionRepository(db *pgxpool.Pool) *AuthSessionRepository {
	return &AuthSessionRepository{db: db}
}

func (r *AuthSessionRepository) Create(
	ctx context.Context,
	userID int64,
	familyID, tokenID uuid.UUID,
	tokenHash []byte,
	now, idleExpiry, absoluteExpiry time.Time,
) error {
	return pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		return r.CreateTx(ctx, tx, userID, familyID, tokenID, tokenHash, now, idleExpiry, absoluteExpiry)
	})
}

func (r *AuthSessionRepository) CreateTx(
	ctx context.Context,
	tx pgx.Tx,
	userID int64,
	familyID, tokenID uuid.UUID,
	tokenHash []byte,
	now, idleExpiry, absoluteExpiry time.Time,
) error {
	if _, err := tx.Exec(ctx, `
			INSERT INTO auth_session_families
				(id,user_id,created_at,last_seen_at,idle_expires_at,absolute_expires_at)
			VALUES ($1,$2,$3,$3,$4,$5)
	`, familyID, userID, now, idleExpiry, absoluteExpiry); err != nil {
		return fmt.Errorf("insert auth session family: %w", err)
	}
	if _, err := tx.Exec(ctx, `
			INSERT INTO auth_refresh_tokens (id,family_id,token_hash,created_at)
			VALUES ($1,$2,$3,$4)
	`, tokenID, familyID, tokenHash, now); err != nil {
		return fmt.Errorf("insert auth refresh token: %w", err)
	}
	return nil
}

type RotatedSession struct {
	UserID   int64
	FamilyID uuid.UUID
}

func (r *AuthSessionRepository) Rotate(
	ctx context.Context,
	oldHash []byte,
	newTokenID uuid.UUID,
	newHash []byte,
	now time.Time,
	idleDuration time.Duration,
) (RotatedSession, error) {
	var result RotatedSession
	var outcome error
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var tokenID uuid.UUID
		var consumedAt, revokedAt *time.Time
		var idleExpiry, absoluteExpiry time.Time
		err := tx.QueryRow(ctx, `
			SELECT t.id,t.family_id,f.user_id,t.consumed_at,f.revoked_at,
				f.idle_expires_at,f.absolute_expires_at
			FROM auth_refresh_tokens t
			JOIN auth_session_families f ON f.id=t.family_id
			WHERE t.token_hash=$1
			FOR UPDATE OF t,f
		`, oldHash).Scan(
			&tokenID, &result.FamilyID, &result.UserID, &consumedAt, &revokedAt,
			&idleExpiry, &absoluteExpiry,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			outcome = ErrInvalidRefreshToken
			return nil
		}
		if err != nil {
			return fmt.Errorf("lock auth refresh token: %w", err)
		}
		if consumedAt != nil {
			if _, err := tx.Exec(ctx, `
				UPDATE auth_session_families
				SET revoked_at=COALESCE(revoked_at,$2),
					revocation_reason=COALESCE(revocation_reason,'refresh_replay')
				WHERE id=$1
			`, result.FamilyID, now); err != nil {
				return fmt.Errorf("revoke replayed session family: %w", err)
			}
			outcome = ErrRefreshTokenReplay
			return nil
		}
		if revokedAt != nil || !idleExpiry.After(now) || !absoluteExpiry.After(now) {
			if revokedAt == nil {
				if _, err := tx.Exec(ctx, `
					UPDATE auth_session_families
					SET revoked_at=$2,revocation_reason='expired' WHERE id=$1
				`, result.FamilyID, now); err != nil {
					return fmt.Errorf("expire auth session family: %w", err)
				}
			}
			outcome = ErrSessionExpired
			return nil
		}
		newIdleExpiry := now.Add(idleDuration)
		if newIdleExpiry.After(absoluteExpiry) {
			newIdleExpiry = absoluteExpiry
		}
		if _, err := tx.Exec(ctx, `
			UPDATE auth_refresh_tokens SET consumed_at=$2 WHERE id=$1
		`, tokenID, now); err != nil {
			return fmt.Errorf("consume auth refresh token: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO auth_refresh_tokens (id,family_id,token_hash,created_at)
			VALUES ($1,$2,$3,$4)
		`, newTokenID, result.FamilyID, newHash, now); err != nil {
			return fmt.Errorf("insert rotated auth refresh token: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE auth_refresh_tokens SET replacement_id=$2 WHERE id=$1
		`, tokenID, newTokenID); err != nil {
			return fmt.Errorf("link rotated auth refresh token: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE auth_session_families
			SET last_seen_at=$2,idle_expires_at=$3 WHERE id=$1
		`, result.FamilyID, now, newIdleExpiry); err != nil {
			return fmt.Errorf("advance auth session family: %w", err)
		}
		return nil
	})
	if err != nil {
		return RotatedSession{}, err
	}
	return result, outcome
}

func (r *AuthSessionRepository) IsActive(
	ctx context.Context,
	userID int64,
	familyID uuid.UUID,
	now time.Time,
) (bool, error) {
	var active bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM auth_session_families
			WHERE id=$1 AND user_id=$2 AND revoked_at IS NULL
				AND idle_expires_at>$3 AND absolute_expires_at>$3
		)
	`, familyID, userID, now).Scan(&active)
	return active, err
}

func (r *AuthSessionRepository) RevokeByRefresh(
	ctx context.Context,
	hash []byte,
	now time.Time,
	reason string,
) error {
	_, err := r.db.Exec(ctx, `
		UPDATE auth_session_families f
		SET revoked_at=COALESCE(f.revoked_at,$2),
			revocation_reason=COALESCE(f.revocation_reason,$3)
		FROM auth_refresh_tokens t
		WHERE t.family_id=f.id AND t.token_hash=$1
	`, hash, now, reason)
	return err
}

func (r *AuthSessionRepository) RevokeAll(
	ctx context.Context,
	userID int64,
	now time.Time,
	reason string,
) error {
	_, err := r.db.Exec(ctx, `
		UPDATE auth_session_families
		SET revoked_at=COALESCE(revoked_at,$2),
			revocation_reason=COALESCE(revocation_reason,$3)
		WHERE user_id=$1
	`, userID, now, reason)
	return err
}
