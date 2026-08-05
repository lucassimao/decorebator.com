package unit

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeGooglePlayClient struct {
	subscription service.GooglePlaySubscription
	err          error
	calls        int
}

func (f *fakeGooglePlayClient) GetSubscription(context.Context, string) (service.GooglePlaySubscription, error) {
	f.calls++
	return f.subscription, f.err
}

type staticGoogleTokenSource struct{ err error }

func (s staticGoogleTokenSource) AccessToken(context.Context) (string, error) {
	return "access-token", s.err
}

type fakeGoogleVerifier struct {
	results  []service.VerifiedGooglePurchase
	errors   []error
	calls    int
	overflow bool
}

func (f *fakeGoogleVerifier) Verify(
	context.Context,
	service.GooglePurchaseVerificationInput,
) (service.VerifiedGooglePurchase, error) {
	index := f.calls
	f.calls++
	if index >= len(f.results) || index >= len(f.errors) {
		f.overflow = true
		return service.VerifiedGooglePurchase{}, errors.New("unexpected verification call")
	}
	return f.results[index], f.errors[index]
}

type fakeGoogleAcknowledger struct {
	calls     int
	productID string
	token     string
	err       error
}

func (f *fakeGoogleAcknowledger) AcknowledgeSubscription(
	_ context.Context,
	productID string,
	token string,
) error {
	f.calls++
	f.productID, f.token = productID, token
	return f.err
}

func validGoogleSubscription(now time.Time, accountID string) service.GooglePlaySubscription {
	return service.GooglePlaySubscription{
		Kind: "androidpublisher#subscriptionPurchaseV2", StartTime: now.Add(-time.Hour).Format(time.RFC3339Nano),
		SubscriptionState: "SUBSCRIPTION_STATE_ACTIVE", AcknowledgementState: "ACKNOWLEDGEMENT_STATE_ACKNOWLEDGED",
		ExternalAccountIdentifiers: &service.GoogleExternalAccountIdentifiers{ObfuscatedExternalAccountID: accountID},
		LineItems: []service.GooglePlaySubscriptionLineItem{{
			ProductID: "decorebator_monthly_premium1", ExpiryTime: now.Add(29 * 24 * time.Hour).Format(time.RFC3339Nano),
			AutoRenewingPlan: &service.GoogleAutoRenewingPlan{AutoRenewEnabled: true},
		}},
		ETag: "provider-version",
	}
}

func newGoogleVerifier(t *testing.T, client service.GooglePlaySubscriptionClient, now time.Time) *service.GooglePurchaseVerifier {
	t.Helper()
	verifier, err := service.NewGooglePurchaseVerifier(client, map[string]string{
		"decorebator_monthly_premium1": "premium",
	}, func() time.Time { return now })
	require.NoError(t, err)
	return verifier
}

func TestGooglePurchaseVerifierAcceptsAccountBoundPurchase(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	account := model.GoogleAccountIdentity{UserID: 42, ObfuscatedExternalAccountID: "server-issued-account"}
	client := &fakeGooglePlayClient{subscription: validGoogleSubscription(now, account.ObfuscatedExternalAccountID)}
	result, err := newGoogleVerifier(t, client, now).Verify(context.Background(), service.GooglePurchaseVerificationInput{
		AuthenticatedUserID: 42, PurchaseToken: "purchase-token", Account: account,
	})
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementStatusActive, result.Status)
	assert.Equal(t, model.StoreEnvironmentProduction, result.Identity.Environment)
	assert.True(t, result.Acknowledged)
	assert.True(t, result.AutoRenewEnabled)
	assert.Equal(t, "premium", result.Entitlement)
}

