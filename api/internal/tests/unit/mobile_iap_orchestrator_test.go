package unit

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMobileIAPStore struct {
	mu            sync.Mutex
	appleAccount  model.AppleAccountIdentity
	googleAccount model.GoogleAccountIdentity
	entitlements  map[model.EntitlementStore]model.StoreEntitlement
	bindings      map[string]repository.StorePurchaseBinding
	known         []repository.StorePurchaseBinding
	acknowledged  bool
	access        bool
	accessErr     error
	candidates    []repository.StoreAcknowledgementRetryCandidate
}

func (f *fakeMobileIAPStore) GetOrCreateAppleAccount(context.Context, int64) (model.AppleAccountIdentity, error) {
	return f.appleAccount, nil
}

func (f *fakeMobileIAPStore) GetOrCreateGoogleAccount(context.Context, int64) (model.GoogleAccountIdentity, error) {
	return f.googleAccount, nil
}

func (f *fakeMobileIAPStore) GetCurrentEntitlement(_ context.Context, _ int64, store model.EntitlementStore, _ model.StoreEnvironment) (model.StoreEntitlement, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	entitlement, found := f.entitlements[store]
	return entitlement, found, nil
}

func (f *fakeMobileIAPStore) CurrentGoogleBindingAcknowledged(context.Context, int64) (bool, error) {
	return f.acknowledged, nil
}

func (f *fakeMobileIAPStore) ListUserPurchaseBindings(context.Context, int64, model.EntitlementStore, model.StoreEnvironment, int) ([]repository.StorePurchaseBinding, error) {
	return append([]repository.StorePurchaseBinding(nil), f.known...), nil
}

func (f *fakeMobileIAPStore) ResolvePurchaseBinding(_ context.Context, store model.EntitlementStore, environment model.StoreEnvironment, evidence string) (repository.StorePurchaseBinding, bool, error) {
	binding, found := f.bindings[evidence]
	if found && (binding.Store != store || binding.Environment != environment) {
		return repository.StorePurchaseBinding{}, false, nil
	}
	return binding, found, nil
}

func (f *fakeMobileIAPStore) GetPurchaseBindingByID(_ context.Context, bindingID int64) (repository.StorePurchaseBinding, bool, error) {
	for _, binding := range f.bindings {
		if binding.ID == bindingID {
			acknowledged := f.acknowledged
			binding.Acknowledged = &acknowledged
			return binding, true, nil
		}
	}
	return repository.StorePurchaseBinding{}, false, nil
}

func (f *fakeMobileIAPStore) ListUnacknowledgedGoogleBindings(context.Context, model.StoreEnvironment, int) ([]repository.StoreAcknowledgementRetryCandidate, error) {
	return append([]repository.StoreAcknowledgementRetryCandidate(nil), f.candidates...), nil
}

func (f *fakeMobileIAPStore) HasEffectivePremiumAccess(context.Context, int64, model.StoreEnvironment, time.Time) (bool, error) {
	return f.access, f.accessErr
}

func (f *fakeMobileIAPStore) Apply(_ context.Context, update repository.StoreEntitlementUpdate, _ pgx.Tx) (repository.StoreEntitlementApplyResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	entitlement := update.Entitlement
	entitlement.ID = 91
	entitlement.CreatedAt = entitlement.LastVerifiedAt
	entitlement.UpdatedAt = entitlement.LastVerifiedAt
	f.entitlements[entitlement.Store] = entitlement
	if update.Acknowledged != nil {
		f.acknowledged = *update.Acknowledged
	}
	return repository.StoreEntitlementApplyResult{
		Operation: model.EntitlementOperationResult{
			Outcome: model.EntitlementOutcomeApplied, Status: entitlement.Status,
			Code: model.EntitlementResultApplied,
		},
		Entitlement: entitlement,
	}, nil
}

type fakeAppleMobileVerifier struct {
	environment model.StoreEnvironment
	account     uuid.UUID
	lineages    map[string]string
	calls       *int
}

