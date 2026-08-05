package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"decorebator.com/internal/model"
)

const (
	googleSubscriptionPending         = "SUBSCRIPTION_STATE_PENDING"
	googleSubscriptionActive          = "SUBSCRIPTION_STATE_ACTIVE"
	googleSubscriptionPaused          = "SUBSCRIPTION_STATE_PAUSED"
	googleSubscriptionGracePeriod     = "SUBSCRIPTION_STATE_IN_GRACE_PERIOD"
	googleSubscriptionOnHold          = "SUBSCRIPTION_STATE_ON_HOLD"
	googleSubscriptionCanceled        = "SUBSCRIPTION_STATE_CANCELED"
	googleSubscriptionExpired         = "SUBSCRIPTION_STATE_EXPIRED"
	googleSubscriptionPendingCanceled = "SUBSCRIPTION_STATE_PENDING_PURCHASE_CANCELED"
	googleAcknowledgementPending      = "ACKNOWLEDGEMENT_STATE_PENDING"
	googleAcknowledged                = "ACKNOWLEDGEMENT_STATE_ACKNOWLEDGED"
	googleExpiryClockSkew             = 5 * time.Minute
)

type GooglePurchaseVerificationInput struct {
	AuthenticatedUserID int64
	PurchaseToken       string
	Account             model.GoogleAccountIdentity
}

type VerifiedGooglePurchase struct {
	Identity            model.GooglePurchaseIdentity
	ProductID           string
	Entitlement         string
	Status              model.EntitlementStatus
	PeriodStart         *time.Time
	PeriodEnd           *time.Time
	AutoRenewEnabled    bool
	CanceledAt          *time.Time
	CancellationReason  *string
	Superseded          bool
	Acknowledged        bool
	LinkedPurchaseToken string `json:"-"`
	ProviderETag        string
}

type GooglePurchaseVerificationError struct {
	Failure model.EntitlementOperationFailure
	Cause   error
}

func (e *GooglePurchaseVerificationError) Error() string {
	return fmt.Sprintf("Google purchase verification failed: %s", e.Failure)
}

func (e *GooglePurchaseVerificationError) Unwrap() error { return e.Cause }

type GooglePurchaseVerifier struct {
	client   GooglePlaySubscriptionClient
	products map[string]string
	now      func() time.Time
}

func NewGooglePurchaseVerifier(
	client GooglePlaySubscriptionClient,
	products map[string]string,
	now func() time.Time,
) (*GooglePurchaseVerifier, error) {
	if client == nil || len(products) == 0 {
		return nil, fmt.Errorf("Google Play client and product catalog are required")
	}
	productCopy := make(map[string]string, len(products))
	for productID, entitlement := range products {
		if strings.TrimSpace(productID) == "" || entitlement != "premium" {
			return nil, fmt.Errorf("invalid Google product catalog entry")
		}
		productCopy[productID] = entitlement
	}
	if now == nil {
		now = time.Now
	}
	return &GooglePurchaseVerifier{client: client, products: productCopy, now: now}, nil
}

func (v *GooglePurchaseVerifier) Verify(
	ctx context.Context,
	input GooglePurchaseVerificationInput,
) (VerifiedGooglePurchase, error) {
	if input.AuthenticatedUserID <= 0 || input.Account.Validate() != nil || input.AuthenticatedUserID != input.Account.UserID {
		return VerifiedGooglePurchase{}, googlePurchaseFailure(model.EntitlementFailureAccountMismatch, nil)
	}
	if !validGooglePurchaseToken(input.PurchaseToken) {
		return VerifiedGooglePurchase{}, googlePurchaseFailure(model.EntitlementFailureInvalidEvidence, nil)
	}
	subscription, err := v.client.GetSubscription(ctx, input.PurchaseToken)
	if err != nil {
		return VerifiedGooglePurchase{}, googleProviderFailure(err)
	}
	return v.validateSubscription(input, subscription)
}

