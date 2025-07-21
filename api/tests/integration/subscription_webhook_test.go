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
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"
)

func TestCustomerSubscriptionCreatedWebhook(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	t.Run("CreatesNewSubscription", func(t *testing.T) {
		// Create test user
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

		// Create Stripe customer
		stripeCustomerID := fmt.Sprintf("cus_test_%d", userID)
		_, err = ts.DB.Exec(ctx,
			"UPDATE users SET stripe_customer_id = $1 WHERE id = $2",
			stripeCustomerID, userID)
		require.NoError(t, err)

		// Create real Stripe subscription object
		stripeSubscription := stripe.Subscription{
			ID:       fmt.Sprintf("sub_test_%d", userID),
			Customer: &stripe.Customer{ID: stripeCustomerID},
			Status:   stripe.SubscriptionStatusActive,
			Currency: stripe.CurrencyUSD,
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{
					{
						Price: &stripe.Price{
							ID:         "price_monthly_test",
							UnitAmount: 699,
						},
						CurrentPeriodStart: time.Now().Unix(),
						CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour).Unix(),
					},
				},
			},
			Metadata: map[string]string{
				"user_id": fmt.Sprintf("%d", userID),
				"plan":    "monthly",
			},
			CancelAtPeriodEnd: false,
		}

		// Create webhook event
		eventData, err := json.Marshal(stripeSubscription)
		require.NoError(t, err)

		event := stripe.Event{
			ID:   setup.GenerateUniqueEventID("evt_test_created"),
			Type: "customer.subscription.created",
			Data: &stripe.EventData{
				Raw: eventData,
			},
		}

		// Send webhook with proper signature
		body, signature, err := setup.CreateSignedStripeEvent(event)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", ts.BaseURL+"/webhook/stripe", bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Stripe-Signature", signature)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Check that a job was enqueued
		var jobCount int
		err = ts.DB.QueryRow(ctx,
			"SELECT COUNT(*) FROM river_job WHERE kind = 'stripe-webhook' AND state = 'available'").
			Scan(&jobCount)
		require.NoError(t, err)
		assert.Greater(t, jobCount, 0, "Should have enqueued at least one job")

		// Process the job manually
		subService := service.NewSubscriptionService(ts.DB, ts.AppContext.MailService, ts.AppContext.RevenueCatService)
		err = ts.ProcessStripeWebhookJob(t, subService)
		require.NoError(t, err)

		// Verify subscription was created
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		subscription, err := subRepo.GetActiveSubscriptionForUser(ctx, userID)
		require.NoError(t, err)
		require.NotNil(t, subscription)

		// Verify subscription details
		assert.Equal(t, model.ProviderStripe, subscription.Provider)
		assert.Equal(t, model.PlanMonthly, subscription.Plan)
		assert.Equal(t, model.StatusActive, subscription.Status)
		assert.NotNil(t, subscription.StripeSubscriptionID)
		assert.Equal(t, stripeSubscription.ID, *subscription.StripeSubscriptionID)
		assert.Equal(t, stripeCustomerID, *subscription.StripeCustomerID)
		assert.Equal(t, 699, subscription.AmountCents)
		assert.Equal(t, "usd", subscription.Currency) // Stripe stores currency as lowercase
	})

	t.Run("HandlesExistingCustomer", func(t *testing.T) {
		// Create test user with existing Stripe customer
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

		stripeCustomerID := fmt.Sprintf("cus_existing_%d", userID)
		_, err = ts.DB.Exec(ctx,
			"UPDATE users SET stripe_customer_id = $1 WHERE id = $2",
			stripeCustomerID, userID)
		require.NoError(t, err)

		// Create subscription with annual plan
		stripeSubscription := stripe.Subscription{
			ID:       fmt.Sprintf("sub_annual_%d", userID),
			Customer: &stripe.Customer{ID: stripeCustomerID},
			Status:   stripe.SubscriptionStatusActive,
			Currency: stripe.CurrencyUSD,
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{
					{
						Price: &stripe.Price{
							ID:         "price_annual_test",
							UnitAmount: 6990,
						},
						CurrentPeriodStart: time.Now().Unix(),
						CurrentPeriodEnd:   time.Now().Add(365 * 24 * time.Hour).Unix(),
					},
				},
			},
			Metadata: map[string]string{
				"user_id": fmt.Sprintf("%d", userID),
				"plan":    "annual",
			},
		}

		// Create webhook event
		eventData, err := json.Marshal(stripeSubscription)
		require.NoError(t, err)

		event := stripe.Event{
			ID:   setup.GenerateUniqueEventID("evt_test_annual"),
			Type: "customer.subscription.created",
			Data: &stripe.EventData{
				Raw: eventData,
			},
		}

		// Send webhook and process with proper signature
		body, signature, err := setup.CreateSignedStripeEvent(event)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", ts.BaseURL+"/webhook/stripe", bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Stripe-Signature", signature)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Process the job
		subService := service.NewSubscriptionService(ts.DB, ts.AppContext.MailService, ts.AppContext.RevenueCatService)
		err = ts.ProcessStripeWebhookJob(t, subService)
		require.NoError(t, err)

		// Verify annual subscription
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		subscription, err := subRepo.GetActiveSubscriptionForUser(ctx, userID)
		require.NoError(t, err)
		require.NotNil(t, subscription)

		assert.Equal(t, model.PlanAnnual, subscription.Plan)
		assert.Equal(t, 6990, subscription.AmountCents)
	})
}

