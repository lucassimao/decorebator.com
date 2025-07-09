package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRevenueCatWebhookEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Initialize test server without mock - use default configuration
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	t.Run("WebhookEndpoint_RequiresAuthentication", func(t *testing.T) {
		webhook := model.RevenueCatWebhook{
			APIVersion: "1.0",
			Event: model.RevenueCatEvent{
				Type:      "INITIAL_PURCHASE",
				ID:        "test_event_123",
				AppUserID: "test_user",
			},
		}

		jsonBody, _ := json.Marshal(webhook)
		req, err := http.NewRequest("POST", ts.BaseURL+"/webhook/revenuecat", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return 401 without authentication
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("WebhookEndpoint_WithValidAuth_Returns200", func(t *testing.T) {
		// Get webhook auth token from environment
		webhookAuthToken := os.Getenv("REVENUECAT_WEBHOOK_AUTHORIZATION")
		require.NotEmpty(t, webhookAuthToken, "REVENUECAT_WEBHOOK_AUTHORIZATION must be set in .env.test")

		webhook := model.RevenueCatWebhook{
			APIVersion: "1.0",
			Event: model.RevenueCatEvent{
				Type:      "INITIAL_PURCHASE",
				ID:        "test_event_123",
				AppUserID: "test_user",
			},
		}

		jsonBody, _ := json.Marshal(webhook)
		req, err := http.NewRequest("POST", ts.BaseURL+"/webhook/revenuecat", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Authorization", webhookAuthToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// RevenueCat webhooks always return 200
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("WebhookEndpoint_EnqueuesJob", func(t *testing.T) {
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

		// Set up RevenueCat customer ID
		revenueCatCustomerID := "rc_test_user_" + time.Now().Format("20060102150405")
		_, err = ts.DB.Exec(ctx,
			"UPDATE users SET revenuecat_customer_id = $1 WHERE id = $2",
			revenueCatCustomerID, userID)
		require.NoError(t, err)

		// Get webhook auth token
		webhookAuthToken := os.Getenv("REVENUECAT_WEBHOOK_AUTHORIZATION")
		require.NotEmpty(t, webhookAuthToken)

		// Create webhook payload
		webhook := model.RevenueCatWebhook{
			APIVersion: "1.0",
			Event: model.RevenueCatEvent{
				Type:                  "INITIAL_PURCHASE",
				ID:                    "test_event_" + time.Now().Format("20060102150405"),
				AppUserID:             revenueCatCustomerID,
				OriginalAppUserID:     revenueCatCustomerID,
				EventTimestampMS:      time.Now().Unix() * 1000,
				ProductID:             "com.decorebator.premium.monthly",
				EntitlementIDs:        []string{"premium"},
				Store:                 "APP_STORE",
				Environment:           "PRODUCTION",
				PurchasedAtMS:         time.Now().Unix() * 1000,
				ExpirationAtMS:        time.Now().Add(30*24*time.Hour).Unix() * 1000,
				PeriodType:            "NORMAL",
				Price:                 6.99,
				Currency:              "USD",
				TransactionID:         "test_transaction_" + time.Now().Format("20060102150405"),
				OriginalTransactionID: "test_original_transaction_" + time.Now().Format("20060102150405"),
				CountryCode:           "US",
			},
		}

		jsonBody, _ := json.Marshal(webhook)
		req, err := http.NewRequest("POST", ts.BaseURL+"/webhook/revenuecat", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Authorization", webhookAuthToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Check that a job was enqueued
		var jobCount int
		err = ts.DB.QueryRow(ctx,
			"SELECT COUNT(*) FROM river_job WHERE kind = 'revenuecat-webhook' AND state = 'available'").
			Scan(&jobCount)
		require.NoError(t, err)
		assert.Greater(t, jobCount, 0, "Should have enqueued at least one job")
	})
}

// TestRevenueCatSubscriptionCreation tests subscription creation without webhook processing
func TestRevenueCatSubscriptionCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	t.Run("CreateSubscription_FromRevenueCat", func(t *testing.T) {
		// Create a test user
		signupInput := setup.GenerateSignupInput()
		ts.Expect.POST("/users").
			WithJSON(signupInput).
			Expect().
			Status(201)

		ctx := context.Background()
		var userID int64
		err := ts.DB.QueryRow(ctx,
			"SELECT id FROM users WHERE email = $1",
			signupInput.Email).Scan(&userID)
		require.NoError(t, err)

		// Create subscription directly using repository
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		revenuecatSubID := "rc_sub_test_" + time.Now().Format("20060102150405")
		productID := "com.decorebator.premium.monthly"
		platform := model.PlatformIOS

		subscription := &model.Subscription{
			UserID:                   userID,
			Provider:                 model.ProviderRevenueCat,
			RevenueCatSubscriptionID: &revenuecatSubID,
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

		// Create test setup event
		testEvent := model.SubscriptionEvent{
			ExternalEventID: setup.GenerateUniqueEventID("test_setup"),
			Provider:        model.ProviderRevenueCat,
			EventType:       "test_setup",
			EventData:       `{"type": "test_setup", "source": "integration_test"}`,
		}
		subID, err := subRepo.CreateSubscription(ctx, subscription, testEvent)
		assert.NoError(t, err)
		assert.Greater(t, subID, int64(0))

		// Verify subscription was created
		createdSub, err := subRepo.GetActiveSubscriptionForUser(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, createdSub)
		assert.Equal(t, model.ProviderRevenueCat, createdSub.Provider)
		assert.Equal(t, model.PlanMonthly, createdSub.Plan)
		assert.Equal(t, model.StatusActive, createdSub.Status)
		assert.Equal(t, 699, createdSub.AmountCents)
	})

	t.Run("UpdateSubscription_Renewal", func(t *testing.T) {
		// Create a test user with existing subscription
		signupInput := setup.GenerateSignupInput()
		ts.Expect.POST("/users").
			WithJSON(signupInput).
			Expect().
			Status(201)

		ctx := context.Background()
		var userID int64
		err := ts.DB.QueryRow(ctx,
			"SELECT id FROM users WHERE email = $1",
			signupInput.Email).Scan(&userID)
		require.NoError(t, err)

		// Create initial subscription
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		revenuecatSubID := "rc_sub_renewal_" + time.Now().Format("20060102150405")
		productID := "com.decorebator.premium.monthly"
		platform := model.PlatformIOS

		subscription := &model.Subscription{
			UserID:                   userID,
			Provider:                 model.ProviderRevenueCat,
			RevenueCatSubscriptionID: &revenuecatSubID,
			AppStoreProductID:        &productID,
			Platform:                 &platform,
			Plan:                     model.PlanMonthly,
			Status:                   model.StatusActive,
			CurrentPeriodStart:       time.Now().Add(-30 * 24 * time.Hour),
			CurrentPeriodEnd:         time.Now().Add(-1 * time.Hour), // Expired
			CancelAtPeriodEnd:        false,
			AmountCents:              699,
			Currency:                 "USD",
		}

		// Create test setup event
		testEvent := model.SubscriptionEvent{
			ExternalEventID: setup.GenerateUniqueEventID("test_setup"),
			Provider:        model.ProviderRevenueCat,
			EventType:       "test_setup",
			EventData:       `{"type": "test_setup", "source": "integration_test"}`,
		}
		_, err = subRepo.CreateSubscription(ctx, subscription, testEvent)
		require.NoError(t, err)

		// Update subscription with new period
		existing, err := subRepo.GetActiveSubscriptionForUser(ctx, userID)
		require.NoError(t, err)
		require.NotNil(t, existing)

		existing.CurrentPeriodStart = time.Now()
		existing.CurrentPeriodEnd = time.Now().Add(30 * 24 * time.Hour)
		existing.Status = model.StatusActive

		// Create test update event
		updateEvent := model.SubscriptionEvent{
			SubscriptionID:  existing.ID,
			ExternalEventID: setup.GenerateUniqueEventID("test_update"),
			Provider:        model.ProviderRevenueCat,
			EventType:       "test_update",
			EventData:       `{"type": "test_update", "source": "integration_test"}`,
		}
		err = subRepo.UpdateSubscription(ctx, existing, updateEvent)
		assert.NoError(t, err)

		// Verify update
		updated, err := subRepo.GetActiveSubscriptionForUser(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.True(t, updated.CurrentPeriodEnd.After(time.Now()))
	})
}
