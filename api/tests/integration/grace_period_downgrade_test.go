package integration

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGracePeriodPlanDowngrade tests the complete grace period plan downgrade functionality
func TestGracePeriodPlanDowngrade(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	t.Run("UserBeyondGracePeriod_PlanDowngradedWithJWTRefresh", func(t *testing.T) {
		ctx := context.Background()

		// Create a test user
		initialToken := ts.WithPremiumUser(t)
		claims := setup.DecodeJWT(t, initialToken)

		subjectRaw, ok := claims["sub"]
		require.True(t, ok, "JWT should contain 'sub' field")
		subject, ok := subjectRaw.(string)
		require.True(t, ok, "JWT 'sub' field should be a string")

		userID, err := strconv.ParseInt(subject, 10, 64)
		require.NoError(t, err)

		// Create a past_due subscription beyond grace period (4 days ago)
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		subscription, err := subRepo.GetActiveSubscriptionForUser(context.Background(), userID)
		require.NoError(t, err)

		// Update subscription to be past_due and beyond grace period
		subscription.Status = model.StatusPastDue
		subscription.CurrentPeriodStart = time.Now().Add(-35 * 24 * time.Hour)
		subscription.CurrentPeriodEnd = time.Now().Add(-4 * 24 * time.Hour) // 4 days ago (beyond 3-day grace period)

		setupEvent := model.SubscriptionEvent{
			ExternalEventID: setup.GenerateUniqueEventID("beyond_grace_test"),
			Provider:        model.ProviderRevenueCat,
			EventType:       "test_setup",
			EventData:       `{"type": "beyond_grace_test"}`,
		}
		err = subRepo.UpdateSubscription(ctx, subscription, setupEvent)
		require.NoError(t, err)

		// Call GetProfile endpoint
		profileResp := ts.Expect.GET("/users").
			WithHeader("Authorization", initialToken).
			Expect().
			Status(200)

		// Verify new JWT was returned in Authorization header
		newToken := profileResp.Header("Authorization").NotEmpty().Raw()
		assert.NotEqual(t, initialToken, newToken, "Should receive new JWT token")

		// Verify response body shows updated plan
		profileResp.JSON().Object().Value("subscriptionPlan").String().IsEqual("free")

		// Verify database was updated
		var updatedPlan model.SubscriptionPlan
		err = ts.DB.QueryRow(ctx,
			"SELECT subscription_plan FROM users WHERE id = $1",
			userID).Scan(&updatedPlan)
		require.NoError(t, err)
		assert.Equal(t, model.PlanFree, updatedPlan, "User plan should be downgraded to free")

		// Verify JWT payload contains updated plan
		decoded := setup.DecodeJWT(t, newToken)
		assert.Equal(t, "free", decoded["subscriptionPlan"], "JWT should contain updated subscription plan")
	})

	t.Run("UserWithinGracePeriod_PlanMaintained", func(t *testing.T) {
		// Create a test user
		signupInput := setup.GenerateSignupInput()
		signupResp := ts.Expect.POST("/users").
			WithJSON(signupInput).
			Expect().
			Status(201)

		// Get user ID and initial auth token
		ctx := context.Background()
		var userID int64
		err := ts.DB.QueryRow(ctx,
			"SELECT id FROM users WHERE email = $1",
			signupInput.Email).Scan(&userID)
		require.NoError(t, err)

		initialToken := signupResp.Header("Authorization").NotEmpty().Raw()

		// Update user to have monthly plan
		_, err = ts.DB.Exec(ctx,
			"UPDATE users SET subscription_plan = $1 WHERE id = $2",
			model.PlanMonthly, userID)
		require.NoError(t, err)

		// Create a past_due subscription within grace period (1 hour ago)
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		revenuecatSubID := fmt.Sprintf("%d", userID)
		subscription := &model.Subscription{
			UserID:                   userID,
			Provider:                 model.ProviderRevenueCat,
			RevenueCatSubscriptionID: &revenuecatSubID,
			Plan:                     model.PlanMonthly,
			Status:                   model.StatusPastDue,
			CurrentPeriodStart:       time.Now().Add(-30 * 24 * time.Hour),
			CurrentPeriodEnd:         time.Now().Add(-1 * time.Hour), // 1 hour ago (within 3-day grace period)
			AmountCents:              model.SubscriptionPrices[model.PlanMonthly].AmountCents,
			Currency:                 "USD",
		}

		setupEvent := model.SubscriptionEvent{
			ExternalEventID: setup.GenerateUniqueEventID("within_grace_test"),
			Provider:        model.ProviderRevenueCat,
			EventType:       "test_setup",
			EventData:       `{"type": "within_grace_test"}`,
		}
		_, err = subRepo.CreateSubscription(ctx, subscription, setupEvent)
		require.NoError(t, err)

		// Call GetProfile endpoint
		profileResp := ts.Expect.GET("/users").
			WithHeader("Authorization", initialToken).
			Expect().
			Status(200)

		// Verify no JWT refresh (no Authorization header)
		profileResp.Header("Authorization").IsEmpty()

		// Verify response body shows original plan
		profileResp.JSON().Object().Value("subscriptionPlan").String().IsEqual("monthly")

		// Verify database was NOT updated
		var updatedPlan model.SubscriptionPlan
		err = ts.DB.QueryRow(ctx,
			"SELECT subscription_plan FROM users WHERE id = $1",
			userID).Scan(&updatedPlan)
		require.NoError(t, err)
		assert.Equal(t, model.PlanMonthly, updatedPlan, "User plan should remain monthly during grace period")
	})

	t.Run("FreeUser_NoChangeNeeded", func(t *testing.T) {
		// Create a test user (defaults to free plan)
		signupInput := setup.GenerateSignupInput()
		signupResp := ts.Expect.POST("/users").
			WithJSON(signupInput).
			Expect().
			Status(201)

		// Get initial auth token
		initialToken := signupResp.Header("Authorization").NotEmpty().Raw()

		// Call GetProfile endpoint
		profileResp := ts.Expect.GET("/users").
			WithHeader("Authorization", initialToken).
			Expect().
			Status(200)

		// Verify no JWT refresh (no Authorization header)
		profileResp.Header("Authorization").IsEmpty()

		// Verify response body shows free plan
		profileResp.JSON().Object().Value("subscriptionPlan").String().IsEqual("free")

		t.Log("✓ Free user requires no plan changes")
	})

	t.Run("ActiveSubscription_NoDowngrade", func(t *testing.T) {
		// Create a test user
		signupInput := setup.GenerateSignupInput()
		signupResp := ts.Expect.POST("/users").
			WithJSON(signupInput).
			Expect().
			Status(201)

		// Get user ID and initial auth token
		ctx := context.Background()
		var userID int64
		err := ts.DB.QueryRow(ctx,
			"SELECT id FROM users WHERE email = $1",
			signupInput.Email).Scan(&userID)
		require.NoError(t, err)

		initialToken := signupResp.Header("Authorization").NotEmpty().Raw()

		// Update user to have monthly plan
		_, err = ts.DB.Exec(ctx,
			"UPDATE users SET subscription_plan = $1 WHERE id = $2",
			model.PlanMonthly, userID)
		require.NoError(t, err)

		// Create an active subscription (future end date)
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		revenuecatSubID := fmt.Sprintf("%d", userID)
		subscription := &model.Subscription{
			UserID:                   userID,
			Provider:                 model.ProviderRevenueCat,
			RevenueCatSubscriptionID: &revenuecatSubID,
			Plan:                     model.PlanMonthly,
			Status:                   model.StatusActive,
			CurrentPeriodStart:       time.Now().Add(-15 * 24 * time.Hour),
			CurrentPeriodEnd:         time.Now().Add(15 * 24 * time.Hour), // Future date (active)
			AmountCents:              model.SubscriptionPrices[model.PlanMonthly].AmountCents,
			Currency:                 "USD",
		}

		setupEvent := model.SubscriptionEvent{
			ExternalEventID: setup.GenerateUniqueEventID("active_sub_test"),
			Provider:        model.ProviderRevenueCat,
			EventType:       "test_setup",
			EventData:       `{"type": "active_sub_test"}`,
		}
		_, err = subRepo.CreateSubscription(ctx, subscription, setupEvent)
		require.NoError(t, err)

		// Call GetProfile endpoint
		profileResp := ts.Expect.GET("/users").
			WithHeader("Authorization", initialToken).
			Expect().
			Status(200)

		// Verify no JWT refresh (no Authorization header)
		profileResp.Header("Authorization").IsEmpty()

		// Verify response body shows original plan
		profileResp.JSON().Object().Value("subscriptionPlan").String().IsEqual("monthly")

		// Verify database was NOT updated
		var updatedPlan model.SubscriptionPlan
		err = ts.DB.QueryRow(ctx,
			"SELECT subscription_plan FROM users WHERE id = $1",
			userID).Scan(&updatedPlan)
		require.NoError(t, err)
		assert.Equal(t, model.PlanMonthly, updatedPlan, "User plan should remain monthly for active subscription")
	})

	t.Run("EdgeCase_ExactlyAtGracePeriodBoundary", func(t *testing.T) {
		ctx := context.Background()

		// Create a test user with premium subscription
		initialToken := ts.WithPremiumUser(t)
		claims := setup.DecodeJWT(t, initialToken)

		subjectRaw, ok := claims["sub"]
		require.True(t, ok, "JWT should contain 'sub' field")
		subject, ok := subjectRaw.(string)
		require.True(t, ok, "JWT 'sub' field should be a string")

		userID, err := strconv.ParseInt(subject, 10, 64)
		require.NoError(t, err)

		// Update the existing subscription to be exactly at grace period boundary (beyond 3 days)
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		subscription, err := subRepo.GetActiveSubscriptionForUser(context.Background(), userID)
		require.NoError(t, err)

		// Update subscription to be past_due and exactly beyond grace period boundary
		subscription.Status = model.StatusPastDue
		subscription.CurrentPeriodStart = time.Now().Add(-33 * 24 * time.Hour)
		subscription.CurrentPeriodEnd = time.Now().Add(-3*24*time.Hour - 1*time.Minute) // Just over 3 days ago (beyond grace period)

		setupEvent := model.SubscriptionEvent{
			ExternalEventID: setup.GenerateUniqueEventID("boundary_grace_test"),
			Provider:        model.ProviderRevenueCat,
			EventType:       "test_setup",
			EventData:       `{"type": "boundary_grace_test"}`,
		}
		err = subRepo.UpdateSubscription(ctx, subscription, setupEvent)
		require.NoError(t, err)

		// Call GetProfile endpoint
		profileResp := ts.Expect.GET("/users").
			WithHeader("Authorization", initialToken).
			Expect().
			Status(200)

		// Verify new JWT was returned (plan should be downgraded)
		newToken := profileResp.Header("Authorization").NotEmpty().Raw()
		assert.NotEqual(t, initialToken, newToken, "Should receive new JWT token at grace period boundary")

		// Verify response body shows updated plan
		profileResp.JSON().Object().Value("subscriptionPlan").String().IsEqual("free")

		// Verify database was updated
		var updatedPlan model.SubscriptionPlan
		err = ts.DB.QueryRow(ctx,
			"SELECT subscription_plan FROM users WHERE id = $1",
			userID).Scan(&updatedPlan)
		require.NoError(t, err)
		assert.Equal(t, model.PlanFree, updatedPlan, "User plan should be downgraded at grace period boundary")
	})
}
