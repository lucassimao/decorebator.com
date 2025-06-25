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

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRevenueCatWebhookSimple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	t.Run("Webhook_CreatesSubscription_DirectApproach", func(t *testing.T) {
		// Create a test user
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
		revenueCatCustomerID := fmt.Sprintf("rc_user_%d", userID)
		_, err = ts.DB.Exec(ctx,
			"UPDATE users SET revenuecat_customer_id = $1 WHERE id = $2",
			revenueCatCustomerID, userID)
		require.NoError(t, err)

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

		// Get webhook auth token from environment
		webhookAuthToken := os.Getenv("REVENUECAT_WEBHOOK_AUTHORIZATION")
		require.NotEmpty(t, webhookAuthToken, "REVENUECAT_WEBHOOK_AUTHORIZATION must be set in .env.test")

		// Send webhook
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

		// Create the RevenueCat service with the test database
		rcService := service.NewRevenueCatService(ts.DB, nil)
		
		// Process the job using our new simplified method
		err = ts.ProcessRevenueCatWebhookJob(t, rcService)
		require.NoError(t, err)
		
		// Verify the subscription was created
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		subscription, err := subRepo.GetActiveSubscriptionForUser(ctx, userID)
		require.NoError(t, err)
		require.NotNil(t, subscription)
		
		// Verify subscription details
		assert.Equal(t, model.ProviderRevenueCat, subscription.Provider)
		assert.Equal(t, model.PlanMonthly, subscription.Plan)
		assert.Equal(t, model.StatusActive, subscription.Status)
		assert.NotNil(t, subscription.RevenueCatSubscriptionID)
		
		// Verify user's subscription status was updated
		var updatedUser model.User
		err = ts.DB.QueryRow(ctx,
			"SELECT subscription_plan, subscription_status FROM users WHERE id = $1",
			userID).Scan(&updatedUser.SubscriptionPlan, &updatedUser.SubscriptionStatus)
		require.NoError(t, err)
		assert.Equal(t, model.PlanMonthly, updatedUser.SubscriptionPlan)
		assert.Equal(t, model.StatusActive, updatedUser.SubscriptionStatus)
	})

	t.Run("DirectSubscriptionCreation_Works", func(t *testing.T) {
		// This test verifies that our subscription creation logic works
		// without going through the webhook flow

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

		// Create subscription directly
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

		subID, err := subRepo.CreateSubscription(ctx, subscription)
		assert.NoError(t, err)
		assert.Greater(t, subID, int64(0))

		// Verify subscription was created
		createdSub, err := subRepo.GetActiveSubscriptionForUser(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, createdSub)
		assert.Equal(t, model.ProviderRevenueCat, createdSub.Provider)
		assert.Equal(t, model.PlanMonthly, createdSub.Plan)
		assert.Equal(t, model.StatusActive, createdSub.Status)
	})
}