func (f fakeAppleMobileVerifier) Verify(_ context.Context, input service.ApplePurchaseVerificationInput) (service.VerifiedApplePurchase, error) {
	if f.calls != nil {
		*f.calls++
	}
	lineage, found := f.lineages[input.TransactionID]
	if !found {
		return service.VerifiedApplePurchase{}, &service.ApplePurchaseVerificationError{Failure: model.EntitlementFailureInvalidEvidence}
	}
	return service.VerifiedApplePurchase{
		Identity: model.ApplePurchaseIdentity{
			UserID: input.AuthenticatedUserID, AppAccountToken: f.account,
			OriginalTransactionID: lineage, Environment: f.environment, LastVerifiedAt: testIAPNow,
		},
		TransactionID: input.TransactionID, ProductID: "apple.monthly", Entitlement: "premium",
	}, nil
}

type fakeAppleMobileRefresher struct {
	mu     sync.Mutex
	calls  []string
	errors map[string]error
}

func (f *fakeAppleMobileRefresher) Refresh(_ context.Context, input service.AppleSubscriptionRefreshInput) (service.VerifiedAppleSubscription, error) {
	f.mu.Lock()
	f.calls = append(f.calls, input.OriginalTransactionID)
	f.mu.Unlock()
	if err := f.errors[input.OriginalTransactionID]; err != nil {
		return service.VerifiedAppleSubscription{}, err
	}
	periodEnd := testIAPNow.Add(30 * 24 * time.Hour)
	return service.VerifiedAppleSubscription{
		Identity: model.ApplePurchaseIdentity{
			UserID: input.AuthenticatedUserID, AppAccountToken: input.Account.AppAccountToken,
			OriginalTransactionID: input.OriginalTransactionID, Environment: input.Environment,
			LastVerifiedAt: testIAPNow,
		},
		TransactionID: "latest", ProductID: "apple.monthly", Entitlement: "premium",
		Status: model.EntitlementStatusActive, PeriodEnd: &periodEnd, AutoRenewEnabled: true,
		ProviderSnapshotSignedAt: testIAPNow,
	}, nil
}

type fakeGoogleMobileProcessor struct {
	result service.GooglePurchaseProcessingResult
	err    error
}

type sequenceGoogleMobileProcessor struct {
	mu      sync.Mutex
	calls   int
	results []service.GooglePurchaseProcessingResult
}

func (f *sequenceGoogleMobileProcessor) VerifyAndAcknowledge(_ context.Context, _ service.GooglePurchaseVerificationInput) (service.GooglePurchaseProcessingResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := f.results[f.calls]
	f.calls++
	return result, nil
}

func (f fakeGoogleMobileProcessor) VerifyAndAcknowledge(context.Context, service.GooglePurchaseVerificationInput) (service.GooglePurchaseProcessingResult, error) {
	return f.result, f.err
}

type fakeGoogleRetryScheduler struct {
	mu      sync.Mutex
	binding []int64
}

func (f *fakeGoogleRetryScheduler) ScheduleGoogleAcknowledgementRetry(_ context.Context, bindingID int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.binding = append(f.binding, bindingID)
	return nil
}

var testIAPNow = time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)

func newTestIAPOrchestrator(t *testing.T, store *fakeMobileIAPStore, apple *fakeAppleMobileRefresher, google service.GooglePurchaseProcessingResult, scheduler *fakeGoogleRetryScheduler) *service.StoreIAPOrchestrator {
	t.Helper()
	persistence, err := service.NewStoreEntitlementPersistence(store)
	require.NoError(t, err)
	orchestrator, err := service.NewStoreIAPOrchestrator(
		store,
		fakeAppleMobileVerifier{environment: model.StoreEnvironmentSandbox, account: store.appleAccount.AppAccountToken, lineages: map[string]string{"tx-1": "lineage-1", "tx-2": "lineage-1"}},
		apple, fakeGoogleMobileProcessor{result: google}, persistence, scheduler,
		model.StoreEnvironmentSandbox,
		[]model.MobileIAPProduct{
			{Store: model.EntitlementStoreApple, ProductID: "apple.monthly", Entitlement: "premium", BillingPeriod: model.IAPBillingPeriodMonthly},
			{Store: model.EntitlementStoreGoogle, ProductID: "google.monthly", Entitlement: "premium", BillingPeriod: model.IAPBillingPeriodMonthly},
		},
		func() time.Time { return testIAPNow },
	)
	require.NoError(t, err)
	return orchestrator
}

