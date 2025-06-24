package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"decorebator.com/internal/common"
	"github.com/redis/go-redis/v9"
)

// BurstThresholds defines the maximum allowed requests per minute for each endpoint
var BurstThresholds = map[string]int64{
	"word_create":     50, // 50 words in 1 minute
	"wordlist_create": 10, // 10 wordlists in 1 minute
	"error_report":    20, // 20 error reports in 1 minute
}

// BurstDetector handles burst detection and progressive blocking
type BurstDetector struct {
	redis *redis.Client
}

// NewBurstDetector creates a new burst detector instance
func NewBurstDetector(redis *redis.Client) *BurstDetector {
	return &BurstDetector{
		redis: redis,
	}
}

// CheckAndTrackBurst checks if the current request is part of a burst pattern and tracks violations
func (b *BurstDetector) CheckAndTrackBurst(ctx context.Context, userID int64, endpoint string) (isBurst bool, dailyViolations int, err error) {
	// Check current burst
	key := fmt.Sprintf("burst:%d:%s", userID, endpoint)
	count, err := b.redis.Incr(ctx, key).Result()
	if err != nil {
		common.Logger.Error("Failed to increment burst counter", "error", err, "user_id", userID, "endpoint", endpoint)
		return false, 0, err
	}

	// Set expiry on first increment
	if count == 1 {
		if err := b.redis.Expire(ctx, key, 60*time.Second).Err(); err != nil {
			common.Logger.Error("Failed to set burst counter expiry", "error", err)
		}
	}

	// Check against threshold
	threshold, exists := BurstThresholds[endpoint]
	if !exists {
		// If no threshold defined, allow the request
		return false, 0, nil
	}

	isBurst = count > threshold

	if isBurst {
		// Track daily violations
		violationKey := fmt.Sprintf("violations:%d:%s", userID, time.Now().Format("2006-01-02"))
		dailyViolations64, err := b.redis.Incr(ctx, violationKey).Result()
		if err != nil {
			common.Logger.Error("Failed to increment violation counter", "error", err)
			return isBurst, 0, err
		}
		dailyViolations = int(dailyViolations64)

		// Set expiry on first violation of the day
		if dailyViolations == 1 {
			if err := b.redis.Expire(ctx, violationKey, 24*time.Hour).Err(); err != nil {
				common.Logger.Error("Failed to set violation counter expiry", "error", err)
			}
		}

		common.Logger.Warn("Burst detected",
			"user_id", userID,
			"endpoint", endpoint,
			"request_count", count,
			"threshold", threshold,
			"daily_violations", dailyViolations,
		)
	}

	return isBurst, dailyViolations, nil
}

// IsUserBlocked checks if user is blocked
func (b *BurstDetector) IsUserBlocked(ctx context.Context, userID int64) bool {
	key := fmt.Sprintf("blocked:%d", userID)
	exists, err := b.redis.Exists(ctx, key).Result()
	if err != nil {
		common.Logger.Error("Failed to check user block status", "error", err, "user_id", userID)
		// In case of error, allow the request
		return false
	}
	return exists > 0
}

// BlockUserFor24Hours blocks the user for exactly 24 hours
func (b *BurstDetector) BlockUserFor24Hours(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("blocked:%d", userID)

	// Block for exactly 24 hours from now
	now := time.Now()
	blockedUntil := now.Add(24 * time.Hour)
	ttl := 24 * time.Hour

	// Store block information
	blockData := map[string]interface{}{
		"blocked_at":    now.Unix(),
		"blocked_until": blockedUntil.Unix(),
	}

	// Use pipeline for atomic operations
	pipe := b.redis.Pipeline()
	pipe.HMSet(ctx, key, blockData)
	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		common.Logger.Error("Failed to block user", "error", err, "user_id", userID, "ttl", ttl)
		return err
	}

	common.Logger.Info("User blocked for burst abuse",
		"user_id", userID,
		"blocked_until", blockedUntil.Format(time.RFC3339),
		"ttl_hours", ttl.Hours(),
	)

	return nil
}

// GetBlockExpiry returns when the user's block will expire
func (b *BurstDetector) GetBlockExpiry(ctx context.Context, userID int64) *time.Time {
	key := fmt.Sprintf("blocked:%d", userID)

	// Get the stored blocked_until timestamp
	blockedUntilStr, err := b.redis.HGet(ctx, key, "blocked_until").Result()
	if err != nil {
		// If we can't get the exact time, try to calculate from TTL
		ttl, ttlErr := b.redis.TTL(ctx, key).Result()
		if ttlErr != nil || ttl <= 0 {
			return nil
		}
		expiry := time.Now().Add(ttl)
		return &expiry
	}

	// Parse the unix timestamp
	blockedUntilUnix, err := strconv.ParseInt(blockedUntilStr, 10, 64)
	if err != nil {
		return nil
	}

	expiry := time.Unix(blockedUntilUnix, 0)
	return &expiry
}

// GetBurstCount returns the current burst count for a user and endpoint
func (b *BurstDetector) GetBurstCount(ctx context.Context, userID int64, endpoint string) (int64, error) {
	key := fmt.Sprintf("burst:%d:%s", userID, endpoint)
	count, err := b.redis.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

// GetDailyViolations returns the number of violations for a user today
func (b *BurstDetector) GetDailyViolations(ctx context.Context, userID int64) (int64, error) {
	violationKey := fmt.Sprintf("violations:%d:%s", userID, time.Now().Format("2006-01-02"))
	count, err := b.redis.Get(ctx, violationKey).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

// UnblockUser removes the block for a specific user
func (b *BurstDetector) UnblockUser(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("blocked:%d", userID)

	err := b.redis.Del(ctx, key).Err()
	if err != nil {
		common.Logger.Error("Failed to unblock user", "error", err, "user_id", userID)
		return err
	}

	common.Logger.Info("User unblocked successfully", "user_id", userID)
	return nil
}

// UnblockAllUsers removes all user blocks
func (b *BurstDetector) UnblockAllUsers(ctx context.Context) (int64, error) {
	// Find all blocked user keys
	pattern := "blocked:*"
	keys, err := b.redis.Keys(ctx, pattern).Result()
	if err != nil {
		common.Logger.Error("Failed to find blocked user keys", "error", err)
		return 0, err
	}

	if len(keys) == 0 {
		common.Logger.Info("No blocked users found")
		return 0, nil
	}

	// Delete all blocked user keys
	deleted, err := b.redis.Del(ctx, keys...).Result()
	if err != nil {
		common.Logger.Error("Failed to unblock all users", "error", err, "keys_count", len(keys))
		return 0, err
	}

	common.Logger.Info("All users unblocked successfully", "unblocked_count", deleted)
	return deleted, nil
}

// GetBlockedUsers returns a list of all currently blocked user IDs
func (b *BurstDetector) GetBlockedUsers(ctx context.Context) ([]int64, error) {
	pattern := "blocked:*"
	keys, err := b.redis.Keys(ctx, pattern).Result()
	if err != nil {
		common.Logger.Error("Failed to find blocked user keys", "error", err)
		return nil, err
	}

	var blockedUsers []int64
	for _, key := range keys {
		// Extract user ID from key format "blocked:123"
		if len(key) > 8 { // "blocked:" is 8 characters
			userIDStr := key[8:] // Remove "blocked:" prefix
			if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
				blockedUsers = append(blockedUsers, userID)
			}
		}
	}

	return blockedUsers, nil
}
