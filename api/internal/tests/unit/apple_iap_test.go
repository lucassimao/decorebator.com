package unit

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAppleTransactionClient struct {
	responses map[model.StoreEnvironment]string
	errors    map[model.StoreEnvironment]error
	calls     []model.StoreEnvironment
}

func (f *fakeAppleTransactionClient) GetTransactionInfo(
	_ context.Context,
	_ string,
	environment model.StoreEnvironment,
) (string, error) {
	f.calls = append(f.calls, environment)
	return f.responses[environment], f.errors[environment]
}

type fakeAppleSignedDataVerifier struct {
	payload service.AppleTransactionPayload
	err     error
}

func (f fakeAppleSignedDataVerifier) VerifyAndDecodeTransaction(string) (service.AppleTransactionPayload, error) {
	return f.payload, f.err
}

func validApplePayload(now time.Time, accountToken uuid.UUID) service.AppleTransactionPayload {
	return service.AppleTransactionPayload{
		OriginalTransactionID: "200000000000001",
		TransactionID:         "200000000000002",
		BundleID:              "com.lsimaocosta.decorebator",
		ProductID:             "decorebator_monthly_premium1",
		PurchaseDateMS:        now.Add(-time.Hour).UnixMilli(),
		ExpiresDateMS:         now.Add(29 * 24 * time.Hour).UnixMilli(),
		ProductType:           "Auto-Renewable Subscription",
		AppAccountToken:       accountToken.String(),
		Environment:           "Production",
		SignedDateMS:          now.UnixMilli(),
	}
}

func newApplePurchaseVerifier(
	t *testing.T,
	client service.AppleTransactionInfoClient,
	payload service.AppleTransactionPayload,
	now time.Time,
) *service.ApplePurchaseVerifier {
	t.Helper()
	verifier, err := service.NewApplePurchaseVerifier(
		client,
		fakeAppleSignedDataVerifier{payload: payload},
		"com.lsimaocosta.decorebator",
		map[string]string{
			"decorebator_monthly_premium1": "premium",
			"decorebator_annual_premium":   "premium",
		},
		func() time.Time { return now },
	)
	require.NoError(t, err)
	return verifier
}

