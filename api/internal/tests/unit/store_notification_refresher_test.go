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

type fakeStoreBindingResolver struct {
	bindings      map[model.StoreEnvironment]repository.StorePurchaseBinding
	err           error
	calls         []model.StoreEnvironment
	voidedBinding repository.StorePurchaseBinding
	voidedAt      time.Time
	voidedRefund  int32
	voidedResult  model.EntitlementOperationResult
}

func (f *fakeStoreBindingResolver) ApplyGoogleVoidedPurchase(
	_ context.Context,
	_ pgx.Tx,
	binding repository.StorePurchaseBinding,
	occurredAt time.Time,
	refundType int32,
) (model.EntitlementOperationResult, error) {
	f.voidedBinding, f.voidedAt, f.voidedRefund = binding, occurredAt, refundType
	return f.voidedResult, nil
}

func (f *fakeStoreBindingResolver) ResolvePurchaseBinding(
	_ context.Context,
	_ model.EntitlementStore,
	environment model.StoreEnvironment,
	_ string,
) (repository.StorePurchaseBinding, bool, error) {
	f.calls = append(f.calls, environment)
	if f.err != nil {
		return repository.StorePurchaseBinding{}, false, f.err
	}
	binding, ok := f.bindings[environment]
	return binding, ok, nil
}

type fakeAppleProviderRefresher struct {
	input  service.AppleSubscriptionRefreshInput
	result service.VerifiedAppleSubscription
	err    error
}

func (f *fakeAppleProviderRefresher) Refresh(
	_ context.Context,
	input service.AppleSubscriptionRefreshInput,
) (service.VerifiedAppleSubscription, error) {
	f.input = input
	return f.result, f.err
}

type fakeGoogleProviderProcessor struct {
	input  service.GooglePurchaseVerificationInput
	result service.GooglePurchaseProcessingResult
	err    error
}

func (f *fakeGoogleProviderProcessor) VerifyAndAcknowledge(
	_ context.Context,
	input service.GooglePurchaseVerificationInput,
) (service.GooglePurchaseProcessingResult, error) {
	f.input = input
	return f.result, f.err
}

func TestStoreNotificationRefresherResolvesAppleOwnershipAndPersistsProviderStatus(t *testing.T) {
	now := time.Date(2026, 8, 5, 15, 0, 0, 0, time.UTC)
	accountToken := uuid.New()
	originalID := "200000000000001"
	resolver := &fakeStoreBindingResolver{bindings: map[model.StoreEnvironment]repository.StorePurchaseBinding{
		model.StoreEnvironmentProduction: {
			UserID: 42, Store: model.EntitlementStoreApple, Environment: model.StoreEnvironmentProduction,
			AccountIdentifier: accountToken.String(), ProviderRecordID: originalID,
		},
	}}
	apple := &fakeAppleProviderRefresher{result: service.VerifiedAppleSubscription{
		Identity: model.ApplePurchaseIdentity{
			UserID: 42, AppAccountToken: accountToken, OriginalTransactionID: originalID,
			Environment: model.StoreEnvironmentProduction, LastVerifiedAt: now,
		},
		ProductID: "premium-monthly", Entitlement: "premium", Status: model.EntitlementStatusActive,
		ProviderSnapshotSignedAt: now,
	}}
	writer := &fakeStoreEntitlementWriter{result: repository.StoreEntitlementApplyResult{
		Operation: model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeApplied, Code: model.EntitlementResultApplied},
	}}
	refresher := newStoreNotificationRefresher(t, resolver, apple, &fakeGoogleProviderProcessor{}, writer)
	processor, err := refresher.RefreshApple(context.Background(), service.VerifiedAppleNotification{
		NotificationType: "DID_RENEW", Environment: model.StoreEnvironmentProduction,
		ProviderSignedAt: now.Add(-time.Minute), Transaction: &service.AppleTransactionPayload{OriginalTransactionID: originalID},
	})
	require.NoError(t, err)
	result, err := processor(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementOutcomeApplied, result.Outcome)
	assert.Equal(t, int64(42), apple.input.AuthenticatedUserID)
	assert.Equal(t, accountToken, apple.input.Account.AppAccountToken)
	assert.Equal(t, model.StoreEnvironmentProduction, apple.input.Environment)
	require.NotNil(t, writer.update.EventOccurredAt)
	assert.Equal(t, now.Add(-time.Minute), *writer.update.EventOccurredAt)
}

