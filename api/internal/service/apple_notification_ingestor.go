package service

import (
	"context"
	"fmt"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
)

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
		return repository.ProviderEventInboxResult{}, ErrInvalidAppleSignedData
	}
	envelope, err := i.verifier.VerifyEnvelope(signedPayload)
	if err != nil {
		return repository.ProviderEventInboxResult{}, err
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
	if err != nil || claimResult.Duplicate {
		return claimResult, err
	}
	if claimResult.Claim == nil {
		return repository.ProviderEventInboxResult{}, fmt.Errorf("Apple notification claim is missing")
	}
	claim := *claimResult.Claim

	verified, err := i.verifier.VerifyNestedEvidence(envelope)
	if err != nil {
		return i.inbox.Complete(ctx, claim, rejectedAppleNotification)
	}
	processor, err := i.refresher(ctx, verified)
	if err != nil {
		if retryErr := i.inbox.MarkRetryable(ctx, claim, now.Add(appleNotificationRetryDelay)); retryErr != nil {
			return repository.ProviderEventInboxResult{}, fmt.Errorf("provider refresh failed and retry state could not be recorded: %w", retryErr)
		}
		return repository.ProviderEventInboxResult{
			Outcome: model.EntitlementOutcomeRetry, Code: model.EntitlementResultRetryableProvider,
		}, err
	}
	return i.inbox.Complete(ctx, claim, processor)
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