func TestGooglePurchaseVerifierFailsClosedForIdentityProductAndProviderErrors(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	account := model.GoogleAccountIdentity{UserID: 42, ObfuscatedExternalAccountID: "server-issued-account"}
	tests := []struct {
		name    string
		failure model.EntitlementOperationFailure
		mutate  func(*fakeGooglePlayClient)
	}{
		{"account mismatch", model.EntitlementFailureAccountMismatch, func(c *fakeGooglePlayClient) {
			c.subscription.ExternalAccountIdentifiers.ObfuscatedExternalAccountID = "other"
		}},
		{"unknown product", model.EntitlementFailureUnknownProduct, func(c *fakeGooglePlayClient) { c.subscription.LineItems[0].ProductID = "other" }},
		{"multiple line items", model.EntitlementFailureInvalidEvidence, func(c *fakeGooglePlayClient) {
			c.subscription.LineItems = append(c.subscription.LineItems, c.subscription.LineItems[0])
		}},
		{"unknown state", model.EntitlementFailureInvalidEvidence, func(c *fakeGooglePlayClient) { c.subscription.SubscriptionState = "FUTURE_STATE" }},
		{"unknown acknowledgement", model.EntitlementFailureInvalidEvidence, func(c *fakeGooglePlayClient) { c.subscription.AcknowledgementState = "FUTURE_STATE" }},
		{"missing etag", model.EntitlementFailureInvalidEvidence, func(c *fakeGooglePlayClient) { c.subscription.ETag = "" }},
		{"both plan types", model.EntitlementFailureInvalidEvidence, func(c *fakeGooglePlayClient) { c.subscription.LineItems[0].PrepaidPlan = &struct{}{} }},
		{"provider retry", model.EntitlementFailureProviderUnavailable, func(c *fakeGooglePlayClient) { c.err = &service.GooglePlayAPIError{StatusCode: 503, Retryable: true} }},
		{"provider permanent", model.EntitlementFailureInvalidEvidence, func(c *fakeGooglePlayClient) { c.err = &service.GooglePlayAPIError{StatusCode: 404} }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &fakeGooglePlayClient{subscription: validGoogleSubscription(now, account.ObfuscatedExternalAccountID)}
			test.mutate(client)
			_, err := newGoogleVerifier(t, client, now).Verify(context.Background(), service.GooglePurchaseVerificationInput{
				AuthenticatedUserID: 42, PurchaseToken: "purchase-token", Account: account,
			})
			var verificationErr *service.GooglePurchaseVerificationError
			require.ErrorAs(t, err, &verificationErr)
			assert.Equal(t, test.failure, verificationErr.Failure)
		})
	}
}

func TestGooglePurchaseVerifierMapsTestPendingCanceledAndExpiredStates(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	account := model.GoogleAccountIdentity{UserID: 42, ObfuscatedExternalAccountID: "server-issued-account"}
	tests := []struct {
		state      string
		status     model.EntitlementStatus
		expiryTime time.Time
	}{
		{"SUBSCRIPTION_STATE_PENDING", model.EntitlementStatusPending, time.Time{}},
		{"SUBSCRIPTION_STATE_CANCELED", model.EntitlementStatusActive, now.Add(time.Hour)},
		{"SUBSCRIPTION_STATE_CANCELED", model.EntitlementStatusExpired, now.Add(-10 * time.Minute)},
		{"SUBSCRIPTION_STATE_ACTIVE", model.EntitlementStatusExpired, now.Add(-10 * time.Minute)},
		{"SUBSCRIPTION_STATE_PENDING_PURCHASE_CANCELED", model.EntitlementStatusExpired, time.Time{}},
		{"SUBSCRIPTION_STATE_PAUSED", model.EntitlementStatusPaused, now.Add(time.Hour)},
		{"SUBSCRIPTION_STATE_IN_GRACE_PERIOD", model.EntitlementStatusGracePeriod, now.Add(time.Hour)},
		{"SUBSCRIPTION_STATE_ON_HOLD", model.EntitlementStatusOnHold, now.Add(-time.Hour)},
		{"SUBSCRIPTION_STATE_EXPIRED", model.EntitlementStatusExpired, now.Add(-time.Hour)},
	}
	for _, test := range tests {
		subscription := validGoogleSubscription(now, account.ObfuscatedExternalAccountID)
		subscription.SubscriptionState = test.state
		if test.state == "SUBSCRIPTION_STATE_PENDING" || test.state == "SUBSCRIPTION_STATE_PENDING_PURCHASE_CANCELED" {
			subscription.StartTime = ""
		}
		if test.state == "SUBSCRIPTION_STATE_CANCELED" {
			subscription.LineItems[0].AutoRenewingPlan.AutoRenewEnabled = false
		}
		if test.expiryTime.IsZero() {
			subscription.LineItems[0].ExpiryTime = ""
		} else {
			subscription.LineItems[0].ExpiryTime = test.expiryTime.Format(time.RFC3339Nano)
		}
		subscription.TestPurchase = &struct{}{}
		result, err := newGoogleVerifier(t, &fakeGooglePlayClient{subscription: subscription}, now).Verify(context.Background(), service.GooglePurchaseVerificationInput{
			AuthenticatedUserID: 42, PurchaseToken: "purchase-token", Account: account,
		})
		require.NoError(t, err)
		assert.Equal(t, test.status, result.Status)
		assert.Equal(t, model.StoreEnvironmentSandbox, result.Identity.Environment)
	}
}