func TestStoreIAPContextKeepsUnacknowledgedGooglePurchasePending(t *testing.T) {
	accountToken := uuid.New()
	periodEnd := testIAPNow.Add(24 * time.Hour)
	store := &fakeMobileIAPStore{
		appleAccount:  model.AppleAccountIdentity{UserID: 7, AppAccountToken: accountToken},
		googleAccount: model.GoogleAccountIdentity{UserID: 7, ObfuscatedExternalAccountID: "opaque-account"},
		entitlements: map[model.EntitlementStore]model.StoreEntitlement{
			model.EntitlementStoreGoogle: {
				ID: 2, UserID: 7, Store: model.EntitlementStoreGoogle, Environment: model.StoreEnvironmentSandbox,
				ProductID: "google.monthly", Entitlement: "premium", Status: model.EntitlementStatusActive,
				PeriodEnd: &periodEnd, LastVerifiedAt: testIAPNow, CreatedAt: testIAPNow,
			},
		},
		bindings: map[string]repository.StorePurchaseBinding{},
	}
	orchestrator := newTestIAPOrchestrator(t, store, &fakeAppleMobileRefresher{}, service.GooglePurchaseProcessingResult{}, &fakeGoogleRetryScheduler{})

	response, err := orchestrator.Context(context.Background(), 7, model.EntitlementStoreGoogle)
	require.NoError(t, err)
	require.NotNil(t, response.Pending)
	assert.Nil(t, response.CurrentEntitlement)
	assert.Equal(t, "opaque-account", *response.PurchaseContext.GoogleObfuscatedAccountID)
}

func TestStoreIAPRestoreDeduplicatesAppleLineageAndPreservesRejection(t *testing.T) {
	accountToken := uuid.New()
	store := &fakeMobileIAPStore{
		appleAccount:  model.AppleAccountIdentity{UserID: 7, AppAccountToken: accountToken},
		googleAccount: model.GoogleAccountIdentity{UserID: 7, ObfuscatedExternalAccountID: "opaque-account"},
		entitlements:  map[model.EntitlementStore]model.StoreEntitlement{},
		bindings: map[string]repository.StorePurchaseBinding{
			"lineage-1": {
				ID: 10, UserID: 7, Store: model.EntitlementStoreApple,
				Environment: model.StoreEnvironmentSandbox, ProviderRecordID: "lineage-1",
				AccountIdentifier: accountToken.String(), Current: true,
			},
		},
	}
	apple := &fakeAppleMobileRefresher{}
	orchestrator := newTestIAPOrchestrator(t, store, apple, service.GooglePurchaseProcessingResult{}, &fakeGoogleRetryScheduler{})

	response, err := orchestrator.Restore(context.Background(), 7, model.EntitlementStoreApple, []string{"tx-1", "tx-2", "invalid"})
	require.NoError(t, err)
	require.NotNil(t, response.Restore)
	assert.Equal(t, model.IAPRestorePartial, response.Restore.Status)
	assert.Equal(t, 1, response.Restore.RestoredPurchases)
	assert.Equal(t, 1, response.Restore.UnchangedPurchases)
	assert.Equal(t, 1, response.Restore.RejectedPurchases)
	assert.Equal(t, []string{"lineage-1"}, apple.calls)
	for _, outcome := range response.Restore.Outcomes {
		if outcome.Error != nil {
			assert.Equal(t, "iap.error.invalid_purchase", outcome.Error.MessageKey)
		}
	}
}

