package unit

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
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

const (
	googleRTDNAudience     = "https://api.decorebator.com/webhook/google-play"
	googleRTDNEmail        = "play-rtdn@project.iam.gserviceaccount.com"
	googleRTDNTopic        = "projects/project/topics/play-rtdn"
	googleRTDNSubscription = "projects/project/subscriptions/play-rtdn-push"
	googleRTDNPackage      = "com.decorebator.app"
)

type fakeGooglePubSubVerifier struct {
	claims   service.GooglePubSubClaims
	err      error
	token    string
	audience string
	calls    int
}

func (f *fakeGooglePubSubVerifier) Verify(
	_ context.Context,
	token string,
	audience string,
) (service.GooglePubSubClaims, error) {
	f.calls++
	f.token, f.audience = token, audience
	return f.claims, f.err
}

type fakeGoogleRTDNInbox struct {
	input          repository.ProviderEventInboxInput
	claimResult    repository.ProviderEventInboxResult
	claimErr       error
	completeResult repository.ProviderEventInboxResult
	completeErr    error
	processor      repository.ProviderEventProcessor
	retryAt        time.Time
	retryErr       error
	claimCalls     int
	completeCalls  int
	retryCalls     int
}

func (f *fakeGoogleRTDNInbox) Claim(
	_ context.Context,
	input repository.ProviderEventInboxInput,
	_ time.Time,
	_ time.Duration,
) (repository.ProviderEventInboxResult, error) {
	f.claimCalls++
	f.input = input
	if f.claimResult.Claim == nil && !f.claimResult.Duplicate && f.claimErr == nil {
		f.claimResult.Claim = &repository.ProviderEventClaim{ID: 1, LeaseToken: uuid.New(), Input: input}
	}
	return f.claimResult, f.claimErr
}

func (f *fakeGoogleRTDNInbox) Complete(
	ctx context.Context,
	_ repository.ProviderEventClaim,
	processor repository.ProviderEventProcessor,
) (repository.ProviderEventInboxResult, error) {
	f.completeCalls++
	f.processor = processor
	operation, err := processor(ctx, nil)
	if err != nil {
		return repository.ProviderEventInboxResult{}, err
	}
	if f.completeResult.Outcome == "" {
		f.completeResult.Outcome, f.completeResult.Code = operation.Outcome, operation.Code
	}
	return f.completeResult, f.completeErr
}

func (f *fakeGoogleRTDNInbox) MarkRetryable(
	_ context.Context,
	_ repository.ProviderEventClaim,
	next time.Time,
) error {
	f.retryCalls++
	f.retryAt = next
	return f.retryErr
}

func validGoogleRTDNVerifier() *fakeGooglePubSubVerifier {
	return &fakeGooglePubSubVerifier{claims: service.GooglePubSubClaims{
		Issuer: "https://accounts.google.com", Email: googleRTDNEmail, EmailVerified: true,
	}}
}

func googleRTDNBody(t *testing.T, messageID string, notification map[string]any) []byte {
	t.Helper()
	data, err := json.Marshal(notification)
	require.NoError(t, err)
	body, err := json.Marshal(map[string]any{
		"message":      map[string]any{"messageId": messageID, "data": base64.StdEncoding.EncodeToString(data)},
		"subscription": googleRTDNSubscription,
	})
	require.NoError(t, err)
	return body
}

func newGoogleRTDNIngestor(
	t *testing.T,
	verifier service.GooglePubSubTokenVerifier,
	inbox service.GoogleRTDNInbox,
	refresher service.GoogleRTDNRefresher,
	now time.Time,
) *service.GoogleRTDNIngestor {
	t.Helper()
	ingestor, err := service.NewGoogleRTDNIngestor(
		verifier, inbox, refresher, googleRTDNAudience, googleRTDNEmail,
		googleRTDNTopic, googleRTDNSubscription, googleRTDNPackage, func() time.Time { return now },
	)
	require.NoError(t, err)
	return ingestor
}

