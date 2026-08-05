package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	googlePlayDefaultBaseURL   = "https://androidpublisher.googleapis.com"
	googlePlayResponseLimit    = 1 << 20
	googlePlayRequestTimeout   = 15 * time.Second
	googlePlaySubscriptionKind = "androidpublisher#subscriptionPurchaseV2"
)

type GoogleAccessTokenSource interface {
	AccessToken(context.Context) (string, error)
}

type GoogleAccessTokenError struct {
	Retryable bool
	Cause     error
}

func (e *GoogleAccessTokenError) Error() string { return "Google access-token acquisition failed" }
func (e *GoogleAccessTokenError) Unwrap() error { return e.Cause }

type GooglePlaySubscriptionClient interface {
	GetSubscription(context.Context, string) (GooglePlaySubscription, error)
}

type GooglePlaySubscriptionAcknowledger interface {
	AcknowledgeSubscription(context.Context, string, string) error
}

type GooglePlaySubscription struct {
	Kind                       string                            `json:"kind"`
	StartTime                  string                            `json:"startTime"`
	SubscriptionState          string                            `json:"subscriptionState"`
	LinkedPurchaseToken        string                            `json:"linkedPurchaseToken"`
	AcknowledgementState       string                            `json:"acknowledgementState"`
	ExternalAccountIdentifiers *GoogleExternalAccountIdentifiers `json:"externalAccountIdentifiers"`
	LineItems                  []GooglePlaySubscriptionLineItem  `json:"lineItems"`
	TestPurchase               *struct{}                         `json:"testPurchase"`
	CanceledStateContext       *GoogleCanceledStateContext       `json:"canceledStateContext"`
	ETag                       string                            `json:"etag"`
}

type GoogleExternalAccountIdentifiers struct {
	ObfuscatedExternalAccountID string `json:"obfuscatedExternalAccountId"`
}

type GooglePlaySubscriptionLineItem struct {
	ProductID        string                  `json:"productId"`
	ExpiryTime       string                  `json:"expiryTime"`
	AutoRenewingPlan *GoogleAutoRenewingPlan `json:"autoRenewingPlan"`
	PrepaidPlan      *struct{}               `json:"prepaidPlan"`
}

type GoogleAutoRenewingPlan struct {
	AutoRenewEnabled bool `json:"autoRenewEnabled"`
}

type GoogleCanceledStateContext struct {
	UserInitiatedCancellation *struct {
		CancelTime string `json:"cancelTime"`
	} `json:"userInitiatedCancellation"`
	SystemInitiatedCancellation    *struct{} `json:"systemInitiatedCancellation"`
	DeveloperInitiatedCancellation *struct{} `json:"developerInitiatedCancellation"`
	ReplacementCancellation        *struct{} `json:"replacementCancellation"`
}

type GooglePlayAPIError struct {
	StatusCode int
	Retryable  bool
	Cause      error
}

func (e *GooglePlayAPIError) Error() string {
	return fmt.Sprintf("Google Play API request failed (status=%d)", e.StatusCode)
}

func (e *GooglePlayAPIError) Unwrap() error { return e.Cause }

var ErrInvalidGoogleAcknowledgeRequest = errors.New("invalid Google acknowledge request")

type GooglePlayClient struct {
	packageName string
	tokens      GoogleAccessTokenSource
	httpClient  *http.Client
	baseURL     string
}

func NewGooglePlayClient(
	packageName string,
	tokens GoogleAccessTokenSource,
	httpClient *http.Client,
	baseURL string,
) (*GooglePlayClient, error) {
	if strings.TrimSpace(packageName) == "" || tokens == nil {
		return nil, fmt.Errorf("Google Play package name and access-token source are required")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: googlePlayRequestTimeout}
	}
	if baseURL == "" {
		baseURL = googlePlayDefaultBaseURL
	}
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil || parsedBaseURL.Scheme == "" || parsedBaseURL.Host == "" || !secureGooglePlayBaseURL(parsedBaseURL) {
		return nil, fmt.Errorf("invalid Google Play base URL")
	}
	return &GooglePlayClient{
		packageName: packageName, tokens: tokens, httpClient: httpClient,
		baseURL: strings.TrimRight(baseURL, "/"),
	}, nil
}