func TestGooglePurchaseVerifierAcceptsPrepaidAndPropagatesCancellationLineage(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	account := model.GoogleAccountIdentity{UserID: 42, ObfuscatedExternalAccountID: "server-issued-account"}
	prepaid := validGoogleSubscription(now, account.ObfuscatedExternalAccountID)
	prepaid.LineItems[0].AutoRenewingPlan = nil
	prepaid.LineItems[0].PrepaidPlan = &struct{}{}
	result, err := newGoogleVerifier(t, &fakeGooglePlayClient{subscription: prepaid}, now).Verify(context.Background(), service.GooglePurchaseVerificationInput{
		AuthenticatedUserID: 42, PurchaseToken: "prepaid-token", Account: account,
	})
	require.NoError(t, err)
	assert.False(t, result.AutoRenewEnabled)

	replacement := validGoogleSubscription(now, account.ObfuscatedExternalAccountID)
	replacement.SubscriptionState = "SUBSCRIPTION_STATE_CANCELED"
	replacement.LineItems[0].AutoRenewingPlan.AutoRenewEnabled = false
	replacement.LinkedPurchaseToken = "replacement-token"
	replacement.CanceledStateContext = &service.GoogleCanceledStateContext{ReplacementCancellation: &struct{}{}}
	result, err = newGoogleVerifier(t, &fakeGooglePlayClient{subscription: replacement}, now).Verify(context.Background(), service.GooglePurchaseVerificationInput{
		AuthenticatedUserID: 42, PurchaseToken: "old-token", Account: account,
	})
	require.NoError(t, err)
	assert.True(t, result.Superseded)
	assert.Equal(t, model.EntitlementStatusExpired, result.Status)
	require.NotNil(t, result.CancellationReason)
	assert.Equal(t, "replacement", *result.CancellationReason)
	assert.Equal(t, "replacement-token", result.LinkedPurchaseToken)
}

func TestGooglePurchaseVerifierRejectsMalformedIdentityAndTokenBeforeProviderCall(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	client := &fakeGooglePlayClient{}
	account := model.GoogleAccountIdentity{UserID: 42, ObfuscatedExternalAccountID: "server-issued-account"}
	for _, input := range []service.GooglePurchaseVerificationInput{
		{AuthenticatedUserID: 99, PurchaseToken: "token", Account: account},
		{AuthenticatedUserID: 42, PurchaseToken: " token", Account: account},
		{AuthenticatedUserID: 42, PurchaseToken: "", Account: account},
		{AuthenticatedUserID: 42, PurchaseToken: "token", Account: model.GoogleAccountIdentity{UserID: 42}},
	} {
		_, err := newGoogleVerifier(t, client, now).Verify(context.Background(), input)
		require.Error(t, err)
	}
	assert.Zero(t, client.calls)
}

func TestVerifiedGooglePurchaseDoesNotSerializeProviderTokens(t *testing.T) {
	value := service.VerifiedGooglePurchase{
		Identity:            model.GooglePurchaseIdentity{PurchaseToken: "raw-purchase-token"},
		LinkedPurchaseToken: "raw-linked-token", ProductID: "premium",
	}
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "raw-purchase-token")
	assert.NotContains(t, string(encoded), "raw-linked-token")
}