func TestStoreIAPRestoreIncludesKnownBindingFailuresWithAndWithoutClientEvidence(t *testing.T) {
	accountToken := uuid.New()
	providerUnavailable := &service.ApplePurchaseVerificationError{Failure: model.EntitlementFailureProviderUnavailable}
	newStore := func() *fakeMobileIAPStore {
		return &fakeMobileIAPStore{
			appleAccount:  model.AppleAccountIdentity{UserID: 7, AppAccountToken: accountToken},
			googleAccount: model.GoogleAccountIdentity{UserID: 7, ObfuscatedExternalAccountID: "opaque-account"},
			entitlements:  map[model.EntitlementStore]model.StoreEntitlement{},
			bindings: map[string]repository.StorePurchaseBinding{
				"lineage-1": {ID: 1, UserID: 7, Store: model.EntitlementStoreApple, Environment: model.StoreEnvironmentSandbox, ProviderRecordID: "lineage-1", AccountIdentifier: accountToken.String(), Current: true},
				"lineage-2": {ID: 2, UserID: 7, Store: model.EntitlementStoreApple, Environment: model.StoreEnvironmentSandbox, ProviderRecordID: "lineage-2", AccountIdentifier: accountToken.String(), Current: true},
			},
			known: []repository.StorePurchaseBinding{
				{ID: 1, UserID: 7, Store: model.EntitlementStoreApple, Environment: model.StoreEnvironmentSandbox, ProviderRecordID: "lineage-1", AccountIdentifier: accountToken.String(), Current: true},
				{ID: 2, UserID: 7, Store: model.EntitlementStoreApple, Environment: model.StoreEnvironmentSandbox, ProviderRecordID: "lineage-2", AccountIdentifier: accountToken.String(), Current: true},
			},
		}
	}

	store := newStore()
	apple := &fakeAppleMobileRefresher{errors: map[string]error{"lineage-2": providerUnavailable}}
	orchestrator := newTestIAPOrchestrator(t, store, apple, service.GooglePurchaseProcessingResult{}, &fakeGoogleRetryScheduler{})
	response, err := orchestrator.Restore(context.Background(), 7, model.EntitlementStoreApple, []string{})
	require.NoError(t, err)
	require.NotNil(t, response.Restore)
	assert.Equal(t, model.IAPRestorePartial, response.Restore.Status)
	assert.Equal(t, 1, response.Restore.RestoredPurchases)
	assert.Equal(t, 1, response.Restore.RetryablePurchases)
	assert.Empty(t, response.Restore.Outcomes, "server-known evidence is never exposed")

	store = newStore()
	apple = &fakeAppleMobileRefresher{errors: map[string]error{"lineage-2": providerUnavailable}}
	orchestrator = newTestIAPOrchestrator(t, store, apple, service.GooglePurchaseProcessingResult{}, &fakeGoogleRetryScheduler{})
	response, err = orchestrator.Restore(context.Background(), 7, model.EntitlementStoreApple, []string{"tx-1"})
	require.NoError(t, err)
	assert.Equal(t, model.IAPRestorePartial, response.Restore.Status)
	assert.Equal(t, 1, response.Restore.RetryablePurchases, "known-binding failures remain visible with client evidence")
}

func TestStoreIAPEmptyRestoreReturnsNoPurchasesInCompleteEnvelope(t *testing.T) {
	accountToken := uuid.New()
	store := &fakeMobileIAPStore{
		appleAccount:  model.AppleAccountIdentity{UserID: 7, AppAccountToken: accountToken},
		googleAccount: model.GoogleAccountIdentity{UserID: 7, ObfuscatedExternalAccountID: "opaque-account"},
		entitlements:  map[model.EntitlementStore]model.StoreEntitlement{},
		bindings:      map[string]repository.StorePurchaseBinding{},
	}
	orchestrator := newTestIAPOrchestrator(t, store, &fakeAppleMobileRefresher{}, service.GooglePurchaseProcessingResult{}, &fakeGoogleRetryScheduler{})
	response, err := orchestrator.Restore(context.Background(), 7, model.EntitlementStoreApple, []string{})
	require.NoError(t, err)
	require.NoError(t, response.Validate())
	require.NotNil(t, response.Restore)
	assert.Equal(t, model.IAPRestoreNoPurchases, response.Restore.Status)
	assert.NotEmpty(t, response.Products)
	assert.NotNil(t, response.PurchaseContext)
	assert.False(t, response.ServerTime.IsZero())
}

