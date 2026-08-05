package config

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	storeEvidenceKeyBytes      = 32
	defaultDLQDeliveryAttempts = 10
	minimumDLQDeliveryAttempts = 5
	maximumDLQDeliveryAttempts = 100
)

type EnvironmentLookup func(string) (string, bool)

type StoreIAPConfig struct {
	Enabled  bool
	Apple    AppleStoreConfig
	Google   GooglePlayConfig
	Evidence StoreEvidenceConfig
}

type AppleStoreConfig struct {
	KeyID      string
	IssuerID   string
	BundleID   string
	AppAppleID int64
	PrivateKey []byte
	Roots      []*x509.Certificate
	Products   map[string]string
}

type GooglePlayConfig struct {
	PackageName             string
	Products                map[string]string
	PushAudience            string
	PushServiceAccountEmail string
	Topic                   string
	Subscription            string
	// Dead-letter fields are a startup contract for the checked-in operator
	// runbook; the application never mutates Google Cloud resources.
	DeadLetterTopic        string
	DeadLetterSubscription string
	MaxDeliveryAttempts    int
}

type StoreEvidenceConfig struct {
	LookupKey     []byte
	ActiveVersion int16
	Keys          map[int16][]byte
}

func LoadStoreIAPConfig() (StoreIAPConfig, error) {
	return LoadStoreIAPConfigFrom(os.LookupEnv)
}

func LoadStoreIAPConfigFrom(lookup EnvironmentLookup) (StoreIAPConfig, error) {
	if lookup == nil {
		return StoreIAPConfig{}, fmt.Errorf("store IAP environment lookup is required")
	}
	enabledRaw, exists := lookup("STORE_IAP_ENABLED")
	if !exists || strings.TrimSpace(enabledRaw) == "" {
		if environment, ok := lookup("ENV"); ok && strings.EqualFold(strings.TrimSpace(environment), "production") {
			return StoreIAPConfig{}, fmt.Errorf("STORE_IAP_ENABLED must be explicitly configured in production")
		}
		return StoreIAPConfig{}, nil
	}
	enabled, err := strconv.ParseBool(strings.TrimSpace(enabledRaw))
	if err != nil {
		return StoreIAPConfig{}, fmt.Errorf("STORE_IAP_ENABLED must be a boolean")
	}
	if !enabled {
		return StoreIAPConfig{}, nil
	}

	apple, err := loadAppleStoreConfig(lookup)
	if err != nil {
		return StoreIAPConfig{}, err
	}
	googleConfig, err := loadGooglePlayConfig(lookup)
	if err != nil {
		return StoreIAPConfig{}, err
	}
	evidence, err := loadStoreEvidenceConfig(lookup)
	if err != nil {
		return StoreIAPConfig{}, err
	}
	return StoreIAPConfig{Enabled: true, Apple: apple, Google: googleConfig, Evidence: evidence}, nil
}

func loadAppleStoreConfig(lookup EnvironmentLookup) (AppleStoreConfig, error) {
	keyID, err := requiredEnv(lookup, "APPLE_IAP_KEY_ID")
	if err != nil {
		return AppleStoreConfig{}, err
	}
	issuerID, err := requiredEnv(lookup, "APPLE_IAP_ISSUER_ID")
	if err != nil {
		return AppleStoreConfig{}, err
	}
	bundleID, err := requiredEnv(lookup, "APPLE_IAP_BUNDLE_ID")
	if err != nil {
		return AppleStoreConfig{}, err
	}
	appIDRaw, err := requiredEnv(lookup, "APPLE_IAP_APP_APPLE_ID")
	if err != nil {
		return AppleStoreConfig{}, err
	}
	appID, err := strconv.ParseInt(appIDRaw, 10, 64)
	if err != nil || appID <= 0 {
		return AppleStoreConfig{}, fmt.Errorf("APPLE_IAP_APP_APPLE_ID must be a positive integer")
	}
	privateKey, err := requiredBase64(lookup, "APPLE_IAP_PRIVATE_KEY_BASE64", 0)
	if err != nil {
		return AppleStoreConfig{}, err
	}
	if err = validateApplePrivateKeyPEM(privateKey); err != nil {
		return AppleStoreConfig{}, err
	}
	roots, err := loadAppleRoots(lookup)
	if err != nil {
		return AppleStoreConfig{}, err
	}
	products, err := loadProductCatalog(lookup, "APPLE_IAP_PRODUCTS_JSON")
	if err != nil {
		return AppleStoreConfig{}, err
	}
	return AppleStoreConfig{
		KeyID: keyID, IssuerID: issuerID, BundleID: bundleID, AppAppleID: appID,
		PrivateKey: privateKey, Roots: roots, Products: products,
	}, nil
}