func TestApplePurchaseVerifierAcceptsAccountBoundProductionPurchase(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	accountToken := uuid.MustParse("db22924d-8f4b-4a7e-845e-3304f319210e")
	client := &fakeAppleTransactionClient{responses: map[model.StoreEnvironment]string{
		model.StoreEnvironmentProduction: "signed-production-transaction",
	}, errors: map[model.StoreEnvironment]error{}}
	verifier := newApplePurchaseVerifier(t, client, validApplePayload(now, accountToken), now)

	result, err := verifier.Verify(context.Background(), service.ApplePurchaseVerificationInput{
		AuthenticatedUserID: 42,
		TransactionID:       "200000000000002",
		Account: model.AppleAccountIdentity{
			UserID: 42, AppAccountToken: accountToken,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, []model.StoreEnvironment{model.StoreEnvironmentProduction}, client.calls)
	assert.Equal(t, "premium", result.Entitlement)
	assert.Equal(t, now.Add(29*24*time.Hour), result.ExpiresDate)
	assert.Equal(t, "200000000000001", result.Identity.OriginalTransactionID)
	assert.Equal(t, model.StoreEnvironmentProduction, result.Identity.Environment)
	assert.True(t, result.Identity.MatchesAccount(model.AppleAccountIdentity{UserID: 42, AppAccountToken: accountToken}))
}

func TestApplePurchaseVerifierFallsBackToSandboxOnlyForDocumentedNotFound(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	accountToken := uuid.New()
	payload := validApplePayload(now, accountToken)
	payload.Environment = "Sandbox"
	client := &fakeAppleTransactionClient{
		responses: map[model.StoreEnvironment]string{model.StoreEnvironmentSandbox: "signed-sandbox-transaction"},
		errors: map[model.StoreEnvironment]error{
			model.StoreEnvironmentProduction: &service.AppleAPIError{StatusCode: http.StatusNotFound, ErrorCode: 4040010},
		},
	}
	verifier := newApplePurchaseVerifier(t, client, payload, now)
	result, err := verifier.Verify(context.Background(), service.ApplePurchaseVerificationInput{
		AuthenticatedUserID: 42, TransactionID: payload.TransactionID,
		Account: model.AppleAccountIdentity{UserID: 42, AppAccountToken: accountToken},
	})
	require.NoError(t, err)
	assert.Equal(t, model.StoreEnvironmentSandbox, result.Identity.Environment)
	assert.Equal(t, []model.StoreEnvironment{model.StoreEnvironmentProduction, model.StoreEnvironmentSandbox}, client.calls)

	client = &fakeAppleTransactionClient{responses: map[model.StoreEnvironment]string{}, errors: map[model.StoreEnvironment]error{
		model.StoreEnvironmentProduction: &service.AppleAPIError{StatusCode: http.StatusTooManyRequests, ErrorCode: 4290000, Retryable: true},
	}}
	verifier = newApplePurchaseVerifier(t, client, payload, now)
	_, err = verifier.Verify(context.Background(), service.ApplePurchaseVerificationInput{
		AuthenticatedUserID: 42, TransactionID: payload.TransactionID,
		Account: model.AppleAccountIdentity{UserID: 42, AppAccountToken: accountToken},
	})
	assertAppleVerificationFailure(t, err, model.EntitlementFailureProviderUnavailable)
	assert.Equal(t, []model.StoreEnvironment{model.StoreEnvironmentProduction}, client.calls)
}

func TestApplePurchaseVerifierFailsClosedForInvalidProviderEvidence(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	accountToken := uuid.New()
	valid := validApplePayload(now, accountToken)
	mutations := []struct {
		name    string
		failure model.EntitlementOperationFailure
		mutate  func(*service.AppleTransactionPayload)
	}{
		{"wrong bundle", model.EntitlementFailureInvalidEvidence, func(p *service.AppleTransactionPayload) { p.BundleID = "attacker.app" }},
		{"wrong environment", model.EntitlementFailureInvalidEvidence, func(p *service.AppleTransactionPayload) { p.Environment = "Sandbox" }},
		{"wrong transaction", model.EntitlementFailureInvalidEvidence, func(p *service.AppleTransactionPayload) { p.TransactionID = "other" }},
		{"missing original transaction", model.EntitlementFailureInvalidEvidence, func(p *service.AppleTransactionPayload) { p.OriginalTransactionID = "" }},
		{"wrong product type", model.EntitlementFailureInvalidEvidence, func(p *service.AppleTransactionPayload) { p.ProductType = "Consumable" }},
		{"unknown product", model.EntitlementFailureUnknownProduct, func(p *service.AppleTransactionPayload) { p.ProductID = "attacker.product" }},
		{"wrong account token", model.EntitlementFailureAccountMismatch, func(p *service.AppleTransactionPayload) { p.AppAccountToken = uuid.NewString() }},
		{"missing account token", model.EntitlementFailureAccountMismatch, func(p *service.AppleTransactionPayload) { p.AppAccountToken = "" }},
		{"invalid period", model.EntitlementFailureInvalidEvidence, func(p *service.AppleTransactionPayload) { p.ExpiresDateMS = p.PurchaseDateMS }},
	}

	for _, test := range mutations {
		t.Run(test.name, func(t *testing.T) {
			payload := valid
			test.mutate(&payload)
			client := &fakeAppleTransactionClient{responses: map[model.StoreEnvironment]string{
				model.StoreEnvironmentProduction: "signed",
			}, errors: map[model.StoreEnvironment]error{}}
			verifier := newApplePurchaseVerifier(t, client, payload, now)
			_, err := verifier.Verify(context.Background(), service.ApplePurchaseVerificationInput{
				AuthenticatedUserID: 42, TransactionID: valid.TransactionID,
				Account: model.AppleAccountIdentity{UserID: 42, AppAccountToken: accountToken},
			})
			assertAppleVerificationFailure(t, err, test.failure)
		})
	}
}

func TestApplePurchaseVerifierReturnsRawVerifiedExpiryAndRevocationEvidence(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	accountToken := uuid.New()
	payload := validApplePayload(now, accountToken)
	payload.ExpiresDateMS = now.Add(-time.Minute).UnixMilli()
	payload.PurchaseDateMS = now.Add(-31 * 24 * time.Hour).UnixMilli()
	client := &fakeAppleTransactionClient{responses: map[model.StoreEnvironment]string{
		model.StoreEnvironmentProduction: "signed",
	}, errors: map[model.StoreEnvironment]error{}}
	verifier := newApplePurchaseVerifier(t, client, payload, now)
	input := service.ApplePurchaseVerificationInput{
		AuthenticatedUserID: 42, TransactionID: payload.TransactionID,
		Account: model.AppleAccountIdentity{UserID: 42, AppAccountToken: accountToken},
	}
	result, err := verifier.Verify(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, time.UnixMilli(payload.ExpiresDateMS).UTC(), result.ExpiresDate)
	assert.Nil(t, result.RevokedAt)

	revokedAt := now.Add(-time.Hour).UnixMilli()
	reason := int32(1)
	payload.RevocationDateMS, payload.RevocationReason = &revokedAt, &reason
	verifier = newApplePurchaseVerifier(t, client, payload, now)
	result, err = verifier.Verify(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, result.RevokedAt)
	require.NotNil(t, result.RevocationReason)
	assert.Equal(t, "apple_1", *result.RevocationReason)
}

func TestApplePurchaseVerifierRejectsMalformedTransactionIDsBeforeProviderCall(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	accountToken := uuid.New()
	for _, transactionID := range []string{"", " 200000000000002", "2000 0002", "abc/def", "abc\tdef"} {
		t.Run(fmt.Sprintf("id_%q", transactionID), func(t *testing.T) {
			client := &fakeAppleTransactionClient{responses: map[model.StoreEnvironment]string{}, errors: map[model.StoreEnvironment]error{}}
			verifier := newApplePurchaseVerifier(t, client, validApplePayload(now, accountToken), now)
			_, err := verifier.Verify(context.Background(), service.ApplePurchaseVerificationInput{
				AuthenticatedUserID: 42, TransactionID: transactionID,
				Account: model.AppleAccountIdentity{UserID: 42, AppAccountToken: accountToken},
			})
			assertAppleVerificationFailure(t, err, model.EntitlementFailureInvalidEvidence)
			assert.Empty(t, client.calls)
		})
	}
}

func TestApplePurchaseVerifierTreatsDoubleNotFoundAsInvalidEvidence(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	accountToken := uuid.New()
	client := &fakeAppleTransactionClient{responses: map[model.StoreEnvironment]string{}, errors: map[model.StoreEnvironment]error{
		model.StoreEnvironmentProduction: &service.AppleAPIError{StatusCode: http.StatusNotFound, ErrorCode: 4040010},
		model.StoreEnvironmentSandbox:    &service.AppleAPIError{StatusCode: http.StatusNotFound, ErrorCode: 4040010},
	}}
	verifier := newApplePurchaseVerifier(t, client, validApplePayload(now, accountToken), now)
	_, err := verifier.Verify(context.Background(), service.ApplePurchaseVerificationInput{
		AuthenticatedUserID: 42, TransactionID: "200000000000002",
		Account: model.AppleAccountIdentity{UserID: 42, AppAccountToken: accountToken},
	})
	assertAppleVerificationFailure(t, err, model.EntitlementFailureInvalidEvidence)
	assert.Equal(t, []model.StoreEnvironment{model.StoreEnvironmentProduction, model.StoreEnvironmentSandbox}, client.calls)
}

func TestApplePurchaseVerifierRejectsInvalidSignatureAndAuthenticatedAccount(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	accountToken := uuid.New()
	client := &fakeAppleTransactionClient{responses: map[model.StoreEnvironment]string{
		model.StoreEnvironmentProduction: "signed",
	}, errors: map[model.StoreEnvironment]error{}}
	verifier, err := service.NewApplePurchaseVerifier(
		client,
		fakeAppleSignedDataVerifier{err: service.ErrInvalidAppleSignedData},
		"com.lsimaocosta.decorebator",
		map[string]string{"decorebator_monthly_premium1": "premium"},
		func() time.Time { return now },
	)
	require.NoError(t, err)
	_, err = verifier.Verify(context.Background(), service.ApplePurchaseVerificationInput{
		AuthenticatedUserID: 42, TransactionID: "200000000000002",
		Account: model.AppleAccountIdentity{UserID: 42, AppAccountToken: accountToken},
	})
	assertAppleVerificationFailure(t, err, model.EntitlementFailureInvalidEvidence)

	verifier = newApplePurchaseVerifier(t, client, validApplePayload(now, accountToken), now)
	_, err = verifier.Verify(context.Background(), service.ApplePurchaseVerificationInput{
		AuthenticatedUserID: 99, TransactionID: "200000000000002",
		Account: model.AppleAccountIdentity{UserID: 42, AppAccountToken: accountToken},
	})
	assertAppleVerificationFailure(t, err, model.EntitlementFailureAccountMismatch)
}

func TestAppleJWSVerifierChecksProfileChainAndSignature(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	accountToken := uuid.New()
	payload := validApplePayload(now, accountToken)
	signed, root := signAppleTestJWS(t, payload, now, true)
	verifier, err := service.NewAppleJWSVerifier([]*x509.Certificate{root})
	require.NoError(t, err)

	decoded, err := verifier.VerifyAndDecodeTransaction(signed)
	require.NoError(t, err)
	assert.Equal(t, payload.TransactionID, decoded.TransactionID)
	assert.Equal(t, payload.AppAccountToken, decoded.AppAccountToken)

	parts := strings.Split(signed, ".")
	require.Len(t, parts, 3)
	parts[1] = base64.RawURLEncoding.EncodeToString([]byte(`{"transactionId":"tampered","signedDate":1}`))
	_, err = verifier.VerifyAndDecodeTransaction(parts[0] + "." + parts[1] + "." + parts[2])
	require.ErrorIs(t, err, service.ErrInvalidAppleSignedData)

	signedWithoutProfile, root := signAppleTestJWS(t, payload, now, false)
	verifier, err = service.NewAppleJWSVerifier([]*x509.Certificate{root})
	require.NoError(t, err)
	_, err = verifier.VerifyAndDecodeTransaction(signedWithoutProfile)
	require.ErrorIs(t, err, service.ErrInvalidAppleSignedData)
}

func TestAppleJWSVerifierRejectsUntrustedRootAlgorithmsAndInvalidSigningDate(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	payload := validApplePayload(now, uuid.New())
	signed, _ := signAppleTestJWS(t, payload, now, true)
	_, unrelatedRoot := signAppleTestJWS(t, payload, now, true)
	verifier, err := service.NewAppleJWSVerifier([]*x509.Certificate{unrelatedRoot})
	require.NoError(t, err)
	_, err = verifier.VerifyAndDecodeTransaction(signed)
	require.ErrorIs(t, err, service.ErrInvalidAppleSignedData)

	algorithmSigned, trustedRoot := signAppleTestJWS(t, payload, now, true)
	verifier, err = service.NewAppleJWSVerifier([]*x509.Certificate{trustedRoot})
	require.NoError(t, err)
	for _, algorithm := range []string{"none", "RS256"} {
		_, err = verifier.VerifyAndDecodeTransaction(replaceJWSAlgorithm(t, algorithmSigned, algorithm))
		require.ErrorIs(t, err, service.ErrInvalidAppleSignedData)
	}

	payload.SignedDateMS = now.Add(2 * time.Hour).UnixMilli()
	signedOutsideValidity, trustedRoot := signAppleTestJWS(t, payload, now, true)
	verifier, err = service.NewAppleJWSVerifier([]*x509.Certificate{trustedRoot})
	require.NoError(t, err)
	_, err = verifier.VerifyAndDecodeTransaction(signedOutsideValidity)
	require.ErrorIs(t, err, service.ErrInvalidAppleSignedData)
}

func TestAppleAppStoreClientSignsBoundedAuthenticatedRequest(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/inApps/v1/transactions/200000000000002", request.URL.Path)
		authorization := request.Header.Get("Authorization")
		require.Contains(t, authorization, "Bearer ")
		token, parseErr := jwt.Parse(strings.TrimPrefix(authorization, "Bearer "), func(token *jwt.Token) (any, error) {
			assert.Equal(t, "APPLEKEY1", token.Header["kid"])
			return &privateKey.PublicKey, nil
		}, jwt.WithAudience("appstoreconnect-v1"), jwt.WithIssuer("issuer-id"), jwt.WithTimeFunc(func() time.Time { return now }))
		require.NoError(t, parseErr)
		assert.True(t, token.Valid)
		claims, ok := token.Claims.(jwt.MapClaims)
		require.True(t, ok)
		assert.Equal(t, "com.lsimaocosta.decorebator", claims["bid"])
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"signedTransactionInfo":"signed-jws"}`))
	}))
	defer server.Close()

	client, err := service.NewAppleAppStoreClient(service.AppleAppStoreClientConfig{
		KeyID: "APPLEKEY1", IssuerID: "issuer-id", BundleID: "com.lsimaocosta.decorebator",
		PrivateKey: encodePKCS8PrivateKey(t, privateKey), HTTPClient: server.Client(),
		Now: func() time.Time { return now },
		BaseURLs: map[model.StoreEnvironment]string{
			model.StoreEnvironmentProduction: server.URL,
		},
	})
	require.NoError(t, err)
	result, err := client.GetTransactionInfo(context.Background(), "200000000000002", model.StoreEnvironmentProduction)
	require.NoError(t, err)
	assert.Equal(t, "signed-jws", result)
}

func TestAppleAppStoreClientClassifiesSafeErrors(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNotFound)
		_, _ = response.Write([]byte(`{"errorCode":4040010,"errorMessage":"transaction details"}`))
	}))
	defer server.Close()
	client, err := service.NewAppleAppStoreClient(service.AppleAppStoreClientConfig{
		KeyID: "key", IssuerID: "issuer", BundleID: "bundle",
		PrivateKey: encodePKCS8PrivateKey(t, privateKey), HTTPClient: server.Client(),
		BaseURLs: map[model.StoreEnvironment]string{model.StoreEnvironmentProduction: server.URL},
	})
	require.NoError(t, err)
	_, err = client.GetTransactionInfo(context.Background(), "sensitive-transaction", model.StoreEnvironmentProduction)
	var apiError *service.AppleAPIError
	require.ErrorAs(t, err, &apiError)
	assert.True(t, apiError.TransactionNotFound())
	assert.False(t, apiError.Retryable)
	assert.NotContains(t, err.Error(), "sensitive-transaction")
	assert.NotContains(t, err.Error(), "transaction details")

	unauthorizedServer := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusUnauthorized)
	}))
	defer unauthorizedServer.Close()
	client, err = service.NewAppleAppStoreClient(service.AppleAppStoreClientConfig{
		KeyID: "key", IssuerID: "issuer", BundleID: "bundle",
		PrivateKey: encodePKCS8PrivateKey(t, privateKey), HTTPClient: unauthorizedServer.Client(),
		BaseURLs: map[model.StoreEnvironment]string{model.StoreEnvironmentProduction: unauthorizedServer.URL},
	})
	require.NoError(t, err)
	_, err = client.GetTransactionInfo(context.Background(), "transaction", model.StoreEnvironmentProduction)
	require.ErrorAs(t, err, &apiError)
	assert.True(t, apiError.Retryable)
}

func TestAppleAppStoreClientUsesDocumentedRetryableErrorCodes(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	for _, errorCode := range []int64{4040002, 4040004, 4040006, 5000001} {
		t.Run(fmt.Sprintf("code_%d", errorCode), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
				response.WriteHeader(http.StatusNotFound)
				_, _ = fmt.Fprintf(response, `{"errorCode":%d}`, errorCode)
			}))
			defer server.Close()
			client, err := service.NewAppleAppStoreClient(service.AppleAppStoreClientConfig{
				KeyID: "key", IssuerID: "issuer", BundleID: "bundle",
				PrivateKey: encodePKCS8PrivateKey(t, privateKey), HTTPClient: server.Client(),
				BaseURLs: map[model.StoreEnvironment]string{model.StoreEnvironmentProduction: server.URL},
			})
			require.NoError(t, err)
			_, err = client.GetTransactionInfo(context.Background(), "transaction", model.StoreEnvironmentProduction)
			var apiError *service.AppleAPIError
			require.ErrorAs(t, err, &apiError)
			assert.True(t, apiError.Retryable)
		})
	}
}

func assertAppleVerificationFailure(t *testing.T, err error, failure model.EntitlementOperationFailure) {
	t.Helper()
	var verificationError *service.ApplePurchaseVerificationError
	require.ErrorAs(t, err, &verificationError)
	assert.Equal(t, failure, verificationError.Failure)
}

func signAppleTestJWS(
	t *testing.T,
	payload service.AppleTransactionPayload,
	now time.Time,
	includeProfileOIDs bool,
) (string, *x509.Certificate) {
	return signAppleTestPayloadJWS(t, payload, now, includeProfileOIDs)
}

func signAppleTestPayloadJWS(
	t *testing.T,
	payload any,
	now time.Time,
	includeProfileOIDs bool,
) (string, *x509.Certificate) {
	t.Helper()
	rootKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	intermediateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	rootTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Test Apple Root"},
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour), IsCA: true,
		BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	rootDER, err := x509.CreateCertificate(rand.Reader, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	require.NoError(t, err)
	root, err := x509.ParseCertificate(rootDER)
	require.NoError(t, err)

	intermediateTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: "Test Apple Intermediate"},
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour), IsCA: true,
		BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	leafTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3), Subject: pkix.Name{CommonName: "Test Apple Leaf"},
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour),
		KeyUsage: x509.KeyUsageDigitalSignature,
	}
	if includeProfileOIDs {
		intermediateTemplate.ExtraExtensions = []pkix.Extension{{Id: []int{1, 2, 840, 113635, 100, 6, 2, 1}, Value: []byte{5, 0}}}
		leafTemplate.ExtraExtensions = []pkix.Extension{{Id: []int{1, 2, 840, 113635, 100, 6, 11, 1}, Value: []byte{5, 0}}}
	}
	intermediateDER, err := x509.CreateCertificate(rand.Reader, intermediateTemplate, root, &intermediateKey.PublicKey, rootKey)
	require.NoError(t, err)
	intermediate, err := x509.ParseCertificate(intermediateDER)
	require.NoError(t, err)
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTemplate, intermediate, &leafKey.PublicKey, intermediateKey)
	require.NoError(t, err)

	headerBytes, err := json.Marshal(map[string]any{
		"alg": "ES256",
		"x5c": []string{
			base64.StdEncoding.EncodeToString(leafDER),
			base64.StdEncoding.EncodeToString(intermediateDER),
			base64.StdEncoding.EncodeToString(rootDER),
		},
	})
	require.NoError(t, err)
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)
	header := base64.RawURLEncoding.EncodeToString(headerBytes)
	body := base64.RawURLEncoding.EncodeToString(payloadBytes)
	digest := sha256.Sum256([]byte(header + "." + body))
	r, s, err := ecdsa.Sign(rand.Reader, leafKey, digest[:])
	require.NoError(t, err)
	signature := make([]byte, 64)
	r.FillBytes(signature[:32])
	s.FillBytes(signature[32:])
	return header + "." + body + "." + base64.RawURLEncoding.EncodeToString(signature), root
}

func encodePKCS8PrivateKey(t *testing.T, privateKey *ecdsa.PrivateKey) []byte {
	t.Helper()
	encoded, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: encoded})
}

func replaceJWSAlgorithm(t *testing.T, signed, algorithm string) string {
	t.Helper()
	parts := strings.Split(signed, ".")
	require.Len(t, parts, 3)
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	require.NoError(t, err)
	var header map[string]any
	require.NoError(t, json.Unmarshal(headerBytes, &header))
	header["alg"] = algorithm
	headerBytes, err = json.Marshal(header)
	require.NoError(t, err)
	parts[0] = base64.RawURLEncoding.EncodeToString(headerBytes)
	return strings.Join(parts, ".")
}

func ExampleAppleAPIError_Error() {
	err := &service.AppleAPIError{StatusCode: 429, ErrorCode: 4290000, Retryable: true}
	fmt.Println(err)
	// Output: Apple App Store API request failed (status=429 code=4290000)
}