func TestStoreIAPRestoreMakesNoProviderCallAfterCancellation(t *testing.T) {
	accountToken := uuid.New()
	store := &fakeMobileIAPStore{
		appleAccount:  model.AppleAccountIdentity{UserID: 7, AppAccountToken: accountToken},
		googleAccount: model.GoogleAccountIdentity{UserID: 7, ObfuscatedExternalAccountID: "opaque-account"},
		entitlements:  map[model.EntitlementStore]model.StoreEntitlement{},
		bindings:      map[string]repository.StorePurchaseBinding{},
	}
	persistence, err := service.NewStoreEntitlementPersistence(store)
	require.NoError(t, err)
	calls := 0
	orchestrator, err := service.NewStoreIAPOrchestrator(
		store,
		fakeAppleMobileVerifier{environment: model.StoreEnvironmentSandbox, account: accountToken, lineages: map[string]string{"tx-1": "lineage-1"}, calls: &calls},
		&fakeAppleMobileRefresher{}, fakeGoogleMobileProcessor{}, persistence,
		&fakeGoogleRetryScheduler{}, model.StoreEnvironmentSandbox,
		[]model.MobileIAPProduct{{Store: model.EntitlementStoreApple, ProductID: "apple.monthly", Entitlement: "premium", BillingPeriod: model.IAPBillingPeriodMonthly}},
		func() time.Time { return testIAPNow },
	)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	response, err := orchestrator.Restore(ctx, 7, model.EntitlementStoreApple, []string{"tx-1"})
	require.NoError(t, err)
	assert.Equal(t, 0, calls)
	assert.Equal(t, model.IAPRestorePending, response.Restore.Status)
}

func TestStoreIAPGoogleRestoreRejectsProviderEnvironmentMismatch(t *testing.T) {
	accountToken := uuid.New()
	periodEnd := testIAPNow.Add(time.Hour)
	store := &fakeMobileIAPStore{
		appleAccount:  model.AppleAccountIdentity{UserID: 7, AppAccountToken: accountToken},
		googleAccount: model.GoogleAccountIdentity{UserID: 7, ObfuscatedExternalAccountID: "opaque-account"},
		entitlements:  map[model.EntitlementStore]model.StoreEntitlement{},
		bindings: map[string]repository.StorePurchaseBinding{
			"token": {ID: 3, UserID: 7, Store: model.EntitlementStoreGoogle, Environment: model.StoreEnvironmentSandbox, ProviderRecordID: "token", AccountIdentifier: "opaque-account", Current: true},
		},
	}
	processed := service.GooglePurchaseProcessingResult{
		Purchase: service.VerifiedGooglePurchase{
			Identity:  model.GooglePurchaseIdentity{UserID: 7, ObfuscatedExternalAccountID: "opaque-account", PurchaseToken: "token", Environment: model.StoreEnvironmentProduction, LastVerifiedAt: testIAPNow},
			ProductID: "google.monthly", Entitlement: "premium", Status: model.EntitlementStatusActive, PeriodEnd: &periodEnd, Acknowledged: true,
		},
		Acknowledgement: service.GoogleAcknowledgementConfirmed,
	}
	orchestrator := newTestIAPOrchestrator(t, store, &fakeAppleMobileRefresher{}, processed, &fakeGoogleRetryScheduler{})
	response, err := orchestrator.Restore(context.Background(), 7, model.EntitlementStoreGoogle, []string{"token"})
	require.NoError(t, err)
	require.Len(t, response.Restore.Outcomes, 1)
	assert.Equal(t, model.IAPRestoreEvidenceRejected, response.Restore.Outcomes[0].Status)
}

