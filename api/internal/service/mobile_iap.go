package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/google/uuid"
)

const MobileIAPRestoreEvidenceLimit = 20

var ErrMobileIAPRestoreLimit = errors.New("restore evidence limit exceeded")

type MobileIAPRepository interface {
	GetOrCreateAppleAccount(context.Context, int64) (model.AppleAccountIdentity, error)
	GetOrCreateGoogleAccount(context.Context, int64) (model.GoogleAccountIdentity, error)
	GetCurrentEntitlement(context.Context, int64, model.EntitlementStore, model.StoreEnvironment) (model.StoreEntitlement, bool, error)
	CurrentGoogleBindingAcknowledged(context.Context, int64) (bool, error)
	ListUserPurchaseBindings(context.Context, int64, model.EntitlementStore, model.StoreEnvironment, int) ([]repository.StorePurchaseBinding, error)
	ResolvePurchaseBinding(context.Context, model.EntitlementStore, model.StoreEnvironment, string) (repository.StorePurchaseBinding, bool, error)
}

type AppleMobilePurchaseVerifier interface {
	Verify(context.Context, ApplePurchaseVerificationInput) (VerifiedApplePurchase, error)
}

type GoogleMobilePurchaseVerifier interface {
	Verify(context.Context, GooglePurchaseVerificationInput) (VerifiedGooglePurchase, error)
}

type GoogleAcknowledgementRetryScheduler interface {
	ScheduleGoogleAcknowledgementRetry(context.Context, int64) error
}

type StoreIAPOrchestrator struct {
	repository      MobileIAPRepository
	appleVerifier   AppleMobilePurchaseVerifier
	appleRefresher  AppleProviderSubscriptionRefresher
	googleProcessor GoogleProviderPurchaseProcessor
	persistence     *StoreEntitlementPersistence
	retryScheduler  GoogleAcknowledgementRetryScheduler
	environment     model.StoreEnvironment
	catalog         map[model.EntitlementStore][]model.MobileIAPProduct
	now             func() time.Time
}

func NewStoreIAPOrchestrator(
	repository MobileIAPRepository,
	appleVerifier AppleMobilePurchaseVerifier,
	appleRefresher AppleProviderSubscriptionRefresher,
	googleProcessor GoogleProviderPurchaseProcessor,
	persistence *StoreEntitlementPersistence,
	retryScheduler GoogleAcknowledgementRetryScheduler,
	environment model.StoreEnvironment,
	products []model.MobileIAPProduct,
	now func() time.Time,
) (*StoreIAPOrchestrator, error) {
	if repository == nil || appleVerifier == nil || appleRefresher == nil ||
		googleProcessor == nil || persistence == nil || retryScheduler == nil || !environment.Valid() {
		return nil, fmt.Errorf("mobile IAP orchestration dependencies are required")
	}
	catalog := map[model.EntitlementStore][]model.MobileIAPProduct{
		model.EntitlementStoreApple: {}, model.EntitlementStoreGoogle: {},
	}
	seen := make(map[string]struct{}, len(products))
	for _, product := range products {
		if err := product.Validate(); err != nil {
			return nil, err
		}
		key := string(product.Store) + "\x00" + product.ProductID
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("duplicate mobile IAP product")
		}
		seen[key] = struct{}{}
		catalog[product.Store] = append(catalog[product.Store], product)
	}
	for store := range catalog {
		sort.Slice(catalog[store], func(i, j int) bool {
			return catalog[store][i].ProductID < catalog[store][j].ProductID
		})
	}
	if now == nil {
		now = time.Now
	}
	return &StoreIAPOrchestrator{
		repository: repository, appleVerifier: appleVerifier, appleRefresher: appleRefresher,
		googleProcessor: googleProcessor, persistence: persistence,
		retryScheduler: retryScheduler, environment: environment, catalog: catalog, now: now,
	}, nil
}

