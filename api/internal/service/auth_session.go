package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"decorebator.com/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	RefreshIdleDuration     = 30 * 24 * time.Hour
	RefreshAbsoluteDuration = 90 * 24 * time.Hour
	refreshTokenBytes       = 32
)

type SessionCredentials struct {
	AccessToken  string
	RefreshToken string
}

type AuthSessionService struct {
	repository *repository.AuthSessionRepository
	access     *AccessTokenService
	now        func() time.Time
	readRandom func([]byte) (int, error)
}

func NewAuthSessionService(
	repository *repository.AuthSessionRepository,
	access *AccessTokenService,
) (*AuthSessionService, error) {
	if repository == nil || access == nil {
		return nil, fmt.Errorf("auth session dependencies are required")
	}
	return &AuthSessionService{repository: repository, access: access, now: time.Now, readRandom: rand.Read}, nil
}

func (s *AuthSessionService) Create(ctx context.Context, userID int64) (SessionCredentials, error) {
	return s.create(ctx, nil, userID)
}

func (s *AuthSessionService) CreateTx(
	ctx context.Context,
	tx pgx.Tx,
	userID int64,
) (SessionCredentials, error) {
	if tx == nil {
		return SessionCredentials{}, fmt.Errorf("auth session transaction is required")
	}
	return s.create(ctx, tx, userID)
}

func (s *AuthSessionService) create(ctx context.Context, tx pgx.Tx, userID int64) (SessionCredentials, error) {
	if userID <= 0 {
		return SessionCredentials{}, fmt.Errorf("positive user ID is required")
	}
	refresh, hash, err := s.newRefreshToken()
	if err != nil {
		return SessionCredentials{}, err
	}
	now := s.now().UTC()
	familyID, tokenID := uuid.New(), uuid.New()
	access, err := s.access.Issue(userID, familyID)
	if err != nil {
		return SessionCredentials{}, fmt.Errorf("issue session access token: %w", err)
	}
	idleExpiry, absoluteExpiry := now.Add(RefreshIdleDuration), now.Add(RefreshAbsoluteDuration)
	var persistErr error
	if tx == nil {
		persistErr = s.repository.Create(ctx, userID, familyID, tokenID, hash[:], now, idleExpiry, absoluteExpiry)
	} else {
		persistErr = s.repository.CreateTx(ctx, tx, userID, familyID, tokenID, hash[:], now, idleExpiry, absoluteExpiry)
	}
	if persistErr != nil {
		return SessionCredentials{}, persistErr
	}
	return SessionCredentials{AccessToken: access, RefreshToken: refresh}, nil
}

func (s *AuthSessionService) Refresh(ctx context.Context, refresh string) (SessionCredentials, error) {
	oldHash, err := parseRefreshToken(refresh)
	if err != nil {
		return SessionCredentials{}, repository.ErrInvalidRefreshToken
	}
	newRefresh, newHash, err := s.newRefreshToken()
	if err != nil {
		return SessionCredentials{}, err
	}
	rotated, err := s.repository.Rotate(
		ctx, oldHash[:], uuid.New(), newHash[:], s.now().UTC(), RefreshIdleDuration,
	)
	if err != nil {
		return SessionCredentials{}, err
	}
	access, err := s.access.Issue(rotated.UserID, rotated.FamilyID)
	if err != nil {
		return SessionCredentials{}, fmt.Errorf("issue rotated access token: %w", err)
	}
	return SessionCredentials{AccessToken: access, RefreshToken: newRefresh}, nil
}

func (s *AuthSessionService) ValidateAccess(ctx context.Context, token string) (AuthIdentity, error) {
	identity, err := s.access.Validate(token)
	if err != nil {
		return AuthIdentity{}, err
	}
	active, err := s.repository.IsActive(ctx, identity.UserID, identity.SessionID, s.now().UTC())
	if err != nil {
		return AuthIdentity{}, fmt.Errorf("read auth session: %w", err)
	}
	if !active {
		return AuthIdentity{}, repository.ErrSessionExpired
	}
	return identity, nil
}

func (s *AuthSessionService) Logout(ctx context.Context, refresh string) error {
	hash, err := parseRefreshToken(refresh)
	if err != nil {
		return nil
	}
	return s.repository.RevokeByRefresh(ctx, hash[:], s.now().UTC(), "logout")
}

func (s *AuthSessionService) RevokeAll(ctx context.Context, userID int64, reason string) error {
	if userID <= 0 || reason == "" {
		return fmt.Errorf("user ID and revocation reason are required")
	}
	return s.repository.RevokeAll(ctx, userID, s.now().UTC(), reason)
}

func (s *AuthSessionService) newRefreshToken() (string, [sha256.Size]byte, error) {
	var raw [refreshTokenBytes]byte
	if _, err := s.readRandom(raw[:]); err != nil {
		return "", [sha256.Size]byte{}, fmt.Errorf("generate refresh token: %w", err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(raw[:])
	return encoded, sha256.Sum256([]byte(encoded)), nil
}

func parseRefreshToken(value string) ([sha256.Size]byte, error) {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(raw) != refreshTokenBytes {
		return [sha256.Size]byte{}, errors.New("invalid refresh token shape")
	}
	return sha256.Sum256([]byte(value)), nil
}
