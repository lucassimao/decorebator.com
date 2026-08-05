package unit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"decorebator.com/internal/app"
	decorebatorhttp "decorebator.com/internal/http"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAppleNotificationReceiver struct {
	result  repository.ProviderEventInboxResult
	err     error
	payload string
	calls   int
}

func (f *fakeAppleNotificationReceiver) Ingest(
	_ context.Context,
	payload string,
) (repository.ProviderEventInboxResult, error) {
	f.calls++
	f.payload = payload
	return f.result, f.err
}

type fakeGoogleRTDNReceiver struct {
	result        repository.ProviderEventInboxResult
	err           error
	authorization string
	body          []byte
	calls         int
}

func (f *fakeGoogleRTDNReceiver) Ingest(
	_ context.Context,
	authorization string,
	body []byte,
) (repository.ProviderEventInboxResult, error) {
	f.calls++
	f.authorization = authorization
	f.body = append([]byte(nil), body...)
	return f.result, f.err
}

func TestAppleStoreWebhookMapsTerminalAndRetryDispositions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	receiver := &fakeAppleNotificationReceiver{result: repository.ProviderEventInboxResult{
		Outcome: model.EntitlementOutcomeApplied, Code: model.EntitlementResultApplied,
	}}
	metrics := decorebatorhttp.NewStoreWebhookMetrics()
	router := gin.New()
	router.POST("/webhook/app-store", decorebatorhttp.HandleAppleStoreNotification(receiver, metrics))

	response := performWebhookRequest(t, router, "/webhook/app-store", `{"signedPayload":"sensitive-jws"}`, "")
	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Equal(t, "sensitive-jws", receiver.payload)
	assert.Equal(t, uint64(1), metrics.Snapshot()[decorebatorhttp.StoreWebhookObservation{
		Provider: "apple", Disposition: "acknowledge",
		Outcome: model.EntitlementOutcomeApplied, ResultCode: model.EntitlementResultApplied,
	}])

	receiver.err = &service.AppleNotificationDeliveryError{
		Disposition: service.AppleNotificationRetry, Cause: errors.New("database unavailable"),
	}
	response = performWebhookRequest(t, router, "/webhook/app-store", `{"signedPayload":"sensitive-jws"}`, "")
	assert.Equal(t, http.StatusServiceUnavailable, response.Code)
}

func TestAppleStoreWebhookAcknowledgesPoisonAndRejectsOversizeBeforeIngestion(t *testing.T) {
	receiver := &fakeAppleNotificationReceiver{}
	router := gin.New()
	router.POST("/webhook/app-store", decorebatorhttp.HandleAppleStoreNotification(receiver, nil))

	response := performWebhookRequest(t, router, "/webhook/app-store", `{"unknown":"value"}`, "")
	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Zero(t, receiver.calls)

	response = performWebhookRequest(t, router, "/webhook/app-store", strings.Repeat("x", 300*1024+1), "")
	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Zero(t, receiver.calls)
}

func TestAppleStoreWebhookToleratesFutureEnvelopeFields(t *testing.T) {
	receiver := &fakeAppleNotificationReceiver{}
	router := gin.New()
	router.POST("/webhook/app-store", decorebatorhttp.HandleAppleStoreNotification(receiver, nil))
	response := performWebhookRequest(
		t, router, "/webhook/app-store", `{"signedPayload":"signed-jws","futureField":{"version":2}}`, "",
	)
	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Equal(t, 1, receiver.calls)
	assert.Equal(t, "signed-jws", receiver.payload)
}

func TestGoogleRTDNWebhookMapsAcknowledgeUnauthorizedAndRetry(t *testing.T) {
	receiver := &fakeGoogleRTDNReceiver{result: repository.ProviderEventInboxResult{
		Outcome: model.EntitlementOutcomeUnchanged, Code: model.EntitlementResultProviderTest,
	}}
	metrics := decorebatorhttp.NewStoreWebhookMetrics()
	router := gin.New()
	router.POST("/webhook/google-play/rtdn", decorebatorhttp.HandleGooglePlayRTDN(receiver, metrics))

	response := performWebhookRequest(t, router, "/webhook/google-play/rtdn", `{"message":{},"deliveryAttempt":3}`, "Bearer signed-token")
	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Equal(t, "Bearer signed-token", receiver.authorization)
	assert.JSONEq(t, `{"message":{},"deliveryAttempt":3}`, string(receiver.body))
	assert.Equal(t, uint64(1), metrics.Snapshot()[decorebatorhttp.StoreWebhookObservation{
		Provider: "google", Disposition: "acknowledge",
		Outcome: model.EntitlementOutcomeUnchanged, ResultCode: model.EntitlementResultProviderTest,
		DeliveryAttempt: 3,
	}])

	receiver.err = &service.GoogleRTDNDeliveryError{
		Disposition: service.GoogleRTDNUnauthorized, Cause: service.ErrInvalidGooglePubSubAuthentication,
	}
	response = performWebhookRequest(t, router, "/webhook/google-play/rtdn", `{}`, "bad")
	assert.Equal(t, http.StatusUnauthorized, response.Code)

	receiver.err = &service.GoogleRTDNDeliveryError{
		Disposition: service.GoogleRTDNRetry, Cause: errors.New("provider unavailable"),
	}
	response = performWebhookRequest(t, router, "/webhook/google-play/rtdn", `{}`, "Bearer signed-token")
	assert.Equal(t, http.StatusServiceUnavailable, response.Code)
}

