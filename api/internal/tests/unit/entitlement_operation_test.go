package unit

import (
	"testing"
	"time"

	"decorebator.com/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecideEntitlementOperationHandlesRequiredOutcomes(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name  string
		input model.EntitlementOperationInput
		want  model.EntitlementOperationResult
	}{
		{
			name:  "verified grace state applies",
			input: model.EntitlementOperationInput{ProviderStatus: model.EntitlementStatusGracePeriod, ProviderOccurredAt: now},
			want:  model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeApplied, Status: model.EntitlementStatusGracePeriod, Code: model.EntitlementResultApplied},
		},
		{
			name: "verified active state applies",
			input: model.EntitlementOperationInput{
				ProviderStatus:     model.EntitlementStatusActive,
				ProviderOccurredAt: now,
			},
			want: model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeApplied, Status: model.EntitlementStatusActive, Code: model.EntitlementResultApplied},
		},
		{
			name: "pending never grants",
			input: model.EntitlementOperationInput{
				ProviderStatus:     model.EntitlementStatusPending,
				ProviderOccurredAt: now,
			},
			want: model.EntitlementOperationResult{Outcome: model.EntitlementOutcomePending, Status: model.EntitlementStatusPending, Code: model.EntitlementResultPurchasePending},
		},
		{
			name: "revocation applies immediately",
			input: model.EntitlementOperationInput{
				ProviderStatus:     model.EntitlementStatusRevoked,
				ProviderOccurredAt: now,
			},
			want: model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeApplied, Status: model.EntitlementStatusRevoked, Code: model.EntitlementResultRevoked},
		},
		{
			name:  "duplicate returns stored outcome",
			input: model.EntitlementOperationInput{Duplicate: true},
			want:  model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeDuplicate, Code: model.EntitlementResultDuplicate},
		},
		{
			name: "transient provider failure retries",
			input: model.EntitlementOperationInput{
				Failure: model.EntitlementFailureProviderUnavailable,
			},
			want: model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeRetry, Code: model.EntitlementResultRetryableProvider},
		},
		{
			name: "invalid evidence is permanently rejected",
			input: model.EntitlementOperationInput{
				Failure: model.EntitlementFailureInvalidEvidence,
			},
			want: model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeRejected, Code: model.EntitlementResultInvalidPurchase},
		},
		{
			name: "account mismatch is permanently rejected",
			input: model.EntitlementOperationInput{
				Failure: model.EntitlementFailureAccountMismatch,
			},
			want: model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeRejected, Code: model.EntitlementResultAccountMismatch},
		},
		{
			name: "unknown product is permanently rejected",
			input: model.EntitlementOperationInput{
				Failure: model.EntitlementFailureUnknownProduct,
			},
			want: model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeRejected, Code: model.EntitlementResultUnknownProduct},
		},
		{
			name: "older revocation is an unchanged no-op",
			input: model.EntitlementOperationInput{
				ProviderStatus:            model.EntitlementStatusRevoked,
				ProviderOccurredAt:        now.Add(-time.Minute),
				CurrentProviderOccurredAt: now,
			},
			want: model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeUnchanged, Code: model.EntitlementResultStaleEvent},
		},
		{
			name:  "equal provider time is unchanged",
			input: model.EntitlementOperationInput{ProviderStatus: model.EntitlementStatusActive, ProviderOccurredAt: now, CurrentProviderOccurredAt: now},
			want:  model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeUnchanged, Code: model.EntitlementResultStaleEvent},
		},
		{
			name:  "unknown verified status is permanently rejected",
			input: model.EntitlementOperationInput{ProviderStatus: "future_status", ProviderOccurredAt: now},
			want:  model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeRejected, Code: model.EntitlementResultInvalidPurchase},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := model.DecideEntitlementOperation(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDecideEntitlementOperationRejectsContradictoryOrUnverifiedInput(t *testing.T) {
	inputs := []model.EntitlementOperationInput{
		{Duplicate: true, Failure: model.EntitlementFailureProviderUnavailable},
		{ProviderStatus: model.EntitlementStatusActive},
		{Failure: "unknown_failure"},
		{Failure: model.EntitlementFailureInvalidEvidence, ProviderStatus: model.EntitlementStatusActive},
	}

	for _, input := range inputs {
		_, err := model.DecideEntitlementOperation(input)
		require.Error(t, err)
	}
}

func TestOnlyAppliedActiveOrGraceResultCanGrantAccess(t *testing.T) {
	for _, result := range []model.EntitlementOperationResult{
		{Outcome: model.EntitlementOutcomeApplied, Status: model.EntitlementStatusActive},
		{Outcome: model.EntitlementOutcomeApplied, Status: model.EntitlementStatusGracePeriod},
	} {
		assert.True(t, result.CanGrantAccess())
	}

	for _, result := range []model.EntitlementOperationResult{
		{Outcome: model.EntitlementOutcomePending, Status: model.EntitlementStatusPending},
		{Outcome: model.EntitlementOutcomeApplied, Status: model.EntitlementStatusRevoked},
		{Outcome: model.EntitlementOutcomeRetry},
		{Outcome: model.EntitlementOutcomeRejected},
		{Outcome: model.EntitlementOutcomeDuplicate, Status: model.EntitlementStatusActive},
	} {
		assert.False(t, result.CanGrantAccess())
	}
}