func TestCustomerSubscriptionUpdatedWebhook(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	t.Run("UpdatesSubscriptionStatus", func(t *testing.T) {
		// Create test user with existing subscription
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

		// Create existing subscription
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		stripeSubID := fmt.Sprintf("sub_test_%d", userID)
		stripeCustomerID := fmt.Sprintf("cus_test_%d", userID)

		sub := &model.Subscription{
			UserID:               userID,
			Provider:             model.ProviderStripe,
			StripeSubscriptionID: &stripeSubID,
			StripeCustomerID:     &stripeCustomerID,
			Plan:                 model.PlanMonthly,
			Status:               model.StatusActive,
			CurrentPeriodStart:   time.Now().UTC(),
			CurrentPeriodEnd:     time.Now().UTC().Add(30 * 24 * time.Hour),
			CancelAtPeriodEnd:    false,
			AmountCents:          699,
			Currency:             "USD",
		}
		// Create test setup event
		testEvent := model.SubscriptionEvent{
			ExternalEventID: setup.GenerateUniqueEventID("test_setup"),
			Provider:        model.ProviderStripe,
			EventType:       "test_setup",
			EventData:       `{"type": "test_setup", "source": "integration_test"}`,
		}
		_, err = subRepo.CreateSubscription(ctx, sub, testEvent)
		require.NoError(t, err)

		// Create updated subscription object (status changed to past_due)
		stripeSubscription := stripe.Subscription{
			ID:       stripeSubID,
			Customer: &stripe.Customer{ID: stripeCustomerID},
			Status:   stripe.SubscriptionStatusPastDue,
			Currency: stripe.CurrencyUSD,
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{
					{
						Price: &stripe.Price{
							ID:         "price_monthly_test",
							UnitAmount: 699,
						},
						CurrentPeriodStart: time.Now().Unix(),
						CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour).Unix(),
					},
				},
			},
			CancelAtPeriodEnd: false,
		}

		// Create webhook event
		eventData, err := json.Marshal(stripeSubscription)
		require.NoError(t, err)

		event := stripe.Event{
			ID:   setup.GenerateUniqueEventID("evt_test_updated"),
			Type: "customer.subscription.updated",
			Data: &stripe.EventData{
				Raw: eventData,
			},
		}

		// Send webhook with proper signature
		body, signature, err := setup.CreateSignedStripeEvent(event)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", ts.BaseURL+"/webhook/stripe", bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Stripe-Signature", signature)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Process the job
		subService := service.NewSubscriptionService(ts.DB, ts.AppContext.MailService, ts.AppContext.RevenueCatService)
		err = ts.ProcessStripeWebhookJob(t, subService)
		require.NoError(t, err)

		// Verify subscription status was updated
		updatedSub, err := subRepo.GetSubscriptionByStripeID(ctx, stripeSubID)
		require.NoError(t, err)
		assert.Equal(t, model.StatusPastDue, updatedSub.Status)
	})

	t.Run("UpdatesCancelAtPeriodEnd", func(t *testing.T) {
		// Create test user with existing subscription
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

		// Create existing subscription
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		stripeSubID := fmt.Sprintf("sub_cancel_%d", userID)
		stripeCustomerID := fmt.Sprintf("cus_cancel_%d", userID)

		sub := &model.Subscription{
			UserID:               userID,
			Provider:             model.ProviderStripe,
			StripeSubscriptionID: &stripeSubID,
			StripeCustomerID:     &stripeCustomerID,
			Plan:                 model.PlanMonthly,
			Status:               model.StatusActive,
			CurrentPeriodStart:   time.Now().UTC(),
			CurrentPeriodEnd:     time.Now().UTC().Add(30 * 24 * time.Hour),
			CancelAtPeriodEnd:    false,
			AmountCents:          699,
			Currency:             "USD",
		}
		// Create test setup event
		testEvent := model.SubscriptionEvent{
			ExternalEventID: setup.GenerateUniqueEventID("test_setup"),
			Provider:        model.ProviderStripe,
			EventType:       "test_setup",
			EventData:       `{"type": "test_setup", "source": "integration_test"}`,
		}
		_, err = subRepo.CreateSubscription(ctx, sub, testEvent)
		require.NoError(t, err)

		// Create updated subscription (cancel at period end = true)
		stripeSubscription := stripe.Subscription{
			ID:       stripeSubID,
			Customer: &stripe.Customer{ID: stripeCustomerID},
			Status:   stripe.SubscriptionStatusActive,
			Currency: stripe.CurrencyUSD,
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{
					{
						Price: &stripe.Price{
							ID:         "price_monthly_test",
							UnitAmount: 699,
						},
						CurrentPeriodStart: time.Now().Unix(),
						CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour).Unix(),
					},
				},
			},
			CancelAtPeriodEnd: true,
		}

		// Create webhook event
		eventData, err := json.Marshal(stripeSubscription)
		require.NoError(t, err)

		event := stripe.Event{
			ID:   setup.GenerateUniqueEventID("evt_test_cancel"),
			Type: "customer.subscription.updated",
			Data: &stripe.EventData{
				Raw: eventData,
			},
		}

		// Send webhook and process with proper signature
		body, signature, err := setup.CreateSignedStripeEvent(event)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", ts.BaseURL+"/webhook/stripe", bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Stripe-Signature", signature)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Process the job
		subService := service.NewSubscriptionService(ts.DB, ts.AppContext.MailService, ts.AppContext.RevenueCatService)
		err = ts.ProcessStripeWebhookJob(t, subService)
		require.NoError(t, err)

		// Verify cancel at period end was updated
		updatedSub, err := subRepo.GetSubscriptionByStripeID(ctx, stripeSubID)
		require.NoError(t, err)
		assert.True(t, updatedSub.CancelAtPeriodEnd)
	})
}

