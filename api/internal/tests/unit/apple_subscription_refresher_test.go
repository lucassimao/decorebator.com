package unit

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppleJWSVerifierAcceptsRenewalWithoutOptionalAutoRenewProduct(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	payload := service.AppleRenewalInfoPayload{
		OriginalTransactionID: "200000000000001", ProductID: "decorebator_monthly_premium1",
		AutoRenewStatus: 0, Environment: "Production", SignedDateMS: now.UnixMilli(),
	}
	signed, root := signAppleTestPayloadJWS(t, payload, now, true)
	verifier, err := service.NewAppleJWSVerifier([]*x509.Certificate{root})
	require.NoError(t, err)
	decoded, err := verifier.VerifyAndDecodeRenewalInfo(signed)
	require.NoError(t, err)
	assert.Empty(t, decoded.AutoRenewProductID)
}

type fakeAppleStatusClient struct {
	response      service.AppleSubscriptionStatusResponse
	err           error
	transactionID string
	environment   model.StoreEnvironment
	calls         int
}

func (f *fakeAppleStatusClient) GetSubscriptionStatuses(
	_ context.Context,
	transactionID string,
	environment model.StoreEnvironment,
) (service.AppleSubscriptionStatusResponse, error) {
	f.calls++
	f.transactionID, f.environment = transactionID, environment
	return f.response, f.err
}

func validAppleStatusEvidence(now time.Time, status int32) (*fakeAppleStatusClient, fakeAppleNotificationSignedData, model.AppleAccountIdentity) {
	account := model.AppleAccountIdentity{
		UserID: 42, AppAccountToken: uuid.MustParse("db22924d-8f4b-4a7e-845e-3304f319210e"),
	}
	transaction := validApplePayload(now, account.AppAccountToken)
	renewal := service.AppleRenewalInfoPayload{
		OriginalTransactionID: transaction.OriginalTransactionID,
		ProductID:             transaction.ProductID, AutoRenewProductID: transaction.ProductID,
		AutoRenewStatus: 1, Environment: "Production", SignedDateMS: now.Add(time.Second).UnixMilli(),
		GracePeriodExpiresDateMS: transaction.ExpiresDateMS + int64((3*24*time.Hour)/time.Millisecond),
	}
	client := &fakeAppleStatusClient{response: service.AppleSubscriptionStatusResponse{
		Environment: "Production", BundleID: transaction.BundleID, AppAppleID: 123456789,
		Data: []service.AppleSubscriptionStatusGroup{{
			SubscriptionGroupIdentifier: "premium-group",
			LastTransactions: []service.AppleSubscriptionLastTransaction{{
				OriginalTransactionID: transaction.OriginalTransactionID, Status: status,
				SignedTransactionInfo: "signed-transaction", SignedRenewalInfo: "signed-renewal",
			}},
		}},
	}}
	return client, fakeAppleNotificationSignedData{transaction: transaction, renewal: renewal}, account
}

func newAppleSubscriptionRefresher(
	t *testing.T,
	client service.AppleSubscriptionStatusClient,
	evidence fakeAppleNotificationSignedData,
	now time.Time,
) *service.AppleSubscriptionRefresher {
	t.Helper()
	refresher, err := service.NewAppleSubscriptionRefresher(
		client, evidence, "com.lsimaocosta.decorebator", 123456789,
		map[string]string{
			"decorebator_monthly_premium1": "premium",
			"decorebator_annual_premium":   "premium",
		}, func() time.Time { return now },
	)
	require.NoError(t, err)
	return refresher
}

func TestAppleSubscriptionRefresherNormalizesEveryProviderStatus(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	statuses := map[int32]model.EntitlementStatus{
		1: model.EntitlementStatusActive,
		2: model.EntitlementStatusExpired,
		3: model.EntitlementStatusOnHold,
		4: model.EntitlementStatusGracePeriod,
		5: model.EntitlementStatusRevoked,
	}
	for providerStatus, expected := range statuses {
		t.Run(string(expected), func(t *testing.T) {
			client, evidence, account := validAppleStatusEvidence(now, providerStatus)
			if providerStatus == 5 {
				revoked := now.Add(-time.Minute).UnixMilli()
				reason := int32(1)
				evidence.transaction.RevocationDateMS = &revoked
				evidence.transaction.RevocationReason = &reason
			}
			result, err := newAppleSubscriptionRefresher(t, client, evidence, now).Refresh(
				context.Background(), service.AppleSubscriptionRefreshInput{
					AuthenticatedUserID: 42, Account: account,
					OriginalTransactionID: evidence.transaction.OriginalTransactionID,
					Environment:           model.StoreEnvironmentProduction,
				},
			)
			require.NoError(t, err)
			assert.Equal(t, expected, result.Status)
			assert.Equal(t, "premium", result.Entitlement)
			assert.Equal(t, evidence.transaction.OriginalTransactionID, result.Identity.OriginalTransactionID)
			assert.Equal(t, evidence.renewal.SignedDateMS, result.ProviderSnapshotSignedAt.UnixMilli())
			if providerStatus == 4 {
				assert.Equal(t, evidence.renewal.GracePeriodExpiresDateMS, result.PeriodEnd.UnixMilli())
			}
			if providerStatus == 5 {
				require.NotNil(t, result.RevokedAt)
				assert.Equal(t, "apple_1", *result.RevocationReason)
			}
		})
	}
}

