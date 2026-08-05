package model

import (
	"fmt"
	"time"
)

type EntitlementOperationOutcome string

const (
	EntitlementOutcomeApplied   EntitlementOperationOutcome = "applied"
	EntitlementOutcomeUnchanged EntitlementOperationOutcome = "unchanged"
	EntitlementOutcomePending   EntitlementOperationOutcome = "pending"
	EntitlementOutcomeDuplicate EntitlementOperationOutcome = "duplicate"
	EntitlementOutcomeRetry     EntitlementOperationOutcome = "retry"
	EntitlementOutcomeRejected  EntitlementOperationOutcome = "rejected"
)

type EntitlementOperationFailure string

const (
	EntitlementFailureNone                EntitlementOperationFailure = ""
	EntitlementFailureProviderUnavailable EntitlementOperationFailure = "provider_unavailable"
	EntitlementFailureInvalidEvidence     EntitlementOperationFailure = "invalid_evidence"
	EntitlementFailureAccountMismatch     EntitlementOperationFailure = "account_mismatch"
	EntitlementFailureUnknownProduct      EntitlementOperationFailure = "unknown_product"
)

type EntitlementResultCode string

const (
	EntitlementResultApplied           EntitlementResultCode = "entitlement_applied"
	EntitlementResultPurchasePending   EntitlementResultCode = "purchase_pending"
	EntitlementResultRevoked           EntitlementResultCode = "entitlement_revoked"
	EntitlementResultDuplicate         EntitlementResultCode = "duplicate_event"
	EntitlementResultRetryableProvider EntitlementResultCode = "retryable_provider_error"
	EntitlementResultInvalidPurchase   EntitlementResultCode = "invalid_purchase"
	EntitlementResultAccountMismatch   EntitlementResultCode = "account_mismatch"
	EntitlementResultUnknownProduct    EntitlementResultCode = "unknown_product"
	EntitlementResultStaleEvent        EntitlementResultCode = "stale_provider_event"
	EntitlementResultProviderTest      EntitlementResultCode = "provider_test_event"
)

// EntitlementOperationInput is evaluated only after transport authentication.
// Failure describes provider verification/catalog/account checks; provider
// status and occurrence time are accepted only when verification succeeded.
type EntitlementOperationInput struct {
	Duplicate                 bool
	Failure                   EntitlementOperationFailure
	ProviderStatus            EntitlementStatus
	ProviderOccurredAt        time.Time
	CurrentProviderOccurredAt time.Time
}

type EntitlementOperationResult struct {
	Outcome EntitlementOperationOutcome
	Status  EntitlementStatus
	Code    EntitlementResultCode
}

func DecideEntitlementOperation(input EntitlementOperationInput) (EntitlementOperationResult, error) {
	if input.Duplicate {
		if input.Failure != EntitlementFailureNone || input.ProviderStatus != "" ||
			!input.ProviderOccurredAt.IsZero() || !input.CurrentProviderOccurredAt.IsZero() {
			return EntitlementOperationResult{}, fmt.Errorf("duplicate operation input must not include verification state")
		}
		return EntitlementOperationResult{
			Outcome: EntitlementOutcomeDuplicate,
			Code:    EntitlementResultDuplicate,
		}, nil
	}

	if input.Failure != EntitlementFailureNone {
		if input.ProviderStatus != "" || !input.ProviderOccurredAt.IsZero() || !input.CurrentProviderOccurredAt.IsZero() {
			return EntitlementOperationResult{}, fmt.Errorf("failed operation input must not include verified provider state")
		}
		return resultForOperationFailure(input.Failure)
	}

	if !input.ProviderStatus.Valid() {
		return EntitlementOperationResult{Outcome: EntitlementOutcomeRejected, Code: EntitlementResultInvalidPurchase}, nil
	}
	if input.ProviderOccurredAt.IsZero() {
		return EntitlementOperationResult{}, fmt.Errorf("provider occurrence time is required")
	}
	if !input.CurrentProviderOccurredAt.IsZero() && !input.ProviderOccurredAt.After(input.CurrentProviderOccurredAt) {
		return EntitlementOperationResult{
			Outcome: EntitlementOutcomeUnchanged,
			Code:    EntitlementResultStaleEvent,
		}, nil
	}
	if input.ProviderStatus == EntitlementStatusPending {
		return EntitlementOperationResult{
			Outcome: EntitlementOutcomePending,
			Status:  EntitlementStatusPending,
			Code:    EntitlementResultPurchasePending,
		}, nil
	}
	if input.ProviderStatus == EntitlementStatusRevoked {
		return EntitlementOperationResult{
			Outcome: EntitlementOutcomeApplied,
			Status:  EntitlementStatusRevoked,
			Code:    EntitlementResultRevoked,
		}, nil
	}

	return EntitlementOperationResult{
		Outcome: EntitlementOutcomeApplied,
		Status:  input.ProviderStatus,
		Code:    EntitlementResultApplied,
	}, nil
}

func resultForOperationFailure(failure EntitlementOperationFailure) (EntitlementOperationResult, error) {
	switch failure {
	case EntitlementFailureProviderUnavailable:
		return EntitlementOperationResult{Outcome: EntitlementOutcomeRetry, Code: EntitlementResultRetryableProvider}, nil
	case EntitlementFailureInvalidEvidence:
		return EntitlementOperationResult{Outcome: EntitlementOutcomeRejected, Code: EntitlementResultInvalidPurchase}, nil
	case EntitlementFailureAccountMismatch:
		return EntitlementOperationResult{Outcome: EntitlementOutcomeRejected, Code: EntitlementResultAccountMismatch}, nil
	case EntitlementFailureUnknownProduct:
		return EntitlementOperationResult{Outcome: EntitlementOutcomeRejected, Code: EntitlementResultUnknownProduct}, nil
	default:
		return EntitlementOperationResult{}, fmt.Errorf("unsupported entitlement operation failure %q", failure)
	}
}

// CanGrantAccess is an operation-level guard. The resulting StoreEntitlement
// must still pass its period-end access check before premium access is returned.
func (r EntitlementOperationResult) CanGrantAccess() bool {
	return r.Outcome == EntitlementOutcomeApplied &&
		(r.Status == EntitlementStatusActive || r.Status == EntitlementStatusGracePeriod)
}