func (c *GooglePlayClient) GetSubscription(
	ctx context.Context,
	purchaseToken string,
) (GooglePlaySubscription, error) {
	if !validGooglePurchaseToken(purchaseToken) {
		return GooglePlaySubscription{}, &GooglePlayAPIError{StatusCode: http.StatusBadRequest}
	}
	requestContext, cancel := context.WithTimeout(ctx, googlePlayRequestTimeout)
	defer cancel()
	accessToken, err := c.tokens.AccessToken(requestContext)
	if err != nil || strings.TrimSpace(accessToken) == "" {
		return GooglePlaySubscription{}, googleTokenFailure(err)
	}
	requestURL := fmt.Sprintf(
		"%s/androidpublisher/v3/applications/%s/purchases/subscriptionsv2/tokens/%s",
		c.baseURL, url.PathEscape(c.packageName), url.PathEscape(purchaseToken),
	)
	request, err := http.NewRequestWithContext(requestContext, http.MethodGet, requestURL, nil)
	if err != nil {
		return GooglePlaySubscription{}, &GooglePlayAPIError{StatusCode: 0, Retryable: true}
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("Accept", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return GooglePlaySubscription{}, &GooglePlayAPIError{StatusCode: 0, Retryable: true}
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, googlePlayResponseLimit))
		return GooglePlaySubscription{}, &GooglePlayAPIError{
			StatusCode: response.StatusCode, Retryable: googlePlayRetryableStatus(response.StatusCode),
		}
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, googlePlayResponseLimit+1))
	if err != nil || len(body) > googlePlayResponseLimit {
		return GooglePlaySubscription{}, &GooglePlayAPIError{StatusCode: response.StatusCode, Retryable: true}
	}
	var subscription GooglePlaySubscription
	if err := json.Unmarshal(body, &subscription); err != nil {
		return GooglePlaySubscription{}, &GooglePlayAPIError{StatusCode: response.StatusCode, Retryable: true}
	}
	if subscription.Kind != googlePlaySubscriptionKind {
		return GooglePlaySubscription{}, &GooglePlayAPIError{StatusCode: response.StatusCode, Retryable: true}
	}
	return subscription, nil
}

func (c *GooglePlayClient) AcknowledgeSubscription(
	ctx context.Context,
	productID string,
	purchaseToken string,
) error {
	if strings.TrimSpace(productID) == "" || !validGooglePurchaseToken(purchaseToken) {
		return ErrInvalidGoogleAcknowledgeRequest
	}
	requestContext, cancel := context.WithTimeout(ctx, googlePlayRequestTimeout)
	defer cancel()
	accessToken, err := c.tokens.AccessToken(requestContext)
	if err != nil || strings.TrimSpace(accessToken) == "" {
		return googleTokenFailure(err)
	}
	requestURL := fmt.Sprintf(
		"%s/androidpublisher/v3/applications/%s/purchases/subscriptions/%s/tokens/%s:acknowledge",
		c.baseURL, url.PathEscape(c.packageName), url.PathEscape(productID), url.PathEscape(purchaseToken),
	)
	request, err := http.NewRequestWithContext(requestContext, http.MethodPost, requestURL, strings.NewReader("{}"))
	if err != nil {
		return &GooglePlayAPIError{StatusCode: 0, Retryable: true}
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return &GooglePlayAPIError{StatusCode: 0, Retryable: true}
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, googlePlayResponseLimit))
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return &GooglePlayAPIError{
			StatusCode: response.StatusCode, Retryable: googlePlayRetryableStatus(response.StatusCode),
		}
	}
	return nil
}

func secureGooglePlayBaseURL(baseURL *url.URL) bool {
	if baseURL.Scheme == "https" {
		return true
	}
	if baseURL.Scheme != "http" {
		return false
	}
	host := baseURL.Hostname()
	return host == "localhost" || net.ParseIP(host) != nil && net.ParseIP(host).IsLoopback()
}

func googlePlayRetryableStatus(status int) bool {
	return status == http.StatusRequestTimeout || status == http.StatusConflict ||
		status == http.StatusUnauthorized ||
		status == http.StatusTooManyRequests || status >= 500
}

func googleTokenFailure(err error) *GooglePlayAPIError {
	retryable := true
	var tokenError *GoogleAccessTokenError
	if errors.As(err, &tokenError) {
		retryable = tokenError.Retryable
	}
	return &GooglePlayAPIError{StatusCode: 0, Retryable: retryable, Cause: err}
}

func validGooglePurchaseToken(token string) bool {
	return token != "" && strings.TrimSpace(token) == token && len(token) <= 4096
}
