package integration

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/tests/integration/setup"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func prepareProviderEventInboxTest(t *testing.T) (*repository.ProviderEventInboxRepository, *pgxpool.Pool, context.Context) {
	t.Helper()
	db := setup.CreateTestDB(t)
	t.Cleanup(db.Close)
	require.NoError(t, setup.RunMigrations(db))
	ctx := context.Background()
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS provider_event_inbox_test_effects (
			idempotency_key TEXT PRIMARY KEY
		);
		TRUNCATE provider_event_inbox, provider_event_inbox_test_effects RESTART IDENTITY;
	`)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = db.Exec(context.Background(), "DROP TABLE IF EXISTS provider_event_inbox_test_effects") })
	return repository.NewProviderEventInboxRepository(db), db, ctx
}

func appleProviderEventInput(t *testing.T, environment model.StoreEnvironment) repository.ProviderEventInboxInput {
	t.Helper()
	key, err := model.AppleNotificationIdempotencyKey(environment, uuid.New())
	require.NoError(t, err)
	return repository.ProviderEventInboxInput{
		Store: model.EntitlementStoreApple, Environment: &environment,
		IdempotencyKey: key, ProviderEventType: "DID_RENEW",
		ProviderOccurredAt: time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC),
	}
}

func TestProviderEventInboxCommitsEffectAndResultAtomically(t *testing.T) {
	repo, db, ctx := prepareProviderEventInboxTest(t)
	now := time.Now().UTC()
	input := appleProviderEventInput(t, model.StoreEnvironmentProduction)
	claimResult, err := repo.Claim(ctx, input, now, time.Minute)
	require.NoError(t, err)
	require.NotNil(t, claimResult.Claim)
	processorCalls := 0
	processor := func(ctx context.Context, tx pgx.Tx) (model.EntitlementOperationResult, error) {
		processorCalls++
		_, execErr := tx.Exec(ctx, "INSERT INTO provider_event_inbox_test_effects (idempotency_key) VALUES ($1)", input.IdempotencyKey.String())
		return model.EntitlementOperationResult{
			Outcome: model.EntitlementOutcomeApplied,
			Status:  model.EntitlementStatusActive,
			Code:    model.EntitlementResultApplied,
		}, execErr
	}

	first, err := repo.Complete(ctx, *claimResult.Claim, processor)
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementOutcomeApplied, first.Outcome)
	second, err := repo.Claim(ctx, input, now.Add(time.Second), time.Minute)
	require.NoError(t, err)
	assert.True(t, second.Duplicate)
	assert.Equal(t, first.Outcome, second.Outcome)
	assert.Equal(t, first.Code, second.Code)
	assert.Equal(t, 1, processorCalls)

	var inboxCount, effectCount int
	require.NoError(t, db.QueryRow(ctx, "SELECT count(*) FROM provider_event_inbox").Scan(&inboxCount))
	require.NoError(t, db.QueryRow(ctx, "SELECT count(*) FROM provider_event_inbox_test_effects").Scan(&effectCount))
	assert.Equal(t, 1, inboxCount)
	assert.Equal(t, 1, effectCount)
}

func TestProviderEventInboxConcurrentClaimRunsOneProcessor(t *testing.T) {
	repo, _, ctx := prepareProviderEventInboxTest(t)
	now := time.Now().UTC()
	input := appleProviderEventInput(t, model.StoreEnvironmentProduction)
	const contenders = 8
	start := make(chan struct{})
	results := make(chan repository.ProviderEventInboxResult, contenders)
	errorsChannel := make(chan error, contenders)
	var wait sync.WaitGroup
	for range contenders {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			result, err := repo.Claim(ctx, input, now, time.Minute)
			results <- result
			errorsChannel <- err
		}()
	}
	close(start)
	wait.Wait()
	close(results)
	close(errorsChannel)
	for err := range errorsChannel {
		require.NoError(t, err)
	}
	var claims int
	var claim repository.ProviderEventClaim
	for result := range results {
		if result.Claim != nil {
			claims++
			claim = *result.Claim
		} else {
			assert.True(t, result.Duplicate)
			assert.Equal(t, model.EntitlementOutcomeRetry, result.Outcome)
		}
	}
	assert.Equal(t, 1, claims)
	var processorCalls atomic.Int32
	_, err := repo.Complete(ctx, claim, func(context.Context, pgx.Tx) (model.EntitlementOperationResult, error) {
		processorCalls.Add(1)
		return model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeUnchanged, Code: model.EntitlementResultStaleEvent}, nil
	})
	require.NoError(t, err)
	assert.Equal(t, int32(1), processorCalls.Load())
}

func TestProviderEventInboxPersistsRetryAndReclaimsAfterBackoff(t *testing.T) {
	repo, _, ctx := prepareProviderEventInboxTest(t)
	now := time.Now().UTC()
	input := appleProviderEventInput(t, model.StoreEnvironmentSandbox)
	first, err := repo.Claim(ctx, input, now, time.Minute)
	require.NoError(t, err)
	require.NotNil(t, first.Claim)
	require.NoError(t, repo.MarkRetryable(ctx, *first.Claim, now.Add(time.Minute)))

	early, err := repo.Claim(ctx, input, now.Add(30*time.Second), time.Minute)
	require.NoError(t, err)
	assert.True(t, early.Duplicate)
	assert.Equal(t, model.EntitlementOutcomeRetry, early.Outcome)

	reclaimed, err := repo.Claim(ctx, input, now.Add(time.Minute), time.Minute)
	require.NoError(t, err)
	require.NotNil(t, reclaimed.Claim)
	assert.NotEqual(t, first.Claim.LeaseToken, reclaimed.Claim.LeaseToken)
}

func TestProviderEventInboxTreatsSemanticKeyConflictAsDuplicate(t *testing.T) {
	repo, _, ctx := prepareProviderEventInboxTest(t)
	now := time.Now().UTC()
	firstTransport, err := model.GoogleRTDNTransportIdempotencyKey("projects/app/topics/subscriptions", "message-1")
	require.NoError(t, err)
	secondTransport, err := model.GoogleRTDNTransportIdempotencyKey("projects/app/topics/subscriptions", "message-2")
	require.NoError(t, err)
	semantic, err := model.GoogleRTDNSemanticIdempotencyKey(model.GoogleRTDNSemanticEvent{
		PackageName: "com.lsimaocosta.decorebator", PurchaseToken: "protected-token",
		NotificationKind: "subscription", NotificationType: 2, EventTimeMillis: now.UnixMilli(),
	})
	require.NoError(t, err)
	firstInput := repository.ProviderEventInboxInput{
		Store: model.EntitlementStoreGoogle, IdempotencyKey: firstTransport, SemanticKey: &semantic,
		ProviderEventType: "2", ProviderOccurredAt: now,
	}
	first, err := repo.Claim(ctx, firstInput, now, time.Minute)
	require.NoError(t, err)
	require.NotNil(t, first.Claim)
	completed, err := repo.Complete(ctx, *first.Claim, func(context.Context, pgx.Tx) (model.EntitlementOperationResult, error) {
		return model.EntitlementOperationResult{
			Outcome: model.EntitlementOutcomeUnchanged, Code: model.EntitlementResultStaleEvent,
		}, nil
	})
	require.NoError(t, err)

	secondInput := firstInput
	secondInput.IdempotencyKey = secondTransport
	duplicate, err := repo.Claim(ctx, secondInput, now.Add(time.Second), time.Minute)
	require.NoError(t, err)
	assert.True(t, duplicate.Duplicate)
	assert.Equal(t, completed.Outcome, duplicate.Outcome)
	assert.Equal(t, completed.Code, duplicate.Code)
}

func TestProviderEventInboxProcessorFailureRollsBackEffectButKeepsLease(t *testing.T) {
	repo, db, ctx := prepareProviderEventInboxTest(t)
	now := time.Now().UTC()
	input := appleProviderEventInput(t, model.StoreEnvironmentSandbox)
	claimed, err := repo.Claim(ctx, input, now, time.Minute)
	require.NoError(t, err)
	require.NotNil(t, claimed.Claim)

	_, err = repo.Complete(ctx, *claimed.Claim, func(ctx context.Context, tx pgx.Tx) (model.EntitlementOperationResult, error) {
		_, insertErr := tx.Exec(ctx, "INSERT INTO provider_event_inbox_test_effects (idempotency_key) VALUES ($1)", input.IdempotencyKey.String())
		require.NoError(t, insertErr)
		return model.EntitlementOperationResult{}, errors.New("database mutation failed")
	})
	require.ErrorContains(t, err, "database mutation failed")
	var effectCount int
	require.NoError(t, db.QueryRow(ctx, "SELECT count(*) FROM provider_event_inbox_test_effects").Scan(&effectCount))
	assert.Zero(t, effectCount)
	require.NoError(t, repo.MarkRetryable(ctx, *claimed.Claim, now.Add(time.Minute)))
}

func TestProviderEventInboxPersistsProviderTestResult(t *testing.T) {
	repo, db, ctx := prepareProviderEventInboxTest(t)
	now := time.Now().UTC()
	key, err := model.GoogleRTDNTransportIdempotencyKey("projects/app/topics/subscriptions", "test-message")
	require.NoError(t, err)
	claimed, err := repo.Claim(ctx, repository.ProviderEventInboxInput{
		Store: model.EntitlementStoreGoogle, IdempotencyKey: key,
		ProviderEventType: "test", ProviderOccurredAt: now,
	}, now, time.Minute)
	require.NoError(t, err)
	require.NotNil(t, claimed.Claim)
	result, err := repo.Complete(ctx, *claimed.Claim, func(context.Context, pgx.Tx) (model.EntitlementOperationResult, error) {
		return model.EntitlementOperationResult{
			Outcome: model.EntitlementOutcomeUnchanged, Code: model.EntitlementResultProviderTest,
		}, nil
	})
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementResultProviderTest, result.Code)
	var stored string
	require.NoError(t, db.QueryRow(ctx, "SELECT result_code FROM provider_event_inbox WHERE id = $1", claimed.Claim.ID).Scan(&stored))
	assert.Equal(t, string(model.EntitlementResultProviderTest), stored)
}

func TestProviderEventInboxPersistsRetainedRevocationResult(t *testing.T) {
	repo, db, ctx := prepareProviderEventInboxTest(t)
	now := time.Now().UTC()
	input := appleProviderEventInput(t, model.StoreEnvironmentProduction)
	claimed, err := repo.Claim(ctx, input, now, time.Minute)
	require.NoError(t, err)
	require.NotNil(t, claimed.Claim)
	result, err := repo.Complete(ctx, *claimed.Claim, func(context.Context, pgx.Tx) (model.EntitlementOperationResult, error) {
		return model.EntitlementOperationResult{
			Outcome: model.EntitlementOutcomeUnchanged,
			Status:  model.EntitlementStatusRevoked,
			Code:    model.EntitlementResultRevocationRetained,
		}, nil
	})
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementResultRevocationRetained, result.Code)
	var stored string
	require.NoError(t, db.QueryRow(ctx, "SELECT result_code FROM provider_event_inbox WHERE id=$1", claimed.Claim.ID).Scan(&stored))
	assert.Equal(t, string(model.EntitlementResultRevocationRetained), stored)
}