func TestAppleSubscriptionRefresherFailsClosedForMismatchedAndAmbiguousEvidence(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name    string
		mutate  func(*fakeAppleStatusClient, *fakeAppleNotificationSignedData, *model.AppleAccountIdentity)
		failure model.EntitlementOperationFailure
	}{
		{name: "wrong bundle", mutate: func(client *fakeAppleStatusClient, _ *fakeAppleNotificationSignedData, _ *model.AppleAccountIdentity) {
			client.response.BundleID = "attacker.bundle"
		}, failure: model.EntitlementFailureInvalidEvidence},
		{name: "wrong account", mutate: func(_ *fakeAppleStatusClient, evidence *fakeAppleNotificationSignedData, _ *model.AppleAccountIdentity) {
			evidence.transaction.AppAccountToken = uuid.NewString()
		}, failure: model.EntitlementFailureAccountMismatch},
		{name: "ambiguous lineage", mutate: func(client *fakeAppleStatusClient, _ *fakeAppleNotificationSignedData, _ *model.AppleAccountIdentity) {
			duplicate := client.response.Data[0].LastTransactions[0]
			client.response.Data[0].LastTransactions = append(client.response.Data[0].LastTransactions, duplicate)
		}, failure: model.EntitlementFailureInvalidEvidence},
		{name: "invalid auto renew state", mutate: func(_ *fakeAppleStatusClient, evidence *fakeAppleNotificationSignedData, _ *model.AppleAccountIdentity) {
			evidence.renewal.AutoRenewStatus = 2
		}, failure: model.EntitlementFailureInvalidEvidence},
		{name: "unknown current product", mutate: func(_ *fakeAppleStatusClient, evidence *fakeAppleNotificationSignedData, _ *model.AppleAccountIdentity) {
			evidence.renewal.ProductID = "unknown"
		}, failure: model.EntitlementFailureUnknownProduct},
		{name: "invalid transaction JWS", mutate: func(_ *fakeAppleStatusClient, evidence *fakeAppleNotificationSignedData, _ *model.AppleAccountIdentity) {
			evidence.transactErr = service.ErrInvalidAppleSignedData
		}, failure: model.EntitlementFailureInvalidEvidence},
		{name: "invalid renewal JWS", mutate: func(_ *fakeAppleStatusClient, evidence *fakeAppleNotificationSignedData, _ *model.AppleAccountIdentity) {
			evidence.renewalErr = service.ErrInvalidAppleSignedData
		}, failure: model.EntitlementFailureInvalidEvidence},
		{name: "future signed date", mutate: func(_ *fakeAppleStatusClient, evidence *fakeAppleNotificationSignedData, _ *model.AppleAccountIdentity) {
			evidence.renewal.SignedDateMS = now.Add(10 * time.Minute).UnixMilli()
		}, failure: model.EntitlementFailureInvalidEvidence},
		{name: "invalid grace end", mutate: func(client *fakeAppleStatusClient, evidence *fakeAppleNotificationSignedData, _ *model.AppleAccountIdentity) {
			client.response.Data[0].LastTransactions[0].Status = 4
			evidence.renewal.GracePeriodExpiresDateMS = evidence.transaction.ExpiresDateMS
		}, failure: model.EntitlementFailureInvalidEvidence},
	} {
		t.Run(test.name, func(t *testing.T) {
			client, evidence, account := validAppleStatusEvidence(now, 1)
			test.mutate(client, &evidence, &account)
			_, err := newAppleSubscriptionRefresher(t, client, evidence, now).Refresh(context.Background(), service.AppleSubscriptionRefreshInput{
				AuthenticatedUserID: 42, Account: account, OriginalTransactionID: "200000000000001",
				Environment: model.StoreEnvironmentProduction,
			})
			assertAppleVerificationFailure(t, err, test.failure)
		})
	}
}

