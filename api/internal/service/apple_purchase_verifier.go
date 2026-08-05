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

const (
	appleAutoRenewableSubscription = "Auto-Renewable Subscription"
	premiumEntitlement             = "premium"
)

type ApplePurchaseVerificationInput struct {
	AuthenticatedUserID int64
	TransactionID       string
	Account             model.AppleAccountIdentity
}

// VerifiedApplePurchase is signature-, app-, product-, environment-, and
// account-bound transaction evidence. It is not an entitlement decision:
// Apple grace, billing-retry, and auto-renew state require verified renewal
// information from Get All Subscription Statuses. Callers must also keep
// sandbox evidence isolated from production access.
type VerifiedApplePurchase struct {
	Identity         model.ApplePurchaseIdentity
	TransactionID    string
	ProductID        string
	Entitlement      string
	PurchaseDate     time.Time
	ExpiresDate      time.Time
	RevokedAt        *time.Time
	RevocationReason *string
	ProviderSignedAt time.Time
}

type ApplePurchaseVerificationError struct {
	Failure model.EntitlementOperationFailure
	Cause   error
}

func (e *ApplePurchaseVerificationError) Error() string {
	return fmt.Sprintf("Apple purchase verification failed: %s", e.Failure)
}

func (e *ApplePurchaseVerificationError) Unwrap() error {
	return e.Cause
}

type ApplePurchaseVerifier struct {
	client      AppleTransactionInfoClient
	signedData  AppleSignedDataVerifier
	bundleID    string
	products    map[string]string
	environment model.StoreEnvironment
	now         func() time.Time
}

func NewApplePurchaseVerifier(
	client AppleTransactionInfoClient,
	signedData AppleSignedDataVerifier,
	bundleID string,
	products map[string]string,
	environment model.StoreEnvironment,
	now func() time.Time,
) (*ApplePurchaseVerifier, error) {
	if client == nil || signedData == nil {
		return nil, fmt.Errorf("Apple API client and signed-data verifier are required")
	}
	if strings.TrimSpace(bundleID) == "" || len(products) == 0 || !environment.Valid() {
		return nil, fmt.Errorf("Apple bundle ID and product catalog are required")
	}
	productCopy := make(map[string]string, len(products))
	for productID, entitlement := range products {
		if strings.TrimSpace(productID) == "" || entitlement != premiumEntitlement {
			return nil, fmt.Errorf("invalid Apple product catalog entry")
		}
		productCopy[productID] = entitlement
	}
	if now == nil {
		now = time.Now
	}
	return &ApplePurchaseVerifier{
		client: client, signedData: signedData, bundleID: bundleID, products: productCopy,
		environment: environment, now: now,
	}, nil
}

func (v *ApplePurchaseVerifier) Verify(
	ctx context.Context,
	input ApplePurchaseVerificationInput,
) (VerifiedApplePurchase, error) {
	if input.AuthenticatedUserID <= 0 || input.Account.Validate() != nil || input.AuthenticatedUserID != input.Account.UserID {
		return VerifiedApplePurchase{}, applePurchaseFailure(model.EntitlementFailureAccountMismatch, nil)
	}
	if !validAppleTransactionID(input.TransactionID) {
		return VerifiedApplePurchase{}, applePurchaseFailure(model.EntitlementFailureInvalidEvidence, nil)
	}

	signedTransaction, environment, err := v.fetchSignedTransaction(ctx, input.TransactionID)
	if err != nil {
		return VerifiedApplePurchase{}, err
	}
	payload, err := v.signedData.VerifyAndDecodeTransaction(signedTransaction)
	if err != nil {
		return VerifiedApplePurchase{}, applePurchaseFailure(model.EntitlementFailureInvalidEvidence, err)
	}

	return v.validatePayload(input, payload, environment)
}

func (v *ApplePurchaseVerifier) fetchSignedTransaction(
	ctx context.Context,
	transactionID string,
) (string, model.StoreEnvironment, error) {
	signedTransaction, err := v.client.GetTransactionInfo(ctx, transactionID, v.environment)
	if err != nil {
		return "", "", appleProviderFailure(err)
	}
	return signedTransaction, v.environment, nil
}

