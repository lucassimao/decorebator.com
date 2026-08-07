package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"decorebator.com/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidAccessToken = errors.New("invalid access token")

const (
	accessTokenType = "access"
	accessClockSkew = 30 * time.Second
)

type AccessTokenClaims struct {
	Environment string `json:"environment"`
	TokenType   string `json:"tokenType"`
	SessionID   string `json:"sid"`
	jwt.RegisteredClaims
}

type AuthIdentity struct {
	UserID    int64
	SessionID uuid.UUID
}

type AccessTokenService struct {
	key         []byte
	environment string
	issuer      string
	audience    string
	ttl         time.Duration
	now         func() time.Time
}

func NewAccessTokenService(authConfig config.AuthConfig) (*AccessTokenService, error) {
	if len(authConfig.SigningKey) < 32 || authConfig.Environment == "" || authConfig.Issuer == "" ||
		authConfig.Audience == "" || authConfig.AccessTTL <= 0 {
		return nil, fmt.Errorf("valid access-token configuration is required")
	}
	return &AccessTokenService{
		key: append([]byte(nil), authConfig.SigningKey...), environment: authConfig.Environment,
		issuer: authConfig.Issuer, audience: authConfig.Audience, ttl: authConfig.AccessTTL,
		now: time.Now,
	}, nil
}

func (s *AccessTokenService) Issue(userID int64, sessionID uuid.UUID) (string, error) {
	if s == nil || userID <= 0 || sessionID == uuid.Nil {
		return "", fmt.Errorf("positive user ID and session ID are required")
	}
	now := s.now().UTC().Truncate(time.Second)
	claims := AccessTokenClaims{
		Environment: s.environment,
		TokenType:   accessTokenType,
		SessionID:   sessionID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: s.issuer, Subject: strconv.FormatInt(userID, 10),
			Audience: jwt.ClaimStrings{s.audience},
			IssuedAt: jwt.NewNumericDate(now), NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.key)
}

func (s *AccessTokenService) Validate(tokenString string) (AuthIdentity, error) {
	if s == nil || tokenString == "" {
		return AuthIdentity{}, ErrInvalidAccessToken
	}
	claims := &AccessTokenClaims{}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(accessClockSkew),
		jwt.WithTimeFunc(func() time.Time { return s.now().UTC() }),
	)
	token, err := parser.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.key, nil
	})
	if err != nil || !token.Valid {
		return AuthIdentity{}, ErrInvalidAccessToken
	}
	return s.validateClaims(claims)
}

func (s *AccessTokenService) validateClaims(claims *AccessTokenClaims) (AuthIdentity, error) {
	if claims.Environment != s.environment || claims.TokenType != accessTokenType ||
		claims.IssuedAt == nil || claims.NotBefore == nil || claims.ExpiresAt == nil {
		return AuthIdentity{}, ErrInvalidAccessToken
	}
	now := s.now().UTC()
	issuedAt, notBefore, expiresAt := claims.IssuedAt.Time, claims.NotBefore.Time, claims.ExpiresAt.Time
	if issuedAt.After(now.Add(accessClockSkew)) || notBefore.Before(issuedAt.Add(-accessClockSkew)) ||
		notBefore.After(now.Add(accessClockSkew)) || !expiresAt.After(now) ||
		expiresAt.Sub(issuedAt) > s.ttl+accessClockSkew || !expiresAt.After(issuedAt) {
		return AuthIdentity{}, ErrInvalidAccessToken
	}
	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || userID <= 0 {
		return AuthIdentity{}, ErrInvalidAccessToken
	}
	sessionID, err := uuid.Parse(claims.SessionID)
	if err != nil || sessionID == uuid.Nil {
		return AuthIdentity{}, ErrInvalidAccessToken
	}
	return AuthIdentity{UserID: userID, SessionID: sessionID}, nil
}
