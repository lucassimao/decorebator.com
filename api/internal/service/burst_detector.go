package service

import (
	"context"
	"fmt"
	"time"

	"decorebator.com/internal/common"
	"github.com/jackc/pgx/v5/pgxpool"
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
	db    *pgxpool.Pool
}

// NewBurstDetector creates a new burst detector instance
func NewBurstDetector(redis *redis.Client, db *pgxpool.Pool) *BurstDetector {
	return &BurstDetector{
		redis: redis,
		db:    db,
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

// IsUserBlocked checks if user is blocked for the current day
func (b *BurstDetector) IsUserBlocked(ctx context.Context, userID int64) bool {
	key := fmt.Sprintf("blocked:%d:%s", userID, time.Now().Format("2006-01-02"))
	exists, err := b.redis.Exists(ctx, key).Result()
	if err != nil {
		common.Logger.Error("Failed to check user block status", "error", err, "user_id", userID)
		// In case of error, allow the request
		return false
	}
	return exists > 0
}

// BlockUserForDay blocks the user until midnight
func (b *BurstDetector) BlockUserForDay(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("blocked:%d:%s", userID, time.Now().Format("2006-01-02"))

	// Calculate time until midnight
	now := time.Now()
	midnight := now.Truncate(24 * time.Hour).Add(24 * time.Hour)
	ttl := time.Until(midnight)

	err := b.redis.Set(ctx, key, "blocked", ttl).Err()
	if err != nil {
		common.Logger.Error("Failed to block user", "error", err, "user_id", userID, "ttl", ttl)
		return err
	}

	common.Logger.Info("User blocked for burst abuse",
		"user_id", userID,
		"blocked_until", midnight.Format(time.RFC3339),
		"ttl_hours", ttl.Hours(),
	)

	return nil
}

// GetBlockExpiry returns when the user's block will expire (midnight)
func (b *BurstDetector) GetBlockExpiry(ctx context.Context, userID int64) *time.Time {
	key := fmt.Sprintf("blocked:%d:%s", userID, time.Now().Format("2006-01-02"))
	ttl, err := b.redis.TTL(ctx, key).Result()
	if err != nil || ttl <= 0 {
		return nil
	}

	expiry := time.Now().Add(ttl)
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

