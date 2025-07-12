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
	"decorebator.com/tests/integration/mocks"
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

		// RevenueCat app_user_id is just the stringified user ID
		appUserID := fmt.Sprintf("%d", userID)

		// Create webhook payload
		webhook := model.RevenueCatWebhook{
			APIVersion: "1.0",
			Event: model.RevenueCatEvent{
				Type:                  "INITIAL_PURCHASE",
				ID:                    "test_event_" + time.Now().Format("20060102150405"),
				AppUserID:             appUserID,
				OriginalAppUserID:     appUserID,
				EventTimestampMS:      time.Now().Unix() * 1000,
				ProductID:             service.ProductMonthlyIOS,
				EntitlementIDs:        []string{service.EntitlementPremium},
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

		// Create a mock API client that returns test data without making external calls
		mockAPIClient := &mocks.MockRevenueCatAPIClient{
			GetCustomerInfoFunc: func(_ context.Context, appUserID string) (*service.CustomerInfo, error) {
				// Return a mock customer info with active subscription
				expiresDateStr := time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339)
				return &service.CustomerInfo{
					RequestDate:   time.Now().Format(time.RFC3339),
					RequestDateMS: time.Now().Unix() * 1000,
					Subscriber: service.Subscriber{
						OriginalAppUserID: appUserID,
						Entitlements: map[string]service.Entitlement{
							service.EntitlementPremium: {
								ProductIdentifier: service.ProductMonthlyIOS,
								PurchaseDate:      time.Now().Format(time.RFC3339),
								ExpiresDate:       &expiresDateStr,
							},
						},
						Subscriptions: map[string]service.Subscription{
							service.ProductMonthlyIOS: {
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

		// Create the real RevenueCat service with the mock API client
		rcServiceWithMockAPI := service.NewRevenueCatService(ts.DB, mockAPIClient)

		// Process the job using the service with mocked API client
		err = ts.ProcessRevenueCatWebhookJob(t, rcServiceWithMockAPI)
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
		require.NotNil(t, updatedUser.SubscriptionStatus)
		assert.Equal(t, model.StatusActive, *updatedUser.SubscriptionStatus)

		// Verify mock API client was called
		assert.Equal(t, 1, len(mockAPIClient.GetCustomerInfoCalls), "GetCustomerInfo should be called once")
		assert.Equal(t, appUserID, mockAPIClient.GetCustomerInfoCalls[0].AppUserID, "GetCustomerInfo should be called with correct app user ID")
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
		productID := service.ProductMonthlyIOS
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
	})

	t.Run("ExpirationEvent_WithBillingError_SetsPastDueStatus", func(t *testing.T) {
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

		// Create an active subscription first
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		subscription := &model.Subscription{
			UserID:                   userID,
			Provider:                 model.ProviderRevenueCat,
			RevenueCatSubscriptionID: &[]string{fmt.Sprintf("%d", userID)}[0],
			Plan:                     model.PlanMonthly,
			Status:                   model.StatusActive,
			CurrentPeriodStart:       time.Now().Add(-30 * 24 * time.Hour),
			CurrentPeriodEnd:         time.Now().Add(-1 * time.Hour), // Expired
			AmountCents:              model.SubscriptionPrices[model.PlanMonthly].AmountCents,
			Currency:                 "USD",
		}

		setupEvent := model.SubscriptionEvent{
			ExternalEventID: setup.GenerateUniqueEventID("setup_active_sub"),
			Provider:        model.ProviderRevenueCat,
			EventType:       "setup",
			EventData:       `{"type": "setup"}`,
		}
		_, err = subRepo.CreateSubscription(ctx, subscription, setupEvent)
		require.NoError(t, err)

		// RevenueCat app_user_id is just the stringified user ID
		appUserID := fmt.Sprintf("%d", userID)
		expirationReason := "BILLING_ERROR"

		// Create expiration webhook payload
		webhook := model.RevenueCatWebhook{
			APIVersion: "1.0",
			Event: model.RevenueCatEvent{
				Type:                  "EXPIRATION",
				ID:                    "test_expiration_" + time.Now().Format("20060102150405"),
				AppUserID:             appUserID,
				OriginalAppUserID:     appUserID,
				EventTimestampMS:      time.Now().Unix() * 1000,
				ProductID:             service.ProductMonthlyAndroid,
				EntitlementIDs:        []string{service.EntitlementPremium},
				Store:                 "PLAY_STORE",
				Environment:           "SANDBOX",
				ExpirationReason:      &expirationReason,
				PurchasedAtMS:         time.Now().Add(-30*24*time.Hour).Unix() * 1000,
				ExpirationAtMS:        time.Now().Add(-1*time.Hour).Unix() * 1000,
				PeriodType:            "NORMAL",
				Price:                 0.0, // Billing error, no payment processed
				Currency:              "USD",
				TransactionID:         "test_transaction_" + time.Now().Format("20060102150405"),
				OriginalTransactionID: "test_original_transaction_" + time.Now().Format("20060102150405"),
				CountryCode:           "US",
			},
		}

		// Process the expiration event directly through the service
		rcService := service.NewRevenueCatService(ts.DB, nil)
		err = rcService.ProcessRevenueCatEvent(ctx, webhook.Event, userID)
		require.NoError(t, err)

		// Verify subscription status changed to past due but is still accessible during grace period
		updatedSub, err := subRepo.GetActiveSubscriptionForUser(ctx, userID)
		require.NoError(t, err)
		assert.NotNil(t, updatedSub, "Should find past_due subscription during grace period")
		assert.Equal(t, model.StatusPastDue, updatedSub.Status)
		assert.Nil(t, updatedSub.CanceledAt, "CanceledAt should not be set for billing errors")
		assert.True(t, updatedSub.IsActive(), "Should be active during grace period")

		// Verify subscription event was recorded
		var eventCount int
		err = ts.DB.QueryRow(ctx,
			"SELECT COUNT(*) FROM subscription_events WHERE subscription_id = $1 AND event_type = 'EXPIRATION'",
			updatedSub.ID).Scan(&eventCount)
		require.NoError(t, err)
		assert.Equal(t, 1, eventCount)
	})

	t.Run("ExpirationEvent_WithVoluntaryCancellation_SetsCanceledStatus", func(t *testing.T) {
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

		// Create an active subscription first
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		subscription := &model.Subscription{
			UserID:                   userID,
			Provider:                 model.ProviderRevenueCat,
			RevenueCatSubscriptionID: &[]string{fmt.Sprintf("%d", userID)}[0],
			Plan:                     model.PlanMonthly,
			Status:                   model.StatusActive,
			CurrentPeriodStart:       time.Now().Add(-30 * 24 * time.Hour),
			CurrentPeriodEnd:         time.Now().Add(-1 * time.Hour), // Expired
			AmountCents:              model.SubscriptionPrices[model.PlanMonthly].AmountCents,
			Currency:                 "USD",
		}

		setupEvent := model.SubscriptionEvent{
			ExternalEventID: setup.GenerateUniqueEventID("setup_active_sub_2"),
			Provider:        model.ProviderRevenueCat,
			EventType:       "setup",
			EventData:       `{"type": "setup"}`,
		}
		_, err = subRepo.CreateSubscription(ctx, subscription, setupEvent)
		require.NoError(t, err)

		// RevenueCat app_user_id is just the stringified user ID
		appUserID := fmt.Sprintf("%d", userID)
		expirationReason := "VOLUNTARY"

		// Create expiration webhook payload
		webhook := model.RevenueCatWebhook{
			APIVersion: "1.0",
			Event: model.RevenueCatEvent{
				Type:                  "EXPIRATION",
				ID:                    "test_expiration_voluntary_" + time.Now().Format("20060102150405"),
				AppUserID:             appUserID,
				OriginalAppUserID:     appUserID,
				EventTimestampMS:      time.Now().Unix() * 1000,
				ProductID:             service.ProductMonthlyAndroid,
				EntitlementIDs:        []string{service.EntitlementPremium},
				Store:                 "PLAY_STORE",
				Environment:           "SANDBOX",
				ExpirationReason:      &expirationReason,
				PurchasedAtMS:         time.Now().Add(-30*24*time.Hour).Unix() * 1000,
				ExpirationAtMS:        time.Now().Add(-1*time.Hour).Unix() * 1000,
				PeriodType:            "NORMAL",
				Price:                 6.99,
				Currency:              "USD",
				TransactionID:         "test_transaction_vol_" + time.Now().Format("20060102150405"),
				OriginalTransactionID: "test_original_transaction_vol_" + time.Now().Format("20060102150405"),
				CountryCode:           "US",
			},
		}

		// Process the expiration event directly through the service
		rcService := service.NewRevenueCatService(ts.DB, nil)
		err = rcService.ProcessRevenueCatEvent(ctx, webhook.Event, userID)
		require.NoError(t, err)

		// Verify subscription status changed to canceled
		updatedSub, err := subRepo.GetActiveSubscriptionForUser(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, updatedSub, "Should not have active subscription after voluntary cancellation")

		// Get all subscriptions for user to verify the canceled one
		var canceledSub model.Subscription
		err = ts.DB.QueryRow(ctx,
			"SELECT id, status, canceled_at FROM subscriptions WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1",
			userID).Scan(&canceledSub.ID, &canceledSub.Status, &canceledSub.CanceledAt)
		require.NoError(t, err)
		assert.Equal(t, model.StatusCanceled, canceledSub.Status)
		assert.NotNil(t, canceledSub.CanceledAt, "CanceledAt should be set for voluntary cancellation")

		// Verify subscription event was recorded
		var eventCount int
		err = ts.DB.QueryRow(ctx,
			"SELECT COUNT(*) FROM subscription_events WHERE subscription_id = $1 AND event_type = 'EXPIRATION'",
			canceledSub.ID).Scan(&eventCount)
		require.NoError(t, err)
		assert.Equal(t, 1, eventCount)
	})

	t.Run("GracePeriod_PastDueSubscription_AccessibleDuringGracePeriod", func(t *testing.T) {
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

		// Create a past_due subscription that should still be in grace period
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		revenuecatSubID := fmt.Sprintf("%d", userID)
		subscription := &model.Subscription{
			UserID:                   userID,
			Provider:                 model.ProviderRevenueCat,
			RevenueCatSubscriptionID: &revenuecatSubID,
			Plan:                     model.PlanMonthly,
			Status:                   model.StatusPastDue,
			CurrentPeriodStart:       time.Now().Add(-30 * 24 * time.Hour),
			CurrentPeriodEnd:         time.Now().Add(-1 * time.Hour), // Expired 1 hour ago (within 3-day grace period)
			AmountCents:              model.SubscriptionPrices[model.PlanMonthly].AmountCents,
			Currency:                 "USD",
		}

		setupEvent := model.SubscriptionEvent{
			ExternalEventID: setup.GenerateUniqueEventID("grace_period_test"),
			Provider:        model.ProviderRevenueCat,
			EventType:       "test_setup",
			EventData:       `{"type": "grace_period_test"}`,
		}
		_, err = subRepo.CreateSubscription(ctx, subscription, setupEvent)
		require.NoError(t, err)

		// Verify subscription is accessible during grace period
		activeSub, err := subRepo.GetActiveSubscriptionForUser(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, activeSub, "Should find subscription during grace period")
		assert.Equal(t, model.StatusPastDue, activeSub.Status)
		assert.True(t, activeSub.IsActive(), "Should be active during grace period")
	})

	t.Run("GracePeriod_ExpiredGracePeriod_NoAccess", func(t *testing.T) {
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

		// Create a past_due subscription that is beyond grace period
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		revenuecatSubID := fmt.Sprintf("%d", userID)
		subscription := &model.Subscription{
			UserID:                   userID,
			Provider:                 model.ProviderRevenueCat,
			RevenueCatSubscriptionID: &revenuecatSubID,
			Plan:                     model.PlanMonthly,
			Status:                   model.StatusPastDue,
			CurrentPeriodStart:       time.Now().Add(-35 * 24 * time.Hour),
			CurrentPeriodEnd:         time.Now().Add(-4 * 24 * time.Hour), // Expired 4 days ago (beyond 3-day grace period)
			AmountCents:              model.SubscriptionPrices[model.PlanMonthly].AmountCents,
			Currency:                 "USD",
		}

		setupEvent := model.SubscriptionEvent{
			ExternalEventID: setup.GenerateUniqueEventID("expired_grace_test"),
			Provider:        model.ProviderRevenueCat,
			EventType:       "test_setup",
			EventData:       `{"type": "expired_grace_test"}`,
		}
		_, err = subRepo.CreateSubscription(ctx, subscription, setupEvent)
		require.NoError(t, err)

		// Verify subscription is NOT accessible beyond grace period
		activeSub, err := subRepo.GetActiveSubscriptionForUser(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, activeSub, "Should not find subscription beyond grace period")

		// Verify IsActive returns false
		var dbSub model.Subscription
		err = ts.DB.QueryRow(ctx,
			"SELECT id, status, current_period_end FROM subscriptions WHERE user_id = $1",
			userID).Scan(&dbSub.ID, &dbSub.Status, &dbSub.CurrentPeriodEnd)
		require.NoError(t, err)
		dbSub.Status = model.StatusPastDue // Set the status we know it should have
		assert.False(t, dbSub.IsActive(), "Should not be active beyond grace period")
	})
}
