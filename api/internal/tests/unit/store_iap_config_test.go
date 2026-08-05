package unit

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"decorebator.com/internal/config"
	"decorebator.com/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreIAPConfigIsExplicitlyDisabledByDefault(t *testing.T) {
	loaded, err := config.LoadStoreIAPConfigFrom(mapLookup(nil))
	require.NoError(t, err)
	assert.False(t, loaded.Enabled)

	loaded, err = config.LoadStoreIAPConfigFrom(mapLookup(map[string]string{"STORE_IAP_ENABLED": "false"}))
	require.NoError(t, err)
	assert.False(t, loaded.Enabled)
}

func TestStoreIAPConfigRequiresExplicitProductionDecision(t *testing.T) {
	_, err := config.LoadStoreIAPConfigFrom(mapLookup(map[string]string{"ENV": "production"}))
	require.ErrorContains(t, err, "explicitly configured in production")

	loaded, err := config.LoadStoreIAPConfigFrom(mapLookup(map[string]string{
		"ENV": "production", "STORE_IAP_ENABLED": "false",
	}))
	require.NoError(t, err)
	assert.False(t, loaded.Enabled)
}

func TestLegacyProviderSurfaceDefaultsEnabledAndRequiresCompleteConfig(t *testing.T) {
	_, err := config.LoadLegacyProviderConfigFrom(mapLookup(nil))
	require.ErrorContains(t, err, "STRIPE_API_KEY")

	loaded, err := config.LoadLegacyProviderConfigFrom(mapLookup(map[string]string{
		"LEGACY_PROVIDER_WEBHOOKS_ENABLED": "false",
	}))
	require.NoError(t, err)
	assert.False(t, loaded.Enabled)
}

func TestLegacyProviderWebhooksRequireExplicitProductionDecision(t *testing.T) {
	_, err := config.LoadLegacyProviderConfigFrom(mapLookup(map[string]string{"ENV": "production"}))
	require.ErrorContains(t, err, "explicitly configured in production")

	loaded, err := config.LoadLegacyProviderConfigFrom(mapLookup(map[string]string{
		"ENV": "production", "LEGACY_PROVIDER_WEBHOOKS_ENABLED": "false",
	}))
	require.NoError(t, err)
	assert.False(t, loaded.Enabled)

	_, err = config.LoadLegacyProviderConfigFrom(mapLookup(map[string]string{
		"ENV": "production", "LEGACY_PROVIDER_WEBHOOKS_ENABLED": "sometimes",
	}))
	require.ErrorContains(t, err, "must be a boolean")
}

func TestLegacyProviderConfigRequiresEverySecretWithoutLeakingValues(t *testing.T) {
	values := validLegacyProviderEnvironment()
	loaded, err := config.LoadLegacyProviderConfigFrom(mapLookup(values))
	require.NoError(t, err)
	require.True(t, loaded.Enabled)
	assert.Equal(t, values["STRIPE_PRICE_ID_MONTHLY"], loaded.Stripe.MonthlyPriceID)

	for _, name := range []string{
		"STRIPE_API_KEY", "STRIPE_WEBHOOK_SECRET", "STRIPE_SUCCESS_URL", "STRIPE_CANCEL_URL",
		"STRIPE_PRICE_ID_MONTHLY", "STRIPE_PRICE_ID_ANNUAL", "REVENUECAT_API_KEY",
		"REVENUECAT_WEBHOOK_AUTHORIZATION",
	} {
		invalid := validLegacyProviderEnvironment()
		secret := invalid[name]
		invalid[name] = "  "
		_, err := config.LoadLegacyProviderConfigFrom(mapLookup(invalid))
		require.ErrorContains(t, err, name)
		assert.NotContains(t, err.Error(), secret)
	}
}

