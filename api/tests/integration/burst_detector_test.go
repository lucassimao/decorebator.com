package integration

import (
	"context"
	"testing"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBurstDetector(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Get Redis client
	redisClient, err := common.GetRedisClient()
	require.NoError(t, err)

	// Get database connection
	db, err := common.GetDBConnection()
	require.NoError(t, err)

	// Create burst detector
	detector := service.NewBurstDetector(redisClient, db)
	ctx := context.Background()

	t.Run("Normal usage should not trigger burst detection", func(t *testing.T) {
		redisClient.FlushAll(ctx)

		userID := int64(12345)
		endpoint := "word_create"

		// Simulate normal usage (well below threshold)
		for i := 0; i < 10; i++ {
			isBurst, violations, err := detector.CheckAndTrackBurst(ctx, userID, endpoint)
			assert.NoError(t, err)
			assert.False(t, isBurst)
			assert.Equal(t, 0, violations)
		}
	})

	t.Run("Burst usage should trigger detection", func(t *testing.T) {
		redisClient.FlushAll(ctx)

		userID := int64(12346)
		endpoint := "word_create"

		// Simulate burst usage (exceeding threshold of 50)
		var firstViolation bool
		var violationCount int

		for i := 0; i < 51; i++ {
			isBurst, violations, err := detector.CheckAndTrackBurst(ctx, userID, endpoint)
			assert.NoError(t, err)

			if i == 50 {
				// 51st request should trigger burst
				assert.True(t, isBurst)
				assert.Equal(t, 1, violations)
				firstViolation = true
				violationCount = violations
			} else {
				assert.False(t, isBurst)
			}
		}

		assert.True(t, firstViolation)
		assert.Equal(t, 1, violationCount)
	})

	t.Run("Second burst violation should result in blocking", func(t *testing.T) {
		redisClient.FlushAll(ctx)

		userID := int64(12347)
		endpoint := "word_create"

		// First burst
		for i := 0; i < 51; i++ {
			detector.CheckAndTrackBurst(ctx, userID, endpoint)
		}

		// Wait for burst window to reset
		time.Sleep(61 * time.Second)

		// Second burst
		for i := 0; i < 51; i++ {
			isBurst, violations, err := detector.CheckAndTrackBurst(ctx, userID, endpoint)
			assert.NoError(t, err)

			if i == 50 {
				// This should be the second violation
				assert.True(t, isBurst)
				assert.Equal(t, 2, violations)
			}
		}

		// User should now be blocked
		isBlocked := detector.IsUserBlocked(ctx, userID)
		assert.True(t, isBlocked)
	})

	t.Run("Different endpoints have different thresholds", func(t *testing.T) {
		redisClient.FlushAll(ctx)

		userID := int64(12348)

		// Test wordlist_create threshold (10)
		for i := 0; i < 11; i++ {
			isBurst, _, err := detector.CheckAndTrackBurst(ctx, userID, "wordlist_create")
			assert.NoError(t, err)

			if i == 10 {
				assert.True(t, isBurst)
			} else {
				assert.False(t, isBurst)
			}
		}

		// Test error_report threshold (20)
		userID = int64(12349)
		for i := 0; i < 21; i++ {
			isBurst, _, err := detector.CheckAndTrackBurst(ctx, userID, "error_report")
			assert.NoError(t, err)

			if i == 20 {
				assert.True(t, isBurst)
			} else {
				assert.False(t, isBurst)
			}
		}
	})

	t.Run("Block expires at midnight", func(t *testing.T) {
		redisClient.FlushAll(ctx)

		userID := int64(12350)

		// Block user
		err := detector.BlockUserForDay(ctx, userID)
		assert.NoError(t, err)

		// Check they are blocked
		isBlocked := detector.IsUserBlocked(ctx, userID)
		assert.True(t, isBlocked)

		// Get block expiry
		expiry := detector.GetBlockExpiry(ctx, userID)
		assert.NotNil(t, expiry)

		// Verify it expires at midnight
		midnight := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
		assert.WithinDuration(t, midnight, *expiry, 2*time.Minute)
	})
}
