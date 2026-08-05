package unit

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAppleNotificationInbox struct {
	claimCalls    int
	completeCalls int
	retryCalls    int
	input         repository.ProviderEventInboxInput
}

func (f *fakeAppleNotificationInbox) Claim(
	_ context.Context,
	input repository.ProviderEventInboxInput,
	_ time.Time,
	_ time.Duration,
) (repository.ProviderEventInboxResult, error) {
	f.claimCalls++
	f.input = input
	return repository.ProviderEventInboxResult{Claim: &repository.ProviderEventClaim{
		ID: 1, LeaseToken: uuid.New(), Input: input,
	}}, nil
}

func (f *fakeAppleNotificationInbox) Complete(
	ctx context.Context,
	_ repository.ProviderEventClaim,
	processor repository.ProviderEventProcessor,
) (repository.ProviderEventInboxResult, error) {
	f.completeCalls++
	result, err := processor(ctx, nil)
	return repository.ProviderEventInboxResult{Outcome: result.Outcome, Code: result.Code}, err
}

func (f *fakeAppleNotificationInbox) MarkRetryable(
	context.Context,
	repository.ProviderEventClaim,
	time.Time,
) error {
	f.retryCalls++
	return nil
}

type fakeAppleNotificationSignedData struct {
	notification service.AppleNotificationPayload
	transaction  service.AppleTransactionPayload
	renewal      service.AppleRenewalInfoPayload
	notifyErr    error
	transactErr  error
	renewalErr   error
}

func (f fakeAppleNotificationSignedData) VerifyAndDecodeNotification(string) (service.AppleNotificationPayload, error) {
	return f.notification, f.notifyErr
}

func (f fakeAppleNotificationSignedData) VerifyAndDecodeTransaction(string) (service.AppleTransactionPayload, error) {
	return f.transaction, f.transactErr
}

func (f fakeAppleNotificationSignedData) VerifyAndDecodeRenewalInfo(string) (service.AppleRenewalInfoPayload, error) {
	return f.renewal, f.renewalErr
}

func validAppleNotificationEvidence(now time.Time) fakeAppleNotificationSignedData {
	accountToken := uuid.MustParse("db22924d-8f4b-4a7e-845e-3304f319210e")
	transaction := validApplePayload(now, accountToken)
	return fakeAppleNotificationSignedData{
		notification: service.AppleNotificationPayload{
			NotificationType: "DID_RENEW",
			NotificationUUID: "c9f2b260-9a60-4e63-a3b8-707f664a5f86",
			Version:          "2.0",
			SignedDateMS:     now.UnixMilli(),
			Data: &service.AppleNotificationData{
				Environment: "Production", AppAppleID: 123456789,
				BundleID:              "com.lsimaocosta.decorebator",
				SignedTransactionInfo: "signed-transaction", SignedRenewalInfo: "signed-renewal",
				Status: 1,
			},
		},
		transaction: transaction,
		renewal: service.AppleRenewalInfoPayload{
			OriginalTransactionID: transaction.OriginalTransactionID,
			ProductID:             transaction.ProductID, AutoRenewProductID: transaction.ProductID,
			AutoRenewStatus: 1, Environment: "Production", SignedDateMS: now.UnixMilli(),
		},
	}
}

func newAppleNotificationVerifier(t *testing.T, evidence fakeAppleNotificationSignedData) *service.AppleNotificationVerifier {
	t.Helper()
	verifier, err := service.NewAppleNotificationVerifier(
		evidence,
		"com.lsimaocosta.decorebator",
		123456789,
		[]string{"decorebator_monthly_premium1", "decorebator_annual_premium"},
		func() time.Time { return time.UnixMilli(evidence.notification.SignedDateMS).UTC() },
	)
	require.NoError(t, err)
	return verifier
}

