package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
)

type AppleNotificationDeliveryDisposition string

const (
	AppleNotificationAcknowledge AppleNotificationDeliveryDisposition = "acknowledge"
	AppleNotificationRetry       AppleNotificationDeliveryDisposition = "retry"
)

type AppleNotificationDeliveryError struct {
	Disposition AppleNotificationDeliveryDisposition
	Cause       error
}

func (e *AppleNotificationDeliveryError) Error() string {
	return "Apple notification delivery failed: " + string(e.Disposition)
}

func (e *AppleNotificationDeliveryError) Unwrap() error { return e.Cause }

func AppleNotificationDispositionOf(err error) AppleNotificationDeliveryDisposition {
	if err == nil {
		return AppleNotificationAcknowledge
	}
	var deliveryError *AppleNotificationDeliveryError
	if errors.As(err, &deliveryError) {
		return deliveryError.Disposition
	}
	return AppleNotificationRetry
}

const (
	appleNotificationMaxSignedPayloadBytes = 256 * 1024
	appleNotificationLeaseDuration         = 30 * time.Second
	appleNotificationRetryDelay            = time.Minute
)

type AppleNotificationInbox interface {
	Claim(context.Context, repository.ProviderEventInboxInput, time.Time, time.Duration) (repository.ProviderEventInboxResult, error)
	Complete(context.Context, repository.ProviderEventClaim, repository.ProviderEventProcessor) (repository.ProviderEventInboxResult, error)
	MarkRetryable(context.Context, repository.ProviderEventClaim, time.Time) error
}

// AppleNotificationRefresher performs provider I/O and returns a database-only
// mutation. The returned processor must not make network calls: Complete runs
// it while holding the short inbox transaction and event lease.
type AppleNotificationRefresher func(
	context.Context,
	VerifiedAppleNotification,
) (repository.ProviderEventProcessor, error)

type AppleNotificationIngestor struct {
	verifier  *AppleNotificationVerifier
	inbox     AppleNotificationInbox
	refresher AppleNotificationRefresher
	now       func() time.Time
}

func NewAppleNotificationIngestor(
	verifier *AppleNotificationVerifier,
	inbox AppleNotificationInbox,
	refresher AppleNotificationRefresher,
	now func() time.Time,
) (*AppleNotificationIngestor, error) {
	if verifier == nil || inbox == nil || refresher == nil {
		return nil, fmt.Errorf("Apple notification ingestion dependencies are required")
	}
	if now == nil {
		now = time.Now
	}
	return &AppleNotificationIngestor{verifier: verifier, inbox: inbox, refresher: refresher, now: now}, nil
}

func (i *AppleNotificationIngestor) Ingest(
	ctx context.Context,
	signedPayload string,
) (repository.ProviderEventInboxResult, error) {
	if len(signedPayload) == 0 || len(signedPayload) > appleNotificationMaxSignedPayloadBytes {
		return repository.ProviderEventInboxResult{}, appleNotificationDeliveryError(AppleNotificationAcknowledge, ErrInvalidAppleSignedData)
	}
	envelope, err := i.verifier.VerifyEnvelope(signedPayload)
	if err != nil {
		return repository.ProviderEventInboxResult{}, appleNotificationDeliveryError(AppleNotificationAcknowledge, err)
	}
	now := i.now().UTC()
	environment := envelope.Environment
	claimResult, err := i.inbox.Claim(ctx, repository.ProviderEventInboxInput{
		Store:              model.EntitlementStoreApple,
		Environment:        &environment,
		IdempotencyKey:     envelope.IdempotencyKey,
		ProviderEventType:  envelope.NotificationType,
		ProviderOccurredAt: envelope.ProviderSignedAt,
	}, now, appleNotificationLeaseDuration)
	if err != nil {
		return claimResult, appleNotificationDeliveryError(AppleNotificationRetry, err)
	}
	if claimResult.Duplicate {
		if claimResult.Outcome == model.EntitlementOutcomeRetry {
			return claimResult, appleNotificationDeliveryError(AppleNotificationRetry, errors.New("Apple notification is already processing or waiting for retry"))
		}
		return claimResult, nil
	}
	if claimResult.Claim == nil {
		return repository.ProviderEventInboxResult{}, appleNotificationDeliveryError(AppleNotificationRetry, errors.New("Apple notification claim is missing"))
	}
	claim := *claimResult.Claim

	verified, err := i.verifier.VerifyNestedEvidence(envelope)
	if err != nil {
		result, completeErr := i.inbox.Complete(ctx, claim, rejectedAppleNotification)
		if completeErr != nil {
			return result, appleNotificationDeliveryError(AppleNotificationRetry, completeErr)
		}
		return result, nil
	}
	processor, err := i.refresher(ctx, verified)
	if err != nil {
		if failureProcessor, permanent := permanentAppleNotificationRefreshFailure(err); permanent {
			result, completeErr := i.inbox.Complete(ctx, claim, failureProcessor)
			if completeErr != nil {
				return result, appleNotificationDeliveryError(AppleNotificationRetry, completeErr)
			}
			return result, nil
		}
		if retryErr := i.inbox.MarkRetryable(ctx, claim, now.Add(appleNotificationRetryDelay)); retryErr != nil {
			return repository.ProviderEventInboxResult{}, appleNotificationDeliveryError(
				AppleNotificationRetry,
				fmt.Errorf("provider refresh failed and retry state could not be recorded: %w", retryErr),
			)
		}
		return repository.ProviderEventInboxResult{
			Outcome: model.EntitlementOutcomeRetry, Code: model.EntitlementResultRetryableProvider,
		}, appleNotificationDeliveryError(AppleNotificationRetry, err)
	}
	result, err := i.inbox.Complete(ctx, claim, processor)
	if err != nil {
		return result, appleNotificationDeliveryError(AppleNotificationRetry, err)
	}
	return result, nil
}

func permanentAppleNotificationRefreshFailure(err error) (repository.ProviderEventProcessor, bool) {
	var verificationError *ApplePurchaseVerificationError
	if !errors.As(err, &verificationError) {
		return nil, false
	}
	switch verificationError.Failure {
	case model.EntitlementFailureInvalidEvidence:
		return rejectedAppleNotification, true
	case model.EntitlementFailureAccountMismatch:
		return fixedAppleNotificationResult(model.EntitlementResultAccountMismatch), true
	case model.EntitlementFailureUnknownProduct:
		return fixedAppleNotificationResult(model.EntitlementResultUnknownProduct), true
	case model.EntitlementFailureProviderUnavailable:
		return nil, false
	default:
		return nil, false
	}
}

func fixedAppleNotificationResult(code model.EntitlementResultCode) repository.ProviderEventProcessor {
	return func(context.Context, pgx.Tx) (model.EntitlementOperationResult, error) {
		return model.EntitlementOperationResult{
			Outcome: model.EntitlementOutcomeRejected,
			Code:    code,
		}, nil
	}
}

func appleNotificationDeliveryError(disposition AppleNotificationDeliveryDisposition, cause error) error {
	return &AppleNotificationDeliveryError{Disposition: disposition, Cause: cause}
}

func rejectedAppleNotification(
	context.Context,
	pgx.Tx,
) (model.EntitlementOperationResult, error) {
	return model.EntitlementOperationResult{
		Outcome: model.EntitlementOutcomeRejected,
		Code:    model.EntitlementResultInvalidPurchase,
	}, nil
}