func (s *StoreIAPOrchestrator) Context(
	ctx context.Context,
	userID int64,
	store model.EntitlementStore,
) (model.MobileIAPResponse, error) {
	if userID <= 0 || !store.Valid() {
		return model.MobileIAPResponse{}, fmt.Errorf("valid mobile IAP context is required")
	}
	response := model.MobileIAPResponse{
		Products: append([]model.MobileIAPProduct(nil), s.catalog[store]...), ServerTime: s.now().UTC(),
	}
	switch store {
	case model.EntitlementStoreApple:
		account, err := s.repository.GetOrCreateAppleAccount(ctx, userID)
		if err != nil {
			return model.MobileIAPResponse{}, err
		}
		response.PurchaseContext = &model.MobileIAPPurchaseContext{
			Store: store, AppleAppAccountToken: &account.AppAccountToken,
		}
	case model.EntitlementStoreGoogle:
		account, err := s.repository.GetOrCreateGoogleAccount(ctx, userID)
		if err != nil {
			return model.MobileIAPResponse{}, err
		}
		response.PurchaseContext = &model.MobileIAPPurchaseContext{
			Store: store, GoogleObfuscatedAccountID: &account.ObfuscatedExternalAccountID,
		}
	}
	if err := s.populateState(ctx, userID, store, &response); err != nil {
		return model.MobileIAPResponse{}, err
	}
	return response, response.Validate()
}

func (s *StoreIAPOrchestrator) VerifyApple(
	ctx context.Context,
	userID int64,
	transactionID string,
) (model.MobileIAPResponse, error) {
	account, err := s.repository.GetOrCreateAppleAccount(ctx, userID)
	if err != nil {
		return s.safeErrorResponse(true), err
	}
	purchase, err := s.appleVerifier.Verify(ctx, ApplePurchaseVerificationInput{
		AuthenticatedUserID: userID, TransactionID: transactionID, Account: account,
	})
	if err != nil || purchase.Identity.Environment != s.environment {
		return s.providerErrorResponse(err), nil
	}
	verified, err := s.appleRefresher.Refresh(ctx, AppleSubscriptionRefreshInput{
		AuthenticatedUserID: userID, Account: account,
		OriginalTransactionID: purchase.Identity.OriginalTransactionID, Environment: s.environment,
	})
	if err != nil {
		return s.providerErrorResponse(err), nil
	}
	result, err := s.persistence.PersistApple(ctx, nil, verified, nil)
	if err != nil {
		return s.safeErrorResponse(true), err
	}
	if result.Operation.Outcome == model.EntitlementOutcomeRejected {
		return s.safeErrorResponse(false), nil
	}
	return s.Context(ctx, userID, model.EntitlementStoreApple)
}

func (s *StoreIAPOrchestrator) VerifyGoogle(
	ctx context.Context,
	userID int64,
	purchaseToken string,
) (model.MobileIAPResponse, error) {
	account, err := s.repository.GetOrCreateGoogleAccount(ctx, userID)
	if err != nil {
		return s.safeErrorResponse(true), err
	}
	processed, err := s.googleProcessor.VerifyAndAcknowledge(ctx, GooglePurchaseVerificationInput{
		AuthenticatedUserID: userID, PurchaseToken: purchaseToken, Account: account,
	})
	if err != nil || processed.Purchase.Identity.Environment != s.environment {
		return s.providerErrorResponse(err), nil
	}
	result, err := s.persistence.PersistGoogle(ctx, nil, processed, nil)
	if err != nil {
		return s.safeErrorResponse(true), err
	}
	if result.Operation.Outcome == model.EntitlementOutcomeRejected {
		return s.safeErrorResponse(false), nil
	}
	if processed.Acknowledgement == GoogleAcknowledgementRetry {
		binding, found, resolveErr := s.repository.ResolvePurchaseBinding(
			ctx, model.EntitlementStoreGoogle, s.environment, purchaseToken,
		)
		if resolveErr != nil || !found || s.retryScheduler.ScheduleGoogleAcknowledgementRetry(ctx, binding.ID) != nil {
			return s.safeErrorResponse(true), resolveErr
		}
		response, contextErr := s.Context(ctx, userID, model.EntitlementStoreGoogle)
		response.Error = retryableMobileIAPError()
		return response, contextErr
	}
	if processed.Acknowledgement == GoogleAcknowledgementFailed {
		return s.safeErrorResponse(false), nil
	}
	return s.Context(ctx, userID, model.EntitlementStoreGoogle)
}

