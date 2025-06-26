package mocks

import (
	"context"

	"decorebator.com/internal/service"
)

// MockRevenueCatAPIClient is a mock implementation of RevenueCatAPIClient for testing.
type MockRevenueCatAPIClient struct {
	// GetCustomerInfoFunc allows setting custom behavior for GetCustomerInfo
	GetCustomerInfoFunc func(ctx context.Context, appUserID string) (*service.CustomerInfo, error)

	// Track calls for verification
	GetCustomerInfoCalls []struct {
		AppUserID string
	}
}

// GetCustomerInfo implements RevenueCatAPIClient interface
func (m *MockRevenueCatAPIClient) GetCustomerInfo(ctx context.Context, appUserID string) (*service.CustomerInfo, error) {
	// Track the call
	m.GetCustomerInfoCalls = append(m.GetCustomerInfoCalls, struct{ AppUserID string }{
		AppUserID: appUserID,
	})

	// Call custom function if provided
	if m.GetCustomerInfoFunc != nil {
		return m.GetCustomerInfoFunc(ctx, appUserID)
	}

	// Default behavior - return empty customer info
	return &service.CustomerInfo{}, nil
}