func TestStoreIAPGoogleAcknowledgementRetryIsDurablyScheduledAndDoesNotGrant(t *testing.T) {
	accountToken := uuid.New()
	periodEnd := testIAPNow.Add(30 * 24 * time.Hour)
	store := &fakeMobileIAPStore{
		appleAccount:  model.AppleAccountIdentity{UserID: 7, AppAccountToken: accountToken},
		googleAccount: model.GoogleAccountIdentity{UserID: 7, ObfuscatedExternalAccountID: "opaque-account"},
		entitlements:  map[model.EntitlementStore]model.StoreEntitlement{},
		bindings: map[string]repository.StorePurchaseBinding{
			"purchase-token": {
				ID: 51, UserID: 7, Store: model.EntitlementStoreGoogle, Environment: model.StoreEnvironmentSandbox,
				ProviderRecordID: "purchase-token", AccountIdentifier: "opaque-account", Current: true,
			},
		},
	}
	processed := service.GooglePurchaseProcessingResult{
		Purchase: service.VerifiedGooglePurchase{
			Identity: model.GooglePurchaseIdentity{
				UserID: 7, ObfuscatedExternalAccountID: "opaque-account", PurchaseToken: "purchase-token",
				Environment: model.StoreEnvironmentSandbox, LastVerifiedAt: testIAPNow,
			},
			ProductID: "google.monthly", Entitlement: "premium", Status: model.EntitlementStatusActive,
			PeriodEnd: &periodEnd,
		},
		Acknowledgement: service.GoogleAcknowledgementRetry,
	}
	scheduler := &fakeGoogleRetryScheduler{}
	orchestrator := newTestIAPOrchestrator(t, store, &fakeAppleMobileRefresher{}, processed, scheduler)

	response, err := orchestrator.VerifyGoogle(context.Background(), 7, "purchase-token")
	require.NoError(t, err)
	require.NotNil(t, response.Error)
	assert.True(t, response.Error.Retryable)
	assert.Nil(t, response.CurrentEntitlement)
	require.NotNil(t, response.Pending)
	assert.Equal(t, []int64{51}, scheduler.binding)

	confirmed := processed
	confirmed.Purchase.Acknowledged = true
	confirmed.Acknowledgement = service.GoogleAcknowledgementConfirmed
	persistence, persistenceErr := service.NewStoreEntitlementPersistence(store)
	require.NoError(t, persistenceErr)
	worker, workerErr := service.NewGoogleAcknowledgementRetryWorker(
		store, fakeGoogleMobileProcessor{result: confirmed}, persistence,
		model.StoreEnvironmentSandbox, func() time.Time { return testIAPNow.Add(time.Hour) },
	)
	require.NoError(t, workerErr)
	require.NoError(t, worker.Work(context.Background(), &river.Job[service.GoogleAcknowledgementRetryArgs]{
		JobRow: &rivertype.JobRow{CreatedAt: testIAPNow, Attempt: 1},
		Args:   service.GoogleAcknowledgementRetryArgs{BindingID: 51},
	}))
	response, err = orchestrator.Context(context.Background(), 7, model.EntitlementStoreGoogle)
	require.NoError(t, err)
	assert.Nil(t, response.Pending)
	require.NotNil(t, response.CurrentEntitlement, "server-side retry grants without a second client verify call")
}

func TestEffectiveAccessIsAdditiveAndFailsClosedForStoreReads(t *testing.T) {
	store := &fakeMobileIAPStore{access: true}
	access, err := service.NewEffectiveAccessService(store, model.StoreEnvironmentProduction, func() time.Time { return testIAPNow })
	require.NoError(t, err)
	plan, err := access.Plan(context.Background(), 7, model.PlanFree)
	require.NoError(t, err)
	assert.Equal(t, model.PlanMonthly, plan)

	store.accessErr = errors.New("database unavailable")
	_, err = access.Plan(context.Background(), 7, model.PlanFree)
	require.Error(t, err)
	plan, err = access.Plan(context.Background(), 7, model.PlanAnnual)
	require.NoError(t, err, "valid legacy access remains additive even if store reads fail")
	assert.Equal(t, model.PlanAnnual, plan)
}

func TestStoreIAPLocalLimiterBoundsRequestsAndRequiresConfiguredRedis(t *testing.T) {
	_, err := service.NewStoreIAPRequestLimiter(nil, true, nil)
	require.Error(t, err)
	limiter, err := service.NewStoreIAPRequestLimiter(nil, false, func() time.Time { return testIAPNow })
	require.NoError(t, err)
	for range 10 {
		release, allowed, acquireErr := limiter.Acquire(context.Background(), 7, "verify")
		require.NoError(t, acquireErr)
		require.True(t, allowed)
		release()
	}
	_, allowed, err := limiter.Acquire(context.Background(), 7, "verify")
	require.NoError(t, err)
	assert.False(t, allowed)
}

