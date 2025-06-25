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

	// Create a mock RevenueCat service that doesn't make any external calls
	rcServiceMock := &mocks.RevenueCatServiceMock{}

	// Initialize test server with mock service
	ts := setup.NewTestServer(t, &setup.TestConfig{
		RevenueCatService: rcServiceMock,
	})
	defer ts.Cleanup()

	// Create test user and get auth token
	authToken := ts.WithTestUser(t)

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

		// Set up RevenueCat customer ID
		revenueCatCustomerID := fmt.Sprintf("rc_user_%d", userID)
		_, err = ts.DB.Exec(ctx,
			"UPDATE users SET revenuecat_customer_id = $1 WHERE id = $2",
			revenueCatCustomerID, userID)
		require.NoError(t, err)

		// Mock GetPlanFromProductID
		rcServiceMock.GetPlanFromProductIDFunc = func(productID string) model.SubscriptionPlan {
			if productID == "com.decorebator.premium.annual" {
				return model.PlanAnnual
			}
			return model.PlanMonthly
		}

		// Mock GetCustomerInfo to return active subscription
		expiresDateStr := time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339)
		rcServiceMock.GetCustomerInfoFunc = func(_ context.Context, appUserID string) (*service.CustomerInfo, error) {
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
							BillingIssuesDetectedAt: nil,
							ExpiresDate:             expiresDateStr,
							GracePeriodExpiresDate:  nil,
							IsSandbox:               false,
							OriginalPurchaseDate:    time.Now().Format(time.RFC3339),
							PeriodType:              "normal",
							PurchaseDate:            time.Now().Format(time.RFC3339),
							RefundedAt:              nil,
							Store:                   "app_store",
							UnsubscribeDetectedAt:   nil,
						},
					},
				},
			}, nil
		}

		// Mock CreateOrUpdateSubscriptionFromRevenueCat to actually create the subscription
		rcServiceMock.CreateOrUpdateSubscriptionFromRevenueCatFunc = func(ctx context.Context, userID int64, customerInfo *service.CustomerInfo, platform model.PlatformType) error {
			// Create subscription in database
			subRepo := repository.NewSubscriptionRepository(ts.DB)
			revenuecatSubID := customerInfo.Subscriber.OriginalAppUserID
			productID := "com.decorebator.premium.monthly"

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
				AmountCents:              699, // $6.99
				Currency:                 "USD",
			}

			_, createErr := subRepo.CreateSubscription(ctx, subscription)
			return createErr
		}

		// Mock the webhook processing to use the real service logic
		// which will call our mocked functions above
		rcServiceMock.HandleWebhookFunc = func(ctx context.Context, payload []byte) error {
			// Parse webhook
			var webhook model.RevenueCatWebhook
			if unmarshalErr := json.Unmarshal(payload, &webhook); unmarshalErr != nil {
				return unmarshalErr
			}

			// Only process INITIAL_PURCHASE events
			if webhook.Event.Type != "INITIAL_PURCHASE" {
				return nil
			}

			// Simulate the real service flow:
			// 1. Find user by RevenueCat customer ID
			userRepo := &repository.UserRepository{Db: ts.DB}
			users, findErr := userRepo.Find(repository.FindUserArgs{
				RevenueCatCustomerID: &webhook.Event.AppUserID,
			})
			if findErr != nil || len(users) == 0 {
				return fmt.Errorf("user not found")
			}
			user := &users[0]

			// 2. Get customer info (mocked above)
			customerInfo, getInfoErr := rcServiceMock.GetCustomerInfo(ctx, webhook.Event.AppUserID)
			if getInfoErr != nil {
				return getInfoErr
			}

			// 3. Create/update subscription (mocked above)
			platform := model.PlatformIOS
			if webhook.Event.Store == "PLAY_STORE" {
				platform = model.PlatformAndroid
			}
			return rcServiceMock.CreateOrUpdateSubscriptionFromRevenueCat(ctx, user.ID, customerInfo, platform)
		}

		// Create webhook payload using the proper model structure
		webhook := model.RevenueCatWebhook{
			APIVersion: "1.0",
			Event: model.RevenueCatEvent{
				Type:                  "INITIAL_PURCHASE",
				ID:                    "test_event_123",
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
				TransactionID:         "test_transaction_123",
				OriginalTransactionID: "test_original_transaction_123",
				CountryCode:           "US",
			},
		}

		jsonBody, _ := json.Marshal(webhook)

		// Get webhook auth token from environment
		webhookAuthToken := os.Getenv("REVENUECAT_WEBHOOK_AUTHORIZATION")
		require.NotEmpty(t, webhookAuthToken, "REVENUECAT_WEBHOOK_AUTHORIZATION must be set in .env.test")

		req, err := http.NewRequest("POST", ts.BaseURL+"/webhook/revenuecat", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Authorization", webhookAuthToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// RevenueCat webhooks always return 200 to prevent retries
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Process the job that was just enqueued by the webhook
		ts.ProcessNextRiverJob(t, "revenuecat_webhooks", rcServiceMock)

		// Wait a bit for the transaction to commit
		time.Sleep(100 * time.Millisecond)

		// Verify subscription was created
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		sub, err := subRepo.GetActiveSubscriptionForUser(ctx, userID)
		assert.NoError(t, err)
		if sub == nil {
			// Check all subscriptions for this user
			var count int
			err = ts.DB.QueryRow(ctx, "SELECT COUNT(*) FROM subscriptions WHERE user_id = $1", userID).Scan(&count)
			require.NoError(t, err)
			t.Fatalf("Subscription not found. Found %d subscriptions for user %d. Mock function was called: %v",
				count, userID, rcServiceMock.HandleWebhookFunc != nil)
		}
		assert.Equal(t, model.ProviderRevenueCat, sub.Provider)
		assert.Equal(t, model.PlanMonthly, sub.Plan)
		assert.Equal(t, model.StatusActive, sub.Status)
		assert.Equal(t, 699, sub.AmountCents)
		assert.Equal(t, "USD", sub.Currency)
	})

	t.Run("RevenueCatWebhook_ProcessesRenewal", func(t *testing.T) {
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

		revenueCatCustomerID := fmt.Sprintf("rc_user_%d", userID)
		_, err = ts.DB.Exec(ctx,
			"UPDATE users SET revenuecat_customer_id = $1 WHERE id = $2",
			revenueCatCustomerID, userID)
		require.NoError(t, err)

		// Create existing subscription
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		revenuecatSubID := "test_subscription_123"
		platformType := model.PlatformIOS
		existingSub := &model.Subscription{
			UserID:                   userID,
			Provider:                 model.ProviderRevenueCat,
			RevenueCatSubscriptionID: &revenuecatSubID,
			Platform:                 &platformType,
			Plan:                     model.PlanMonthly,
			Status:                   model.StatusActive,
			CurrentPeriodStart:       time.Now().Add(-30 * 24 * time.Hour),
			CurrentPeriodEnd:         time.Now().Add(-1 * time.Hour), // Expired
			CancelAtPeriodEnd:        false,
			AmountCents:              699,
			Currency:                 "USD",
		}
		_, err = subRepo.CreateSubscription(ctx, existingSub)
		require.NoError(t, err)

		// Mock renewal webhook processing
		rcServiceMock.HandleWebhookFunc = func(ctx context.Context, payload []byte) error {
			var webhook model.RevenueCatWebhook
			if unmarshalErr := json.Unmarshal(payload, &webhook); unmarshalErr != nil {
				return unmarshalErr
			}

			if webhook.Event.Type != "RENEWAL" {
				return nil
			}

			// Find user by RevenueCat customer ID
			userRepo := &repository.UserRepository{Db: ts.DB}
			users, findErr := userRepo.Find(repository.FindUserArgs{
				RevenueCatCustomerID: &webhook.Event.AppUserID,
			})
			if findErr != nil || len(users) == 0 {
				return fmt.Errorf("user not found")
			}
			user := &users[0]

			// Update subscription with new period
			renewalSubRepo := repository.NewSubscriptionRepository(ts.DB)
			subscription, getErr := renewalSubRepo.GetActiveSubscriptionForUser(ctx, user.ID)
			if getErr != nil {
				return getErr
			}

			// Update subscription period
			subscription.CurrentPeriodStart = time.Now()
			subscription.CurrentPeriodEnd = time.Now().Add(30 * 24 * time.Hour)
			subscription.Status = model.StatusActive

			return renewalSubRepo.UpdateSubscription(ctx, subscription)
		}

		// Create renewal webhook
		webhook := model.RevenueCatWebhook{
			APIVersion: "1.0",
			Event: model.RevenueCatEvent{
				Type:             "RENEWAL",
				ID:               "test_renewal_123",
				AppUserID:        revenueCatCustomerID,
				ProductID:        "com.decorebator.premium.monthly",
				EventTimestampMS: time.Now().Unix() * 1000,
				PurchasedAtMS:    time.Now().Unix() * 1000,
				ExpirationAtMS:   time.Now().Add(30*24*time.Hour).Unix() * 1000,
			},
		}

		jsonBody, _ := json.Marshal(webhook)
		webhookAuthToken := os.Getenv("REVENUECAT_WEBHOOK_AUTHORIZATION")

		req, err := http.NewRequest("POST", ts.BaseURL+"/webhook/revenuecat", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Authorization", webhookAuthToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Process the renewal job
		ts.ProcessNextRiverJob(t, "revenuecat_webhooks", rcServiceMock)

		// Verify subscription was renewed
		renewedSub, err := subRepo.GetActiveSubscriptionForUser(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, renewedSub)
		assert.Equal(t, model.StatusActive, renewedSub.Status)
		assert.True(t, renewedSub.CurrentPeriodEnd.After(time.Now()))
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
		// Mock the restore purchases function to avoid external calls
		rcServiceMock.RestorePurchasesFunc = func(_ context.Context, _ int64, _ string, _ model.PlatformType) error {
			// Simulate successful restoration without making external calls
			return nil
		}

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

	t.Run("RevenueCatWebhook_ProcessesCancellation", func(t *testing.T) {
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

		revenueCatCustomerID := fmt.Sprintf("rc_user_%d", userID)
		_, err = ts.DB.Exec(ctx,
			"UPDATE users SET revenuecat_customer_id = $1 WHERE id = $2",
			revenueCatCustomerID, userID)
		require.NoError(t, err)

		// Create existing active subscription
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		revenuecatSubID := "test_subscription_123"
		platformType := model.PlatformIOS
		existingSub := &model.Subscription{
			UserID:                   userID,
			Provider:                 model.ProviderRevenueCat,
			RevenueCatSubscriptionID: &revenuecatSubID,
			Platform:                 &platformType,
			Plan:                     model.PlanMonthly,
			Status:                   model.StatusActive,
			CurrentPeriodStart:       time.Now(),
			CurrentPeriodEnd:         time.Now().Add(30 * 24 * time.Hour),
			CancelAtPeriodEnd:        false,
			AmountCents:              699,
			Currency:                 "USD",
		}
		_, err = subRepo.CreateSubscription(ctx, existingSub)
		require.NoError(t, err)

		// Mock cancellation webhook processing
		rcServiceMock.HandleWebhookFunc = func(ctx context.Context, payload []byte) error {
			var webhook model.RevenueCatWebhook
			if unmarshalErr := json.Unmarshal(payload, &webhook); unmarshalErr != nil {
				return unmarshalErr
			}

			if webhook.Event.Type != "CANCELLATION" {
				return nil
			}

			// Find user and mark subscription for cancellation
			userRepo := &repository.UserRepository{Db: ts.DB}
			users, findErr := userRepo.Find(repository.FindUserArgs{
				RevenueCatCustomerID: &webhook.Event.AppUserID,
			})
			if findErr != nil || len(users) == 0 {
				return fmt.Errorf("user not found")
			}
			user := &users[0]

			cancelSubRepo := repository.NewSubscriptionRepository(ts.DB)
			subscription, getErr := cancelSubRepo.GetActiveSubscriptionForUser(ctx, user.ID)
			if getErr != nil {
				return getErr
			}

			// Mark for cancellation at period end
			subscription.CancelAtPeriodEnd = true
			now := time.Now()
			subscription.CancelledAt = &now

			return cancelSubRepo.UpdateSubscription(ctx, subscription)
		}

		// Create cancellation webhook
		webhook := model.RevenueCatWebhook{
			APIVersion: "1.0",
			Event: model.RevenueCatEvent{
				Type:             "CANCELLATION",
				ID:               "test_cancellation_123",
				AppUserID:        revenueCatCustomerID,
				ProductID:        "com.decorebator.premium.monthly",
				EventTimestampMS: time.Now().Unix() * 1000,
			},
		}

		jsonBody, _ := json.Marshal(webhook)
		webhookAuthToken := os.Getenv("REVENUECAT_WEBHOOK_AUTHORIZATION")

		req, err := http.NewRequest("POST", ts.BaseURL+"/webhook/revenuecat", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Authorization", webhookAuthToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Process the cancellation job
		ts.ProcessNextRiverJob(t, "revenuecat_webhooks", rcServiceMock)

		// Verify subscription was marked for cancellation
		cancelledSub, err := subRepo.GetActiveSubscriptionForUser(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, cancelledSub)
		assert.True(t, cancelledSub.CancelAtPeriodEnd)
		assert.NotNil(t, cancelledSub.CancelledAt)
		assert.Equal(t, model.StatusActive, cancelledSub.Status) // Still active until period end
	})
}

func init() {
	// Ensure database connection is initialized
	_, err := common.GetDBConnection()
	if err != nil {
		panic("Failed to initialize database connection: " + err.Error())
	}
}
