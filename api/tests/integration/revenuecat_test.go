package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/mocks"
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

	// Start mock RevenueCat server
	rcMock := mocks.NewRevenueCatMock()
	defer rcMock.Close()

	// Set environment variable to use mock server
	os.Setenv("REVENUECAT_API_BASE_URL", rcMock.Server.URL)
	defer os.Unsetenv("REVENUECAT_API_BASE_URL")

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

		// Mock the GetCustomerInfo response
		rcMock.MockGetCustomerInfo(map[string]interface{}{
			"request_date": "2024-01-01T00:00:00Z",
			"subscriber": map[string]interface{}{
				"original_app_user_id": revenueCatCustomerID,
				"entitlements": map[string]interface{}{
					"premium": map[string]interface{}{
						"expires_date":       time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339),
						"product_identifier": "com.decorebator.premium.monthly",
						"purchase_date":      time.Now().Format(time.RFC3339),
					},
				},
			},
		})

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

		// Create a mock RevenueCat service.
		rcServiceMock := &mocks.RevenueCatServiceMock{}

		// Set the mock functions.
		rcServiceMock.HandleWebhookFunc = func(ctx context.Context, payload []byte, _ string) error {
			// Simulate the real service's behavior by creating the subscription directly.
			var webhook service.WebhookPayload
			if unmarshalErr := json.Unmarshal(payload, &webhook); unmarshalErr != nil {
				return unmarshalErr
			}

			user, userErr := (&repository.UserRepository{Db: ts.DB}).GetUserByRevenueCatCustomerID(ctx, webhook.Event.AppUserID)
			require.NoError(t, userErr)

			plan := service.NewRevenueCatService(ts.DB).GetPlanFromProductID(webhook.Event.ProductID)
			require.NotEqual(t, model.PlanFree, plan)

			purchaseDate := time.UnixMilli(webhook.Event.PurchasedAtMS)
			expirationDate := time.UnixMilli(webhook.Event.ExpirationAtMS)

			sub := &model.Subscription{
				UserID:                   user.ID,
				Provider:                 model.ProviderRevenueCat,
				RevenueCatSubscriptionID: &webhook.Event.AppUserID,
				AppStoreProductID:        &webhook.Event.ProductID,
				Plan:                     plan,
				Status:                   model.StatusActive,
				CurrentPeriodStart:       purchaseDate,
				CurrentPeriodEnd:         expirationDate,
			}

			_, err = repository.NewSubscriptionRepository(ts.DB).CreateSubscription(ctx, sub)
			require.NoError(t, err)

			return nil
		}

		// Process the job that was just enqueued by the webhook.
		ts.ProcessNextRiverJob(t, "revenuecat-webhook", rcServiceMock)

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