func TestAppleSubscriptionRefresherPreservesProviderStatusAndHandlesCatalogTransitions(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	client, evidence, account := validAppleStatusEvidence(now, 1)
	evidence.transaction.ExpiresDateMS = now.Add(-time.Hour).UnixMilli()
	evidence.transaction.PurchaseDateMS = now.Add(-31 * 24 * time.Hour).UnixMilli()
	evidence.renewal.AutoRenewProductID = "future-premium-product"
	result, err := newAppleSubscriptionRefresher(t, client, evidence, now).Refresh(context.Background(), service.AppleSubscriptionRefreshInput{
		AuthenticatedUserID: 42, Account: account, OriginalTransactionID: "200000000000001",
		Environment: model.StoreEnvironmentProduction,
	})
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementStatusActive, result.Status)
	assert.False(t, result.PeriodEnd.After(now))

	client, evidence, account = validAppleStatusEvidence(now, 1)
	evidence.renewal.AutoRenewStatus = 0
	evidence.renewal.AutoRenewProductID = ""
	result, err = newAppleSubscriptionRefresher(t, client, evidence, now).Refresh(context.Background(), service.AppleSubscriptionRefreshInput{
		AuthenticatedUserID: 42, Account: account, OriginalTransactionID: "200000000000001",
		Environment: model.StoreEnvironmentProduction,
	})
	require.NoError(t, err)
	assert.False(t, result.AutoRenewEnabled)
	require.NotNil(t, result.CancellationObservedAt)
	assert.Equal(t, evidence.renewal.SignedDateMS, result.CancellationObservedAt.UnixMilli())
}

func TestAppleSubscriptionRefresherRevocationDominatesStatusAndUsesCanonicalReason(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	client, evidence, account := validAppleStatusEvidence(now, 1)
	revoked := now.Add(-time.Minute).UnixMilli()
	reason := int32(99)
	evidence.transaction.RevocationDateMS = &revoked
	evidence.transaction.RevocationReason = &reason
	result, err := newAppleSubscriptionRefresher(t, client, evidence, now).Refresh(context.Background(), service.AppleSubscriptionRefreshInput{
		AuthenticatedUserID: 42, Account: account, OriginalTransactionID: "200000000000001",
		Environment: model.StoreEnvironmentProduction,
	})
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementStatusRevoked, result.Status)
	require.NotNil(t, result.RevocationReason)
	assert.Equal(t, "apple_other", *result.RevocationReason)

	client, evidence, account = validAppleStatusEvidence(now, 5)
	result, err = newAppleSubscriptionRefresher(t, client, evidence, now).Refresh(context.Background(), service.AppleSubscriptionRefreshInput{
		AuthenticatedUserID: 42, Account: account, OriginalTransactionID: "200000000000001",
		Environment: model.StoreEnvironmentProduction,
	})
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementStatusRevoked, result.Status)
	require.NotNil(t, result.RevokedAt)
	assert.Equal(t, evidence.transaction.SignedDateMS, result.RevokedAt.UnixMilli())
}

func TestAppleSubscriptionRefresherRetriesUnknownFutureStatus(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	client, evidence, account := validAppleStatusEvidence(now, 6)
	_, err := newAppleSubscriptionRefresher(t, client, evidence, now).Refresh(context.Background(), service.AppleSubscriptionRefreshInput{
		AuthenticatedUserID: 42, Account: account, OriginalTransactionID: "200000000000001",
		Environment: model.StoreEnvironmentProduction,
	})
	assertAppleVerificationFailure(t, err, model.EntitlementFailureProviderUnavailable)
}

func TestAppleSubscriptionRefresherAcceptsBoundSandboxAndRejectsCrossEnvironment(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	client, evidence, account := validAppleStatusEvidence(now, 1)
	client.response.Environment, client.response.AppAppleID = "Sandbox", 0
	evidence.transaction.Environment, evidence.renewal.Environment = "Sandbox", "Sandbox"
	result, err := newAppleSubscriptionRefresher(t, client, evidence, now).Refresh(context.Background(), service.AppleSubscriptionRefreshInput{
		AuthenticatedUserID: 42, Account: account, OriginalTransactionID: "200000000000001",
		Environment: model.StoreEnvironmentSandbox,
	})
	require.NoError(t, err)
	assert.Equal(t, model.StoreEnvironmentSandbox, result.Identity.Environment)
	assert.Equal(t, "200000000000001", client.transactionID)
	assert.Equal(t, model.StoreEnvironmentSandbox, client.environment)

	_, err = newAppleSubscriptionRefresher(t, client, evidence, now).Refresh(context.Background(), service.AppleSubscriptionRefreshInput{
		AuthenticatedUserID: 42, Account: account, OriginalTransactionID: "200000000000001",
		Environment: model.StoreEnvironmentProduction,
	})
	assertAppleVerificationFailure(t, err, model.EntitlementFailureInvalidEvidence)
}

