package service

import (
	"context"
	"fmt"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/repository"
	"github.com/riverqueue/river"
)

const providerInboxHealthGracePeriod = 5 * time.Minute

type ProviderEventInboxHealthArgs struct{}

func (ProviderEventInboxHealthArgs) Kind() string { return "provider-event-inbox-health" }

type ProviderEventInboxHealthReader interface {
	StrandedHealth(context.Context, time.Time, time.Duration) (repository.ProviderEventInboxHealth, error)
}

type ProviderEventInboxHealthWorker struct {
	river.WorkerDefaults[ProviderEventInboxHealthArgs]
	repository ProviderEventInboxHealthReader
	now        func() time.Time
	grace      time.Duration
}

func NewProviderEventInboxHealthWorker(
	repository ProviderEventInboxHealthReader,
	now func() time.Time,
	grace time.Duration,
) (*ProviderEventInboxHealthWorker, error) {
	if repository == nil || grace < 0 {
		return nil, fmt.Errorf("provider inbox health repository and non-negative grace are required")
	}
	if now == nil {
		now = time.Now
	}
	return &ProviderEventInboxHealthWorker{repository: repository, now: now, grace: grace}, nil
}

func (w *ProviderEventInboxHealthWorker) Work(
	ctx context.Context,
	_ *river.Job[ProviderEventInboxHealthArgs],
) error {
	health, err := w.repository.StrandedHealth(ctx, w.now().UTC(), w.grace)
	if err != nil {
		return fmt.Errorf("inspect provider event inbox health: %w", err)
	}
	if health.OverdueRetryable == 0 && health.ExpiredProcessing == 0 {
		common.Logger.InfoContext(ctx, "provider event inbox health check completed",
			"overdue_retryable", int64(0), "expired_processing", int64(0))
		return nil
	}
	common.Logger.ErrorContext(ctx, "provider event inbox contains stranded events",
		"overdue_retryable", health.OverdueRetryable,
		"expired_processing", health.ExpiredProcessing,
		"oldest_retryable_due_at", health.OldestRetryableDueAt,
		"oldest_processing_expiry", health.OldestProcessingExpiry,
	)
	return nil
}
