package service

import (
	"context"
	"fmt"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ErrorReportRateLimitService struct {
	repo *repository.ErrorReportRepository
}

func NewErrorReportRateLimitService(db *pgxpool.Pool) *ErrorReportRateLimitService {
	return &ErrorReportRateLimitService{
		repo: repository.NewErrorReportRepository(db),
	}
}

// Rate limit configuration
const (
	// Free tier limits
	FreeHourlyLimit = 3
	FreeDailyLimit  = 5

	// Premium tier limits
	PremiumHourlyLimit = 10
	PremiumDailyLimit  = 30

	// Time windows
	HourlyWindow = time.Hour
	DailyWindow  = 24 * time.Hour
)

type RateLimitError struct {
	Message    string
	RetryAfter time.Duration
	Limit      int
	Remaining  int
	WindowType string // "hourly" or "daily"
}

func (e RateLimitError) Error() string {
	return e.Message
}

// CheckRateLimit checks if a user has exceeded their error report rate limits
func (s *ErrorReportRateLimitService) CheckRateLimit(ctx context.Context, user *model.User) error {
	// Determine limits based on subscription
	hourlyLimit := FreeHourlyLimit
	dailyLimit := FreeDailyLimit
	if user.SubscriptionPlan == model.PlanMonthly || user.SubscriptionPlan == model.PlanAnnual {
		hourlyLimit = PremiumHourlyLimit
		dailyLimit = PremiumDailyLimit
	}

	now := time.Now()
	hourAgo := now.Add(-HourlyWindow)
	dayAgo := now.Add(-DailyWindow)

	// Check hourly limit
	hourlyCount, err := s.repo.CountUserReportsInWindow(ctx, user.ID, hourAgo)
	if err != nil {
		common.Logger.Error("Failed to check hourly rate limit", "error", err)
		return err
	}

	if hourlyCount >= hourlyLimit {
		// Find when the oldest report in the hour window will expire
		oldestReportTime, err := s.repo.GetOldestReportTimeInWindow(ctx, user.ID, hourAgo)

		retryAfter := time.Hour
		if err == nil && !oldestReportTime.IsZero() {
			retryAfter = oldestReportTime.Add(HourlyWindow).Sub(now)
		}

		// Log rate limit hit for monitoring
		common.Logger.Warn("Error report rate limit hit",
			"userId", user.ID,
			"limitType", "hourly",
			"limit", hourlyLimit,
			"attempts", hourlyCount,
			"subscription", user.SubscriptionPlan,
			"action", "rate_limit_hourly_exceeded",
		)

		return RateLimitError{
			Message:    fmt.Sprintf("Hourly limit exceeded. You can report %d errors per hour.", hourlyLimit),
			RetryAfter: retryAfter,
			Limit:      hourlyLimit,
			Remaining:  0,
			WindowType: "hourly",
		}
	}

	// Check daily limit
	dailyCount, err := s.repo.CountUserReportsInWindow(ctx, user.ID, dayAgo)
	if err != nil {
		common.Logger.Error("Failed to check daily rate limit", "error", err)
		return nil // Don't block on database failures
	}

	if dailyCount >= dailyLimit {
		// Find when the oldest report in the day window will expire
		oldestReportTime, err := s.repo.GetOldestReportTimeInWindow(ctx, user.ID, dayAgo)

		retryAfter := DailyWindow
		if err == nil && !oldestReportTime.IsZero() {
			retryAfter = oldestReportTime.Add(DailyWindow).Sub(now)
		}

		// Log rate limit hit for monitoring
		common.Logger.Warn("Error report rate limit hit",
			"userId", user.ID,
			"limitType", "daily",
			"limit", dailyLimit,
			"attempts", dailyCount,
			"subscription", user.SubscriptionPlan,
			"action", "rate_limit_daily_exceeded",
		)

		return RateLimitError{
			Message:    fmt.Sprintf("Daily limit exceeded. You can report %d errors per day.", dailyLimit),
			RetryAfter: retryAfter,
			Limit:      dailyLimit,
			Remaining:  0,
			WindowType: "daily",
		}
	}

	// Log successful rate limit check
	common.Logger.Info("Error report rate limit check passed",
		"userId", user.ID,
		"hourlyCount", hourlyCount,
		"hourlyLimit", hourlyLimit,
		"dailyCount", dailyCount,
		"dailyLimit", dailyLimit,
	)

	// Update rate limit tracking table (for backup/analytics)
	err = s.repo.UpdateRateLimitTracking(ctx, user.ID, hourlyCount, dailyCount, now)
	if err != nil {
		// Log but don't fail
		common.Logger.Error("Failed to update rate limit tracking", "error", err)
	}

	return nil
}

// GetRateLimitStatus returns the current rate limit status for a user
func (s *ErrorReportRateLimitService) GetRateLimitStatus(ctx context.Context, user *model.User) (map[string]interface{}, error) {
	// Determine limits based on subscription
	hourlyLimit := FreeHourlyLimit
	dailyLimit := FreeDailyLimit
	if user.SubscriptionPlan == model.PlanMonthly || user.SubscriptionPlan == model.PlanAnnual {
		hourlyLimit = PremiumHourlyLimit
		dailyLimit = PremiumDailyLimit
	}

	now := time.Now()
	hourAgo := now.Add(-HourlyWindow)
	dayAgo := now.Add(-DailyWindow)

	// Get current counts
	hourlyCount, dailyCount, err := s.repo.CountUserReportsInWindows(ctx, user.ID, hourAgo, dayAgo)
	if err != nil {
		hourlyCount = 0
		dailyCount = 0
	}

	// Calculate reset times
	var hourlyResetIn, dailyResetIn int

	if hourlyCount > 0 {
		oldestHourly, err := s.repo.GetOldestReportTimeInWindow(ctx, user.ID, hourAgo)
		if err == nil && !oldestHourly.IsZero() {
			hourlyResetIn = int(oldestHourly.Add(HourlyWindow).Sub(now).Seconds())
		}
	}

	if dailyCount > 0 {
		oldestDaily, err := s.repo.GetOldestReportTimeInWindow(ctx, user.ID, dayAgo)
		if err == nil && !oldestDaily.IsZero() {
			dailyResetIn = int(oldestDaily.Add(DailyWindow).Sub(now).Seconds())
		}
	}

	return map[string]interface{}{
		"hourly": map[string]interface{}{
			"limit":     hourlyLimit,
			"used":      hourlyCount,
			"remaining": max(0, hourlyLimit-hourlyCount),
			"resetsIn":  max(0, hourlyResetIn),
		},
		"daily": map[string]interface{}{
			"limit":     dailyLimit,
			"used":      dailyCount,
			"remaining": max(0, dailyLimit-dailyCount),
			"resetsIn":  max(0, dailyResetIn),
		},
	}, nil
}

// GetUserActiveCooldowns returns all active cooldowns for a user
func (s *ErrorReportRateLimitService) GetUserActiveCooldowns(ctx context.Context, userID int64) ([]map[string]interface{}, error) {
	cooldowns, err := s.repo.GetUserActiveCooldowns(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(cooldowns))

	for i, c := range cooldowns {
		result[i] = map[string]interface{}{
			"wordId":           c.WordID,
			"definitionId":     c.DefinitionID,
			"errorType":        c.ErrorType,
			"cooldownUntil":    c.CooldownUntil,
			"remainingSeconds": int(time.Until(c.CooldownUntil).Seconds()),
		}
	}

	return result, nil
}

// GetErrorReportStats returns error reporting statistics for monitoring
func (s *ErrorReportRateLimitService) GetErrorReportStats(ctx context.Context, hours int) (*repository.ErrorReportStats, error) {
	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	return s.repo.GetErrorReportStats(ctx, startTime, endTime)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