func (s *StoreIAPOrchestrator) Restore(
	ctx context.Context,
	userID int64,
	store model.EntitlementStore,
	evidence []string,
) (model.MobileIAPResponse, error) {
	if err := validateMobileIAPRestoreInput(userID, store, evidence); err != nil {
		return model.MobileIAPResponse{}, err
	}
	known, err := s.repository.ListUserPurchaseBindings(ctx, userID, store, s.environment, MobileIAPRestoreEvidenceLimit)
	if err != nil {
		return s.safeErrorResponse(true), err
	}
	result := model.MobileIAPRestoreResult{Outcomes: make([]model.MobileIAPRestoreEvidenceResult, 0, len(evidence))}
	if len(evidence) == 0 && len(known) == 0 {
		response, contextErr := s.Context(ctx, userID, store)
		if contextErr != nil {
			return response, contextErr
		}
		result.Status = model.IAPRestoreNoPurchases
		response.Restore = &result
		return response, response.Validate()
	}

	seen := make(map[string]struct{}, len(evidence)+len(known))
	s.processClientRestoreEvidence(ctx, userID, store, evidence, len(known), seen, &result)
	s.processKnownRestoreBindings(ctx, store, known, seen, &result)
	for _, outcome := range result.Outcomes {
		incrementRestoreCount(&result, outcome.Status)
	}
	finalizeRestoreResult(&result)
	response, err := s.Context(ctx, userID, store)
	if err != nil {
		return response, err
	}
	response.Restore = &result
	return response, response.Validate()
}

func validateMobileIAPRestoreInput(userID int64, store model.EntitlementStore, evidence []string) error {
	if len(evidence) > MobileIAPRestoreEvidenceLimit {
		return ErrMobileIAPRestoreLimit
	}
	if userID <= 0 || !store.Valid() {
		return fmt.Errorf("valid restore request is required")
	}
	return nil
}

func (s *StoreIAPOrchestrator) processClientRestoreEvidence(
	ctx context.Context,
	userID int64,
	store model.EntitlementStore,
	evidence []string,
	knownCount int,
	seen map[string]struct{},
	result *model.MobileIAPRestoreResult,
) {
	for index, item := range evidence {
		outcome, key := retryableRestoreEvidence(), ""
		if ctx.Err() == nil {
			callCtx, cancel := restoreCallContext(ctx, len(evidence)-index+knownCount)
			outcome, key = s.restoreEvidence(callCtx, userID, store, item, seen)
			cancel()
		}
		outcome.Index = index
		result.Outcomes = append(result.Outcomes, outcome)
		if key != "" {
			seen[key] = struct{}{}
		}
	}
}

func (s *StoreIAPOrchestrator) processKnownRestoreBindings(
	ctx context.Context,
	store model.EntitlementStore,
	known []repository.StorePurchaseBinding,
	seen map[string]struct{},
	result *model.MobileIAPRestoreResult,
) {
	for index, binding := range known {
		key := string(store) + "\x00" + binding.ProviderRecordID
		if _, exists := seen[key]; exists {
			continue
		}
		outcome := retryableRestoreEvidence()
		if ctx.Err() == nil {
			callCtx, cancel := restoreCallContext(ctx, len(known)-index)
			operation, refreshErr := s.refreshKnownBinding(callCtx, binding)
			cancel()
			outcome = restoreOperationEvidence(operation, refreshErr)
		}
		incrementRestoreCount(result, outcome.Status)
		seen[key] = struct{}{}
	}
}

func incrementRestoreCount(result *model.MobileIAPRestoreResult, status model.IAPRestoreEvidenceStatus) {
	switch status {
	case model.IAPRestoreEvidenceRestored:
		result.RestoredPurchases++
	case model.IAPRestoreEvidenceUnchanged:
		result.UnchangedPurchases++
	case model.IAPRestoreEvidenceRejected:
		result.RejectedPurchases++
	case model.IAPRestoreEvidenceRetryable:
		result.RetryablePurchases++
	}
}

func finalizeRestoreResult(result *model.MobileIAPRestoreResult) {
	result.Status = restoreAggregate(*result)
	if result.Status == model.IAPRestoreFailed {
		if result.RetryablePurchases > 0 {
			result.Error = retryableMobileIAPError()
		} else {
			result.Error = invalidMobileIAPError()
		}
	}
}

