package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"decorebator.com/internal/model"
	"github.com/google/uuid"
)

const appleSubscriptionExpirySkew = 5 * time.Minute

var errUnsupportedAppleSubscriptionStatus = errors.New("unsupported Apple subscription status")

type AppleSubscriptionRefreshInput struct {
	AuthenticatedUserID   int64
	Account               model.AppleAccountIdentity
	OriginalTransactionID string
	Environment           model.StoreEnvironment
}

type VerifiedAppleSubscription struct {
	Identity         model.ApplePurchaseIdentity
	TransactionID    string
	ProductID        string
	Entitlement      string
	Status           model.EntitlementStatus
	PeriodStart      *time.Time
	PeriodEnd        *time.Time
	AutoRenewEnabled bool
	// CancellationObservedAt is the signed snapshot observation time, not
	// Apple's unavailable exact user-cancellation time.
	CancellationObservedAt *time.Time
	RevokedAt              *time.Time
	RevocationReason       *string
	ExpirationIntent       *int32
	// ProviderSnapshotSignedAt must never be compared with notification event
	// occurrence time; status refreshes are authoritative snapshots.
	ProviderSnapshotSignedAt time.Time
}

type AppleSubscriptionRefresher struct {
	client     AppleSubscriptionStatusClient
	signedData AppleNotificationSignedDataVerifier
	bundleID   string
	appAppleID int64
	products   map[string]string
	now        func() time.Time
}

func NewAppleSubscriptionRefresher(
	client AppleSubscriptionStatusClient,
	signedData AppleNotificationSignedDataVerifier,
	bundleID string,
	appAppleID int64,
	products map[string]string,
	now func() time.Time,
) (*AppleSubscriptionRefresher, error) {
	if client == nil || signedData == nil || strings.TrimSpace(bundleID) == "" || appAppleID <= 0 || len(products) == 0 {
		return nil, fmt.Errorf("Apple subscription refresher configuration is incomplete")
	}
	productCopy := make(map[string]string, len(products))
	for productID, entitlement := range products {
		if strings.TrimSpace(productID) == "" || entitlement != premiumEntitlement {
			return nil, fmt.Errorf("invalid Apple subscription product catalog entry")
		}
		productCopy[productID] = entitlement
	}
	if now == nil {
		now = time.Now
	}
	return &AppleSubscriptionRefresher{
		client: client, signedData: signedData, bundleID: bundleID, appAppleID: appAppleID,
		products: productCopy, now: now,
	}, nil
}

func (r *AppleSubscriptionRefresher) Refresh(
	ctx context.Context,
	input AppleSubscriptionRefreshInput,
) (VerifiedAppleSubscription, error) {
	if input.AuthenticatedUserID <= 0 || input.Account.Validate() != nil || input.AuthenticatedUserID != input.Account.UserID {
		return VerifiedAppleSubscription{}, applePurchaseFailure(model.EntitlementFailureAccountMismatch, nil)
	}
	if !validAppleTransactionID(input.OriginalTransactionID) || !input.Environment.Valid() {
		return VerifiedAppleSubscription{}, applePurchaseFailure(model.EntitlementFailureInvalidEvidence, nil)
	}
	response, err := r.client.GetSubscriptionStatuses(ctx, input.OriginalTransactionID, input.Environment)
	if err != nil {
		return VerifiedAppleSubscription{}, appleProviderFailure(err)
	}
	return r.validateResponse(input, response)
}

func (r *AppleSubscriptionRefresher) validateResponse(
	input AppleSubscriptionRefreshInput,
	response AppleSubscriptionStatusResponse,
) (VerifiedAppleSubscription, error) {
	environment, ok := appleEnvironment(response.Environment)
	if !ok || environment != input.Environment || response.BundleID != r.bundleID ||
		(environment == model.StoreEnvironmentProduction && response.AppAppleID != r.appAppleID) ||
		(environment == model.StoreEnvironmentSandbox && response.AppAppleID != 0 && response.AppAppleID != r.appAppleID) {
		return VerifiedAppleSubscription{}, applePurchaseFailure(model.EntitlementFailureInvalidEvidence, nil)
	}
	var selected *AppleSubscriptionLastTransaction
	for groupIndex := range response.Data {
		for itemIndex := range response.Data[groupIndex].LastTransactions {
			item := &response.Data[groupIndex].LastTransactions[itemIndex]
			if item.OriginalTransactionID == input.OriginalTransactionID {
				if selected != nil {
					return VerifiedAppleSubscription{}, applePurchaseFailure(model.EntitlementFailureInvalidEvidence, nil)
				}
				selected = item
			}
		}
	}
	if selected == nil || selected.SignedTransactionInfo == "" || selected.SignedRenewalInfo == "" {
		return VerifiedAppleSubscription{}, applePurchaseFailure(model.EntitlementFailureInvalidEvidence, nil)
	}
	transaction, err := r.signedData.VerifyAndDecodeTransaction(selected.SignedTransactionInfo)
	if err != nil {
		return VerifiedAppleSubscription{}, applePurchaseFailure(model.EntitlementFailureInvalidEvidence, err)
	}
	renewal, err := r.signedData.VerifyAndDecodeRenewalInfo(selected.SignedRenewalInfo)
	if err != nil {
		return VerifiedAppleSubscription{}, applePurchaseFailure(model.EntitlementFailureInvalidEvidence, err)
	}
	return r.normalize(input, *selected, transaction, renewal)
}

