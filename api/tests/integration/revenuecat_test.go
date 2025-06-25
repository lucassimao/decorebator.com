package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRevenueCatIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Initialize test server
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	// Create test user and get auth token
	authToken := ts.WithTestUser(t)

	// Provider detection tests removed - now handled locally in mobile app

	t.Run("RevenueCatWebhook_ProcessesInitialPurchase", func(t *testing.T) {
		// Create a test user to link with RevenueCat
		signupInput := setup.GenerateSignupInput()
		ts.Expect.POST("/users").
			WithJSON(signupInput).
			Expect().
			Status(201)

		// Get user ID from database
		ctx := context.Background()
		var userID int64
		err := ts.DB.QueryRow(ctx,
			"SELECT id FROM users WHERE email = $1",
			signupInput.Email).Scan(&userID)
		require.NoError(t, err)

		// First, link the user to a RevenueCat customer ID
		revenueCatCustomerID := fmt.Sprintf("rc_user_%d", userID)
		_, err = ts.DB.Exec(ctx,
			"UPDATE users SET revenuecat_customer_id = $1 WHERE id = $2",
			revenueCatCustomerID, userID)
		require.NoError(t, err)

		// Create webhook payload following exact RevenueCat structure
		webhook := map[string]interface{}{
			"api_version": "1.0",
			"event": map[string]interface{}{
				"type":                    "INITIAL_PURCHASE",
				"id":                      "test_event_123",
				"app_id":                  "app_123",
				"app_user_id":             revenueCatCustomerID,
				"original_app_user_id":    revenueCatCustomerID,
				"event_timestamp_ms":      time.Now().Unix() * 1000,
				"product_id":              "com.decorebator.premium.monthly",
				"entitlement_ids":         []string{"premium"},
				"store":                   "APP_STORE",
				"environment":             "PRODUCTION",
				"purchased_at_ms":         time.Now().Unix() * 1000,
				"expiration_at_ms":        time.Now().Add(30*24*time.Hour).Unix() * 1000,
				"period_type":             "NORMAL",
				"price":                   6.99,
				"currency":                "USD",
				"transaction_id":          "test_transaction_123",
				"original_transaction_id": "test_original_transaction_123",
				"country_code":            "US",
			},
		}

		jsonBody, _ := json.Marshal(webhook)

		// Get webhook secret from config (would be set in test config)
		webhookSecret := "test_webhook_secret"

		req, err := http.NewRequest("POST", ts.BaseURL+"/webhook/revenuecat", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Authorization", webhookSecret)
		req.Header.Set("Content-Type", "application/json")

		resp2, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp2.Body.Close()

		// RevenueCat webhooks always return 200 to prevent retries
		assert.Equal(t, http.StatusOK, resp2.StatusCode)

		// Verify subscription was created
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		sub, err := subRepo.GetActiveSubscriptionForUser(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, sub)
		assert.Equal(t, model.ProviderRevenueCat, sub.Provider)
		assert.Equal(t, model.PlanMonthly, sub.Plan)
	})

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

	t.Run("RestorePurchases_WithAuthentication_Succeeds", func(t *testing.T) {
		body := map[string]interface{}{
			"appUserId": "test_user_123",
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

		// Should succeed with authentication
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func init() {
	// Ensure database connection is initialized
	_, err := common.GetDBConnection()
	if err != nil {
		panic("Failed to initialize database connection: " + err.Error())
	}
}
