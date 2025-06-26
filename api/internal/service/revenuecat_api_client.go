package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// revenueCatAPIClient is the default implementation of RevenueCatAPIClient.
type revenueCatAPIClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewRevenueCatAPIClient creates a new RevenueCat API client with the given API key.
func NewRevenueCatAPIClient(apiKey string) RevenueCatAPIClient {
	return &revenueCatAPIClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetCustomerInfo fetches customer info from RevenueCat API.
func (c *revenueCatAPIClient) GetCustomerInfo(ctx context.Context, appUserID string) (*CustomerInfo, error) {
	url := fmt.Sprintf("https://api.revenuecat.com/v1/subscribers/%s", appUserID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch customer info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RevenueCat API error: %d - %s", resp.StatusCode, string(body))
	}

	var customerInfo CustomerInfo
	if err := json.NewDecoder(resp.Body).Decode(&customerInfo); err != nil {
		return nil, fmt.Errorf("failed to decode customer info: %w", err)
	}

	return &customerInfo, nil
}
