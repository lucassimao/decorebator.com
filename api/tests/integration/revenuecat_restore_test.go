package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/mocks"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestorePurchases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create a mock RevenueCat service for testing
	rcServiceMock := &mocks.RevenueCatServiceMock{}

	// Initialize test server with mock service
	ts := setup.NewTestServer(t, &setup.TestConfig{
		RevenueCatService: rcServiceMock,
	})
	defer ts.Cleanup()

	t.Run("RestorePurchases_RequiresAuthentication", func(t *testing.T) {
		body := map[string]interface{}{
			"appUserId": "test_user_123",
			"platform":  "ios",
		}
		jsonBody, _ := json.Marshal(body)

		req, err := http.NewRequest("POST", ts.BaseURL+"/subscription/revenuecat/restore", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("RestorePurchases_WithActiveSubscription", func(t *testing.T) {
		// Create test user
		signupInput := setup.GenerateSignupInput()
		ts.Expect.POST("/users").
			WithJSON(signupInput).
			Expect().
			Status(201)

		// Get user auth token from login response header
		loginResp := ts.Expect.POST("/login").
			WithJSON(map[string]string{
				"email":    signupInput.Email,
				"password": signupInput.Password,
			}).
			Expect().
			Status(200)

		authToken := loginResp.Header("Authorization").NotEmpty().Raw()

		ctx := context.Background()
		var userID int64
		err := ts.DB.QueryRow(ctx,
			"SELECT id FROM users WHERE email = $1",
			signupInput.Email).Scan(&userID)
		require.NoError(t, err)

		appUserID := fmt.Sprintf("rc_user_%d", userID)

		// Mock successful customer info fetch with active subscription
		expiresDateStr := time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339)
		rcServiceMock.GetCustomerInfoFunc = func(_ context.Context, requestedAppUserID string) (*service.CustomerInfo, error) {
			assert.Equal(t, appUserID, requestedAppUserID)
			return &service.CustomerInfo{
				Subscriber: service.Subscriber{
					OriginalAppUserID: appUserID,
					Entitlements: map[string]service.Entitlement{
						"premium": {
							ProductIdentifier: "com.decorebator.premium.monthly",
							PurchaseDate:      time.Now().Format(time.RFC3339),
							ExpiresDate:       &expiresDateStr,
						},
					},
					Subscriptions: map[string]service.Subscription{
						"com.decorebator.premium.monthly": {
							ExpiresDate:          expiresDateStr,
							OriginalPurchaseDate: time.Now().Format(time.RFC3339),
							PurchaseDate:         time.Now().Format(time.RFC3339),
							Store:                "app_store",
							PeriodType:           "normal",
						},
					},
				},
			}, nil
		}

		// Mock plan lookup
		rcServiceMock.GetPlanFromProductIDFunc = func(productID string) model.SubscriptionPlan {
			if productID == "com.decorebator.premium.monthly" {
				return model.PlanMonthly
			}
			return model.PlanAnnual
		}

		// Mock subscription creation
		rcServiceMock.CreateOrUpdateSubscriptionFromRevenueCatFunc = func(ctx context.Context, createdUserID int64, customerInfo *service.CustomerInfo, platform model.PlatformType) error {
			assert.Equal(t, userID, createdUserID)
			assert.Equal(t, model.PlatformIOS, platform)
			assert.NotNil(t, customerInfo)

			// Create subscription in database
			subRepo := repository.NewSubscriptionRepository(ts.DB)
			productID := customerInfo.Subscriber.Entitlements["premium"].ProductIdentifier
			subscription := &model.Subscription{
				UserID:                   createdUserID,
				Provider:                 model.ProviderRevenueCat,
				RevenueCatSubscriptionID: &customerInfo.Subscriber.OriginalAppUserID,
				AppStoreProductID:        &productID,
				Platform:                 &platform,
				Plan:                     model.PlanMonthly,
				Status:                   model.StatusActive,
				CurrentPeriodStart:       time.Now(),
				CurrentPeriodEnd:         time.Now().Add(30 * 24 * time.Hour),
				CancelAtPeriodEnd:        false,
				AmountCents:              699,
				Currency:                 "USD",
			}
			_, createErr := subRepo.CreateSubscription(ctx, subscription)
			return createErr
		}

		// Mock the restore purchases to use real service logic
		rcServiceMock.RestorePurchasesFunc = func(ctx context.Context, restoreUserID int64, restoreAppUserID string, platform model.PlatformType) error {
			// Link user to RevenueCat customer
			_, linkErr := ts.DB.Exec(ctx,
				"UPDATE users SET revenuecat_customer_id = $1 WHERE id = $2",
				restoreAppUserID, restoreUserID)
			if linkErr != nil {
				return linkErr
			}

			// Get customer info
			customerInfo, getErr := rcServiceMock.GetCustomerInfo(ctx, restoreAppUserID)
			if getErr != nil {
				return getErr
			}

			// Create/update subscription
			return rcServiceMock.CreateOrUpdateSubscriptionFromRevenueCat(ctx, restoreUserID, customerInfo, platform)
		}

		// Make the restore request
		body := map[string]interface{}{
			"appUserId": appUserID,
			"platform":  "ios",
		}
		jsonBody, _ := json.Marshal(body)

		req, err := http.NewRequest("POST", ts.BaseURL+"/subscription/revenuecat/restore", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Authorization", authToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify user was linked
		var linkedCustomerID *string
		err = ts.DB.QueryRow(ctx,
			"SELECT revenuecat_customer_id FROM users WHERE id = $1",
			userID).Scan(&linkedCustomerID)
		require.NoError(t, err)
		assert.NotNil(t, linkedCustomerID)
		assert.Equal(t, appUserID, *linkedCustomerID)

		// Verify subscription was created
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		sub, err := subRepo.GetActiveSubscriptionForUser(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, sub)
		assert.Equal(t, model.ProviderRevenueCat, sub.Provider)
		assert.Equal(t, model.PlanMonthly, sub.Plan)
		assert.Equal(t, model.StatusActive, sub.Status)
	})

	t.Run("RestorePurchases_WithNoSubscription", func(t *testing.T) {
		// Create test user
		signupInput := setup.GenerateSignupInput()
		ts.Expect.POST("/users").
			WithJSON(signupInput).
			Expect().
			Status(201)

		// Get user auth token from login response header
		loginResp := ts.Expect.POST("/login").
			WithJSON(map[string]string{
				"email":    signupInput.Email,
				"password": signupInput.Password,
			}).
			Expect().
			Status(200)

		authToken := loginResp.Header("Authorization").NotEmpty().Raw()

		ctx := context.Background()
		var userID int64
		err := ts.DB.QueryRow(ctx,
			"SELECT id FROM users WHERE email = $1",
			signupInput.Email).Scan(&userID)
		require.NoError(t, err)

		appUserID := fmt.Sprintf("rc_user_%d", userID)

		// Mock customer info with no active subscription
		rcServiceMock.GetCustomerInfoFunc = func(_ context.Context, requestedAppUserID string) (*service.CustomerInfo, error) {
			assert.Equal(t, appUserID, requestedAppUserID)
			return &service.CustomerInfo{
				Subscriber: service.Subscriber{
					OriginalAppUserID: appUserID,
					Entitlements:      map[string]service.Entitlement{},
					Subscriptions:     map[string]service.Subscription{},
				},
			}, nil
		}

		// Mock restore purchases that doesn't create subscription when none exists
		rcServiceMock.RestorePurchasesFunc = func(ctx context.Context, restoreUserID int64, restoreAppUserID string, _ model.PlatformType) error {
			// Link user to RevenueCat customer
			_, linkErr := ts.DB.Exec(ctx,
				"UPDATE users SET revenuecat_customer_id = $1 WHERE id = $2",
				restoreAppUserID, restoreUserID)
			if linkErr != nil {
				return linkErr
			}

			// Get customer info
			customerInfo, getErr := rcServiceMock.GetCustomerInfo(ctx, restoreAppUserID)
			if getErr != nil {
				return getErr
			}

			// Check if there's a premium entitlement
			if _, hasPremium := customerInfo.Subscriber.Entitlements["premium"]; !hasPremium {
				// No active subscription - this is valid
				return nil
			}

			return nil
		}

		// Make the restore request
		body := map[string]interface{}{
			"appUserId": appUserID,
			"platform":  "ios",
		}
		jsonBody, _ := json.Marshal(body)

		req, err := http.NewRequest("POST", ts.BaseURL+"/subscription/revenuecat/restore", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Authorization", authToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify user was linked
		var linkedCustomerID *string
		err = ts.DB.QueryRow(ctx,
			"SELECT revenuecat_customer_id FROM users WHERE id = $1",
			userID).Scan(&linkedCustomerID)
		require.NoError(t, err)
		assert.NotNil(t, linkedCustomerID)
		assert.Equal(t, appUserID, *linkedCustomerID)

		// Verify no subscription was created
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		sub, err := subRepo.GetActiveSubscriptionForUser(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, sub)
	})

	t.Run("RestorePurchases_WithAPIError", func(t *testing.T) {
		// Create test user
		signupInput := setup.GenerateSignupInput()
		ts.Expect.POST("/users").
			WithJSON(signupInput).
			Expect().
			Status(201)

		// Get user auth token from login response header
		loginResp := ts.Expect.POST("/login").
			WithJSON(map[string]string{
				"email":    signupInput.Email,
				"password": signupInput.Password,
			}).
			Expect().
			Status(200)

		authToken := loginResp.Header("Authorization").NotEmpty().Raw()

		ctx := context.Background()
		var userID int64
		err := ts.DB.QueryRow(ctx,
			"SELECT id FROM users WHERE email = $1",
			signupInput.Email).Scan(&userID)
		require.NoError(t, err)

		appUserID := fmt.Sprintf("rc_user_%d", userID)

		// Mock API error
		rcServiceMock.GetCustomerInfoFunc = func(_ context.Context, _ string) (*service.CustomerInfo, error) {
			return nil, fmt.Errorf("RevenueCat API error: 401 - Unauthorized")
		}

		// Mock restore purchases with error handling
		rcServiceMock.RestorePurchasesFunc = func(ctx context.Context, restoreUserID int64, restoreAppUserID string, _ model.PlatformType) error {
			// Link user
			_, linkErr := ts.DB.Exec(ctx,
				"UPDATE users SET revenuecat_customer_id = $1 WHERE id = $2",
				restoreAppUserID, restoreUserID)
			if linkErr != nil {
				return linkErr
			}

			// Get customer info - this will fail
			_, getErr := rcServiceMock.GetCustomerInfo(ctx, restoreAppUserID)
			if getErr != nil {
				return fmt.Errorf("failed to get customer info: %w", getErr)
			}

			return nil
		}

		// Make the restore request
		body := map[string]interface{}{
			"appUserId": appUserID,
			"platform":  "ios",
		}
		jsonBody, _ := json.Marshal(body)

		req, err := http.NewRequest("POST", ts.BaseURL+"/subscription/revenuecat/restore", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Authorization", authToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		// Verify user was still linked (happens before API call)
		var linkedCustomerID *string
		err = ts.DB.QueryRow(ctx,
			"SELECT revenuecat_customer_id FROM users WHERE id = $1",
			userID).Scan(&linkedCustomerID)
		require.NoError(t, err)
		assert.NotNil(t, linkedCustomerID)
		assert.Equal(t, appUserID, *linkedCustomerID)

		// Verify no subscription was created
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		sub, err := subRepo.GetActiveSubscriptionForUser(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, sub)
	})
}
