# RevenueCat Service Testing Guide

## Overview

The RevenueCat service has been refactored to follow dependency injection patterns, making it highly testable by separating external API calls from business logic.

## Architecture

### Interfaces

1. **RevenueCatAPIClient** - Interface for external API calls
   - Located in: `internal/service/interfaces.go`
   - Methods: `GetCustomerInfo(ctx, appUserID) (*CustomerInfo, error)`

2. **RevenueCatService** - Main service interface with business logic
   - Located in: `internal/service/interfaces.go`
   - Contains all RevenueCat-related business operations

### Implementations

1. **revenueCatAPIClient** - Default implementation for production
   - Located in: `internal/service/revenuecat_api_client.go`
   - Makes actual HTTP calls to RevenueCat API

2. **revenueCatService** - Main service implementation
   - Located in: `internal/service/revenuecat.go`
   - Uses RevenueCatAPIClient for external calls
   - Contains all business logic

3. **MockRevenueCatAPIClient** - Mock implementation for testing
   - Located in: `internal/service/revenuecat_api_client_mock.go`
   - Allows custom behavior injection for tests

## Usage

### Production Usage

```go
// Standard initialization with real API client
service := NewRevenueCatService(db)
```

### Test Usage

```go
// Create mock client with custom behavior
mockClient := &MockRevenueCatAPIClient{
    GetCustomerInfoFunc: func(ctx context.Context, appUserID string) (*CustomerInfo, error) {
        // Return test data
        return &CustomerInfo{
            Subscriber: Subscriber{
                OriginalAppUserID: appUserID,
                Entitlements: map[string]Entitlement{
                    EntitlementPremium: {
                        ProductIdentifier: ProductMonthlyIOS,
                        PurchaseDate:      "2025-01-01T00:00:00Z",
                        ExpiresDate:       &expiresDate,
                    },
                },
            },
        }, nil
    },
}

// Create service with mock
service := NewRevenueCatServiceWithClient(db, mockClient)

// Use service normally
err := service.RestorePurchases(ctx, userID, appUserID, platform)

// Verify calls
assert.Len(t, mockClient.GetCustomerInfoCalls, 1)
assert.Equal(t, appUserID, mockClient.GetCustomerInfoCalls[0].AppUserID)
```

## Testing Scenarios

### 1. Success Case
```go
mockClient := &MockRevenueCatAPIClient{
    GetCustomerInfoFunc: func(ctx context.Context, appUserID string) (*CustomerInfo, error) {
        return &CustomerInfo{/* valid data */}, nil
    },
}
```

### 2. API Error
```go
mockClient := &MockRevenueCatAPIClient{
    GetCustomerInfoFunc: func(ctx context.Context, appUserID string) (*CustomerInfo, error) {
        return nil, errors.New("API error: rate limit exceeded")
    },
}
```

### 3. Expired Subscription
```go
pastDate := "2024-01-01T00:00:00Z"
mockClient := &MockRevenueCatAPIClient{
    GetCustomerInfoFunc: func(ctx context.Context, appUserID string) (*CustomerInfo, error) {
        return &CustomerInfo{
            Subscriber: Subscriber{
                Entitlements: map[string]Entitlement{
                    EntitlementPremium: {
                        ExpiresDate: &pastDate, // Past date
                    },
                },
            },
        }, nil
    },
}
```

### 4. No Active Subscription
```go
mockClient := &MockRevenueCatAPIClient{
    GetCustomerInfoFunc: func(ctx context.Context, appUserID string) (*CustomerInfo, error) {
        return &CustomerInfo{
            Subscriber: Subscriber{
                Entitlements: map[string]Entitlement{}, // Empty
            },
        }, nil
    },
}
```

## Benefits

1. **Isolated Testing** - Test business logic without external API calls
2. **Fast Tests** - No network latency or API rate limits
3. **Predictable Results** - Mock returns exactly what you specify
4. **Error Simulation** - Easy to test error handling paths
5. **Call Verification** - Track what methods were called and with what parameters

## Migration Guide

For existing tests:

1. Replace direct service creation:
   ```go
   // Old
   service := NewRevenueCatService(db)
   
   // New (for tests)
   mockClient := &MockRevenueCatAPIClient{/* setup */}
   service := NewRevenueCatServiceWithClient(db, mockClient)
   ```

2. Mock API responses instead of HTTP endpoints
3. Verify business logic separately from API integration

## Best Practices

1. **Keep mocks simple** - Only mock what's needed for the test
2. **Test one thing** - Each test should focus on one scenario
3. **Use descriptive names** - Make test intent clear
4. **Verify calls** - Check that the right API methods were called
5. **Test error paths** - Don't just test the happy path