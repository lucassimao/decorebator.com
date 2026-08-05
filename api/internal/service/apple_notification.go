package service

import (
	"fmt"
	"strings"
	"time"

	"decorebator.com/internal/model"
	"github.com/google/uuid"
)

const appleNotificationVersion2 = "2.0"

const appleNotificationMaxFutureSkew = 5 * time.Minute

type VerifiedAppleNotification struct {
	IdempotencyKey   model.ProviderEventKey
	NotificationUUID uuid.UUID
	NotificationType string
	Subtype          string
	Environment      model.StoreEnvironment
	ProviderSignedAt time.Time
	Transaction      *AppleTransactionPayload
	RenewalInfo      *AppleRenewalInfoPayload
	data             *AppleNotificationData
}

type AppleNotificationVerifier struct {
	signedData AppleNotificationSignedDataVerifier
	bundleID   string
	appAppleID int64
	products   map[string]struct{}
	now        func() time.Time
}

func NewAppleNotificationVerifier(
	signedData AppleNotificationSignedDataVerifier,
	bundleID string,
	appAppleID int64,
	productIDs []string,
	now func() time.Time,
) (*AppleNotificationVerifier, error) {
	if signedData == nil || strings.TrimSpace(bundleID) == "" || appAppleID <= 0 || len(productIDs) == 0 {
		return nil, fmt.Errorf("Apple notification verifier configuration is incomplete")
	}
	products := make(map[string]struct{}, len(productIDs))
	for _, productID := range productIDs {
		if strings.TrimSpace(productID) == "" {
			return nil, fmt.Errorf("Apple notification product ID is required")
		}
		products[productID] = struct{}{}
	}
	if now == nil {
		now = time.Now
	}
	return &AppleNotificationVerifier{
		signedData: signedData, bundleID: bundleID, appAppleID: appAppleID, products: products, now: now,
	}, nil
}

func (v *AppleNotificationVerifier) Verify(signedPayload string) (VerifiedAppleNotification, error) {
	result, err := v.VerifyEnvelope(signedPayload)
	if err != nil {
		return VerifiedAppleNotification{}, err
	}
	return v.VerifyNestedEvidence(result)
}

func (v *AppleNotificationVerifier) VerifyEnvelope(signedPayload string) (VerifiedAppleNotification, error) {
	payload, err := v.signedData.VerifyAndDecodeNotification(signedPayload)
	if err != nil {
		return VerifiedAppleNotification{}, ErrInvalidAppleSignedData
	}
	result, err := v.validateEnvelope(payload)
	if err != nil {
		return VerifiedAppleNotification{}, err
	}
	return result, nil
}

func (v *AppleNotificationVerifier) VerifyNestedEvidence(
	result VerifiedAppleNotification,
) (VerifiedAppleNotification, error) {
	if result.data == nil {
		return VerifiedAppleNotification{}, ErrInvalidAppleSignedData
	}
	if err := v.verifyNestedEvidence(result.data, &result); err != nil {
		return VerifiedAppleNotification{}, err
	}
	result.data = nil
	return result, nil
}

func (v *AppleNotificationVerifier) validateEnvelope(
	payload AppleNotificationPayload,
) (VerifiedAppleNotification, error) {
	if payload.Version != appleNotificationVersion2 || strings.TrimSpace(payload.NotificationType) == "" || payload.Data == nil {
		return VerifiedAppleNotification{}, ErrInvalidAppleSignedData
	}
	notificationUUID, err := uuid.Parse(payload.NotificationUUID)
	if err != nil || notificationUUID == uuid.Nil {
		return VerifiedAppleNotification{}, ErrInvalidAppleSignedData
	}
	environment, ok := appleEnvironment(payload.Data.Environment)
	if !ok || payload.Data.BundleID != v.bundleID ||
		(environment == model.StoreEnvironmentProduction && payload.Data.AppAppleID != v.appAppleID) ||
		(environment == model.StoreEnvironmentSandbox && payload.Data.AppAppleID != 0 && payload.Data.AppAppleID != v.appAppleID) {
		return VerifiedAppleNotification{}, ErrInvalidAppleSignedData
	}
	signedAt := time.UnixMilli(payload.SignedDateMS).UTC()
	if payload.SignedDateMS <= 0 || signedAt.After(v.now().UTC().Add(appleNotificationMaxFutureSkew)) {
		return VerifiedAppleNotification{}, ErrInvalidAppleSignedData
	}
	key, err := model.AppleNotificationIdempotencyKey(environment, notificationUUID)
	if err != nil {
		return VerifiedAppleNotification{}, ErrInvalidAppleSignedData
	}

	result := VerifiedAppleNotification{
		IdempotencyKey: key, NotificationUUID: notificationUUID,
		NotificationType: payload.NotificationType, Subtype: payload.Subtype,
		Environment: environment, ProviderSignedAt: signedAt, data: payload.Data,
	}
	return result, nil
}

func (v *AppleNotificationVerifier) verifyNestedEvidence(
	data *AppleNotificationData,
	result *VerifiedAppleNotification,
) error {
	if data.SignedTransactionInfo != "" {
		transaction, verifyErr := v.signedData.VerifyAndDecodeTransaction(data.SignedTransactionInfo)
		if verifyErr != nil || v.validateTransaction(transaction, result.Environment) != nil {
			return ErrInvalidAppleSignedData
		}
		result.Transaction = &transaction
	}
	if data.SignedRenewalInfo != "" {
		renewalInfo, verifyErr := v.signedData.VerifyAndDecodeRenewalInfo(data.SignedRenewalInfo)
		if verifyErr != nil || v.validateRenewalInfo(renewalInfo, result.Environment, result.Transaction) != nil {
			return ErrInvalidAppleSignedData
		}
		result.RenewalInfo = &renewalInfo
	}
	if result.Transaction == nil && result.NotificationType != "TEST" {
		return ErrInvalidAppleSignedData
	}
	return nil
}

func (v *AppleNotificationVerifier) validateTransaction(
	payload AppleTransactionPayload,
	environment model.StoreEnvironment,
) error {
	payloadEnvironment, ok := appleEnvironment(payload.Environment)
	_, knownProduct := v.products[payload.ProductID]
	if !ok || payloadEnvironment != environment || payload.BundleID != v.bundleID || !knownProduct ||
		payload.OriginalTransactionID == "" || payload.TransactionID == "" ||
		payload.ProductType != appleAutoRenewableSubscription || payload.SignedDateMS <= 0 {
		return ErrInvalidAppleSignedData
	}
	return nil
}

func (v *AppleNotificationVerifier) validateRenewalInfo(
	payload AppleRenewalInfoPayload,
	environment model.StoreEnvironment,
	transaction *AppleTransactionPayload,
) error {
	payloadEnvironment, ok := appleEnvironment(payload.Environment)
	_, knownProduct := v.products[payload.ProductID]
	if !ok || payloadEnvironment != environment || !knownProduct || payload.OriginalTransactionID == "" ||
		payload.TransactionID != "" || payload.SignedDateMS <= 0 || payload.AutoRenewStatus < 0 || payload.AutoRenewStatus > 1 ||
		(payload.AutoRenewStatus == 1 && payload.AutoRenewProductID == "") {
		return ErrInvalidAppleSignedData
	}
	if transaction != nil && (payload.OriginalTransactionID != transaction.OriginalTransactionID || payload.ProductID != transaction.ProductID) {
		return ErrInvalidAppleSignedData
	}
	return nil
}
