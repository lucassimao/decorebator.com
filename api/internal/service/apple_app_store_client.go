package service

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	"decorebator.com/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

const (
	appleProductionBaseURL = "https://api.storekit.apple.com"
	appleSandboxBaseURL    = "https://api.storekit-sandbox.apple.com"
	appleResponseLimit     = 1 << 20
)

type AppleTransactionInfoClient interface {
	GetTransactionInfo(ctx context.Context, transactionID string, environment model.StoreEnvironment) (string, error)
}

type AppleAPIError struct {
	StatusCode int
	ErrorCode  int64
	Retryable  bool
}

func (e *AppleAPIError) Error() string {
	return fmt.Sprintf("Apple App Store API request failed (status=%d code=%d)", e.StatusCode, e.ErrorCode)
}

func (e *AppleAPIError) TransactionNotFound() bool {
	return e != nil && e.ErrorCode == 4040010
}

type AppleAppStoreClientConfig struct {
	KeyID      string
	IssuerID   string
	BundleID   string
	PrivateKey []byte
	HTTPClient *http.Client
	Now        func() time.Time
	BaseURLs   map[model.StoreEnvironment]string
}

type AppleAppStoreClient struct {
	keyID      string
	issuerID   string
	bundleID   string
	privateKey *ecdsa.PrivateKey
	httpClient *http.Client
	now        func() time.Time
	baseURLs   map[model.StoreEnvironment]string
}

func NewAppleAppStoreClient(config AppleAppStoreClientConfig) (*AppleAppStoreClient, error) {
	if strings.TrimSpace(config.KeyID) == "" || strings.TrimSpace(config.IssuerID) == "" || strings.TrimSpace(config.BundleID) == "" {
		return nil, fmt.Errorf("Apple key ID, issuer ID, and bundle ID are required")
	}
	if strings.TrimSpace(config.KeyID) != config.KeyID || strings.TrimSpace(config.IssuerID) != config.IssuerID ||
		strings.TrimSpace(config.BundleID) != config.BundleID {
		return nil, fmt.Errorf("Apple key ID, issuer ID, and bundle ID must not contain surrounding whitespace")
	}
	privateKey, err := parseApplePrivateKey(config.PrivateKey)
	if err != nil {
		return nil, err
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	now := config.Now
	if now == nil {
		now = time.Now
	}
	baseURLs := map[model.StoreEnvironment]string{
		model.StoreEnvironmentProduction: appleProductionBaseURL,
		model.StoreEnvironmentSandbox:    appleSandboxBaseURL,
	}
	for environment, baseURL := range config.BaseURLs {
		if !environment.Valid() {
			return nil, fmt.Errorf("unsupported Apple API environment %q", environment)
		}
		parsed, err := url.Parse(baseURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return nil, fmt.Errorf("invalid Apple API base URL for %s", environment)
		}
		baseURLs[environment] = strings.TrimRight(baseURL, "/")
	}

	return &AppleAppStoreClient{
		keyID: config.KeyID, issuerID: config.IssuerID, bundleID: config.BundleID,
		privateKey: privateKey, httpClient: httpClient, now: now, baseURLs: baseURLs,
	}, nil
}

func (c *AppleAppStoreClient) GetTransactionInfo(
	ctx context.Context,
	transactionID string,
	environment model.StoreEnvironment,
) (string, error) {
	if !validAppleTransactionID(transactionID) {
		return "", fmt.Errorf("Apple transaction ID is invalid")
	}
	baseURL, ok := c.baseURLs[environment]
	if !ok {
		return "", fmt.Errorf("unsupported Apple API environment %q", environment)
	}
	token, err := c.authorizationToken()
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		baseURL+"/inApps/v1/transactions/"+url.PathEscape(transactionID),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("create Apple transaction request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("call Apple transaction API: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, appleResponseLimit+1))
	if err != nil {
		return "", fmt.Errorf("read Apple transaction response: %w", err)
	}
	if len(body) > appleResponseLimit {
		return "", fmt.Errorf("Apple transaction response exceeds limit")
	}
	if response.StatusCode != http.StatusOK {
		var apiBody struct {
			ErrorCode int64 `json:"errorCode"`
		}
		_ = json.Unmarshal(body, &apiBody)
		return "", &AppleAPIError{
			StatusCode: response.StatusCode,
			ErrorCode:  apiBody.ErrorCode,
			Retryable: response.StatusCode == http.StatusUnauthorized ||
				response.StatusCode == http.StatusForbidden ||
				response.StatusCode == http.StatusTooManyRequests ||
				response.StatusCode >= http.StatusInternalServerError ||
				appleErrorCodeIsRetryable(apiBody.ErrorCode),
		}
	}
	var result struct {
		SignedTransactionInfo string `json:"signedTransactionInfo"`
	}
	if err := json.Unmarshal(body, &result); err != nil || result.SignedTransactionInfo == "" {
		return "", fmt.Errorf("decode Apple transaction response")
	}
	return result.SignedTransactionInfo, nil
}

func appleErrorCodeIsRetryable(errorCode int64) bool {
	switch errorCode {
	case 4040002, 4040004, 4040006, 5000001:
		return true
	default:
		return false
	}
}

func validAppleTransactionID(transactionID string) bool {
	return transactionID != "" &&
		strings.TrimSpace(transactionID) == transactionID &&
		!strings.Contains(transactionID, "/") &&
		!strings.ContainsFunc(transactionID, unicode.IsSpace)
}

func (c *AppleAppStoreClient) authorizationToken() (string, error) {
	now := c.now().UTC()
	claims := struct {
		BundleID string `json:"bid"`
		jwt.RegisteredClaims
	}{
		BundleID: c.bundleID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    c.issuerID,
			Audience:  jwt.ClaimStrings{"appstoreconnect-v1"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = c.keyID
	token.Header["typ"] = "JWT"
	signed, err := token.SignedString(c.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign Apple API authorization token: %w", err)
	}
	return signed, nil
}

func parseApplePrivateKey(encoded []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(encoded)
	if block == nil {
		return nil, errors.New("Apple private key is not PEM encoded")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse Apple private key: %w", err)
	}
	privateKey, ok := parsed.(*ecdsa.PrivateKey)
	if !ok || privateKey.Curve.Params().BitSize != 256 {
		return nil, errors.New("Apple private key must use P-256")
	}
	return privateKey, nil
}