func TestGooglePlayClientAuthenticatesBoundsAndClassifiesSafeErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "Bearer access-token", request.Header.Get("Authorization"))
		assert.Equal(t, "/androidpublisher/v3/applications/com.lsimaocosta.decorebator/purchases/subscriptionsv2/tokens/sensitive%2Ftoken", request.RequestURI)
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"kind":"androidpublisher#subscriptionPurchaseV2","subscriptionState":"SUBSCRIPTION_STATE_PENDING"}`))
	}))
	defer server.Close()
	client, err := service.NewGooglePlayClient("com.lsimaocosta.decorebator", staticGoogleTokenSource{}, server.Client(), server.URL)
	require.NoError(t, err)
	_, err = client.GetSubscription(context.Background(), "sensitive/token")
	require.NoError(t, err)

	errorServer := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusTooManyRequests)
		_, _ = response.Write([]byte(`{"error":{"message":"sensitive provider body"}}`))
	}))
	defer errorServer.Close()
	client, err = service.NewGooglePlayClient("package", staticGoogleTokenSource{}, errorServer.Client(), errorServer.URL)
	require.NoError(t, err)
	_, err = client.GetSubscription(context.Background(), "sensitive-token")
	var apiErr *service.GooglePlayAPIError
	require.ErrorAs(t, err, &apiErr)
	assert.True(t, apiErr.Retryable)
	assert.NotContains(t, err.Error(), "sensitive")

	client, err = service.NewGooglePlayClient("package", staticGoogleTokenSource{err: &service.GoogleAccessTokenError{Retryable: true, Cause: errors.New("auth")}}, http.DefaultClient, server.URL)
	require.NoError(t, err)
	_, err = client.GetSubscription(context.Background(), "sensitive-token")
	require.ErrorAs(t, err, &apiErr)
	assert.True(t, apiErr.Retryable)
	assert.NotContains(t, err.Error(), "sensitive-token")
}

func TestGooglePlayClientRejectsInsecureRemoteBaseAndTreatsUnexpectedKindAsRetryable(t *testing.T) {
	_, err := service.NewGooglePlayClient("package", staticGoogleTokenSource{}, http.DefaultClient, "http://example.com")
	require.Error(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{"kind":"unexpected"}`))
	}))
	defer server.Close()
	client, err := service.NewGooglePlayClient("package", staticGoogleTokenSource{}, server.Client(), server.URL)
	require.NoError(t, err)
	_, err = client.GetSubscription(context.Background(), "token")
	var apiErr *service.GooglePlayAPIError
	require.ErrorAs(t, err, &apiErr)
	assert.True(t, apiErr.Retryable)
}

func TestGooglePlayClientHonorsBoundedRequestContextWithInjectedClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		<-request.Context().Done()
	}))
	defer server.Close()
	client, err := service.NewGooglePlayClient("package", staticGoogleTokenSource{}, server.Client(), server.URL)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	started := time.Now()
	_, err = client.GetSubscription(ctx, "token")
	var apiErr *service.GooglePlayAPIError
	require.ErrorAs(t, err, &apiErr)
	assert.True(t, apiErr.Retryable)
	assert.Less(t, time.Since(started), time.Second)
}

