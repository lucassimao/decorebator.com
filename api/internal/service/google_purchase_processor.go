package service

import (
	"context"
	"errors"
	"fmt"

	"decorebator.com/internal/model"
)

type GooglePurchaseVerificationService interface {
	Verify(context.Context, GooglePurchaseVerificationInput) (VerifiedGooglePurchase, error)
}

type GoogleAcknowledgementOutcome string

const (
	GoogleAcknowledgementNotRequired GoogleAcknowledgementOutcome = "not_required"
	GoogleAcknowledgementConfirmed   GoogleAcknowledgementOutcome = "acknowledged"
	GoogleAcknowledgementRetry       GoogleAcknowledgementOutcome = "retry_required"
	GoogleAcknowledgementFailed      GoogleAcknowledgementOutcome = "permanent_failure"
)

type GooglePurchaseProcessingResult struct {
	Purchase             VerifiedGooglePurchase
	Acknowledgement      GoogleAcknowledgementOutcome
	AcknowledgementError error `json:"-"`
}

type GooglePurchaseProcessor struct {
	verifier     GooglePurchaseVerificationService
	acknowledger GooglePlaySubscriptionAcknowledger
}

func NewGooglePurchaseProcessor(
	verifier GooglePurchaseVerificationService,
	acknowledger GooglePlaySubscriptionAcknowledger,
) (*GooglePurchaseProcessor, error) {
	if verifier == nil || acknowledger == nil {
		return nil, fmt.Errorf("Google purchase verifier and acknowledger are required")
	}
	return &GooglePurchaseProcessor{verifier: verifier, acknowledger: acknowledger}, nil
}

// VerifyAndAcknowledge never trusts client purchase state. Acknowledgement
// failure does not erase a provider-verified purchase: the explicit outcome is
// persisted/returned so the later API and mobile retry queue can retry safely.
func (p *GooglePurchaseProcessor) VerifyAndAcknowledge(
	ctx context.Context,
	input GooglePurchaseVerificationInput,
) (GooglePurchaseProcessingResult, error) {
	verified, err := p.verifier.Verify(ctx, input)
	if err != nil {
		return GooglePurchaseProcessingResult{}, err
	}
	if verified.Acknowledged {
		return googleProcessingResult(verified, GoogleAcknowledgementConfirmed, nil), nil
	}
	if !googleStatusRequiresAcknowledgement(verified.Status) {
		return googleProcessingResult(verified, GoogleAcknowledgementNotRequired, nil), nil
	}
	err = p.acknowledger.AcknowledgeSubscription(ctx, verified.ProductID, input.PurchaseToken)
	if err == nil {
		verified.Acknowledged = true
		return googleProcessingResult(verified, GoogleAcknowledgementConfirmed, nil), nil
	}
	if errors.Is(err, ErrInvalidGoogleAcknowledgeRequest) {
		return googleProcessingResult(verified, GoogleAcknowledgementFailed, err), nil
	}
	var apiError *GooglePlayAPIError
	if !errors.As(err, &apiError) || apiError.Retryable {
		return googleProcessingResult(verified, GoogleAcknowledgementRetry, err), nil
	}
	refreshed, refreshErr := p.verifier.Verify(ctx, input)
	if refreshErr == nil && refreshed.Acknowledged {
		return googleProcessingResult(refreshed, GoogleAcknowledgementConfirmed, nil), nil
	}
	if refreshErr != nil {
		return googleProcessingResult(verified, GoogleAcknowledgementRetry, refreshErr), nil
	}
	return googleProcessingResult(refreshed, GoogleAcknowledgementFailed, err), nil
}

func googleProcessingResult(
	purchase VerifiedGooglePurchase,
	outcome GoogleAcknowledgementOutcome,
	err error,
) GooglePurchaseProcessingResult {
	return GooglePurchaseProcessingResult{
		Purchase: purchase, Acknowledgement: outcome, AcknowledgementError: err,
	}
}

func googleStatusRequiresAcknowledgement(status model.EntitlementStatus) bool {
	switch status {
	case model.EntitlementStatusActive,
		model.EntitlementStatusGracePeriod,
		model.EntitlementStatusPaused,
		model.EntitlementStatusOnHold:
		return true
	default:
		return false
	}
}