func TestAppleSubscriptionRefresherClassifiesProviderFailureAndNeverFallsBackEnvironment(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	client, evidence, account := validAppleStatusEvidence(now, 1)
	client.err = &service.AppleAPIError{StatusCode: http.StatusServiceUnavailable, Retryable: true}
	_, err := newAppleSubscriptionRefresher(t, client, evidence, now).Refresh(context.Background(), service.AppleSubscriptionRefreshInput{
		AuthenticatedUserID: 42, Account: account, OriginalTransactionID: "200000000000001",
		Environment: model.StoreEnvironmentProduction,
	})
	assertAppleVerificationFailure(t, err, model.EntitlementFailureProviderUnavailable)
	assert.Equal(t, 1, client.calls)
	assert.Equal(t, model.StoreEnvironmentProduction, client.environment)

	client.err = errors.New("transport")
	_, err = newAppleSubscriptionRefresher(t, client, evidence, now).Refresh(context.Background(), service.AppleSubscriptionRefreshInput{
		AuthenticatedUserID: 42, Account: account, OriginalTransactionID: "200000000000001",
		Environment: model.StoreEnvironmentProduction,
	})
	assertAppleVerificationFailure(t, err, model.EntitlementFailureProviderUnavailable)
}

func TestAppleAppStoreClientFetchesBoundedSubscriptionStatuses(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/inApps/v1/subscriptions/200000000000001", request.URL.Path)
		assert.Contains(t, request.Header.Get("Authorization"), "Bearer ")
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(service.AppleSubscriptionStatusResponse{
			Environment: "Production", BundleID: "com.lsimaocosta.decorebator", AppAppleID: 123456789,
			Data: []service.AppleSubscriptionStatusGroup{},
		})
	}))
	defer server.Close()
	client, err := service.NewAppleAppStoreClient(service.AppleAppStoreClientConfig{
		KeyID: "key", IssuerID: "issuer", BundleID: "com.lsimaocosta.decorebator",
		PrivateKey: encodePKCS8PrivateKey(t, privateKey), HTTPClient: server.Client(), Now: func() time.Time { return now },
		BaseURLs: map[model.StoreEnvironment]string{model.StoreEnvironmentProduction: server.URL},
	})
	require.NoError(t, err)
	result, err := client.GetSubscriptionStatuses(context.Background(), "200000000000001", model.StoreEnvironmentProduction)
	require.NoError(t, err)
	assert.Equal(t, int64(123456789), result.AppAppleID)
}

func TestAppleAppStoreClientClassifiesStatusErrorsAndBoundsBody(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	responses := [][]byte{
		[]byte(`{"errorCode":4040002,"errorMessage":"sensitive provider detail"}`),
		bytes.Repeat([]byte("x"), (1<<20)+1),
	}
	statusCodes := []int{http.StatusNotFound, http.StatusOK}
	for index := range responses {
		server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(statusCodes[index])
			_, _ = response.Write(responses[index])
		}))
		client, createErr := service.NewAppleAppStoreClient(service.AppleAppStoreClientConfig{
			KeyID: "key", IssuerID: "issuer", BundleID: "bundle", PrivateKey: encodePKCS8PrivateKey(t, privateKey),
			HTTPClient: server.Client(), BaseURLs: map[model.StoreEnvironment]string{model.StoreEnvironmentProduction: server.URL},
		})
		require.NoError(t, createErr)
		_, requestErr := client.GetSubscriptionStatuses(context.Background(), "sensitive-transaction", model.StoreEnvironmentProduction)
		require.Error(t, requestErr)
		assert.NotContains(t, requestErr.Error(), "sensitive")
		if index == 0 {
			var apiError *service.AppleAPIError
			require.ErrorAs(t, requestErr, &apiError)
			assert.True(t, apiError.Retryable)
		} else {
			assert.ErrorContains(t, requestErr, "exceeds limit")
		}
		server.Close()
	}
}