func TestGooglePlayClientClassifiesDocumentedStatusesAndRejectsOversizeResponses(t *testing.T) {
	for _, test := range []struct {
		status    int
		retryable bool
	}{
		{http.StatusBadRequest, false}, {http.StatusNotFound, false}, {http.StatusGone, false},
		{http.StatusConflict, true}, {http.StatusTooManyRequests, true}, {http.StatusServiceUnavailable, true},
	} {
		server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(test.status)
		}))
		client, err := service.NewGooglePlayClient("package", staticGoogleTokenSource{}, server.Client(), server.URL)
		require.NoError(t, err)
		_, err = client.GetSubscription(context.Background(), "token")
		var apiErr *service.GooglePlayAPIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, test.retryable, apiErr.Retryable)
		server.Close()
	}

	const responseLimit = 1 << 20
	prefix := `{"kind":"androidpublisher#subscriptionPurchaseV2","padding":"`
	suffix := `"}`
	exactBody := prefix + strings.Repeat("x", responseLimit-len(prefix)-len(suffix)) + suffix
	exactServer := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(exactBody))
	}))
	client, err := service.NewGooglePlayClient("package", staticGoogleTokenSource{}, exactServer.Client(), exactServer.URL)
	require.NoError(t, err)
	_, err = client.GetSubscription(context.Background(), "token")
	require.NoError(t, err)
	exactServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{"kind":"androidpublisher#subscriptionPurchaseV2","padding":"` + strings.Repeat("x", 1<<20) + `"}`))
	}))
	defer server.Close()
	client, err = service.NewGooglePlayClient("package", staticGoogleTokenSource{}, server.Client(), server.URL)
	require.NoError(t, err)
	_, err = client.GetSubscription(context.Background(), "token")
	var apiErr *service.GooglePlayAPIError
	require.ErrorAs(t, err, &apiErr)
	assert.True(t, apiErr.Retryable)
}

func TestGooglePurchaseProcessorVerifiesBeforeAcknowledgingPaidPurchase(t *testing.T) {
	verified := service.VerifiedGooglePurchase{
		ProductID: "decorebator_monthly_premium1", Status: model.EntitlementStatusActive,
	}
	verifier := &fakeGoogleVerifier{results: []service.VerifiedGooglePurchase{verified}, errors: []error{nil}}
	acknowledger := &fakeGoogleAcknowledger{}
	processor, err := service.NewGooglePurchaseProcessor(verifier, acknowledger)
	require.NoError(t, err)
	result, err := processor.VerifyAndAcknowledge(context.Background(), service.GooglePurchaseVerificationInput{PurchaseToken: "token"})
	require.NoError(t, err)
	assert.True(t, result.Purchase.Acknowledged)
	assert.Equal(t, service.GoogleAcknowledgementConfirmed, result.Acknowledgement)
	assert.Equal(t, 1, verifier.calls)
	assert.Equal(t, 1, acknowledger.calls)
	assert.Equal(t, verified.ProductID, acknowledger.productID)
	assert.Equal(t, "token", acknowledger.token)
}

func TestGooglePurchaseProcessorNeverAcknowledgesPendingExpiredOrFailedVerification(t *testing.T) {
	for _, status := range []model.EntitlementStatus{
		model.EntitlementStatusPending, model.EntitlementStatusExpired, model.EntitlementStatusRevoked,
	} {
		verifier := &fakeGoogleVerifier{
			results: []service.VerifiedGooglePurchase{{ProductID: "premium", Status: status}}, errors: []error{nil},
		}
		acknowledger := &fakeGoogleAcknowledger{}
		processor, err := service.NewGooglePurchaseProcessor(verifier, acknowledger)
		require.NoError(t, err)
		result, err := processor.VerifyAndAcknowledge(context.Background(), service.GooglePurchaseVerificationInput{PurchaseToken: "token"})
		require.NoError(t, err)
		assert.Equal(t, status, result.Purchase.Status)
		assert.False(t, result.Purchase.Acknowledged)
		assert.Equal(t, service.GoogleAcknowledgementNotRequired, result.Acknowledgement)
		assert.Zero(t, acknowledger.calls)
	}

	verificationErr := errors.New("verification failed")
	verifier := &fakeGoogleVerifier{results: []service.VerifiedGooglePurchase{{}}, errors: []error{verificationErr}}
	acknowledger := &fakeGoogleAcknowledger{}
	processor, err := service.NewGooglePurchaseProcessor(verifier, acknowledger)
	require.NoError(t, err)
	_, err = processor.VerifyAndAcknowledge(context.Background(), service.GooglePurchaseVerificationInput{PurchaseToken: "token"})
	require.ErrorIs(t, err, verificationErr)
	assert.Zero(t, acknowledger.calls)
}

func TestGooglePurchaseProcessorConfirmsIdempotentPermanentAcknowledgeRace(t *testing.T) {
	initial := service.VerifiedGooglePurchase{ProductID: "premium", Status: model.EntitlementStatusActive}
	refreshed := initial
	refreshed.Acknowledged = true
	verifier := &fakeGoogleVerifier{
		results: []service.VerifiedGooglePurchase{initial, refreshed}, errors: []error{nil, nil},
	}
	acknowledger := &fakeGoogleAcknowledger{err: &service.GooglePlayAPIError{StatusCode: http.StatusBadRequest}}
	processor, err := service.NewGooglePurchaseProcessor(verifier, acknowledger)
	require.NoError(t, err)
	result, err := processor.VerifyAndAcknowledge(context.Background(), service.GooglePurchaseVerificationInput{PurchaseToken: "token"})
	require.NoError(t, err)
	assert.True(t, result.Purchase.Acknowledged)
	assert.Equal(t, service.GoogleAcknowledgementConfirmed, result.Acknowledgement)
	assert.Equal(t, 2, verifier.calls)
	assert.Equal(t, 1, acknowledger.calls)
}

func TestGooglePlayClientSendsBoundedAuthenticatedAcknowledge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "Bearer access-token", request.Header.Get("Authorization"))
		assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
		assert.Equal(t, "/androidpublisher/v3/applications/com.app/purchases/subscriptions/premium%2Fmonthly/tokens/token%2Fvalue:acknowledge", request.RequestURI)
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{}`, string(body))
		response.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client, err := service.NewGooglePlayClient("com.app", staticGoogleTokenSource{}, server.Client(), server.URL)
	require.NoError(t, err)
	require.NoError(t, client.AcknowledgeSubscription(context.Background(), "premium/monthly", "token/value"))
}

