package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type StorePurchaseBindingResolver interface {
	ResolvePurchaseBinding(context.Context, model.EntitlementStore, model.StoreEnvironment, string) (repository.StorePurchaseBinding, bool, error)
}

type GoogleVoidedPurchaseWriter interface {
	ApplyGoogleVoidedPurchase(context.Context, pgx.Tx, repository.StorePurchaseBinding, time.Time, int32) (model.EntitlementOperationResult, error)
}

type AppleProviderSubscriptionRefresher interface {
	Refresh(context.Context, AppleSubscriptionRefreshInput) (VerifiedAppleSubscription, error)
}

type GoogleProviderPurchaseProcessor interface {
	VerifyAndAcknowledge(context.Context, GooglePurchaseVerificationInput) (GooglePurchaseProcessingResult, error)
}

type StoreNotificationRefresher struct {
	bindings    StorePurchaseBindingResolver
	voided      GoogleVoidedPurchaseWriter
	apple       AppleProviderSubscriptionRefresher
	google      GoogleProviderPurchaseProcessor
	persistence *StoreEntitlementPersistence
}

func NewStoreNotificationRefresher(
	bindings StorePurchaseBindingResolver,
	voided GoogleVoidedPurchaseWriter,
	apple AppleProviderSubscriptionRefresher,
	google GoogleProviderPurchaseProcessor,
	persistence *StoreEntitlementPersistence,
) (*StoreNotificationRefresher, error) {
	if bindings == nil || voided == nil || apple == nil || google == nil || persistence == nil {
		return nil, fmt.Errorf("store notification refresh dependencies are required")
	}
	return &StoreNotificationRefresher{
		bindings: bindings, voided: voided, apple: apple, google: google, persistence: persistence,
	}, nil
}

// RefreshApple resolves the server-generated account binding before invoking
// the provider-status refresher. The resolved user is an internal ownership
// assertion, not an HTTP-authenticated user supplied by the notification.
func (r *StoreNotificationRefresher) RefreshApple(
	ctx context.Context,
	notification VerifiedAppleNotification,
) (repository.ProviderEventProcessor, error) {
	if notification.NotificationType == "TEST" {
		return acceptedStoreProviderTest, nil
	}
	if notification.Transaction == nil || notification.Transaction.OriginalTransactionID == "" ||
		!notification.Environment.Valid() {
		return nil, ErrInvalidAppleSignedData
	}
	originalTransactionID := notification.Transaction.OriginalTransactionID
	binding, found, err := r.bindings.ResolvePurchaseBinding(
		ctx, model.EntitlementStoreApple, notification.Environment, originalTransactionID,
	)
	if err != nil {
		return nil, fmt.Errorf("resolve Apple notification binding: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("Apple notification binding is not available yet")
	}
	if binding.Store != model.EntitlementStoreApple || binding.Environment != notification.Environment ||
		binding.UserID <= 0 || binding.ProviderRecordID != originalTransactionID {
		return nil, fmt.Errorf("resolved Apple notification binding is inconsistent")
	}
	accountToken, err := uuid.Parse(binding.AccountIdentifier)
	if err != nil || accountToken == uuid.Nil {
		return nil, fmt.Errorf("resolved Apple notification account is invalid")
	}
	verified, err := r.apple.Refresh(ctx, AppleSubscriptionRefreshInput{
		AuthenticatedUserID: binding.UserID,
		Account: model.AppleAccountIdentity{
			UserID: binding.UserID, AppAccountToken: accountToken,
		},
		OriginalTransactionID: originalTransactionID,
		Environment:           notification.Environment,
	})
	if err != nil {
		return nil, err
	}
	eventOccurredAt := notification.ProviderSignedAt.UTC()
	return func(ctx context.Context, tx pgx.Tx) (model.EntitlementOperationResult, error) {
		result, persistErr := r.persistence.PersistApple(ctx, tx, verified, &eventOccurredAt)
		return result.Operation, persistErr
	}, nil
}

// RefreshGoogle resolves both canonical environments because RTDN does not
// include a sandbox/production discriminator. A token present in both scopes
// is rejected as ambiguous rather than guessed.
func (r *StoreNotificationRefresher) RefreshGoogle(
	ctx context.Context,
	event GoogleRTDNEvent,
) (repository.ProviderEventProcessor, error) {
	if event.Kind == googleRTDNTestEventKind {
		return acceptedStoreProviderTest, nil
	}
	if event.PurchaseToken == "" {
		return nil, ErrInvalidGoogleRTDNEnvelope
	}
	binding, found, err := r.resolveGoogleBinding(ctx, event.PurchaseToken)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("Google notification binding is not available yet")
	}
	account := model.GoogleAccountIdentity{
		UserID: binding.UserID, ObfuscatedExternalAccountID: binding.AccountIdentifier,
	}
	if binding.Store != model.EntitlementStoreGoogle || binding.UserID <= 0 ||
		binding.ProviderRecordID != event.PurchaseToken || account.Validate() != nil {
		return nil, fmt.Errorf("resolved Google notification binding is inconsistent")
	}
	if event.Kind == "voided_subscription" {
		eventOccurredAt := event.OccurredAt.UTC()
		return func(ctx context.Context, tx pgx.Tx) (model.EntitlementOperationResult, error) {
			return r.voided.ApplyGoogleVoidedPurchase(
				ctx, tx, binding, eventOccurredAt, event.RefundType,
			)
		}, nil
	}
	processed, err := r.google.VerifyAndAcknowledge(ctx, GooglePurchaseVerificationInput{
		AuthenticatedUserID: binding.UserID, PurchaseToken: event.PurchaseToken, Account: account,
	})
	if err != nil {
		return nil, err
	}
	if processed.Purchase.Identity.Environment != binding.Environment {
		return nil, fmt.Errorf("Google notification provider environment does not match its binding")
	}
	if processed.Acknowledgement == GoogleAcknowledgementRetry {
		if processed.AcknowledgementError != nil {
			return nil, processed.AcknowledgementError
		}
		return nil, errors.New("Google purchase acknowledgement requires retry")
	}
	eventOccurredAt := event.OccurredAt.UTC()
	return func(ctx context.Context, tx pgx.Tx) (model.EntitlementOperationResult, error) {
		result, persistErr := r.persistence.PersistGoogle(ctx, tx, processed, &eventOccurredAt)
		return result.Operation, persistErr
	}, nil
}

func (r *StoreNotificationRefresher) resolveGoogleBinding(
	ctx context.Context,
	purchaseToken string,
) (repository.StorePurchaseBinding, bool, error) {
	var selected repository.StorePurchaseBinding
	foundAny := false
	for _, environment := range []model.StoreEnvironment{
		model.StoreEnvironmentProduction,
		model.StoreEnvironmentSandbox,
	} {
		binding, found, err := r.bindings.ResolvePurchaseBinding(
			ctx, model.EntitlementStoreGoogle, environment, purchaseToken,
		)
		if err != nil {
			return repository.StorePurchaseBinding{}, false, fmt.Errorf("resolve Google notification binding: %w", err)
		}
		if !found {
			continue
		}
		if foundAny {
			return repository.StorePurchaseBinding{}, false, fmt.Errorf("Google notification binding is ambiguous across environments")
		}
		selected, foundAny = binding, true
	}
	return selected, foundAny, nil
}

func acceptedStoreProviderTest(context.Context, pgx.Tx) (model.EntitlementOperationResult, error) {
	return model.EntitlementOperationResult{
		Outcome: model.EntitlementOutcomeUnchanged,
		Code:    model.EntitlementResultProviderTest,
	}, nil
}
