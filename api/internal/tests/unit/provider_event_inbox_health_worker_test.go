package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeProviderInboxHealthReader struct {
	health repository.ProviderEventInboxHealth
	err    error
	now    time.Time
	grace  time.Duration
}

func (f *fakeProviderInboxHealthReader) StrandedHealth(
	_ context.Context,
	now time.Time,
	grace time.Duration,
) (repository.ProviderEventInboxHealth, error) {
	f.now, f.grace = now, grace
	return f.health, f.err
}

func TestProviderEventInboxHealthWorkerChecksBothStrandedStatesWithoutMutation(t *testing.T) {
	now := time.Date(2026, 8, 5, 16, 0, 0, 0, time.UTC)
	reader := &fakeProviderInboxHealthReader{health: repository.ProviderEventInboxHealth{
		OverdueRetryable: 2, ExpiredProcessing: 3,
	}}
	worker, err := service.NewProviderEventInboxHealthWorker(reader, func() time.Time { return now }, 5*time.Minute)
	require.NoError(t, err)
	require.NoError(t, worker.Work(context.Background(), nil))
	assert.Equal(t, now, reader.now)
	assert.Equal(t, 5*time.Minute, reader.grace)
}

func TestProviderEventInboxHealthWorkerRetriesDatabaseInspectionFailure(t *testing.T) {
	reader := &fakeProviderInboxHealthReader{err: errors.New("database unavailable")}
	worker, err := service.NewProviderEventInboxHealthWorker(reader, nil, 0)
	require.NoError(t, err)
	err = worker.Work(context.Background(), nil)
	require.ErrorContains(t, err, "inspect provider event inbox health")
}