func validLegacyProviderEnvironment() map[string]string {
	return map[string]string{
		"LEGACY_PROVIDER_WEBHOOKS_ENABLED": "true",
		"STRIPE_API_KEY":                   "stripe-secret", "STRIPE_WEBHOOK_SECRET": "webhook-secret",
		"STRIPE_SUCCESS_URL": "https://example.com/success", "STRIPE_CANCEL_URL": "https://example.com/cancel",
		"STRIPE_PRICE_ID_MONTHLY": "price-monthly", "STRIPE_PRICE_ID_ANNUAL": "price-annual",
		"REVENUECAT_API_KEY": "revenuecat-secret", "REVENUECAT_WEBHOOK_AUTHORIZATION": "revenuecat-webhook-secret",
	}
}

func TestStoreIAPConfigLoadsValidatedSecretsAndBindings(t *testing.T) {
	values := validStoreIAPEnvironment(t)
	loaded, err := config.LoadStoreIAPConfigFrom(mapLookup(values))
	require.NoError(t, err)
	require.True(t, loaded.Enabled)
	require.Equal(t, model.StoreEnvironmentSandbox, loaded.ClientEnvironment)
	require.Equal(t, model.IAPBillingPeriodMonthly, loaded.Apple.Products["com.example.premium.monthly"].BillingPeriod)
	assert.Equal(t, int16(1), loaded.Evidence.ActiveVersion)
	assert.Len(t, loaded.Evidence.LookupKey, 32)
	assert.Equal(t, 10, loaded.Google.MaxDeliveryAttempts)
	assert.Equal(t, "premium", loaded.Apple.Products["com.example.premium.monthly"].Entitlement)
	assert.Equal(t, "premium", loaded.Google.Products["premium-monthly"].Entitlement)
}

func TestStoreIAPConfigPinsProductionEnvironmentAndRequiresRedis(t *testing.T) {
	values := validStoreIAPEnvironment(t)
	values["ENV"] = "production"
	_, err := config.LoadStoreIAPConfigFrom(mapLookup(values))
	require.ErrorContains(t, err, "STORE_IAP_CLIENT_ENVIRONMENT")

	values["STORE_IAP_CLIENT_ENVIRONMENT"] = "sandbox"
	_, err = config.LoadStoreIAPConfigFrom(mapLookup(values))
	require.ErrorContains(t, err, "production")

	values["STORE_IAP_CLIENT_ENVIRONMENT"] = "production"
	_, err = config.LoadStoreIAPConfigFrom(mapLookup(values))
	require.ErrorContains(t, err, "REDIS_HOST")

	values["REDIS_HOST"] = "redis.example.invalid"
	loaded, err := config.LoadStoreIAPConfigFrom(mapLookup(values))
	require.NoError(t, err)
	require.Equal(t, model.StoreEnvironmentProduction, loaded.ClientEnvironment)
}

func TestStoreIAPConfigFailsClosedWithoutLeakingSecretValues(t *testing.T) {
	values := validStoreIAPEnvironment(t)
	secret := values["STORE_EVIDENCE_LOOKUP_KEY_BASE64"]
	delete(values, "APPLE_IAP_KEY_ID")
	_, err := config.LoadStoreIAPConfigFrom(mapLookup(values))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "APPLE_IAP_KEY_ID")
	assert.NotContains(t, err.Error(), secret)

	values = validStoreIAPEnvironment(t)
	values["STORE_EVIDENCE_ENCRYPTION_KEYS_JSON"] = `{"1":"sensitive-invalid-value"}`
	_, err = config.LoadStoreIAPConfigFrom(mapLookup(values))
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "sensitive-invalid-value")
}