func TestAppleNotificationVerifierAcceptsBoundProductionEvidence(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	evidence := validAppleNotificationEvidence(now)
	verified, err := newAppleNotificationVerifier(t, evidence).Verify("signed-notification")
	require.NoError(t, err)
	assert.Equal(t, model.StoreEnvironmentProduction, verified.Environment)
	assert.Equal(t, "DID_RENEW", verified.NotificationType)
	assert.Equal(t, now, verified.ProviderSignedAt)
	require.NotNil(t, verified.Transaction)
	require.NotNil(t, verified.RenewalInfo)
	assert.Equal(t, evidence.transaction.OriginalTransactionID, verified.RenewalInfo.OriginalTransactionID)
	assert.Contains(t, verified.IdempotencyKey.String(), "apple_notification:production:")
}

func TestAppleNotificationVerifierSeparatesSandboxAndAllowsMissingSandboxAppID(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	evidence := validAppleNotificationEvidence(now)
	evidence.notification.Data.Environment = "Sandbox"
	evidence.notification.Data.AppAppleID = 0
	evidence.transaction.Environment = "Sandbox"
	evidence.renewal.Environment = "Sandbox"
	verified, err := newAppleNotificationVerifier(t, evidence).Verify("signed-notification")
	require.NoError(t, err)
	assert.Equal(t, model.StoreEnvironmentSandbox, verified.Environment)
	assert.Contains(t, verified.IdempotencyKey.String(), "apple_notification:sandbox:")
}

func TestAppleNotificationVerifierFailsClosedForOuterAndNestedMismatch(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		mutate func(*fakeAppleNotificationSignedData)
	}{
		{"invalid outer signature", func(e *fakeAppleNotificationSignedData) { e.notifyErr = service.ErrInvalidAppleSignedData }},
		{"wrong version", func(e *fakeAppleNotificationSignedData) { e.notification.Version = "1.0" }},
		{"missing uuid", func(e *fakeAppleNotificationSignedData) { e.notification.NotificationUUID = "" }},
		{"wrong production app id", func(e *fakeAppleNotificationSignedData) { e.notification.Data.AppAppleID++ }},
		{"missing production app id", func(e *fakeAppleNotificationSignedData) { e.notification.Data.AppAppleID = 0 }},
		{"wrong bundle", func(e *fakeAppleNotificationSignedData) { e.notification.Data.BundleID = "attacker.app" }},
		{"invalid transaction signature", func(e *fakeAppleNotificationSignedData) { e.transactErr = service.ErrInvalidAppleSignedData }},
		{"transaction environment mismatch", func(e *fakeAppleNotificationSignedData) { e.transaction.Environment = "Sandbox" }},
		{"unknown transaction product", func(e *fakeAppleNotificationSignedData) { e.transaction.ProductID = "attacker.product" }},
		{"renewal lineage mismatch", func(e *fakeAppleNotificationSignedData) { e.renewal.OriginalTransactionID = "other" }},
		{"renewal product mismatch", func(e *fakeAppleNotificationSignedData) { e.renewal.ProductID = "decorebator_annual_premium" }},
		{"unknown auto-renew product", func(e *fakeAppleNotificationSignedData) { e.renewal.AutoRenewProductID = "attacker.product" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			evidence := validAppleNotificationEvidence(now)
			test.mutate(&evidence)
			_, err := newAppleNotificationVerifier(t, evidence).Verify("signed-notification")
			require.ErrorIs(t, err, service.ErrInvalidAppleSignedData)
		})
	}
}

func TestAppleNotificationVerifierRejectsFutureEnvelopeAndRenewalWithoutTransaction(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	evidence := validAppleNotificationEvidence(now)
	verifier, err := service.NewAppleNotificationVerifier(
		evidence, "com.lsimaocosta.decorebator", 123456789,
		[]string{"decorebator_monthly_premium1"}, func() time.Time { return now.Add(-10 * time.Minute) },
	)
	require.NoError(t, err)
	_, err = verifier.Verify("signed-notification")
	require.ErrorIs(t, err, service.ErrInvalidAppleSignedData)

	evidence = validAppleNotificationEvidence(now)
	evidence.notification.Data.SignedTransactionInfo = ""
	_, err = newAppleNotificationVerifier(t, evidence).Verify("signed-notification")
	require.ErrorIs(t, err, service.ErrInvalidAppleSignedData)
}