func (r *AppleSubscriptionRefresher) normalize(
	input AppleSubscriptionRefreshInput,
	item AppleSubscriptionLastTransaction,
	transaction AppleTransactionPayload,
	renewal AppleRenewalInfoPayload,
) (VerifiedAppleSubscription, error) {
	now := r.now().UTC()
	entitlement, accountToken, transactionSignedAt, renewalSignedAt, err :=
		r.validateSignedStatusEvidence(input, item, transaction, renewal, now)
	if err != nil {
		return VerifiedAppleSubscription{}, err
	}
	periodStart := time.UnixMilli(transaction.PurchaseDateMS).UTC()
	revokedAt, revocationReason, err := appleRevocation(item.Status, transaction, periodStart, transactionSignedAt)
	if err != nil {
		return VerifiedAppleSubscription{}, applePurchaseFailure(model.EntitlementFailureInvalidEvidence, err)
	}
	status, periodEnd, err := appleStatusPeriod(item.Status, transaction, renewal, revokedAt != nil)
	if errors.Is(err, errUnsupportedAppleSubscriptionStatus) {
		return VerifiedAppleSubscription{}, applePurchaseFailure(model.EntitlementFailureProviderUnavailable, err)
	}
	if err != nil {
		return VerifiedAppleSubscription{}, applePurchaseFailure(model.EntitlementFailureInvalidEvidence, err)
	}
	var cancellationObservedAt *time.Time
	if renewal.AutoRenewStatus == 0 &&
		(status == model.EntitlementStatusActive || status == model.EntitlementStatusGracePeriod) {
		value := renewalSignedAt
		cancellationObservedAt = &value
	}
	var expirationIntent *int32
	if renewal.ExpirationIntent > 0 {
		value := renewal.ExpirationIntent
		expirationIntent = &value
	}
	identity := model.ApplePurchaseIdentity{
		UserID: input.AuthenticatedUserID, AppAccountToken: accountToken,
		OriginalTransactionID: input.OriginalTransactionID, Environment: input.Environment, LastVerifiedAt: now,
	}
	providerSnapshotSignedAt := transactionSignedAt
	if renewalSignedAt.After(transactionSignedAt) {
		providerSnapshotSignedAt = renewalSignedAt
	}
	return VerifiedAppleSubscription{
		Identity: identity, TransactionID: transaction.TransactionID, ProductID: transaction.ProductID,
		Entitlement: entitlement, Status: status, PeriodStart: &periodStart, PeriodEnd: &periodEnd,
		AutoRenewEnabled: renewal.AutoRenewStatus == 1, CancellationObservedAt: cancellationObservedAt,
		RevokedAt: revokedAt, RevocationReason: revocationReason, ExpirationIntent: expirationIntent,
		ProviderSnapshotSignedAt: providerSnapshotSignedAt,
	}, nil
}

func (r *AppleSubscriptionRefresher) validateSignedStatusEvidence(
	input AppleSubscriptionRefreshInput,
	item AppleSubscriptionLastTransaction,
	transaction AppleTransactionPayload,
	renewal AppleRenewalInfoPayload,
	now time.Time,
) (string, uuid.UUID, time.Time, time.Time, error) {
	if err := r.validateStatusLinkage(input, item, transaction, renewal); err != nil {
		return "", uuid.Nil, time.Time{}, time.Time{}, err
	}
	entitlement, err := r.validateStatusProductsAndTimes(transaction, renewal)
	if err != nil {
		return "", uuid.Nil, time.Time{}, time.Time{}, err
	}
	accountToken, accountErr := uuid.Parse(transaction.AppAccountToken)
	if accountErr != nil || accountToken != input.Account.AppAccountToken {
		return "", uuid.Nil, time.Time{}, time.Time{}, applePurchaseFailure(model.EntitlementFailureAccountMismatch, accountErr)
	}
	transactionSignedAt := time.UnixMilli(transaction.SignedDateMS).UTC()
	renewalSignedAt := time.UnixMilli(renewal.SignedDateMS).UTC()
	if transactionSignedAt.After(now.Add(appleSubscriptionExpirySkew)) ||
		renewalSignedAt.After(now.Add(appleSubscriptionExpirySkew)) || transactionSignedAt.Before(time.UnixMilli(transaction.PurchaseDateMS).UTC()) {
		return "", uuid.Nil, time.Time{}, time.Time{}, applePurchaseFailure(model.EntitlementFailureInvalidEvidence, nil)
	}
	return entitlement, accountToken, transactionSignedAt, renewalSignedAt, nil
}