func TestStoreIAPConfigRejectsInvalidDLQAndBoolean(t *testing.T) {
	_, err := config.LoadStoreIAPConfigFrom(mapLookup(map[string]string{"STORE_IAP_ENABLED": "sometimes"}))
	require.Error(t, err)

	values := validStoreIAPEnvironment(t)
	values["GOOGLE_PUBSUB_DEAD_LETTER_TOPIC"] = values["GOOGLE_PUBSUB_TOPIC"]
	_, err = config.LoadStoreIAPConfigFrom(mapLookup(values))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must differ")

	values = validStoreIAPEnvironment(t)
	values["GOOGLE_PUBSUB_MAX_DELIVERY_ATTEMPTS"] = "101"
	_, err = config.LoadStoreIAPConfigFrom(mapLookup(values))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "between 5 and 100")

	values = validStoreIAPEnvironment(t)
	values["GOOGLE_PUBSUB_PUSH_AUDIENCE"] = "http://api.example.com/webhook/google-play/rtdn"
	_, err = config.LoadStoreIAPConfigFrom(mapLookup(values))
	require.ErrorContains(t, err, "absolute HTTPS URL")

	values = validStoreIAPEnvironment(t)
	values["GOOGLE_PUBSUB_DEAD_LETTER_SUBSCRIPTION"] = values["GOOGLE_PUBSUB_SUBSCRIPTION"]
	_, err = config.LoadStoreIAPConfigFrom(mapLookup(values))
	require.ErrorContains(t, err, "must differ")
}

func mapLookup(values map[string]string) config.EnvironmentLookup {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}

func validStoreIAPEnvironment(t *testing.T) map[string]string {
	t.Helper()
	rootDER := testRootCertificate(t)
	roots, err := json.Marshal([]string{base64.StdEncoding.EncodeToString(rootDER)})
	require.NoError(t, err)
	key := base64.StdEncoding.EncodeToString(make([]byte, 32))
	privateKey := base64.StdEncoding.EncodeToString(testApplePrivateKeyPEM(t))
	return map[string]string{
		"STORE_IAP_ENABLED":                        "true",
		"APPLE_IAP_KEY_ID":                         "KEY123",
		"APPLE_IAP_ISSUER_ID":                      "issuer-id",
		"APPLE_IAP_BUNDLE_ID":                      "com.example.app",
		"APPLE_IAP_APP_APPLE_ID":                   "123456789",
		"APPLE_IAP_PRIVATE_KEY_BASE64":             privateKey,
		"APPLE_IAP_ROOT_CERTIFICATES_BASE64_JSON":  string(roots),
		"APPLE_IAP_PRODUCTS_JSON":                  `{"com.example.premium.monthly":{"entitlement":"premium","billingPeriod":"monthly"}}`,
		"GOOGLE_PLAY_PACKAGE_NAME":                 "com.example.app",
		"GOOGLE_PLAY_PRODUCTS_JSON":                `{"premium-monthly":{"entitlement":"premium","billingPeriod":"monthly"}}`,
		"GOOGLE_PUBSUB_PUSH_AUDIENCE":              "https://api.example.com/webhook/google-play/rtdn",
		"GOOGLE_PUBSUB_PUSH_SERVICE_ACCOUNT_EMAIL": "pubsub@example.iam.gserviceaccount.com",
		"GOOGLE_PUBSUB_TOPIC":                      "projects/example/topics/store-rtdn",
		"GOOGLE_PUBSUB_SUBSCRIPTION":               "projects/example/subscriptions/store-rtdn-push",
		"GOOGLE_PUBSUB_DEAD_LETTER_TOPIC":          "projects/example/topics/store-rtdn-dead-letter",
		"GOOGLE_PUBSUB_DEAD_LETTER_SUBSCRIPTION":   "projects/example/subscriptions/store-rtdn-dead-letter",
		"STORE_EVIDENCE_LOOKUP_KEY_BASE64":         key,
		"STORE_EVIDENCE_ACTIVE_KEY_VERSION":        "1",
		"STORE_EVIDENCE_ENCRYPTION_KEYS_JSON":      `{"1":"` + key + `"}`,
	}
}

func testApplePrivateKeyPEM(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

func testRootCertificate(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "test root"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	return der
}