func TestStoreIAPLimiterRejectsConcurrentCallsAndConfiguredRedisOutage(t *testing.T) {
	limiter, err := service.NewStoreIAPRequestLimiter(nil, false, func() time.Time { return testIAPNow })
	require.NoError(t, err)
	release, allowed, err := limiter.Acquire(context.Background(), 7, "verify")
	require.NoError(t, err)
	require.True(t, allowed)
	_, allowed, err = limiter.Acquire(context.Background(), 7, "restore")
	require.NoError(t, err)
	assert.False(t, allowed, "one user cannot fan out provider calls")
	release()

	redisClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:1", DialTimeout: time.Millisecond, ReadTimeout: time.Millisecond,
		WriteTimeout: time.Millisecond, MaxRetries: 0,
	})
	t.Cleanup(func() { _ = redisClient.Close() })
	configured, err := service.NewStoreIAPRequestLimiter(redisClient, true, nil)
	require.NoError(t, err)
	_, allowed, err = configured.Acquire(context.Background(), 7, "verify")
	require.Error(t, err)
	assert.False(t, allowed, "configured runtime Redis failures fail closed")
}

func TestGoogleAcknowledgementSweepAttemptsEveryCandidate(t *testing.T) {
	periodEnd := testIAPNow.Add(24 * time.Hour)
	store := &fakeMobileIAPStore{
		entitlements: map[model.EntitlementStore]model.StoreEntitlement{},
		bindings: map[string]repository.StorePurchaseBinding{
			"first":  {ID: 1, UserID: 7, Store: model.EntitlementStoreGoogle, Environment: model.StoreEnvironmentSandbox, ProviderRecordID: "first", AccountIdentifier: "opaque-account", Current: true},
			"second": {ID: 2, UserID: 8, Store: model.EntitlementStoreGoogle, Environment: model.StoreEnvironmentSandbox, ProviderRecordID: "second", AccountIdentifier: "opaque-account-2", Current: true},
		},
		candidates: []repository.StoreAcknowledgementRetryCandidate{
			{BindingID: 1, CreatedAt: testIAPNow}, {BindingID: 2, CreatedAt: testIAPNow},
		},
	}
	makePurchase := func(userID int64, account, token string, outcome service.GoogleAcknowledgementOutcome) service.GooglePurchaseProcessingResult {
		return service.GooglePurchaseProcessingResult{
			Purchase: service.VerifiedGooglePurchase{
				Identity:  model.GooglePurchaseIdentity{UserID: userID, ObfuscatedExternalAccountID: account, PurchaseToken: token, Environment: model.StoreEnvironmentSandbox, LastVerifiedAt: testIAPNow},
				ProductID: "google.monthly", Entitlement: "premium", Status: model.EntitlementStatusActive, PeriodEnd: &periodEnd, Acknowledged: outcome == service.GoogleAcknowledgementConfirmed,
			},
			Acknowledgement: outcome,
		}
	}
	processor := &sequenceGoogleMobileProcessor{results: []service.GooglePurchaseProcessingResult{
		makePurchase(7, "opaque-account", "first", service.GoogleAcknowledgementRetry),
		makePurchase(8, "opaque-account-2", "second", service.GoogleAcknowledgementConfirmed),
	}}
	persistence, err := service.NewStoreEntitlementPersistence(store)
	require.NoError(t, err)
	retry, err := service.NewGoogleAcknowledgementRetryWorker(store, processor, persistence, model.StoreEnvironmentSandbox, func() time.Time { return testIAPNow.Add(time.Hour) })
	require.NoError(t, err)
	sweep, err := service.NewGoogleAcknowledgementSweepWorker(store, retry, model.StoreEnvironmentSandbox)
	require.NoError(t, err)
	err = sweep.Work(context.Background(), &river.Job[service.GoogleAcknowledgementSweepArgs]{JobRow: &rivertype.JobRow{CreatedAt: testIAPNow}})
	require.Error(t, err, "the sweep remains retryable after processing the full batch")
	assert.Equal(t, 2, processor.calls)
}