func TestStoreNotificationRefresherRetriesMissingAppleBindingAndRejectsCrossEnvironmentState(t *testing.T) {
	now := time.Date(2026, 8, 5, 15, 0, 0, 0, time.UTC)
	resolver := &fakeStoreBindingResolver{bindings: map[model.StoreEnvironment]repository.StorePurchaseBinding{}}
	apple := &fakeAppleProviderRefresher{}
	refresher := newStoreNotificationRefresher(t, resolver, apple, &fakeGoogleProviderProcessor{}, &fakeStoreEntitlementWriter{})
	_, err := refresher.RefreshApple(context.Background(), service.VerifiedAppleNotification{
		NotificationType: "DID_RENEW", Environment: model.StoreEnvironmentSandbox,
		ProviderSignedAt: now, Transaction: &service.AppleTransactionPayload{OriginalTransactionID: "200000000000001"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not available yet")
	assert.Zero(t, apple.input.AuthenticatedUserID)

	accountToken := uuid.New()
	resolver.bindings[model.StoreEnvironmentSandbox] = repository.StorePurchaseBinding{
		UserID: 42, Store: model.EntitlementStoreApple, Environment: model.StoreEnvironmentProduction,
		AccountIdentifier: accountToken.String(), ProviderRecordID: "200000000000001",
	}
	_, err = refresher.RefreshApple(context.Background(), service.VerifiedAppleNotification{
		NotificationType: "DID_RENEW", Environment: model.StoreEnvironmentSandbox,
		ProviderSignedAt: now, Transaction: &service.AppleTransactionPayload{OriginalTransactionID: "200000000000001"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "inconsistent")
}

func TestStoreNotificationRefresherResolvesGoogleAcrossEnvironmentsAndRequiresAcknowledgement(t *testing.T) {
	now := time.Date(2026, 8, 5, 15, 0, 0, 0, time.UTC)
	resolver := &fakeStoreBindingResolver{bindings: map[model.StoreEnvironment]repository.StorePurchaseBinding{
		model.StoreEnvironmentSandbox: {
			UserID: 42, Store: model.EntitlementStoreGoogle, Environment: model.StoreEnvironmentSandbox,
			AccountIdentifier: "opaque-account", ProviderRecordID: "purchase-token",
		},
	}}
	google := &fakeGoogleProviderProcessor{result: service.GooglePurchaseProcessingResult{
		Purchase: service.VerifiedGooglePurchase{
			Identity: model.GooglePurchaseIdentity{
				UserID: 42, ObfuscatedExternalAccountID: "opaque-account", PurchaseToken: "purchase-token",
				Environment: model.StoreEnvironmentSandbox, LastVerifiedAt: now,
			},
			ProductID: "premium-monthly", Entitlement: "premium", Status: model.EntitlementStatusActive,
			Acknowledged: true,
		},
		Acknowledgement: service.GoogleAcknowledgementConfirmed,
	}}
	writer := &fakeStoreEntitlementWriter{result: repository.StoreEntitlementApplyResult{
		Operation: model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeApplied},
	}}
	refresher := newStoreNotificationRefresher(t, resolver, &fakeAppleProviderRefresher{}, google, writer)
	processor, err := refresher.RefreshGoogle(context.Background(), service.GoogleRTDNEvent{
		Kind: "subscription", PurchaseToken: "purchase-token", OccurredAt: now.Add(-time.Minute),
	})
	require.NoError(t, err)
	_, err = processor(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, []model.StoreEnvironment{model.StoreEnvironmentProduction, model.StoreEnvironmentSandbox}, resolver.calls)
	assert.Equal(t, int64(42), google.input.AuthenticatedUserID)
	assert.Equal(t, "opaque-account", google.input.Account.ObfuscatedExternalAccountID)

	google.result.Acknowledgement = service.GoogleAcknowledgementRetry
	google.result.AcknowledgementError = errors.New("temporary acknowledge outage")
	_, err = refresher.RefreshGoogle(context.Background(), service.GoogleRTDNEvent{
		Kind: "subscription", PurchaseToken: "purchase-token", OccurredAt: now,
	})
	require.ErrorContains(t, err, "temporary acknowledge outage")
}

func TestStoreNotificationRefresherRejectsAmbiguousGoogleBinding(t *testing.T) {
	binding := repository.StorePurchaseBinding{
		UserID: 42, Store: model.EntitlementStoreGoogle,
		AccountIdentifier: "opaque-account", ProviderRecordID: "purchase-token",
	}
	production, sandbox := binding, binding
	production.Environment = model.StoreEnvironmentProduction
	sandbox.Environment = model.StoreEnvironmentSandbox
	resolver := &fakeStoreBindingResolver{bindings: map[model.StoreEnvironment]repository.StorePurchaseBinding{
		model.StoreEnvironmentProduction: production,
		model.StoreEnvironmentSandbox:    sandbox,
	}}
	refresher := newStoreNotificationRefresher(
		t, resolver, &fakeAppleProviderRefresher{}, &fakeGoogleProviderProcessor{}, &fakeStoreEntitlementWriter{},
	)
	_, err := refresher.RefreshGoogle(context.Background(), service.GoogleRTDNEvent{
		Kind: "subscription", PurchaseToken: "purchase-token", OccurredAt: time.Now(),
	})
	require.ErrorContains(t, err, "ambiguous")
}

func TestStoreNotificationRefresherAppliesVoidedGooglePurchaseWithoutStatusLookup(t *testing.T) {
	now := time.Date(2026, 8, 5, 15, 0, 0, 0, time.UTC)
	resolver := &fakeStoreBindingResolver{
		bindings: map[model.StoreEnvironment]repository.StorePurchaseBinding{
			model.StoreEnvironmentProduction: {
				EntitlementID: 8, UserID: 42, Store: model.EntitlementStoreGoogle,
				Environment: model.StoreEnvironmentProduction, AccountIdentifier: "opaque-account",
				ProviderRecordID: "purchase-token", Current: true, MatchedPrimary: true,
			},
		},
		voidedResult: model.EntitlementOperationResult{
			Outcome: model.EntitlementOutcomeApplied, Status: model.EntitlementStatusRevoked,
			Code: model.EntitlementResultRevoked,
		},
	}
	google := &fakeGoogleProviderProcessor{err: errors.New("must not call Google status for a voided purchase")}
	refresher := newStoreNotificationRefresher(t, resolver, &fakeAppleProviderRefresher{}, google, &fakeStoreEntitlementWriter{})
	processor, err := refresher.RefreshGoogle(context.Background(), service.GoogleRTDNEvent{
		Kind: "voided_subscription", PurchaseToken: "purchase-token", RefundType: 2, OccurredAt: now,
	})
	require.NoError(t, err)
	result, err := processor(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementResultRevoked, result.Code)
	assert.Equal(t, int64(8), resolver.voidedBinding.EntitlementID)
	assert.Equal(t, now, resolver.voidedAt)
	assert.Equal(t, int32(2), resolver.voidedRefund)
	assert.Zero(t, google.input.AuthenticatedUserID)
}

func newStoreNotificationRefresher(
	t *testing.T,
	resolver service.StorePurchaseBindingResolver,
	apple service.AppleProviderSubscriptionRefresher,
	google service.GoogleProviderPurchaseProcessor,
	writer service.StoreEntitlementWriter,
) *service.StoreNotificationRefresher {
	t.Helper()
	persistence, err := service.NewStoreEntitlementPersistence(writer)
	require.NoError(t, err)
	voided, ok := resolver.(service.GoogleVoidedPurchaseWriter)
	require.True(t, ok)
	refresher, err := service.NewStoreNotificationRefresher(resolver, voided, apple, google, persistence)
	require.NoError(t, err)
	return refresher
}