func (s *StoreIAPOrchestrator) restoreEvidence(
	ctx context.Context,
	userID int64,
	store model.EntitlementStore,
	evidence string,
	seen map[string]struct{},
) (model.MobileIAPRestoreEvidenceResult, string) {
	if store == model.EntitlementStoreApple {
		account, err := s.repository.GetOrCreateAppleAccount(ctx, userID)
		if err != nil {
			return retryableRestoreEvidence(), ""
		}
		purchase, err := s.appleVerifier.Verify(ctx, ApplePurchaseVerificationInput{
			AuthenticatedUserID: userID, TransactionID: evidence, Account: account,
		})
		if err != nil || purchase.Identity.Environment != s.environment {
			return restoreErrorEvidence(err), ""
		}
		key := string(store) + "\x00" + purchase.Identity.OriginalTransactionID
		if _, exists := seen[key]; exists {
			return model.MobileIAPRestoreEvidenceResult{Status: model.IAPRestoreEvidenceUnchanged}, key
		}
		binding, found, err := s.repository.ResolvePurchaseBinding(ctx, store, s.environment, purchase.Identity.OriginalTransactionID)
		if err != nil {
			return retryableRestoreEvidence(), key
		}
		if !found || binding.UserID != userID {
			return rejectedRestoreEvidence(), key
		}
		operation, err := s.refreshKnownBinding(ctx, binding)
		return restoreOperationEvidence(operation, err), key
	}

	binding, found, err := s.repository.ResolvePurchaseBinding(ctx, store, s.environment, evidence)
	if err != nil {
		return retryableRestoreEvidence(), ""
	}
	if !found || binding.UserID != userID {
		return rejectedRestoreEvidence(), ""
	}
	key := string(store) + "\x00" + binding.ProviderRecordID
	if _, exists := seen[key]; exists {
		return model.MobileIAPRestoreEvidenceResult{Status: model.IAPRestoreEvidenceUnchanged}, key
	}
	operation, err := s.refreshKnownBinding(ctx, binding)
	return restoreOperationEvidence(operation, err), key
}

func (s *StoreIAPOrchestrator) refreshKnownBinding(
	ctx context.Context,
	binding repository.StorePurchaseBinding,
) (model.EntitlementOperationResult, error) {
	if binding.Store == model.EntitlementStoreApple {
		token, err := uuid.Parse(binding.AccountIdentifier)
		if err != nil {
			return model.EntitlementOperationResult{}, err
		}
		verified, err := s.appleRefresher.Refresh(ctx, AppleSubscriptionRefreshInput{
			AuthenticatedUserID:   binding.UserID,
			Account:               model.AppleAccountIdentity{UserID: binding.UserID, AppAccountToken: token},
			OriginalTransactionID: binding.ProviderRecordID, Environment: s.environment,
		})
		if err != nil {
			return model.EntitlementOperationResult{}, err
		}
		result, err := s.persistence.PersistApple(ctx, nil, verified, nil)
		return result.Operation, err
	}
	account := model.GoogleAccountIdentity{
		UserID: binding.UserID, ObfuscatedExternalAccountID: binding.AccountIdentifier,
	}
	processed, err := s.googleProcessor.VerifyAndAcknowledge(ctx, GooglePurchaseVerificationInput{
		AuthenticatedUserID: binding.UserID, PurchaseToken: binding.ProviderRecordID, Account: account,
	})
	if err != nil {
		return model.EntitlementOperationResult{}, err
	}
	if processed.Purchase.Identity.Environment != s.environment {
		return model.EntitlementOperationResult{}, &GooglePurchaseVerificationError{
			Failure: model.EntitlementFailureInvalidEvidence,
		}
	}
	result, err := s.persistence.PersistGoogle(ctx, nil, processed, nil)
	if err != nil {
		return model.EntitlementOperationResult{}, err
	}
	if processed.Acknowledgement == GoogleAcknowledgementRetry {
		if err = s.retryScheduler.ScheduleGoogleAcknowledgementRetry(ctx, binding.ID); err != nil {
			return model.EntitlementOperationResult{}, err
		}
		return result.Operation, &GooglePurchaseVerificationError{Failure: model.EntitlementFailureProviderUnavailable}
	}
	return result.Operation, nil
}

func (s *StoreIAPOrchestrator) populateState(
	ctx context.Context,
	userID int64,
	store model.EntitlementStore,
	response *model.MobileIAPResponse,
) error {
	entitlement, found, err := s.repository.GetCurrentEntitlement(ctx, userID, store, s.environment)
	if err != nil || !found {
		return err
	}
	pending := entitlement.Status == model.EntitlementStatusPending
	if store == model.EntitlementStoreGoogle && !pending &&
		(entitlement.Status == model.EntitlementStatusActive || entitlement.Status == model.EntitlementStatusGracePeriod) {
		acknowledged, ackErr := s.repository.CurrentGoogleBindingAcknowledged(ctx, entitlement.ID)
		if ackErr != nil {
			return ackErr
		}
		pending = !acknowledged
	}
	if pending {
		response.Pending = &model.MobileIAPPending{
			Store: store, ProductID: entitlement.ProductID, StartedAt: entitlement.CreatedAt,
		}
		return nil
	}
	response.CurrentEntitlement = mobileEntitlement(entitlement)
	return nil
}