func validateApplePrivateKeyPEM(value []byte) error {
	block, rest := pem.Decode(value)
	if block == nil || len(rest) != 0 || block.Type != "PRIVATE KEY" {
		return fmt.Errorf("APPLE_IAP_PRIVATE_KEY_BASE64 must encode one PKCS8 EC private key PEM")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("APPLE_IAP_PRIVATE_KEY_BASE64 must encode one PKCS8 EC private key PEM")
	}
	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok || key.Curve == nil || key.Curve.Params().BitSize != 256 {
		return fmt.Errorf("APPLE_IAP_PRIVATE_KEY_BASE64 must encode one PKCS8 EC private key PEM")
	}
	return nil
}

func loadAppleRoots(lookup EnvironmentLookup) ([]*x509.Certificate, error) {
	raw, err := requiredEnv(lookup, "APPLE_IAP_ROOT_CERTIFICATES_BASE64_JSON")
	if err != nil {
		return nil, err
	}
	var encoded []string
	if err := json.Unmarshal([]byte(raw), &encoded); err != nil || len(encoded) == 0 {
		return nil, fmt.Errorf("APPLE_IAP_ROOT_CERTIFICATES_BASE64_JSON must be a non-empty JSON string array")
	}
	roots := make([]*x509.Certificate, 0, len(encoded))
	for _, value := range encoded {
		der, decodeErr := base64.StdEncoding.DecodeString(value)
		if decodeErr != nil {
			return nil, fmt.Errorf("APPLE_IAP_ROOT_CERTIFICATES_BASE64_JSON contains invalid base64")
		}
		certificate, parseErr := x509.ParseCertificate(der)
		if parseErr != nil || !certificate.IsCA {
			return nil, fmt.Errorf("APPLE_IAP_ROOT_CERTIFICATES_BASE64_JSON contains an invalid CA certificate")
		}
		roots = append(roots, certificate)
	}
	return roots, nil
}

func loadGooglePlayConfig(lookup EnvironmentLookup) (GooglePlayConfig, error) {
	packageName, err := requiredEnv(lookup, "GOOGLE_PLAY_PACKAGE_NAME")
	if err != nil {
		return GooglePlayConfig{}, err
	}
	products, err := loadProductCatalog(lookup, "GOOGLE_PLAY_PRODUCTS_JSON")
	if err != nil {
		return GooglePlayConfig{}, err
	}
	pubSub, err := loadGooglePubSubConfig(lookup)
	if err != nil {
		return GooglePlayConfig{}, err
	}
	pubSub.PackageName = packageName
	pubSub.Products = products
	return pubSub, nil
}

func loadGooglePubSubConfig(lookup EnvironmentLookup) (GooglePlayConfig, error) {
	audience, err := requiredEnv(lookup, "GOOGLE_PUBSUB_PUSH_AUDIENCE")
	if err != nil {
		return GooglePlayConfig{}, err
	}
	audienceURL, parseErr := url.Parse(audience)
	if parseErr != nil || audienceURL.Scheme != "https" || audienceURL.Host == "" || audienceURL.User != nil {
		return GooglePlayConfig{}, fmt.Errorf("GOOGLE_PUBSUB_PUSH_AUDIENCE must be an absolute HTTPS URL")
	}
	serviceAccount, err := requiredEnv(lookup, "GOOGLE_PUBSUB_PUSH_SERVICE_ACCOUNT_EMAIL")
	if err != nil {
		return GooglePlayConfig{}, err
	}
	if !strings.Contains(serviceAccount, "@") {
		return GooglePlayConfig{}, fmt.Errorf("GOOGLE_PUBSUB_PUSH_SERVICE_ACCOUNT_EMAIL must be a valid email")
	}
	topic, err := requiredGoogleResource(lookup, "GOOGLE_PUBSUB_TOPIC", "topics")
	if err != nil {
		return GooglePlayConfig{}, err
	}
	subscription, err := requiredGoogleResource(lookup, "GOOGLE_PUBSUB_SUBSCRIPTION", "subscriptions")
	if err != nil {
		return GooglePlayConfig{}, err
	}
	deadLetterTopic, err := requiredGoogleResource(lookup, "GOOGLE_PUBSUB_DEAD_LETTER_TOPIC", "topics")
	if err != nil {
		return GooglePlayConfig{}, err
	}
	if deadLetterTopic == topic {
		return GooglePlayConfig{}, fmt.Errorf("Google Pub/Sub dead-letter topic must differ from the RTDN topic")
	}
	deadLetterSubscription, err := requiredGoogleResource(lookup, "GOOGLE_PUBSUB_DEAD_LETTER_SUBSCRIPTION", "subscriptions")
	if err != nil {
		return GooglePlayConfig{}, err
	}
	if deadLetterSubscription == subscription {
		return GooglePlayConfig{}, fmt.Errorf("Google Pub/Sub dead-letter subscription must differ from the RTDN subscription")
	}
	attempts := defaultDLQDeliveryAttempts
	if raw, ok := lookup("GOOGLE_PUBSUB_MAX_DELIVERY_ATTEMPTS"); ok && strings.TrimSpace(raw) != "" {
		attempts, err = strconv.Atoi(strings.TrimSpace(raw))
		if err != nil || attempts < minimumDLQDeliveryAttempts || attempts > maximumDLQDeliveryAttempts {
			return GooglePlayConfig{}, fmt.Errorf("GOOGLE_PUBSUB_MAX_DELIVERY_ATTEMPTS must be between 5 and 100")
		}
	}
	return GooglePlayConfig{
		PushAudience: audience, PushServiceAccountEmail: serviceAccount,
		Topic: topic, Subscription: subscription,
		DeadLetterTopic: deadLetterTopic, DeadLetterSubscription: deadLetterSubscription,
		MaxDeliveryAttempts: attempts,
	}, nil
}