func TestGoogleRTDNAuthenticatesBeforeParsingOrClaiming(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name   string
		header string
		claims service.GooglePubSubClaims
	}{
		{name: "missing bearer"},
		{name: "wrong issuer", header: "Bearer signed", claims: service.GooglePubSubClaims{Issuer: "attacker", Email: googleRTDNEmail, EmailVerified: true}},
		{name: "wrong email", header: "Bearer signed", claims: service.GooglePubSubClaims{Issuer: "accounts.google.com", Email: "other@example.com", EmailVerified: true}},
		{name: "unverified email", header: "Bearer signed", claims: service.GooglePubSubClaims{Issuer: "accounts.google.com", Email: googleRTDNEmail}},
	} {
		t.Run(test.name, func(t *testing.T) {
			verifier := &fakeGooglePubSubVerifier{claims: test.claims}
			inbox := &fakeGoogleRTDNInbox{}
			ingestor := newGoogleRTDNIngestor(t, verifier, inbox, func(context.Context, service.GoogleRTDNEvent) (repository.ProviderEventProcessor, error) {
				t.Fatal("refresher must not run")
				return nil, nil
			}, now)
			_, err := ingestor.Ingest(context.Background(), test.header, []byte(`not-json`))
			require.Error(t, err)
			assert.Zero(t, inbox.claimCalls)
		})
	}

	verifier := validGoogleRTDNVerifier()
	inbox := &fakeGoogleRTDNInbox{}
	ingestor := newGoogleRTDNIngestor(t, verifier, inbox, func(context.Context, service.GoogleRTDNEvent) (repository.ProviderEventProcessor, error) {
		return nil, errors.New("unused")
	}, now)
	_, _ = ingestor.Ingest(context.Background(), "Bearer signed-token", []byte(`not-json`))
	assert.Equal(t, "signed-token", verifier.token)
	assert.Equal(t, googleRTDNAudience, verifier.audience)
}

func TestGoogleRTDNDispositionContractFailsUnknownErrorsSafe(t *testing.T) {
	assert.Equal(t, service.GoogleRTDNAcknowledge, service.GoogleRTDNDispositionOf(nil))
	assert.Equal(t, service.GoogleRTDNRetry, service.GoogleRTDNDispositionOf(errors.New("unknown")))
	assert.Equal(t, service.GoogleRTDNUnauthorized, service.GoogleRTDNDispositionOf(&service.GoogleRTDNDeliveryError{
		Disposition: service.GoogleRTDNUnauthorized,
	}))
}

func TestGoogleRTDNClaimsTransportAndSemanticKeysThenRefreshes(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	inbox := &fakeGoogleRTDNInbox{}
	refreshCalls := 0
	ingestor := newGoogleRTDNIngestor(t, validGoogleRTDNVerifier(), inbox, func(_ context.Context, event service.GoogleRTDNEvent) (repository.ProviderEventProcessor, error) {
		refreshCalls++
		assert.Equal(t, "subscription", event.Kind)
		assert.Equal(t, int32(4), event.NotificationType)
		assert.Equal(t, "purchase-token", event.PurchaseToken)
		return func(context.Context, pgx.Tx) (model.EntitlementOperationResult, error) {
			return model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeApplied, Code: model.EntitlementResultApplied}, nil
		}, nil
	}, now)
	body := googleRTDNBody(t, "message-1", map[string]any{
		"version": "1.0", "packageName": googleRTDNPackage, "eventTimeMillis": now.Add(-time.Minute).UnixMilli(),
		"subscriptionNotification":   map[string]any{"version": "1.0", "notificationType": 4, "purchaseToken": "purchase-token"},
		"oneTimeProductNotification": nil,
	})
	result, err := ingestor.Ingest(context.Background(), "Bearer signed", body)
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementOutcomeApplied, result.Outcome)
	assert.Equal(t, 1, refreshCalls)
	assert.Equal(t, model.EntitlementStoreGoogle, inbox.input.Store)
	assert.Nil(t, inbox.input.Environment)
	require.NotNil(t, inbox.input.SemanticKey)
	assert.Contains(t, inbox.input.IdempotencyKey.String(), "google_rtdn_transport:")
	assert.Contains(t, inbox.input.SemanticKey.String(), "google_rtdn_semantic:")
	assert.NotContains(t, inbox.input.SemanticKey.String(), "purchase-token")
}

