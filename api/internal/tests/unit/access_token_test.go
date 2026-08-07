package unit

import (
	"testing"
	"time"

	"decorebator.com/internal/config"
	"decorebator.com/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const authTestKey = "0123456789abcdef0123456789abcdef"

func TestAccessTokenContainsOnlyBoundAuthorizationClaims(t *testing.T) {
	tokens := newTestAccessTokens(t)
	signed, err := tokens.Issue(42)
	require.NoError(t, err)

	claims := jwt.MapClaims{}
	_, _, err = jwt.NewParser().ParseUnverified(signed, claims)
	require.NoError(t, err)
	assert.Equal(t, "42", claims["sub"])
	assert.Equal(t, config.AuthIssuer, claims["iss"])
	assert.Equal(t, "test", claims["environment"])
	assert.Equal(t, "access", claims["tokenType"])
	assert.NotContains(t, claims, "email")
	assert.NotContains(t, claims, "subscriptionPlan")

	userID, err := tokens.Validate(signed)
	require.NoError(t, err)
	assert.Equal(t, int64(42), userID)
}

func TestAccessTokenRejectsWrongAlgorithmAndLegacyOrMismatchedClaims(t *testing.T) {
	tokens := newTestAccessTokens(t)
	now := time.Now().UTC().Truncate(time.Second)
	valid := service.AccessTokenClaims{
		Environment: "test", TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: config.AuthIssuer, Subject: "7", Audience: jwt.ClaimStrings{config.AuthAudience},
			IssuedAt: jwt.NewNumericDate(now), NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(config.AccessTokenDuration)),
		},
	}

	tests := []struct {
		name   string
		claims service.AccessTokenClaims
		method jwt.SigningMethod
		key    string
	}{
		{name: "wrong environment", claims: mutateAccessClaims(valid, func(c *service.AccessTokenClaims) { c.Environment = "production" }), method: jwt.SigningMethodHS256},
		{name: "wrong type", claims: mutateAccessClaims(valid, func(c *service.AccessTokenClaims) { c.TokenType = "refresh" }), method: jwt.SigningMethodHS256},
		{name: "wrong audience", claims: mutateAccessClaims(valid, func(c *service.AccessTokenClaims) { c.Audience = jwt.ClaimStrings{"attacker"} }), method: jwt.SigningMethodHS256},
		{name: "missing not-before", claims: mutateAccessClaims(valid, func(c *service.AccessTokenClaims) { c.NotBefore = nil }), method: jwt.SigningMethodHS256},
		{name: "expired", claims: mutateAccessClaims(valid, func(c *service.AccessTokenClaims) { c.ExpiresAt = jwt.NewNumericDate(now.Add(-time.Minute)) }), method: jwt.SigningMethodHS256},
		{name: "future issued-at", claims: mutateAccessClaims(valid, func(c *service.AccessTokenClaims) {
			c.IssuedAt = jwt.NewNumericDate(now.Add(2 * time.Minute))
			c.NotBefore = jwt.NewNumericDate(now.Add(2 * time.Minute))
			c.ExpiresAt = jwt.NewNumericDate(now.Add(17 * time.Minute))
		}), method: jwt.SigningMethodHS256},
		{name: "overlong lifetime", claims: mutateAccessClaims(valid, func(c *service.AccessTokenClaims) {
			c.ExpiresAt = jwt.NewNumericDate(now.Add(config.AccessTokenDuration + 2*time.Minute))
		}), method: jwt.SigningMethodHS256},
		{name: "nonpositive subject", claims: mutateAccessClaims(valid, func(c *service.AccessTokenClaims) { c.Subject = "0" }), method: jwt.SigningMethodHS256},
		{name: "wrong key", claims: valid, method: jwt.SigningMethodHS256, key: "abcdef0123456789abcdef0123456789"},
		{name: "legacy token", claims: service.AccessTokenClaims{RegisteredClaims: jwt.RegisteredClaims{Issuer: "Decorebator", Subject: "7", IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(365 * 24 * time.Hour))}}, method: jwt.SigningMethodHS256},
		{name: "none algorithm", claims: valid, method: jwt.SigningMethodNone},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			signingKey := test.key
			if signingKey == "" {
				signingKey = authTestKey
			}
			key := any([]byte(signingKey))
			if test.method == jwt.SigningMethodNone {
				key = jwt.UnsafeAllowNoneSignatureType
			}
			signed, err := jwt.NewWithClaims(test.method, test.claims).SignedString(key)
			require.NoError(t, err)
			_, err = tokens.Validate(signed)
			require.Error(t, err)
		})
	}
}

func newTestAccessTokens(t *testing.T) *service.AccessTokenService {
	t.Helper()
	tokens, err := service.NewAccessTokenService(config.AuthConfig{
		SigningKey: []byte(authTestKey), Environment: "test", Issuer: config.AuthIssuer,
		Audience: config.AuthAudience, AccessTTL: config.AccessTokenDuration,
	})
	require.NoError(t, err)
	return tokens
}

func mutateAccessClaims(
	claims service.AccessTokenClaims,
	mutate func(*service.AccessTokenClaims),
) service.AccessTokenClaims {
	mutate(&claims)
	return claims
}
