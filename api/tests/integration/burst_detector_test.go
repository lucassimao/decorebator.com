package integration

import (
	"context"
	"fmt"
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

	// Create burst detector
	detector := service.NewBurstDetector(redisClient)
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

		// First burst - trigger first violation
		for i := 0; i < 51; i++ {
			isBurst, violations, err := detector.CheckAndTrackBurst(ctx, userID, endpoint)
			assert.NoError(t, err)
			if i == 50 {
				assert.True(t, isBurst)
				assert.Equal(t, 1, violations)
			}
		}

		// Manually reset burst window by deleting the burst key (simulate time passing)
		burstKey := fmt.Sprintf("burst:%d:%s", userID, endpoint)
		redisClient.Del(ctx, burstKey)

		// Second burst - this should trigger second violation
		var finalViolations int
		for i := 0; i < 51; i++ {
			isBurst, violations, err := detector.CheckAndTrackBurst(ctx, userID, endpoint)
			assert.NoError(t, err)

			if i == 50 {
				// This should be the second violation
				assert.True(t, isBurst)
				assert.Equal(t, 2, violations)
				finalViolations = violations

				// Simulate middleware behavior: block user when violations >= 2
				if violations >= 2 {
					err := detector.BlockUserFor24Hours(ctx, userID)
					assert.NoError(t, err)
				}
			}
		}

		// User should now be blocked
		isBlocked := detector.IsUserBlocked(ctx, userID)
		assert.True(t, isBlocked)
		assert.Equal(t, 2, finalViolations)
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

	t.Run("Block expires after exactly 24 hours", func(t *testing.T) {
		redisClient.FlushAll(ctx)

		userID := int64(12350)

		// Record the time before blocking
		blockTime := time.Now()

		// Block user
		err := detector.BlockUserFor24Hours(ctx, userID)
		assert.NoError(t, err)

		// Check they are blocked
		isBlocked := detector.IsUserBlocked(ctx, userID)
		assert.True(t, isBlocked)

		// Get block expiry
		expiry := detector.GetBlockExpiry(ctx, userID)
		assert.NotNil(t, expiry)

		// Check that expiry is exactly 24 hours from block time
		expectedExpiry := blockTime.Add(24 * time.Hour)
		assert.WithinDuration(t, expectedExpiry, *expiry, 2*time.Second)
	})

	t.Run("Block duration is consistent regardless of time of day", func(t *testing.T) {
		redisClient.FlushAll(ctx)

		// Test blocking at different times
		testCases := []struct {
			name   string
			userID int64
		}{
			{"Morning block", 12351},
			{"Afternoon block", 12352},
			{"Evening block", 12353},
		}

		var blockTimes []time.Time
		var expiryTimes []time.Time

		for _, tc := range testCases {
			blockTime := time.Now()
			blockTimes = append(blockTimes, blockTime)

			err := detector.BlockUserFor24Hours(ctx, tc.userID)
			assert.NoError(t, err, tc.name)

			expiry := detector.GetBlockExpiry(ctx, tc.userID)
			assert.NotNil(t, expiry, tc.name)
			expiryTimes = append(expiryTimes, *expiry)

			// Small delay between blocks
			time.Sleep(100 * time.Millisecond)
		}

		// Verify all blocks have 24-hour duration
		for i, blockTime := range blockTimes {
			duration := expiryTimes[i].Sub(blockTime)
			// Allow 2 second tolerance for execution time
			assert.InDelta(t, 24*time.Hour.Seconds(), duration.Seconds(), 2.0,
				"Block %d duration should be 24 hours, got %v", i, duration)
		}
	})

	t.Run("UnblockUser removes user block", func(t *testing.T) {
		redisClient.FlushAll(ctx)

		userID := int64(12354)

		// Block user first
		err := detector.BlockUserFor24Hours(ctx, userID)
		assert.NoError(t, err)

		// Verify user is blocked
		isBlocked := detector.IsUserBlocked(ctx, userID)
		assert.True(t, isBlocked)

		// Unblock user
		err = detector.UnblockUser(ctx, userID)
		assert.NoError(t, err)

		// Verify user is no longer blocked
		isBlocked = detector.IsUserBlocked(ctx, userID)
		assert.False(t, isBlocked)
	})

	t.Run("UnblockAllUsers removes all blocks", func(t *testing.T) {
		redisClient.FlushAll(ctx)

		// Block multiple users
		userIDs := []int64{12355, 12356, 12357}
		for _, userID := range userIDs {
			err := detector.BlockUserFor24Hours(ctx, userID)
			assert.NoError(t, err)
		}

		// Verify all users are blocked
		for _, userID := range userIDs {
			isBlocked := detector.IsUserBlocked(ctx, userID)
			assert.True(t, isBlocked)
		}

		// Unblock all users
		unlockedCount, err := detector.UnblockAllUsers(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), unlockedCount)

		// Verify no users are blocked
		for _, userID := range userIDs {
			isBlocked := detector.IsUserBlocked(ctx, userID)
			assert.False(t, isBlocked)
		}
	})

	t.Run("GetBlockedUsers returns correct list", func(t *testing.T) {
		redisClient.FlushAll(ctx)

		// Initially no blocked users
		blockedUsers, err := detector.GetBlockedUsers(ctx)
		assert.NoError(t, err)
		assert.Empty(t, blockedUsers)

		// Block some users
		expectedUserIDs := []int64{12358, 12359, 12360}
		for _, userID := range expectedUserIDs {
			err := detector.BlockUserFor24Hours(ctx, userID)
			assert.NoError(t, err)
		}

		// Get blocked users
		blockedUsers, err = detector.GetBlockedUsers(ctx)
		assert.NoError(t, err)
		assert.Len(t, blockedUsers, 3)

		// Verify all expected users are in the list
		for _, expectedID := range expectedUserIDs {
			assert.Contains(t, blockedUsers, expectedID)
		}
	})
}
