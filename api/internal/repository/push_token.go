package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PushTokenRepository struct {
	Db *pgxpool.Pool
}

type UpsertPushTokenInput struct {
	UserID    int64
	ExpoToken string
	Platform  string
	DeviceID  *string
	Timezone  string
	Locale    *string
}

func (r *PushTokenRepository) Upsert(ctx context.Context, input UpsertPushTokenInput) error {
	_, err := r.Db.Exec(ctx, `
		INSERT INTO push_tokens (
			user_id, expo_token, platform, device_id, timezone, locale,
			is_active, last_seen_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, true, NOW(), NOW(), NOW())
		ON CONFLICT (expo_token)
		DO UPDATE SET
			user_id = EXCLUDED.user_id,
			platform = EXCLUDED.platform,
			device_id = EXCLUDED.device_id,
			timezone = EXCLUDED.timezone,
			locale = EXCLUDED.locale,
			is_active = true,
			last_seen_at = NOW(),
			updated_at = NOW()
	`, input.UserID, input.ExpoToken, input.Platform, input.DeviceID, input.Timezone, input.Locale)
	return err
}

func (r *PushTokenRepository) Deactivate(ctx context.Context, userID int64, expoToken string) error {
	_, err := r.Db.Exec(ctx, `
		UPDATE push_tokens
		SET is_active = false,
			updated_at = NOW()
		WHERE user_id = $1 AND expo_token = $2
	`, userID, expoToken)
	return err
}

func (r *PushTokenRepository) DeactivateByTokens(ctx context.Context, expoTokens []string) error {
	if len(expoTokens) == 0 {
		return nil
	}
	tokenArray := pgtype.FlatArray[string](expoTokens)
	_, err := r.Db.Exec(ctx, `
		UPDATE push_tokens
		SET is_active = false,
			updated_at = NOW()
		WHERE expo_token = ANY($1)
	`, tokenArray)
	return err
}

func (r *PushTokenRepository) MarkNotified(ctx context.Context, expoTokens []string, notifiedAt time.Time) error {
	if len(expoTokens) == 0 {
		return nil
	}
	tokenArray := pgtype.FlatArray[string](expoTokens)
	_, err := r.Db.Exec(ctx, `
		UPDATE push_tokens
		SET last_notified_at = $2,
			updated_at = NOW()
		WHERE expo_token = ANY($1)
	`, tokenArray, notifiedAt)
	return err
}