func TestAppleNotificationVerifierAcceptsSignedTestWithoutTransaction(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	evidence := validAppleNotificationEvidence(now)
	evidence.notification.NotificationType = "TEST"
	evidence.notification.Data.SignedTransactionInfo = ""
	evidence.notification.Data.SignedRenewalInfo = ""
	verified, err := newAppleNotificationVerifier(t, evidence).Verify("signed-notification")
	require.NoError(t, err)
	assert.Nil(t, verified.Transaction)
	assert.Nil(t, verified.RenewalInfo)
}

func TestAppleJWSVerifierAuthenticatesNotificationPayloadBeforeDecoding(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	payload := validAppleNotificationEvidence(now).notification
	signed, root := signAppleTestPayloadJWS(t, payload, now, true)
	verifier, err := service.NewAppleJWSVerifier([]*x509.Certificate{root})
	require.NoError(t, err)
	decoded, err := verifier.VerifyAndDecodeNotification(signed)
	require.NoError(t, err)
	assert.Equal(t, payload.NotificationUUID, decoded.NotificationUUID)

	parts := strings.Split(signed, ".")
	require.Len(t, parts, 3)
	parts[1] = base64.RawURLEncoding.EncodeToString([]byte(`{"notificationUUID":"attacker","signedDate":1}`))
	_, err = verifier.VerifyAndDecodeNotification(strings.Join(parts, "."))
	require.ErrorIs(t, err, service.ErrInvalidAppleSignedData)
}

func TestAppleJWSVerifierRejectsSignedPayloadTypeSubstitution(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	transaction := validApplePayload(now, uuid.New())
	signedTransaction, root := signAppleTestPayloadJWS(t, transaction, now, true)
	verifier, err := service.NewAppleJWSVerifier([]*x509.Certificate{root})
	require.NoError(t, err)
	_, err = verifier.VerifyAndDecodeRenewalInfo(signedTransaction)
	require.ErrorIs(t, err, service.ErrInvalidAppleSignedData)
	_, err = verifier.VerifyAndDecodeNotification(signedTransaction)
	require.ErrorIs(t, err, service.ErrInvalidAppleSignedData)

	renewal := validAppleNotificationEvidence(now).renewal
	signedRenewal, renewalRoot := signAppleTestPayloadJWS(t, renewal, now, true)
	renewalVerifier, err := service.NewAppleJWSVerifier([]*x509.Certificate{renewalRoot})
	require.NoError(t, err)
	_, err = renewalVerifier.VerifyAndDecodeTransaction(signedRenewal)
	require.ErrorIs(t, err, service.ErrInvalidAppleSignedData)
}

func TestAppleNotificationIngestorClaimsOnlyVerifiedDerivedKey(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	evidence := validAppleNotificationEvidence(now)
	verifier := newAppleNotificationVerifier(t, evidence)
	inbox := &fakeAppleNotificationInbox{}
	processorCalls := 0
	ingestor, err := service.NewAppleNotificationIngestor(
		verifier,
		inbox,
		func(_ context.Context, verified service.VerifiedAppleNotification) (repository.ProviderEventProcessor, error) {
			processorCalls++
			assert.Equal(t, evidence.notification.NotificationUUID, verified.NotificationUUID.String())
			return func(_ context.Context, tx pgx.Tx) (model.EntitlementOperationResult, error) {
				assert.Nil(t, tx)
				return model.EntitlementOperationResult{
					Outcome: model.EntitlementOutcomeApplied,
					Status:  model.EntitlementStatusActive,
					Code:    model.EntitlementResultApplied,
				}, nil
			}, nil
		},
		func() time.Time { return now },
	)
	require.NoError(t, err)
	result, err := ingestor.Ingest(context.Background(), "signed-notification")
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementOutcomeApplied, result.Outcome)
	assert.Equal(t, 1, inbox.claimCalls)
	assert.Equal(t, 1, inbox.completeCalls)
	assert.Equal(t, 1, processorCalls)
	assert.Equal(t, model.EntitlementStoreApple, inbox.input.Store)
	assert.Equal(t, now, inbox.input.ProviderOccurredAt)

	evidence.notifyErr = service.ErrInvalidAppleSignedData
	invalidVerifier := newAppleNotificationVerifier(t, evidence)
	invalidIngestor, err := service.NewAppleNotificationIngestor(
		invalidVerifier, inbox,
		func(context.Context, service.VerifiedAppleNotification) (repository.ProviderEventProcessor, error) {
			t.Fatal("processor must not run for invalid evidence")
			return nil, nil
		},
		func() time.Time { return now },
	)
	require.NoError(t, err)
	_, err = invalidIngestor.Ingest(context.Background(), "signed-notification")
	require.ErrorIs(t, err, service.ErrInvalidAppleSignedData)
	assert.Equal(t, 1, inbox.claimCalls)
}

