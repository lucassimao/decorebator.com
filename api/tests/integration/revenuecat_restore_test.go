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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestorePurchases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	t.Run("RestorePurchases_RequiresAuthentication", func(t *testing.T) {
		// Initialize test server
		ts := setup.NewTestServer(t)
		defer ts.Cleanup()

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
		// Create a mock API client with predefined response
		expiresDateStr := time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339)
		mockAPIClient := &mocks.MockRevenueCatAPIClient{
			GetCustomerInfoFunc: func(_ context.Context, appUserID string) (*service.CustomerInfo, error) {
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
			},
		}

		// Create test server with a config function that creates the service with the mock API client
		ts := setup.NewTestServer(t, func(db *pgxpool.Pool) *setup.TestConfig {
			return &setup.TestConfig{
				Database:          db,
				RevenueCatService: service.NewRevenueCatService(db, mockAPIClient),
			}
		})
		defer ts.Cleanup()

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
		// Create a mock API client that returns no subscription
		mockAPIClient := &mocks.MockRevenueCatAPIClient{
			GetCustomerInfoFunc: func(_ context.Context, appUserID string) (*service.CustomerInfo, error) {
				return &service.CustomerInfo{
					Subscriber: service.Subscriber{
						OriginalAppUserID: appUserID,
						Entitlements:      map[string]service.Entitlement{},
						Subscriptions:     map[string]service.Subscription{},
					},
				}, nil
			},
		}

		// Create test server with a config function that creates the service with the mock API client
		ts := setup.NewTestServer(t, func(db *pgxpool.Pool) *setup.TestConfig {
			return &setup.TestConfig{
				Database:          db,
				RevenueCatService: service.NewRevenueCatService(db, mockAPIClient),
			}
		})
		defer ts.Cleanup()

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
		// Create a mock API client that returns an error
		mockAPIClient := &mocks.MockRevenueCatAPIClient{
			GetCustomerInfoFunc: func(_ context.Context, _ string) (*service.CustomerInfo, error) {
				return nil, fmt.Errorf("RevenueCat API error: 401 - Unauthorized")
			},
		}

		// Create test server with a config function that creates the service with the mock API client
		ts := setup.NewTestServer(t, func(db *pgxpool.Pool) *setup.TestConfig {
			return &setup.TestConfig{
				Database:          db,
				RevenueCatService: service.NewRevenueCatService(db, mockAPIClient),
			}
		})
		defer ts.Cleanup()

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
