package service

import (
	"context"
	"fmt"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
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

// ErrorReportQuotaUnavailableError means a durable quota read or write could
// not be completed. Callers must fail closed and may safely retry later; it
// intentionally does not expose the underlying storage error to clients.
type ErrorReportQuotaUnavailableError struct {
	cause   error
	message string
}

func (e ErrorReportQuotaUnavailableError) Error() string {
	if e.message != "" {
		return e.message
	}
	return "error report quota temporarily unavailable"
}

func (e ErrorReportQuotaUnavailableError) Unwrap() error {
	return e.cause
}

// CheckRateLimit checks if a user has exceeded their error report rate limits
func (s *ErrorReportRateLimitService) CheckRateLimit(ctx context.Context, user *model.User) error {
	usage, err := s.repo.GetQuotaUsage(ctx, user.ID)
	if err != nil {
		common.Logger.Error("failed to read error report quota", "error", err)
		return ErrorReportQuotaUnavailableError{cause: err}
	}
	return checkErrorReportQuota(user, usage)
}

// CheckRateLimitTx performs the authoritative check inside the error report
// transaction. The transaction must hold LockUserQuota first so concurrent
// submissions cannot pass the same count before either records its event.
func (s *ErrorReportRateLimitService) CheckRateLimitTx(ctx context.Context, tx pgx.Tx, user *model.User) (repository.ErrorReportQuotaUsage, error) {
	usage, err := s.repo.GetQuotaUsageTx(ctx, tx, user.ID)
	if err != nil {
		common.Logger.Error("failed to read error report quota", "error", err)
		return repository.ErrorReportQuotaUsage{}, ErrorReportQuotaUnavailableError{cause: err}
	}
	if err := checkErrorReportQuota(user, usage); err != nil {
		return repository.ErrorReportQuotaUsage{}, err
	}
	return usage, nil
}

// LockUserQuota serializes the authoritative quota check with its committed
// event write and retrieves the current subscription tier under that same row
// lock. A missing user is an unavailable/invalid authenticated state, so
// callers must not continue with an unbounded report submission.
func (s *ErrorReportRateLimitService) LockUserQuota(ctx context.Context, tx pgx.Tx, userID int64) (*model.User, error) {
	var subscriptionPlan model.SubscriptionPlan
	err := tx.QueryRow(ctx, `
		SELECT subscription_plan
		FROM users
		WHERE id=$1
		FOR UPDATE
	`, userID).Scan(&subscriptionPlan)
	if err != nil {
		common.Logger.Error("failed to lock error report quota", "error", err)
		return nil, ErrorReportQuotaUnavailableError{cause: err}
	}
	return &model.User{ID: userID, SubscriptionPlan: subscriptionPlan}, nil
}

// RecordCommittedReportTx records one successful submission and updates the
// derived legacy counters in the same transaction. If either storage write
// fails, the caller rolls every report side effect back rather than allowing a
// counter bypass.
func (s *ErrorReportRateLimitService) RecordCommittedReportTx(ctx context.Context, tx pgx.Tx, userID int64, usage repository.ErrorReportQuotaUsage) error {
	if err := s.repo.RecordQuotaEventTx(ctx, tx, userID); err != nil {
		common.Logger.Error("failed to record error report quota event", "error", err)
		return ErrorReportQuotaUnavailableError{cause: err}
	}
	if err := s.repo.UpdateRateLimitTrackingTx(ctx, tx, userID, usage.HourlyCount+1, usage.DailyCount+1, usage.CheckedAt); err != nil {
		common.Logger.Error("failed to update error report quota counters", "error", err)
		return ErrorReportQuotaUnavailableError{cause: err}
	}
	return nil
}

func checkErrorReportQuota(user *model.User, usage repository.ErrorReportQuotaUsage) error {
	hourlyLimit, dailyLimit := errorReportLimitsFor(user)
	if usage.HourlyCount >= hourlyLimit {
		return RateLimitError{
			Message:    fmt.Sprintf("Hourly limit exceeded. You can report %d errors per hour.", hourlyLimit),
			RetryAfter: quotaRetryAfter(usage.OldestHourly, HourlyWindow, usage.CheckedAt),
			Limit:      hourlyLimit,
			Remaining:  0,
			WindowType: "hourly",
		}
	}
	if usage.DailyCount >= dailyLimit {
		return RateLimitError{
			Message:    fmt.Sprintf("Daily limit exceeded. You can report %d errors per day.", dailyLimit),
			RetryAfter: quotaRetryAfter(usage.OldestDaily, DailyWindow, usage.CheckedAt),
			Limit:      dailyLimit,
			Remaining:  0,
			WindowType: "daily",
		}
	}
	return nil
}

func errorReportLimitsFor(user *model.User) (hourlyLimit, dailyLimit int) {
	if user.SubscriptionPlan == model.PlanMonthly || user.SubscriptionPlan == model.PlanAnnual {
		return PremiumHourlyLimit, PremiumDailyLimit
	}
	return FreeHourlyLimit, FreeDailyLimit
}

func quotaRetryAfter(oldest *time.Time, window time.Duration, now time.Time) time.Duration {
	if oldest == nil {
		return window
	}
	retryAfter := oldest.Add(window).Sub(now)
	if retryAfter < time.Second {
		return time.Second
	}
	return retryAfter
}

// GetRateLimitStatus returns the current rate limit status for a user
func (s *ErrorReportRateLimitService) GetRateLimitStatus(ctx context.Context, user *model.User) (map[string]interface{}, error) {
	hourlyLimit, dailyLimit := errorReportLimitsFor(user)
	usage, err := s.repo.GetQuotaUsage(ctx, user.ID)
	if err != nil {
		return nil, ErrorReportQuotaUnavailableError{cause: err}
	}

	hourlyResetIn := 0
	if usage.OldestHourly != nil {
		hourlyResetIn = int(quotaRetryAfter(usage.OldestHourly, HourlyWindow, usage.CheckedAt).Seconds())
	}
	dailyResetIn := 0
	if usage.OldestDaily != nil {
		dailyResetIn = int(quotaRetryAfter(usage.OldestDaily, DailyWindow, usage.CheckedAt).Seconds())
	}

	return map[string]interface{}{
		"hourly": map[string]interface{}{
			"limit":     hourlyLimit,
			"used":      usage.HourlyCount,
			"remaining": max(0, hourlyLimit-usage.HourlyCount),
			"resetsIn":  max(0, hourlyResetIn),
		},
		"daily": map[string]interface{}{
			"limit":     dailyLimit,
			"used":      usage.DailyCount,
			"remaining": max(0, dailyLimit-usage.DailyCount),
			"resetsIn":  max(0, dailyResetIn),
		},
	}, nil
}

// ErrorReportCooldownCursor is the stable keyset for active cooldowns.
type ErrorReportCooldownCursor = repository.ErrorReportCooldownCursor

// GetUserActiveCooldowns returns a bounded, keyset-paginated cooldown page.
// It fetches one extra row so callers can publish a next cursor without a
// count query.
func (s *ErrorReportRateLimitService) GetUserActiveCooldowns(ctx context.Context, userID int64, limit int, cursor *ErrorReportCooldownCursor) ([]map[string]interface{}, *ErrorReportCooldownCursor, error) {
	cooldowns, err := s.repo.GetUserActiveCooldowns(ctx, userID, limit+1, cursor)
	if err != nil {
		return nil, nil, err
	}

	var next *ErrorReportCooldownCursor
	if len(cooldowns) > limit {
		lastVisible := cooldowns[limit-1]
		next = &ErrorReportCooldownCursor{
			CooldownUntil: lastVisible.CooldownUntil,
			WordID:        lastVisible.WordID,
			DefinitionID:  lastVisible.DefinitionID,
			ErrorType:     lastVisible.ErrorType,
		}
		cooldowns = cooldowns[:limit]
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

	return result, next, nil
}

// GetErrorReportStats returns error reporting statistics for monitoring
func (s *ErrorReportRateLimitService) GetErrorReportStats(ctx context.Context, hours int) (*repository.ErrorReportStats, error) {
	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	return s.repo.GetErrorReportStats(ctx, startTime, endTime)
}