func (r *AppleSubscriptionRefresher) validateStatusLinkage(
	input AppleSubscriptionRefreshInput,
	item AppleSubscriptionLastTransaction,
	transaction AppleTransactionPayload,
	renewal AppleRenewalInfoPayload,
) error {
	transactionEnvironment, transactionEnvironmentOK := appleEnvironment(transaction.Environment)
	renewalEnvironment, renewalEnvironmentOK := appleEnvironment(renewal.Environment)
	if !transactionEnvironmentOK || !renewalEnvironmentOK || transactionEnvironment != input.Environment || renewalEnvironment != input.Environment {
		return applePurchaseFailure(model.EntitlementFailureInvalidEvidence, nil)
	}
	if transaction.BundleID != r.bundleID || transaction.OriginalTransactionID != input.OriginalTransactionID ||
		renewal.OriginalTransactionID != input.OriginalTransactionID || item.OriginalTransactionID != input.OriginalTransactionID ||
		transaction.ProductType != appleAutoRenewableSubscription || transaction.TransactionID == "" {
		return applePurchaseFailure(model.EntitlementFailureInvalidEvidence, nil)
	}
	return nil
}

func (r *AppleSubscriptionRefresher) validateStatusProductsAndTimes(
	transaction AppleTransactionPayload,
	renewal AppleRenewalInfoPayload,
) (string, error) {
	entitlement, knownProduct := r.products[transaction.ProductID]
	_, knownRenewalProduct := r.products[renewal.ProductID]
	if !knownProduct || !knownRenewalProduct {
		return "", applePurchaseFailure(model.EntitlementFailureUnknownProduct, nil)
	}
	if renewal.ProductID != transaction.ProductID || (renewal.AutoRenewStatus == 1 && renewal.AutoRenewProductID == "") {
		return "", applePurchaseFailure(model.EntitlementFailureInvalidEvidence, nil)
	}
	if transaction.PurchaseDateMS <= 0 || transaction.ExpiresDateMS <= transaction.PurchaseDateMS ||
		transaction.SignedDateMS <= 0 || renewal.SignedDateMS <= 0 || renewal.AutoRenewStatus < 0 || renewal.AutoRenewStatus > 1 {
		return "", applePurchaseFailure(model.EntitlementFailureInvalidEvidence, nil)
	}
	return entitlement, nil
}

func appleStatusPeriod(
	providerStatus int32,
	transaction AppleTransactionPayload,
	renewal AppleRenewalInfoPayload,
	revoked bool,
) (model.EntitlementStatus, time.Time, error) {
	if revoked {
		return model.EntitlementStatusRevoked, time.UnixMilli(transaction.ExpiresDateMS).UTC(), nil
	}
	status, ok := appleCanonicalSubscriptionStatus(providerStatus)
	if !ok {
		return "", time.Time{}, errUnsupportedAppleSubscriptionStatus
	}
	periodEnd := time.UnixMilli(transaction.ExpiresDateMS).UTC()
	if status == model.EntitlementStatusGracePeriod {
		graceEnd := time.UnixMilli(renewal.GracePeriodExpiresDateMS).UTC()
		if renewal.GracePeriodExpiresDateMS <= transaction.ExpiresDateMS {
			return "", time.Time{}, fmt.Errorf("invalid Apple grace-period end")
		}
		periodEnd = graceEnd
	}
	return status, periodEnd, nil
}

func appleRevocation(
	providerStatus int32,
	transaction AppleTransactionPayload,
	periodStart time.Time,
	transactionSignedAt time.Time,
) (*time.Time, *string, error) {
	if providerStatus == 5 {
		if transaction.RevocationDateMS == nil {
			value := transactionSignedAt
			return &value, nil, nil
		}
		value := time.UnixMilli(*transaction.RevocationDateMS).UTC()
		if value.Before(periodStart) || value.After(transactionSignedAt.Add(appleSubscriptionExpirySkew)) {
			return nil, nil, fmt.Errorf("invalid Apple revocation date")
		}
		return &value, appleRevocationReason(transaction.RevocationReason), nil
	}
	if transaction.RevocationDateMS != nil {
		value := time.UnixMilli(*transaction.RevocationDateMS).UTC()
		if value.Before(periodStart) || value.After(transactionSignedAt.Add(appleSubscriptionExpirySkew)) {
			return nil, nil, fmt.Errorf("invalid Apple revocation date")
		}
		return &value, appleRevocationReason(transaction.RevocationReason), nil
	}
	return nil, nil, nil
}

func appleCanonicalSubscriptionStatus(status int32) (model.EntitlementStatus, bool) {
	switch status {
	case 1:
		return model.EntitlementStatusActive, true
	case 2:
		return model.EntitlementStatusExpired, true
	case 3:
		return model.EntitlementStatusOnHold, true
	case 4:
		return model.EntitlementStatusGracePeriod, true
	case 5:
		return model.EntitlementStatusRevoked, true
	default:
		return "", false
	}
}