func TestGoogleRTDNWebhookAcknowledgesOversizePoisonBeforeIngestion(t *testing.T) {
	receiver := &fakeGoogleRTDNReceiver{}
	router := gin.New()
	router.POST("/webhook/google-play/rtdn", decorebatorhttp.HandleGooglePlayRTDN(receiver, nil))
	response := performWebhookRequest(t, router, "/webhook/google-play/rtdn", strings.Repeat("x", 64*1024+1), "Bearer token")
	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Zero(t, receiver.calls)
}

func TestStoreWebhookRoutesAreAbsentWhenIAPIsDisabled(t *testing.T) {
	router := gin.New()
	decorebatorhttp.RegisterStoreWebhookRoutes(router, &app.Context{StoreIAPEnabled: false}, nil)
	for _, route := range router.Routes() {
		assert.NotEqual(t, "/webhook/app-store", route.Path)
		assert.NotEqual(t, "/webhook/google-play/rtdn", route.Path)
	}
}

func TestStoreWebhookRoutesFailClosedWhenEnabledReceiversAreMissing(t *testing.T) {
	assert.Panics(t, func() {
		decorebatorhttp.RegisterStoreWebhookRoutes(
			gin.New(), &app.Context{StoreIAPEnabled: true}, decorebatorhttp.NewStoreWebhookMetrics(),
		)
	})
}

func TestLegacyProviderWebhookRoutesAreAbsentWhenDisabled(t *testing.T) {
	previous, existed := os.LookupEnv("REVENUECAT_WEBHOOK_AUTHORIZATION")
	require.NoError(t, os.Unsetenv("REVENUECAT_WEBHOOK_AUTHORIZATION"))
	t.Cleanup(func() {
		if existed {
			require.NoError(t, os.Setenv("REVENUECAT_WEBHOOK_AUTHORIZATION", previous))
		}
	})
	router := gin.New()
	appCtx := &app.Context{
		LegacyProviderSurfaceEnabled: false,
	}
	decorebatorhttp.RegisterLegacyProviderPublicRoutes(router, appCtx)
	decorebatorhttp.RegisterLegacyProviderAuthenticatedRoutes(router.Group("/"), appCtx)

	for _, path := range []string{
		"/webhook/stripe", "/webhook/revenuecat", "/subscription/checkout-session",
		"/subscription/revenuecat/restore", "/subscription/checkout-redirect",
	} {
		request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`))
		response := httptest.NewRecorder()
		assert.NotPanics(t, func() { router.ServeHTTP(response, request) })
		assert.Equal(t, http.StatusNotFound, response.Code)
	}
}

func TestLegacyProviderWebhookRoutesArePresentWhenEnabled(t *testing.T) {
	router := gin.New()
	appCtx := &app.Context{
		LegacyProviderSurfaceEnabled: true,
	}
	decorebatorhttp.RegisterLegacyProviderPublicRoutes(router, appCtx)
	decorebatorhttp.RegisterLegacyProviderAuthenticatedRoutes(router.Group("/"), appCtx)

	paths := make(map[string]bool)
	for _, route := range router.Routes() {
		paths[route.Path] = true
	}
	assert.True(t, paths["/webhook/stripe"])
	assert.True(t, paths["/webhook/revenuecat"])
	assert.True(t, paths["/subscription/checkout-session"])
	assert.True(t, paths["/subscription/revenuecat/restore"])
	assert.True(t, paths["/subscription/checkout-redirect"])
}

func TestStoreWebhookMetricsExposeOnlyBoundedOperationalLabels(t *testing.T) {
	metrics := decorebatorhttp.NewStoreWebhookMetrics()
	metrics.Observe(decorebatorhttp.StoreWebhookObservation{
		Provider: "google", Disposition: "retry", Outcome: model.EntitlementOutcomeRetry,
		ResultCode: model.EntitlementResultRetryableProvider, Duplicate: true,
	})
	router := gin.New()
	router.GET("/metrics", decorebatorhttp.GetStoreWebhookMetrics(metrics))
	request := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"metrics":[{"provider":"google","disposition":"retry","outcome":"retry","resultCode":"retryable_provider_error","duplicate":true,"count":1}]}`, response.Body.String())
	assert.NotContains(t, response.Body.String(), "purchase-token")
	assert.NotContains(t, response.Body.String(), "signedPayload")
}

func TestStoreWebhookErrorKindIsBoundedAndNeverContainsRawErrorText(t *testing.T) {
	tests := []struct {
		err  error
		kind string
	}{
		{&service.AppleNotificationDeliveryError{
			Disposition: service.AppleNotificationAcknowledge,
			Cause:       service.ErrInvalidAppleSignedData,
		}, "invalid_apple_signed_data"},
		{&service.GoogleRTDNDeliveryError{
			Disposition: service.GoogleRTDNUnauthorized,
			Cause:       service.ErrInvalidGooglePubSubAuthentication,
		}, "invalid_google_authentication"},
		{&service.ApplePurchaseVerificationError{
			Failure: model.EntitlementFailureProviderUnavailable,
			Cause:   errors.New("sensitive provider response"),
		}, "provider_unavailable"},
		{errors.New("purchase-token sensitive-database-detail"), "internal_error"},
	}
	for _, test := range tests {
		kind := decorebatorhttp.StoreWebhookErrorKind(test.err)
		assert.Equal(t, test.kind, kind)
		assert.NotContains(t, kind, "purchase-token")
		assert.NotContains(t, kind, "sensitive")
	}
}

func performWebhookRequest(
	t *testing.T,
	handler http.Handler,
	path string,
	body string,
	authorization string,
) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	require.Empty(t, response.Body.String())
	return response
}