func (v *GooglePurchaseVerifier) validateSubscription(
	input GooglePurchaseVerificationInput,
	subscription GooglePlaySubscription,
) (VerifiedGooglePurchase, error) {
	lineItem, entitlement, err := v.validateSubscriptionIdentityAndProduct(input, subscription)
	if err != nil {
		return VerifiedGooglePurchase{}, err
	}
	lifecycle, err := v.validateSubscriptionLifecycle(subscription, lineItem)
	if err != nil {
		return VerifiedGooglePurchase{}, googlePurchaseFailure(model.EntitlementFailureInvalidEvidence, err)
	}
	environment := model.StoreEnvironmentProduction
	if subscription.TestPurchase != nil {
		environment = model.StoreEnvironmentSandbox
	}
	verifiedAt := v.now().UTC()
	identity := model.GooglePurchaseIdentity{
		UserID: input.AuthenticatedUserID, ObfuscatedExternalAccountID: input.Account.ObfuscatedExternalAccountID,
		PurchaseToken: input.PurchaseToken, Environment: environment, LastVerifiedAt: verifiedAt,
	}
	if !identity.MatchesAccount(input.Account) {
		return VerifiedGooglePurchase{}, googlePurchaseFailure(model.EntitlementFailureAccountMismatch, nil)
	}
	return VerifiedGooglePurchase{
		Identity: identity, ProductID: lineItem.ProductID, Entitlement: entitlement, Status: lifecycle.status,
		PeriodStart: lifecycle.periodStart, PeriodEnd: lifecycle.periodEnd, AutoRenewEnabled: lifecycle.autoRenewEnabled,
		CanceledAt: lifecycle.canceledAt, CancellationReason: lifecycle.cancellationReason,
		Superseded: lifecycle.superseded, Acknowledged: lifecycle.acknowledged,
		LinkedPurchaseToken: subscription.LinkedPurchaseToken, ProviderETag: subscription.ETag,
	}, nil
}

func (v *GooglePurchaseVerifier) validateSubscriptionIdentityAndProduct(
	input GooglePurchaseVerificationInput,
	subscription GooglePlaySubscription,
) (GooglePlaySubscriptionLineItem, string, error) {
	if subscription.Kind != googlePlaySubscriptionKind || strings.TrimSpace(subscription.ETag) == "" {
		return GooglePlaySubscriptionLineItem{}, "", googlePurchaseFailure(model.EntitlementFailureInvalidEvidence, nil)
	}
	if subscription.ExternalAccountIdentifiers == nil ||
		subscription.ExternalAccountIdentifiers.ObfuscatedExternalAccountID != input.Account.ObfuscatedExternalAccountID {
		return GooglePlaySubscriptionLineItem{}, "", googlePurchaseFailure(model.EntitlementFailureAccountMismatch, nil)
	}
	if len(subscription.LineItems) != 1 {
		return GooglePlaySubscriptionLineItem{}, "", googlePurchaseFailure(model.EntitlementFailureInvalidEvidence, nil)
	}
	lineItem := subscription.LineItems[0]
	if (lineItem.AutoRenewingPlan == nil) == (lineItem.PrepaidPlan == nil) {
		return GooglePlaySubscriptionLineItem{}, "", googlePurchaseFailure(model.EntitlementFailureInvalidEvidence, nil)
	}
	entitlement, knownProduct := v.products[lineItem.ProductID]
	if !knownProduct {
		return GooglePlaySubscriptionLineItem{}, "", googlePurchaseFailure(model.EntitlementFailureUnknownProduct, nil)
	}
	return lineItem, entitlement, nil
}

type googleVerifiedLifecycle struct {
	status             model.EntitlementStatus
	periodStart        *time.Time
	periodEnd          *time.Time
	autoRenewEnabled   bool
	canceledAt         *time.Time
	cancellationReason *string
	superseded         bool
	acknowledged       bool
}

func (v *GooglePurchaseVerifier) validateSubscriptionLifecycle(
	subscription GooglePlaySubscription,
	lineItem GooglePlaySubscriptionLineItem,
) (googleVerifiedLifecycle, error) {
	status, err := googleCanonicalStatus(subscription.SubscriptionState, lineItem.ExpiryTime)
	if err != nil {
		return googleVerifiedLifecycle{}, err
	}
	periodStart, periodEnd, err := validateGoogleSubscriptionPeriod(subscription, lineItem, status)
	if err != nil {
		return googleVerifiedLifecycle{}, err
	}
	status = googleExpiryNormalizedStatus(status, periodEnd, v.now().UTC())
	canceledAt, cancellationReason, superseded, err := googleCancellation(subscription.CanceledStateContext)
	if err != nil {
		return googleVerifiedLifecycle{}, err
	}
	autoRenewEnabled := lineItem.AutoRenewingPlan != nil && lineItem.AutoRenewingPlan.AutoRenewEnabled
	if subscription.SubscriptionState == googleSubscriptionCanceled && autoRenewEnabled {
		return googleVerifiedLifecycle{}, fmt.Errorf("canceled Google subscription still auto-renews")
	}
	if superseded {
		status = model.EntitlementStatusExpired
	}
	acknowledged, err := googleAcknowledgement(subscription.AcknowledgementState)
	if err != nil {
		return googleVerifiedLifecycle{}, err
	}
	return googleVerifiedLifecycle{
		status: status, periodStart: periodStart, periodEnd: periodEnd,
		autoRenewEnabled: autoRenewEnabled, canceledAt: canceledAt,
		cancellationReason: cancellationReason, superseded: superseded, acknowledged: acknowledged,
	}, nil
}