func TestGooglePurchaseProcessorPreservesVerifiedPurchaseAcrossAcknowledgeFailures(t *testing.T) {
	initial := service.VerifiedGooglePurchase{ProductID: "premium", Status: model.EntitlementStatusActive}
	for _, test := range []struct {
		name        string
		ackErr      error
		results     []service.VerifiedGooglePurchase
		errors      []error
		outcome     service.GoogleAcknowledgementOutcome
		verifyCalls int
	}{
		{
			name: "retryable skips refresh", ackErr: &service.GooglePlayAPIError{StatusCode: 503, Retryable: true},
			results: []service.VerifiedGooglePurchase{initial}, errors: []error{nil},
			outcome: service.GoogleAcknowledgementRetry, verifyCalls: 1,
		},
		{
			name: "permanent still unacknowledged", ackErr: &service.GooglePlayAPIError{StatusCode: 403},
			results: []service.VerifiedGooglePurchase{initial, initial}, errors: []error{nil, nil},
			outcome: service.GoogleAcknowledgementFailed, verifyCalls: 2,
		},
		{
			name: "permanent refresh unavailable", ackErr: &service.GooglePlayAPIError{StatusCode: 400},
			results: []service.VerifiedGooglePurchase{initial, {}}, errors: []error{nil, errors.New("refresh failed")},
			outcome: service.GoogleAcknowledgementRetry, verifyCalls: 2,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			verifier := &fakeGoogleVerifier{results: test.results, errors: test.errors}
			acknowledger := &fakeGoogleAcknowledger{err: test.ackErr}
			processor, err := service.NewGooglePurchaseProcessor(verifier, acknowledger)
			require.NoError(t, err)
			result, err := processor.VerifyAndAcknowledge(context.Background(), service.GooglePurchaseVerificationInput{PurchaseToken: "token"})
			require.NoError(t, err)
			assert.Equal(t, initial.ProductID, result.Purchase.ProductID)
			assert.Equal(t, test.outcome, result.Acknowledgement)
			assert.Error(t, result.AcknowledgementError)
			assert.Equal(t, test.verifyCalls, verifier.calls)
			assert.False(t, verifier.overflow)
		})
	}
}

func TestGooglePlayClientClassifiesAcknowledgeFailuresAndRejectsLocalInput(t *testing.T) {
	for _, test := range []struct {
		status int
		retry  bool
	}{{503, true}, {403, false}} {
		server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(test.status)
		}))
		client, err := service.NewGooglePlayClient("com.app", staticGoogleTokenSource{}, server.Client(), server.URL)
		require.NoError(t, err)
		err = client.AcknowledgeSubscription(context.Background(), "premium", "token")
		var apiErr *service.GooglePlayAPIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, test.retry, apiErr.Retryable)
		server.Close()
	}
	client, err := service.NewGooglePlayClient("com.app", staticGoogleTokenSource{}, http.DefaultClient, "https://example.com")
	require.NoError(t, err)
	for _, input := range [][2]string{{"", "token"}, {"premium", " token"}} {
		err = client.AcknowledgeSubscription(context.Background(), input[0], input[1])
		require.ErrorIs(t, err, service.ErrInvalidGoogleAcknowledgeRequest)
	}
	requestCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requestCalls++ }))
	defer server.Close()
	client, err = service.NewGooglePlayClient(
		"com.app",
		staticGoogleTokenSource{err: &service.GoogleAccessTokenError{Retryable: false, Cause: errors.New("invalid grant")}},
		server.Client(), server.URL,
	)
	require.NoError(t, err)
	err = client.AcknowledgeSubscription(context.Background(), "premium", "token")
	var apiErr *service.GooglePlayAPIError
	require.ErrorAs(t, err, &apiErr)
	assert.False(t, apiErr.Retryable)
	assert.Zero(t, requestCalls)
	assert.NotContains(t, err.Error(), "invalid grant")
}