func TestAppleNotificationIngestorRecordsVerifiedEnvelopeNestedRejection(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	evidence := validAppleNotificationEvidence(now)
	evidence.transactErr = service.ErrInvalidAppleSignedData
	inbox := &fakeAppleNotificationInbox{}
	refresherCalls := 0
	ingestor, err := service.NewAppleNotificationIngestor(
		newAppleNotificationVerifier(t, evidence), inbox,
		func(context.Context, service.VerifiedAppleNotification) (repository.ProviderEventProcessor, error) {
			refresherCalls++
			return nil, nil
		},
		func() time.Time { return now },
	)
	require.NoError(t, err)
	result, err := ingestor.Ingest(context.Background(), "signed-notification")
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementOutcomeRejected, result.Outcome)
	assert.Equal(t, model.EntitlementResultInvalidPurchase, result.Code)
	assert.Equal(t, 1, inbox.claimCalls)
	assert.Equal(t, 1, inbox.completeCalls)
	assert.Zero(t, refresherCalls)
}

func TestAppleNotificationIngestorPersistsRetryAfterProviderFailure(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	evidence := validAppleNotificationEvidence(now)
	inbox := &fakeAppleNotificationInbox{}
	providerErr := errors.New("temporary Apple outage")
	ingestor, err := service.NewAppleNotificationIngestor(
		newAppleNotificationVerifier(t, evidence), inbox,
		func(context.Context, service.VerifiedAppleNotification) (repository.ProviderEventProcessor, error) {
			return nil, providerErr
		},
		func() time.Time { return now },
	)
	require.NoError(t, err)
	result, err := ingestor.Ingest(context.Background(), "signed-notification")
	require.ErrorIs(t, err, providerErr)
	assert.Equal(t, model.EntitlementOutcomeRetry, result.Outcome)
	assert.Equal(t, model.EntitlementResultRetryableProvider, result.Code)
	assert.Equal(t, 1, inbox.claimCalls)
	assert.Equal(t, 1, inbox.retryCalls)
	assert.Zero(t, inbox.completeCalls)
}

func TestAppleNotificationIngestorRejectsEmptyAndOversizeBeforeClaim(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	evidence := validAppleNotificationEvidence(now)
	inbox := &fakeAppleNotificationInbox{}
	ingestor, err := service.NewAppleNotificationIngestor(
		newAppleNotificationVerifier(t, evidence), inbox,
		func(context.Context, service.VerifiedAppleNotification) (repository.ProviderEventProcessor, error) {
			return nil, nil
		},
		func() time.Time { return now },
	)
	require.NoError(t, err)
	for _, payload := range []string{"", strings.Repeat("x", 256*1024+1)} {
		_, err = ingestor.Ingest(context.Background(), payload)
		require.ErrorIs(t, err, service.ErrInvalidAppleSignedData)
	}
	assert.Zero(t, inbox.claimCalls)
}