func TestGoogleRTDNSeparatesVoidsAndRedeliveryWithoutExtraRefresh(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	inbox := &fakeGoogleRTDNInbox{claimResult: repository.ProviderEventInboxResult{
		Duplicate: true, Outcome: model.EntitlementOutcomeApplied, Code: model.EntitlementResultRevoked,
	}}
	refreshCalls := 0
	ingestor := newGoogleRTDNIngestor(t, validGoogleRTDNVerifier(), inbox, func(context.Context, service.GoogleRTDNEvent) (repository.ProviderEventProcessor, error) {
		refreshCalls++
		return nil, nil
	}, now)
	body := googleRTDNBody(t, "message-void", map[string]any{
		"version": "1.0", "packageName": googleRTDNPackage, "eventTimeMillis": now.UnixMilli(),
		"voidedPurchaseNotification": map[string]any{"purchaseToken": "purchase-token", "productType": 1, "refundType": 1},
	})
	result, err := ingestor.Ingest(context.Background(), "Bearer signed", body)
	require.NoError(t, err)
	assert.True(t, result.Duplicate)
	assert.Equal(t, "voided_subscription", inbox.input.ProviderEventType)
	assert.Zero(t, refreshCalls)
}

func TestGoogleRTDNRecordsTestAndPoisonEventsWithoutRefresh(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name         string
		notification map[string]any
		wantOutcome  model.EntitlementOperationOutcome
		wantCode     model.EntitlementResultCode
	}{
		{name: "test", notification: map[string]any{
			"version": "1.0", "packageName": googleRTDNPackage, "eventTimeMillis": now.UnixMilli(),
			"testNotification": map[string]any{"version": "1.0"},
		}, wantOutcome: model.EntitlementOutcomeUnchanged, wantCode: model.EntitlementResultProviderTest},
		{name: "unsupported one-time", notification: map[string]any{
			"version": "1.0", "packageName": googleRTDNPackage, "eventTimeMillis": now.UnixMilli(),
			"oneTimeProductNotification": map[string]any{"version": "1.0", "notificationType": 1, "purchaseToken": "token", "sku": "sku"},
		}, wantOutcome: model.EntitlementOutcomeRejected, wantCode: model.EntitlementResultInvalidPurchase},
		{name: "wrong package", notification: map[string]any{
			"version": "1.0", "packageName": "attacker.package", "eventTimeMillis": now.UnixMilli(),
			"testNotification": map[string]any{"version": "1.0"},
		}, wantOutcome: model.EntitlementOutcomeRejected, wantCode: model.EntitlementResultInvalidPurchase},
	} {
		t.Run(test.name, func(t *testing.T) {
			inbox := &fakeGoogleRTDNInbox{}
			ingestor := newGoogleRTDNIngestor(t, validGoogleRTDNVerifier(), inbox, func(context.Context, service.GoogleRTDNEvent) (repository.ProviderEventProcessor, error) {
				t.Fatal("refresher must not run")
				return nil, nil
			}, now)
			result, err := ingestor.Ingest(context.Background(), "Bearer signed", googleRTDNBody(t, "message-"+test.name, test.notification))
			require.NoError(t, err)
			assert.Equal(t, test.wantOutcome, result.Outcome)
			assert.Equal(t, test.wantCode, result.Code)
			assert.Equal(t, 1, inbox.completeCalls)
		})
	}
}

func TestGoogleRTDNUnknownPositiveSubscriptionTypeStillRefreshes(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	inbox := &fakeGoogleRTDNInbox{}
	refreshed := false
	ingestor := newGoogleRTDNIngestor(t, validGoogleRTDNVerifier(), inbox, func(_ context.Context, event service.GoogleRTDNEvent) (repository.ProviderEventProcessor, error) {
		refreshed = true
		assert.Equal(t, int32(999), event.NotificationType)
		return func(context.Context, pgx.Tx) (model.EntitlementOperationResult, error) {
			return model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeUnchanged, Code: model.EntitlementResultStaleEvent}, nil
		}, nil
	}, now)
	body := googleRTDNBody(t, "message-future-type", map[string]any{
		"version": "1.0", "packageName": googleRTDNPackage, "eventTimeMillis": now.UnixMilli(),
		"subscriptionNotification": map[string]any{"version": "1.0", "notificationType": 999, "purchaseToken": "token"},
	})
	_, err := ingestor.Ingest(context.Background(), "Bearer signed", body)
	require.NoError(t, err)
	assert.True(t, refreshed)
}

