package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthRateLimiterSeparatesHashedAccountAndSourceBuckets(t *testing.T) {
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	limiter, err := NewAuthRateLimiter(nil, false, "test", func() time.Time { return now })
	require.NoError(t, err)
	assert.NotContains(t, limiter.key(AuthLimitLogin, AuthLimitAccount, "person@example.com"), "person@example.com")

	for range 5 {
		decision, checkErr := limiter.Check(context.Background(), AuthLimitLogin, AuthLimitAccount, "person@example.com")
		require.NoError(t, checkErr)
		assert.True(t, decision.Allowed)
	}
	decision, err := limiter.Check(context.Background(), AuthLimitLogin, AuthLimitAccount, "person@example.com")
	require.NoError(t, err)
	assert.False(t, decision.Allowed)
	assert.Equal(t, 10*time.Minute, decision.RetryAfter)

	otherAccount, err := limiter.Check(context.Background(), AuthLimitLogin, AuthLimitAccount, "other@example.com")
	require.NoError(t, err)
	assert.True(t, otherAccount.Allowed)
	source, err := limiter.Check(context.Background(), AuthLimitLogin, AuthLimitSource, "192.0.2.1")
	require.NoError(t, err)
	assert.True(t, source.Allowed)

	now = now.Add(10 * time.Minute)
	reset, err := limiter.Check(context.Background(), AuthLimitLogin, AuthLimitAccount, "person@example.com")
	require.NoError(t, err)
	assert.True(t, reset.Allowed)
	assert.False(t, strings.Contains(limiter.key(AuthLimitLogin, AuthLimitSource, "192.0.2.1"), "192.0.2.1"))
}

func TestAuthRateLimiterFailsClosedWhenRedisIsRequired(t *testing.T) {
	_, err := NewAuthRateLimiter(nil, true, "production", nil)
	require.Error(t, err)
}

func TestAuthRateLimiterReleasesInfrastructureReservationsAndClearsSuccess(t *testing.T) {
	limiter, err := NewAuthRateLimiter(nil, false, "test", nil)
	require.NoError(t, err)
	const account = "reservation@example.com"

	for range 5 {
		decision, checkErr := limiter.Check(context.Background(), AuthLimitLogin, AuthLimitAccount, account)
		require.NoError(t, checkErr)
		require.True(t, decision.Allowed)
		require.NoError(t, limiter.Release(context.Background(), AuthLimitLogin, AuthLimitAccount, account))
	}
	decision, err := limiter.Check(context.Background(), AuthLimitLogin, AuthLimitAccount, account)
	require.NoError(t, err)
	require.True(t, decision.Allowed, "refunded infrastructure attempts must not poison the bucket")
	require.NoError(t, limiter.Clear(context.Background(), AuthLimitLogin, AuthLimitAccount, account))

	for range 5 {
		decision, err = limiter.Check(context.Background(), AuthLimitLogin, AuthLimitAccount, account)
		require.NoError(t, err)
		require.True(t, decision.Allowed)
	}
	decision, err = limiter.Check(context.Background(), AuthLimitLogin, AuthLimitAccount, account)
	require.NoError(t, err)
	require.False(t, decision.Allowed)
}

func TestAuthRateLimiterNamespaceIsFleetStableOutsideTests(t *testing.T) {
	first, err := NewAuthRateLimiter(nil, false, "staging", nil)
	require.NoError(t, err)
	second, err := NewAuthRateLimiter(nil, false, "staging", nil)
	require.NoError(t, err)
	assert.Equal(t,
		first.key(AuthLimitLogin, AuthLimitAccount, "person@example.com"),
		second.key(AuthLimitLogin, AuthLimitAccount, "person@example.com"),
	)

	testFirst, err := NewAuthRateLimiter(nil, false, "test", nil)
	require.NoError(t, err)
	testSecond, err := NewAuthRateLimiter(nil, false, "test", nil)
	require.NoError(t, err)
	assert.NotEqual(t,
		testFirst.key(AuthLimitLogin, AuthLimitAccount, "person@example.com"),
		testSecond.key(AuthLimitLogin, AuthLimitAccount, "person@example.com"),
	)
}