func (v *ApplePurchaseVerifier) validatePayload(
	input ApplePurchaseVerificationInput,
	payload AppleTransactionPayload,
	environment model.StoreEnvironment,
) (VerifiedApplePurchase, error) {
	payloadEnvironment, ok := appleEnvironment(payload.Environment)
	if !ok || payloadEnvironment != environment || payload.BundleID != v.bundleID ||
		payload.TransactionID != input.TransactionID || payload.OriginalTransactionID == "" ||
		payload.ProductType != appleAutoRenewableSubscription {
		return VerifiedApplePurchase{}, applePurchaseFailure(model.EntitlementFailureInvalidEvidence, nil)
	}
	entitlement, knownProduct := v.products[payload.ProductID]
	if !knownProduct {
		return VerifiedApplePurchase{}, applePurchaseFailure(model.EntitlementFailureUnknownProduct, nil)
	}
	appAccountToken, err := uuid.Parse(payload.AppAccountToken)
	if err != nil || appAccountToken != input.Account.AppAccountToken {
		return VerifiedApplePurchase{}, applePurchaseFailure(model.EntitlementFailureAccountMismatch, err)
	}
	if payload.PurchaseDateMS <= 0 || payload.ExpiresDateMS <= payload.PurchaseDateMS || payload.SignedDateMS <= 0 {
		return VerifiedApplePurchase{}, applePurchaseFailure(model.EntitlementFailureInvalidEvidence, nil)
	}

	purchaseDate := time.UnixMilli(payload.PurchaseDateMS).UTC()
	expiresDate := time.UnixMilli(payload.ExpiresDateMS).UTC()
	signedAt := time.UnixMilli(payload.SignedDateMS).UTC()
	now := v.now().UTC()
	var revokedAt *time.Time
	var revocationReason *string
	if payload.RevocationDateMS != nil {
		value := time.UnixMilli(*payload.RevocationDateMS).UTC()
		revokedAt = &value
		revocationReason = appleRevocationReason(payload.RevocationReason)
	}

	identity := model.ApplePurchaseIdentity{
		UserID: input.AuthenticatedUserID, AppAccountToken: appAccountToken,
		OriginalTransactionID: payload.OriginalTransactionID, Environment: environment,
		LastVerifiedAt: now,
	}
	if !identity.MatchesAccount(input.Account) {
		return VerifiedApplePurchase{}, applePurchaseFailure(model.EntitlementFailureAccountMismatch, nil)
	}

	return VerifiedApplePurchase{
		Identity: identity, TransactionID: payload.TransactionID, ProductID: payload.ProductID,
		Entitlement: entitlement, PurchaseDate: purchaseDate, ExpiresDate: expiresDate,
		RevokedAt: revokedAt, RevocationReason: revocationReason, ProviderSignedAt: signedAt,
	}, nil
}

func appleRevocationReason(reason *int32) *string {
	if reason == nil {
		return nil
	}
	value := "apple_other"
	if *reason == 0 || *reason == 1 {
		value = fmt.Sprintf("apple_%d", *reason)
	}
	return &value
}

func appleEnvironment(value string) (model.StoreEnvironment, bool) {
	switch value {
	case "Production":
		return model.StoreEnvironmentProduction, true
	case "Sandbox":
		return model.StoreEnvironmentSandbox, true
	default:
		return "", false
	}
}

func appleProviderFailure(err error) error {
	var apiError *AppleAPIError
	if errors.As(err, &apiError) && !apiError.Retryable {
		return applePurchaseFailure(model.EntitlementFailureInvalidEvidence, err)
	}
	return applePurchaseFailure(model.EntitlementFailureProviderUnavailable, err)
}

func applePurchaseFailure(failure model.EntitlementOperationFailure, cause error) error {
	return &ApplePurchaseVerificationError{Failure: failure, Cause: cause}
}