func TestGoogleRTDNNonterminalDuplicateRequiresRedelivery(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	inbox := &fakeGoogleRTDNInbox{claimResult: repository.ProviderEventInboxResult{
		Duplicate: true, Outcome: model.EntitlementOutcomeRetry, Code: model.EntitlementResultRetryableProvider,
	}}
	ingestor := newGoogleRTDNIngestor(t, validGoogleRTDNVerifier(), inbox, func(context.Context, service.GoogleRTDNEvent) (repository.ProviderEventProcessor, error) {
		t.Fatal("refresher must not run")
		return nil, nil
	}, now)
	body := googleRTDNBody(t, "message-processing", map[string]any{
		"version": "1.0", "packageName": googleRTDNPackage, "eventTimeMillis": now.UnixMilli(),
		"subscriptionNotification": map[string]any{"version": "1.0", "notificationType": 2, "purchaseToken": "token"},
	})
	result, err := ingestor.Ingest(context.Background(), "Bearer signed", body)
	assert.Equal(t, model.EntitlementOutcomeRetry, result.Outcome)
	var deliveryError *service.GoogleRTDNDeliveryError
	require.ErrorAs(t, err, &deliveryError)
	assert.Equal(t, service.GoogleRTDNRetry, deliveryError.Disposition)
	assert.NotContains(t, err.Error(), "token")
}

func TestGoogleRTDNClassifiesRefreshFailuresAndBoundsRefresh(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	body := googleRTDNBody(t, "message-failure", map[string]any{
		"version": "1.0", "packageName": googleRTDNPackage, "eventTimeMillis": now.UnixMilli(),
		"subscriptionNotification": map[string]any{"version": "1.0", "notificationType": 2, "purchaseToken": "token"},
	})

	permanentInbox := &fakeGoogleRTDNInbox{}
	permanent := newGoogleRTDNIngestor(t, validGoogleRTDNVerifier(), permanentInbox, func(ctx context.Context, _ service.GoogleRTDNEvent) (repository.ProviderEventProcessor, error) {
		deadline, ok := ctx.Deadline()
		assert.True(t, ok)
		assert.LessOrEqual(t, time.Until(deadline), 20*time.Second)
		return nil, &service.GooglePurchaseVerificationError{Failure: model.EntitlementFailureUnknownProduct}
	}, now)
	result, err := permanent.Ingest(context.Background(), "Bearer signed", body)
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementOutcomeRejected, result.Outcome)
	assert.Equal(t, model.EntitlementResultUnknownProduct, result.Code)
	assert.Zero(t, permanentInbox.retryCalls)

	retryInbox := &fakeGoogleRTDNInbox{}
	retry := newGoogleRTDNIngestor(t, validGoogleRTDNVerifier(), retryInbox, func(context.Context, service.GoogleRTDNEvent) (repository.ProviderEventProcessor, error) {
		return nil, &service.GooglePurchaseVerificationError{Failure: model.EntitlementFailureProviderUnavailable}
	}, now)
	result, err = retry.Ingest(context.Background(), "Bearer signed", body)
	assert.Equal(t, model.EntitlementOutcomeRetry, result.Outcome)
	var deliveryError *service.GoogleRTDNDeliveryError
	require.ErrorAs(t, err, &deliveryError)
	assert.Equal(t, service.GoogleRTDNRetry, deliveryError.Disposition)
	assert.Equal(t, 1, retryInbox.retryCalls)
}

func TestGoogleRTDNVoidedSubscriptionRefreshesWithForwardCompatibleFields(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	for _, voided := range []map[string]any{
		{"purchaseToken": "token"},
		{"purchaseToken": "token", "productType": 1, "refundType": 2},
	} {
		inbox := &fakeGoogleRTDNInbox{}
		refreshed := false
		ingestor := newGoogleRTDNIngestor(t, validGoogleRTDNVerifier(), inbox, func(_ context.Context, event service.GoogleRTDNEvent) (repository.ProviderEventProcessor, error) {
			refreshed = true
			assert.Equal(t, "voided_subscription", event.Kind)
			return func(context.Context, pgx.Tx) (model.EntitlementOperationResult, error) {
				return model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeApplied, Code: model.EntitlementResultRevoked}, nil
			}, nil
		}, now)
		body := googleRTDNBody(t, "message-void-forward", map[string]any{
			"version": "1.0", "packageName": googleRTDNPackage, "eventTimeMillis": now.UnixMilli(),
			"voidedPurchaseNotification": voided,
		})
		_, err := ingestor.Ingest(context.Background(), "Bearer signed", body)
		require.NoError(t, err)
		assert.True(t, refreshed)
	}
}

