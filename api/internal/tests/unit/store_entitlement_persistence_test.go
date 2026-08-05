package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStoreEntitlementWriter struct {
	update repository.StoreEntitlementUpdate
	tx     pgx.Tx
	result repository.StoreEntitlementApplyResult
	err    error
}

func (f *fakeStoreEntitlementWriter) Apply(
	_ context.Context,
	update repository.StoreEntitlementUpdate,
	tx pgx.Tx,
) (repository.StoreEntitlementApplyResult, error) {
	f.update, f.tx = update, tx
	return f.result, f.err
}

func TestStoreEntitlementPersistenceKeepsAppleSnapshotAndEventClocksSeparate(t *testing.T) {
	observedAt := time.Date(2026, 8, 5, 14, 0, 0, 0, time.UTC)
	signedAt := observedAt.Add(-time.Minute)
	eventAt := observedAt.Add(-time.Hour)
	periodStart, periodEnd := observedAt.Add(-24*time.Hour), observedAt.Add(30*24*time.Hour)
	writer := &fakeStoreEntitlementWriter{result: repository.StoreEntitlementApplyResult{
		Operation: model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeApplied},
	}}
	persistence, err := service.NewStoreEntitlementPersistence(writer)
	require.NoError(t, err)
	_, err = persistence.PersistApple(context.Background(), nil, service.VerifiedAppleSubscription{
		Identity: model.ApplePurchaseIdentity{
			UserID: 42, AppAccountToken: uuid.New(), OriginalTransactionID: "200000000000001",
			Environment: model.StoreEnvironmentProduction, LastVerifiedAt: observedAt,
		},
		ProductID: "decorebator_monthly_premium1", Entitlement: "premium",
		Status: model.EntitlementStatusActive, PeriodStart: &periodStart, PeriodEnd: &periodEnd,
		AutoRenewEnabled: true, ProviderSnapshotSignedAt: signedAt,
	}, &eventAt)
	require.NoError(t, err)
	assert.Equal(t, observedAt, writer.update.SnapshotObservedAt)
	require.NotNil(t, writer.update.EventOccurredAt)
	assert.Equal(t, eventAt, *writer.update.EventOccurredAt)
	assert.NotEqual(t, signedAt, writer.update.SnapshotObservedAt)
}

func TestStoreEntitlementPersistenceMapsOwnershipConflictForEveryCaller(t *testing.T) {
	writer := &fakeStoreEntitlementWriter{err: repository.ErrStoreEvidenceOwnershipConflict}
	persistence, err := service.NewStoreEntitlementPersistence(writer)
	require.NoError(t, err)
	now := time.Date(2026, 8, 5, 14, 0, 0, 0, time.UTC)
	periodEnd := now.Add(30 * 24 * time.Hour)
	result, err := persistence.PersistGoogle(context.Background(), nil, service.GooglePurchaseProcessingResult{
		Purchase: service.VerifiedGooglePurchase{
			Identity: model.GooglePurchaseIdentity{
				UserID: 42, ObfuscatedExternalAccountID: "opaque-account", PurchaseToken: "protected-token",
				Environment: model.StoreEnvironmentProduction, LastVerifiedAt: now,
			},
			ProductID: "decorebator_monthly_premium1", Entitlement: "premium",
			Status: model.EntitlementStatusActive, PeriodEnd: &periodEnd, Acknowledged: true,
		},
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementOutcomeRejected, result.Operation.Outcome)
	assert.Equal(t, model.EntitlementResultAccountMismatch, result.Operation.Code)
	assert.False(t, errors.Is(err, repository.ErrStoreEvidenceOwnershipConflict))
}

func TestStoreEntitlementPersistenceMarksSupersededGoogleTokenHistorical(t *testing.T) {
	now := time.Date(2026, 8, 5, 14, 0, 0, 0, time.UTC)
	writer := &fakeStoreEntitlementWriter{}
	persistence, err := service.NewStoreEntitlementPersistence(writer)
	require.NoError(t, err)
	_, err = persistence.PersistGoogle(context.Background(), nil, service.GooglePurchaseProcessingResult{
		Purchase: service.VerifiedGooglePurchase{
			Identity: model.GooglePurchaseIdentity{
				UserID: 42, ObfuscatedExternalAccountID: "opaque-account", PurchaseToken: "old-token",
				Environment: model.StoreEnvironmentProduction, LastVerifiedAt: now,
			},
			ProductID: "decorebator_monthly_premium1", Entitlement: "premium",
			Status: model.EntitlementStatusExpired, Superseded: true, Acknowledged: true,
		},
	}, nil)
	require.NoError(t, err)
	assert.True(t, writer.update.HistoricalBinding)
	assert.False(t, writer.update.CurrentBinding)
}