func mobileEntitlement(entitlement model.StoreEntitlement) *model.MobileIAPEntitlement {
	return &model.MobileIAPEntitlement{
		Store: entitlement.Store, ProductID: entitlement.ProductID, Entitlement: entitlement.Entitlement,
		Status: entitlement.Status, PeriodEnd: entitlement.PeriodEnd,
		AutoRenewEnabled: entitlement.AutoRenewEnabled, CanceledAt: entitlement.CanceledAt,
		RevokedAt: entitlement.RevokedAt,
	}
}

func (s *StoreIAPOrchestrator) providerErrorResponse(err error) model.MobileIAPResponse {
	retryable := providerFailureRetryable(err)
	return s.safeErrorResponse(retryable)
}

func (s *StoreIAPOrchestrator) safeErrorResponse(retryable bool) model.MobileIAPResponse {
	response := model.MobileIAPResponse{Products: []model.MobileIAPProduct{}, ServerTime: s.now().UTC()}
	if retryable {
		response.Error = retryableMobileIAPError()
	} else {
		response.Error = invalidMobileIAPError()
	}
	return response
}

func providerFailureRetryable(err error) bool {
	var apple *ApplePurchaseVerificationError
	if errors.As(err, &apple) {
		return apple.Failure == model.EntitlementFailureProviderUnavailable
	}
	var google *GooglePurchaseVerificationError
	if errors.As(err, &google) {
		return google.Failure == model.EntitlementFailureProviderUnavailable
	}
	return err != nil
}

func retryableMobileIAPError() *model.MobileIAPError {
	return &model.MobileIAPError{
		Code: model.EntitlementResultRetryableProvider, MessageKey: "iap.error.unavailable", Retryable: true,
	}
}

func invalidMobileIAPError() *model.MobileIAPError {
	return &model.MobileIAPError{
		Code: model.EntitlementResultInvalidPurchase, MessageKey: "iap.error.invalid_purchase",
	}
}

func rejectedRestoreEvidence() model.MobileIAPRestoreEvidenceResult {
	return model.MobileIAPRestoreEvidenceResult{Status: model.IAPRestoreEvidenceRejected, Error: invalidMobileIAPError()}
}

func retryableRestoreEvidence() model.MobileIAPRestoreEvidenceResult {
	return model.MobileIAPRestoreEvidenceResult{Status: model.IAPRestoreEvidenceRetryable, Error: retryableMobileIAPError()}
}

func restoreErrorEvidence(err error) model.MobileIAPRestoreEvidenceResult {
	if providerFailureRetryable(err) {
		return retryableRestoreEvidence()
	}
	return rejectedRestoreEvidence()
}

func restoreOperationEvidence(
	operation model.EntitlementOperationResult,
	err error,
) model.MobileIAPRestoreEvidenceResult {
	if err != nil {
		return restoreErrorEvidence(err)
	}
	if operation.Outcome == model.EntitlementOutcomeApplied {
		return model.MobileIAPRestoreEvidenceResult{Status: model.IAPRestoreEvidenceRestored}
	}
	if operation.Outcome == model.EntitlementOutcomeRejected {
		return rejectedRestoreEvidence()
	}
	return model.MobileIAPRestoreEvidenceResult{Status: model.IAPRestoreEvidenceUnchanged}
}

func restoreAggregate(result model.MobileIAPRestoreResult) model.IAPRestoreStatus {
	success := result.RestoredPurchases + result.UnchangedPurchases
	failures := result.RejectedPurchases + result.RetryablePurchases
	if success > 0 && failures > 0 {
		return model.IAPRestorePartial
	}
	if result.RejectedPurchases > 0 && result.RetryablePurchases > 0 {
		return model.IAPRestorePartial
	}
	if result.RetryablePurchases > 0 {
		return model.IAPRestorePending
	}
	if success > 0 {
		return model.IAPRestoreRestored
	}
	return model.IAPRestoreFailed
}

func restoreCallContext(parent context.Context, callsRemaining int) (context.Context, context.CancelFunc) {
	budget := 5 * time.Second
	if callsRemaining < 1 {
		callsRemaining = 1
	}
	if deadline, exists := parent.Deadline(); exists {
		remaining := time.Until(deadline)
		shared := remaining / time.Duration(callsRemaining)
		if callsRemaining == 1 {
			shared = remaining - 500*time.Millisecond
			if shared > 15*time.Second {
				shared = 15 * time.Second
			}
		}
		if shared < budget {
			budget = shared
		}
	}
	if budget <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, budget)
}