func TestGoogleRTDNPersistsRetryBeforeRequestingRedelivery(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	inbox := &fakeGoogleRTDNInbox{}
	providerErr := errors.New("provider unavailable")
	ingestor := newGoogleRTDNIngestor(t, validGoogleRTDNVerifier(), inbox, func(context.Context, service.GoogleRTDNEvent) (repository.ProviderEventProcessor, error) {
		return nil, providerErr
	}, now)
	body := googleRTDNBody(t, "message-retry", map[string]any{
		"version": "1.0", "packageName": googleRTDNPackage, "eventTimeMillis": now.UnixMilli(),
		"subscriptionNotification": map[string]any{"version": "1.0", "notificationType": 2, "purchaseToken": "token"},
	})
	result, err := ingestor.Ingest(context.Background(), "Bearer signed", body)
	assert.ErrorIs(t, err, providerErr)
	assert.Equal(t, model.EntitlementOutcomeRetry, result.Outcome)
	assert.Equal(t, 1, inbox.retryCalls)
	assert.Equal(t, now.Add(time.Minute), inbox.retryAt)
	assert.Zero(t, inbox.completeCalls)
}

func TestGoogleRTDNRetryPersistenceFailureAndTokenFailureRemainRetryableAndSafe(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	body := googleRTDNBody(t, "message-retry-error", map[string]any{
		"version": "1.0", "packageName": googleRTDNPackage, "eventTimeMillis": now.UnixMilli(),
		"subscriptionNotification": map[string]any{"version": "1.0", "notificationType": 2, "purchaseToken": "sensitive-token"},
	})
	inbox := &fakeGoogleRTDNInbox{retryErr: errors.New("database unavailable")}
	ingestor := newGoogleRTDNIngestor(t, validGoogleRTDNVerifier(), inbox, func(context.Context, service.GoogleRTDNEvent) (repository.ProviderEventProcessor, error) {
		return nil, errors.New("provider unavailable")
	}, now)
	_, err := ingestor.Ingest(context.Background(), "Bearer signed", body)
	var deliveryError *service.GoogleRTDNDeliveryError
	require.ErrorAs(t, err, &deliveryError)
	assert.Equal(t, service.GoogleRTDNRetry, deliveryError.Disposition)
	assert.NotContains(t, err.Error(), "sensitive-token")

	verifier := &fakeGooglePubSubVerifier{err: &service.GooglePubSubTokenError{Retryable: true, Cause: errors.New("sensitive certificate failure")}}
	inbox = &fakeGoogleRTDNInbox{}
	ingestor = newGoogleRTDNIngestor(t, verifier, inbox, func(context.Context, service.GoogleRTDNEvent) (repository.ProviderEventProcessor, error) {
		return nil, nil
	}, now)
	_, err = ingestor.Ingest(context.Background(), "Bearer signed", body)
	require.ErrorAs(t, err, &deliveryError)
	assert.Equal(t, service.GoogleRTDNRetry, deliveryError.Disposition)
	assert.NotContains(t, err.Error(), "sensitive")
	assert.Zero(t, inbox.claimCalls)
}

func TestGoogleRTDNRejectsMalformedEnvelopeBeforeInbox(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	for _, body := range [][]byte{
		[]byte(`{}`),
		[]byte(`not-json`),
		make([]byte, 64*1024+1),
	} {
		inbox := &fakeGoogleRTDNInbox{}
		ingestor := newGoogleRTDNIngestor(t, validGoogleRTDNVerifier(), inbox, func(context.Context, service.GoogleRTDNEvent) (repository.ProviderEventProcessor, error) {
			return nil, errors.New("unused")
		}, now)
		_, err := ingestor.Ingest(context.Background(), "Bearer signed", body)
		require.Error(t, err)
		var deliveryError *service.GoogleRTDNDeliveryError
		require.ErrorAs(t, err, &deliveryError)
		assert.Equal(t, service.GoogleRTDNAcknowledge, deliveryError.Disposition)
		assert.Zero(t, inbox.claimCalls)
	}
}