func TestCustomerSubscriptionDeletedWebhook(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	t.Run("CancelsSubscription", func(t *testing.T) {
		// Create test user with existing subscription
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

		// Create existing subscription
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		stripeSubID := fmt.Sprintf("sub_delete_%d", userID)
		stripeCustomerID := fmt.Sprintf("cus_delete_%d", userID)

		sub := &model.Subscription{
			UserID:               userID,
			Provider:             model.ProviderStripe,
			StripeSubscriptionID: &stripeSubID,
			StripeCustomerID:     &stripeCustomerID,
			Plan:                 model.PlanMonthly,
			Status:               model.StatusActive,
			CurrentPeriodStart:   time.Now().UTC(),
			CurrentPeriodEnd:     time.Now().UTC().Add(30 * 24 * time.Hour),
			CancelAtPeriodEnd:    false,
			AmountCents:          699,
			Currency:             "USD",
		}
		// Create test setup event
		testEvent := model.SubscriptionEvent{
			ExternalEventID: setup.GenerateUniqueEventID("test_setup"),
			Provider:        model.ProviderStripe,
			EventType:       "test_setup",
			EventData:       `{"type": "test_setup", "source": "integration_test"}`,
		}
		_, err = subRepo.CreateSubscription(ctx, sub, testEvent)
		require.NoError(t, err)

		// Create deleted subscription object
		canceledAt := time.Now().Unix()
		stripeSubscription := stripe.Subscription{
			ID:         stripeSubID,
			Customer:   &stripe.Customer{ID: stripeCustomerID},
			Status:     stripe.SubscriptionStatusCanceled,
			Currency:   stripe.CurrencyUSD,
			CanceledAt: canceledAt,
		}

		// Create webhook event
		eventData, err := json.Marshal(stripeSubscription)
		require.NoError(t, err)

		event := stripe.Event{
			ID:   setup.GenerateUniqueEventID("evt_test_deleted"),
			Type: "customer.subscription.deleted",
			Data: &stripe.EventData{
				Raw: eventData,
			},
		}

		// Send webhook with proper signature
		body, signature, err := setup.CreateSignedStripeEvent(event)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", ts.BaseURL+"/webhook/stripe", bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Stripe-Signature", signature)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Process the job
		subService := service.NewSubscriptionService(ts.DB, ts.AppContext.MailService, ts.AppContext.RevenueCatService)
		err = ts.ProcessStripeWebhookJob(t, subService)
		require.NoError(t, err)

		// Verify subscription was canceled
		canceledSub, err := subRepo.GetSubscriptionByStripeID(ctx, stripeSubID)
		require.NoError(t, err)
		assert.Equal(t, model.StatusCanceled, canceledSub.Status)
		assert.NotNil(t, canceledSub.CanceledAt)
	})
}