func loadStoreEvidenceConfig(lookup EnvironmentLookup) (StoreEvidenceConfig, error) {
	lookupKey, err := requiredBase64(lookup, "STORE_EVIDENCE_LOOKUP_KEY_BASE64", storeEvidenceKeyBytes)
	if err != nil {
		return StoreEvidenceConfig{}, err
	}
	activeRaw, err := requiredEnv(lookup, "STORE_EVIDENCE_ACTIVE_KEY_VERSION")
	if err != nil {
		return StoreEvidenceConfig{}, err
	}
	active, err := strconv.ParseInt(activeRaw, 10, 16)
	if err != nil || active <= 0 {
		return StoreEvidenceConfig{}, fmt.Errorf("STORE_EVIDENCE_ACTIVE_KEY_VERSION must be a positive 16-bit integer")
	}
	keyringRaw, err := requiredEnv(lookup, "STORE_EVIDENCE_ENCRYPTION_KEYS_JSON")
	if err != nil {
		return StoreEvidenceConfig{}, err
	}
	var encoded map[string]string
	if err := json.Unmarshal([]byte(keyringRaw), &encoded); err != nil || len(encoded) == 0 {
		return StoreEvidenceConfig{}, fmt.Errorf("STORE_EVIDENCE_ENCRYPTION_KEYS_JSON must be a non-empty JSON object")
	}
	keys := make(map[int16][]byte, len(encoded))
	for versionRaw, value := range encoded {
		version, parseErr := strconv.ParseInt(versionRaw, 10, 16)
		key, decodeErr := base64.StdEncoding.DecodeString(value)
		if parseErr != nil || version <= 0 || decodeErr != nil || len(key) != storeEvidenceKeyBytes {
			return StoreEvidenceConfig{}, fmt.Errorf("STORE_EVIDENCE_ENCRYPTION_KEYS_JSON contains an invalid version or key")
		}
		keys[int16(version)] = key
	}
	if _, ok := keys[int16(active)]; !ok {
		return StoreEvidenceConfig{}, fmt.Errorf("active store evidence encryption key is missing")
	}
	return StoreEvidenceConfig{LookupKey: lookupKey, ActiveVersion: int16(active), Keys: keys}, nil
}

func loadProductCatalog(lookup EnvironmentLookup, name string) (map[string]string, error) {
	raw, err := requiredEnv(lookup, name)
	if err != nil {
		return nil, err
	}
	var products map[string]string
	if err := json.Unmarshal([]byte(raw), &products); err != nil || len(products) == 0 {
		return nil, fmt.Errorf("%s must be a non-empty JSON object", name)
	}
	for product, entitlement := range products {
		if strings.TrimSpace(product) != product || product == "" || entitlement != "premium" {
			return nil, fmt.Errorf("%s contains an invalid product mapping", name)
		}
	}
	return products, nil
}

func requiredGoogleResource(lookup EnvironmentLookup, name, kind string) (string, error) {
	value, err := requiredEnv(lookup, name)
	if err != nil {
		return "", err
	}
	parts := strings.Split(value, "/")
	if len(parts) != 4 || parts[0] != "projects" || parts[1] == "" || parts[2] != kind || parts[3] == "" {
		return "", fmt.Errorf("%s must be a fully-qualified Google %s resource", name, kind)
	}
	return value, nil
}

func requiredBase64(lookup EnvironmentLookup, name string, exactBytes int) ([]byte, error) {
	raw, err := requiredEnv(lookup, name)
	if err != nil {
		return nil, err
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil || (exactBytes > 0 && len(decoded) != exactBytes) || len(decoded) == 0 {
		return nil, fmt.Errorf("%s contains invalid base64 key material", name)
	}
	return decoded, nil
}

func requiredEnv(lookup EnvironmentLookup, name string) (string, error) {
	value, exists := lookup(name)
	if !exists || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s is required when store IAP is enabled", name)
	}
	if value != strings.TrimSpace(value) {
		return "", fmt.Errorf("%s must not contain surrounding whitespace", name)
	}
	return value, nil
}