func validateGoogleSubscriptionPeriod(
	subscription GooglePlaySubscription,
	lineItem GooglePlaySubscriptionLineItem,
	status model.EntitlementStatus,
) (*time.Time, *time.Time, error) {
	periodStart, err := optionalGoogleTime(subscription.StartTime)
	if err != nil {
		return nil, nil, err
	}
	periodEnd, err := optionalGoogleTime(lineItem.ExpiryTime)
	if err != nil {
		return nil, nil, err
	}
	if (status == model.EntitlementStatusActive || status == model.EntitlementStatusGracePeriod) && periodEnd == nil {
		return nil, nil, fmt.Errorf("Google subscription expiry is required")
	}
	if periodStart != nil && periodEnd != nil && periodEnd.Before(*periodStart) {
		return nil, nil, fmt.Errorf("Google subscription period ends before it starts")
	}
	if subscription.SubscriptionState != googleSubscriptionPending &&
		subscription.SubscriptionState != googleSubscriptionPendingCanceled && periodStart == nil {
		return nil, nil, fmt.Errorf("non-pending Google subscription lacks a start time")
	}
	return periodStart, periodEnd, nil
}

func googleAcknowledgement(state string) (bool, error) {
	switch state {
	case googleAcknowledged:
		return true, nil
	case googleAcknowledgementPending:
		return false, nil
	default:
		return false, fmt.Errorf("unknown Google acknowledgement state")
	}
}

func googleCanonicalStatus(state, expiry string) (model.EntitlementStatus, error) {
	switch state {
	case googleSubscriptionPending:
		return model.EntitlementStatusPending, nil
	case googleSubscriptionActive:
		return model.EntitlementStatusActive, nil
	case googleSubscriptionPaused:
		return model.EntitlementStatusPaused, nil
	case googleSubscriptionGracePeriod:
		return model.EntitlementStatusGracePeriod, nil
	case googleSubscriptionOnHold:
		return model.EntitlementStatusOnHold, nil
	case googleSubscriptionExpired, googleSubscriptionPendingCanceled:
		return model.EntitlementStatusExpired, nil
	case googleSubscriptionCanceled:
		periodEnd, err := optionalGoogleTime(expiry)
		if err != nil || periodEnd == nil {
			return "", fmt.Errorf("canceled Google subscription requires expiry")
		}
		return model.EntitlementStatusActive, nil
	default:
		return "", fmt.Errorf("unknown Google subscription state")
	}
}

func optionalGoogleTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil, err
	}
	parsed = parsed.UTC()
	return &parsed, nil
}

func googleExpiryNormalizedStatus(
	status model.EntitlementStatus,
	periodEnd *time.Time,
	now time.Time,
) model.EntitlementStatus {
	if periodEnd != nil && (status == model.EntitlementStatusActive || status == model.EntitlementStatusGracePeriod) &&
		!now.Add(-googleExpiryClockSkew).Before(*periodEnd) {
		return model.EntitlementStatusExpired
	}
	return status
}

func googleCancellation(cancellation *GoogleCanceledStateContext) (*time.Time, *string, bool, error) {
	if cancellation == nil {
		return nil, nil, false, nil
	}
	variants := 0
	var canceledAt *time.Time
	reason := ""
	if cancellation.UserInitiatedCancellation != nil {
		variants++
		reason = "user"
		parsed, err := optionalGoogleTime(cancellation.UserInitiatedCancellation.CancelTime)
		if err != nil {
			return nil, nil, false, err
		}
		canceledAt = parsed
	}
	if cancellation.SystemInitiatedCancellation != nil {
		variants++
		reason = "system"
	}
	if cancellation.DeveloperInitiatedCancellation != nil {
		variants++
		reason = "developer"
	}
	superseded := cancellation.ReplacementCancellation != nil
	if superseded {
		variants++
		reason = "replacement"
	}
	if variants != 1 {
		return nil, nil, false, fmt.Errorf("Google cancellation context must contain exactly one reason")
	}
	return canceledAt, &reason, superseded, nil
}

func googleProviderFailure(err error) error {
	var apiError *GooglePlayAPIError
	if errors.As(err, &apiError) && !apiError.Retryable &&
		apiError.StatusCode != 0 && apiError.StatusCode != http.StatusUnauthorized && apiError.StatusCode != http.StatusForbidden {
		return googlePurchaseFailure(model.EntitlementFailureInvalidEvidence, err)
	}
	return googlePurchaseFailure(model.EntitlementFailureProviderUnavailable, err)
}

func googlePurchaseFailure(failure model.EntitlementOperationFailure, cause error) error {
	return &GooglePurchaseVerificationError{Failure: failure, Cause: cause}
}