func TestInvoicePaymentFailedWebhook(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	t.Run("UpdatesSubscriptionToPastDue", func(t *testing.T) {
		// Create test user with active subscription
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

		// Create existing subscription
		subRepo := repository.NewSubscriptionRepository(ts.DB)
		stripeSubID := fmt.Sprintf("sub_failed_%d", userID)
		stripeCustomerID := fmt.Sprintf("cus_failed_%d", userID)

		sub := &model.Subscription{
			UserID:               userID,
			Provider:             model.ProviderStripe,
			StripeSubscriptionID: &stripeSubID,
			StripeCustomerID:     &stripeCustomerID,
			Plan:                 model.PlanMonthly,
			Status:               model.StatusActive,
			CurrentPeriodStart:   time.Now().UTC(),
			CurrentPeriodEnd:     time.Now().UTC().Add(30 * 24 * time.Hour),
			CancelAtPeriodEnd:    false,
			AmountCents:          699,
			Currency:             "USD",
		}
		// Create test setup event
		testEvent := model.SubscriptionEvent{
			ExternalEventID: setup.GenerateUniqueEventID("test_setup"),
			Provider:        model.ProviderStripe,
			EventType:       "test_setup",
			EventData:       `{"type": "test_setup", "source": "integration_test"}`,
		}
		_, err = subRepo.CreateSubscription(ctx, sub, testEvent)
		require.NoError(t, err)

		// Update user's stripe customer ID for email testing
		_, err = ts.DB.Exec(ctx,
			"UPDATE users SET stripe_customer_id = $1 WHERE id = $2",
			stripeCustomerID, userID)
		require.NoError(t, err)

		// Create failed invoice
		invoiceID := fmt.Sprintf("in_failed_%d", userID)
		nextAttempt := time.Now().Add(24 * time.Hour).Unix()
		stripeInvoice := stripe.Invoice{
			ID:                 invoiceID,
			AmountDue:          699,
			Currency:           stripe.CurrencyUSD,
			Customer:           &stripe.Customer{ID: stripeCustomerID},
			AttemptCount:       1,
			NextPaymentAttempt: nextAttempt,
			Lines: &stripe.InvoiceLineItemList{
				Data: []*stripe.InvoiceLineItem{
					{
						Subscription: &stripe.Subscription{ID: stripeSubID},
					},
				},
			},
		}

		// Create webhook event
		eventData, err := json.Marshal(stripeInvoice)
		require.NoError(t, err)

		event := stripe.Event{
			ID:   setup.GenerateUniqueEventID("evt_payment_failed"),
			Type: "invoice.payment_failed",
			Data: &stripe.EventData{
				Raw: eventData,
			},
		}

		// Send webhook with proper signature
		body, signature, err := setup.CreateSignedStripeEvent(event)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", ts.BaseURL+"/webhook/stripe", bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Stripe-Signature", signature)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Check that job was enqueued
		var jobCount int
		err = ts.DB.QueryRow(ctx,
			"SELECT COUNT(*) FROM river_job WHERE kind = 'stripe-webhook' AND state = 'available'").
			Scan(&jobCount)
		require.NoError(t, err)
		assert.Greater(t, jobCount, 0, "Should have enqueued at least one job")

		// Process the job
		subService := service.NewSubscriptionService(ts.DB, ts.AppContext.MailService, ts.AppContext.RevenueCatService)
		err = ts.ProcessStripeWebhookJob(t, subService)
		require.NoError(t, err)

		// Verify subscription status was updated to past_due
		updatedSub, err := subRepo.GetSubscriptionByStripeID(ctx, stripeSubID)
		require.NoError(t, err)
		assert.Equal(t, model.StatusPastDue, updatedSub.Status)
	})

	t.Run("HandlesNonexistentSubscription", func(t *testing.T) {
		// Create invoice payment failed event for non-existent subscription
		stripeInvoice := stripe.Invoice{
			ID:        "in_nonexistent",
			AmountDue: 699,
			Currency:  stripe.CurrencyUSD,
			Customer:  &stripe.Customer{ID: "cus_nonexistent"},
			Lines: &stripe.InvoiceLineItemList{
				Data: []*stripe.InvoiceLineItem{
					{
						Subscription: &stripe.Subscription{ID: "sub_nonexistent"},
					},
				},
			},
		}

		// Create webhook event
		eventData, err := json.Marshal(stripeInvoice)
		require.NoError(t, err)

		event := stripe.Event{
			ID:   setup.GenerateUniqueEventID("evt_nonexistent"),
			Type: "invoice.payment_failed",
			Data: &stripe.EventData{
				Raw: eventData,
			},
		}

		// Send webhook with proper signature
		body, signature, err := setup.CreateSignedStripeEvent(event)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", ts.BaseURL+"/webhook/stripe", bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Stripe-Signature", signature)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return OK even if subscription doesn't exist
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Process the job (should handle the error gracefully)
		subService := service.NewSubscriptionService(ts.DB, ts.AppContext.MailService, ts.AppContext.RevenueCatService)
		err = ts.ProcessStripeWebhookJob(t, subService)
		// The job should fail but not crash the test
		assert.Error(t, err, "Should error when subscription not found")
	})

	t.Run("HandlesInvoiceWithoutSubscription", func(t *testing.T) {
		// Create invoice without subscription (one-time payment)
		stripeInvoice := stripe.Invoice{
			ID:        "in_no_subscription",
			AmountDue: 1000,
			Currency:  stripe.CurrencyUSD,
			Customer:  &stripe.Customer{ID: "cus_no_sub"},
			Lines:     nil, // No line items indicating one-time payment
		}

		// Create webhook event
		eventData, err := json.Marshal(stripeInvoice)
		require.NoError(t, err)

		eventID := setup.GenerateUniqueEventID("evt_no_sub")
		event := stripe.Event{
			ID:   eventID,
			Type: "invoice.payment_failed",
			Data: &stripe.EventData{
				Raw: eventData,
			},
		}

		// Send webhook with proper signature
		body, signature, err := setup.CreateSignedStripeEvent(event)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", ts.BaseURL+"/webhook/stripe", bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Stripe-Signature", signature)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle gracefully and return OK
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Process the specific job for this event (should handle gracefully)
		subService := service.NewSubscriptionService(ts.DB, ts.AppContext.MailService, ts.AppContext.RevenueCatService)
		err = ts.ProcessStripeWebhookJobWithEventID(t, subService, eventID)
		require.NoError(t, err, "Should handle invoice without subscription gracefully")
	})
}
