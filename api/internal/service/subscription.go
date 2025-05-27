package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/mail"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/subscription"
	"github.com/stripe/stripe-go/v82/webhook"
)

type SubscriptionService struct {
	db            *pgxpool.Pool
	subRepo       *repository.SubscriptionRepository
	userRepo      *repository.UserRepository
	stripeKey     string
	webhookSecret string
	successURL    string
	cancelURL     string
}

func NewSubscriptionService(db *pgxpool.Pool) *SubscriptionService {
	// Validate required environment variables
	stripeKey := os.Getenv("STRIPE_API_KEY")
	if stripeKey == "" {
		panic("STRIPE_API_KEY environment variable is required for subscription service")
	}

	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		panic("STRIPE_WEBHOOK_SECRET environment variable is required for subscription service")
	}

	successURL := os.Getenv("STRIPE_SUCCESS_URL")
	if successURL == "" {
		panic("STRIPE_SUCCESS_URL environment variable is required for subscription service")
	}

	cancelURL := os.Getenv("STRIPE_CANCEL_URL")
	if cancelURL == "" {
		panic("STRIPE_CANCEL_URL environment variable is required for subscription service")
	}

	// Set Stripe API key globally
	stripe.Key = stripeKey

	return &SubscriptionService{
		db:            db,
		subRepo:       repository.NewSubscriptionRepository(db),
		userRepo:      &repository.UserRepository{Db: db},
		stripeKey:     stripeKey,
		webhookSecret: webhookSecret,
		successURL:    successURL,
		cancelURL:     cancelURL,
	}
}

// CreateCheckoutSession creates a Stripe checkout session for subscription
func (s *SubscriptionService) CreateCheckoutSession(ctx context.Context, userID int64, email string, plan model.SubscriptionPlan, redirectUri string) (*stripe.CheckoutSession, error) {
	// Get or create Stripe customer
	stripeCustomerID, err := s.getOrCreateStripeCustomer(ctx, userID, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create stripe customer: %w", err)
	}

	// Get price ID based on plan
	priceID := s.getPriceIDForPlan(plan)
	if priceID == "" {
		return nil, fmt.Errorf("invalid subscription plan: %s", plan)
	}

	successURL := fmt.Sprintf("%s?session_id={CHECKOUT_SESSION_ID}&redirect_uri=%s", s.successURL, url.QueryEscape(redirectUri))

	// Create checkout session params
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(stripeCustomerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL:        stripe.String(successURL),
		CancelURL:         stripe.String(s.cancelURL),
		ClientReferenceID: stripe.String(fmt.Sprintf("%d", userID)),
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			// TrialPeriodDays: stripe.Int64(7), // 7-day free trial
			Metadata: map[string]string{
				"user_id": fmt.Sprintf("%d", userID),
				"plan":    string(plan),
			},
		},
	}

	// Create the session
	checkoutSession, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	return checkoutSession, nil
}

// HandleWebhook processes Stripe webhook events
func (s *SubscriptionService) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	// Verify webhook signature
	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		return fmt.Errorf("webhook signature verification failed: %w", err)
	}

	// Check if we've already processed this event
	processed, err := s.subRepo.HasProcessedEvent(ctx, event.ID)
	if err != nil {
		return fmt.Errorf("failed to check event processing status: %w", err)
	}
	if processed {
		return nil // Already processed, skip
	}

	// Handle different event types
	switch event.Type {
	case "checkout.session.completed":
		return s.handleCheckoutSessionCompleted(ctx, event)
	case "customer.subscription.created":
		return s.handleSubscriptionCreated(ctx, event)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(ctx, event)
	case "invoice.payment_succeeded":
		return s.handleInvoicePaymentSucceeded(ctx, event)
	case "invoice.payment_failed":
		return s.handleInvoicePaymentFailed(ctx, event)
	default:
		// Log unhandled event types
		fmt.Printf("Unhandled webhook event type: %s\n", event.Type)
	}

	return nil
}

// handleCheckoutSessionCompleted handles successful checkout completion
func (s *SubscriptionService) handleCheckoutSessionCompleted(ctx context.Context, event stripe.Event) error {
	var checkoutSession stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &checkoutSession); err != nil {
		return fmt.Errorf("failed to unmarshal checkout session: %w", err)
	}

	// Session completed successfully - subscription will be created via subscription.created event
	// Just log for now
	fmt.Printf("Checkout session completed: %s\n", checkoutSession.ID)

	return nil
}

// handleSubscriptionCreated handles new subscription creation
func (s *SubscriptionService) handleSubscriptionCreated(ctx context.Context, event stripe.Event) error {
	var stripeSubscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSubscription); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	// Extract user ID from metadata
	userIDStr := stripeSubscription.Metadata["user_id"]
	if userIDStr == "" {
		return fmt.Errorf("missing user_id in subscription metadata")
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user_id in subscription metadata: %w", err)
	}

	// Determine plan from price ID
	plan := s.getPlanFromPriceID(stripeSubscription.Items.Data[0].Price.ID)

	fmt.Println(stripeSubscription.Items.Data[0])

	// Create subscription record
	sub := &model.Subscription{
		UserID:               userID,
		StripeSubscriptionID: stripeSubscription.ID,
		StripeCustomerID:     stripeSubscription.Customer.ID,
		Plan:                 plan,
		Status:               model.SubscriptionStatus(stripeSubscription.Status),
		CurrentPeriodStart:   time.Unix(stripeSubscription.Items.Data[0].CurrentPeriodStart, 0),
		CurrentPeriodEnd:     time.Unix(stripeSubscription.Items.Data[0].CurrentPeriodEnd, 0),
		CancelAtPeriodEnd:    stripeSubscription.CancelAtPeriodEnd,
		AmountCents:          int(stripeSubscription.Items.Data[0].Price.UnitAmount),
		Currency:             string(stripeSubscription.Currency),
	}

	if stripeSubscription.TrialEnd > 0 {
		trialEnd := time.Unix(stripeSubscription.TrialEnd, 0)
		sub.TrialEnd = &trialEnd
	}

	// Save subscription
	if err := s.subRepo.CreateSubscription(ctx, sub); err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	// Record event
	eventData, err := json.Marshal(stripeSubscription)
	if err != nil {
		common.Logger.Error("Failed to marshal subscription data", "error", err, "subscription_id", sub.ID)
		// Continue without failing - subscription is created
	} else {
		if err := s.subRepo.CreateSubscriptionEvent(ctx, &model.SubscriptionEvent{
			SubscriptionID: sub.ID,
			StripeEventID:  event.ID,
			EventType:      string(event.Type),
			EventData:      string(eventData),
		}); err != nil {
			// Log error but don't fail - subscription is already created
			common.Logger.Error("Failed to record subscription event", "error", err, "subscription_id", sub.ID)
		}
	}

	// Get user details for email
	users, err := s.userRepo.Find(repository.FindUserArgs{ID: &userID})
	if err != nil || len(users) == 0 {
		// Log error but don't fail the webhook
		fmt.Printf("Failed to get user for email notification: %v\n", err)
		return nil
	}
	user := &users[0]

	// Send subscription activated email
	emailData := mail.SubscriptionEmailData{
		PlanName:        string(plan),
		AmountCents:     sub.AmountCents,
		Currency:        sub.Currency,
		NextBillingDate: sub.CurrentPeriodEnd,
		SubscriptionID:  stripeSubscription.ID,
	}

	if err := mail.SendSubscriptionActivatedEmail(user, emailData); err != nil {
		// Log error but don't fail the webhook
		fmt.Printf("Failed to send subscription activated email: %v\n", err)
	}

	return nil
}

// handleSubscriptionUpdated handles subscription updates
func (s *SubscriptionService) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
	var stripeSubscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSubscription); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	// Find existing subscription
	sub, err := s.subRepo.GetSubscriptionByStripeID(ctx, stripeSubscription.ID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}
	if sub == nil {
		return fmt.Errorf("subscription not found: %s", stripeSubscription.ID)
	}

	// Update subscription
	sub.Status = model.SubscriptionStatus(stripeSubscription.Status)
	sub.CurrentPeriodStart = time.Unix(stripeSubscription.Items.Data[0].CurrentPeriodStart, 0)
	sub.CurrentPeriodEnd = time.Unix(stripeSubscription.Items.Data[0].CurrentPeriodEnd, 0)
	sub.CancelAtPeriodEnd = stripeSubscription.CancelAtPeriodEnd

	if stripeSubscription.CanceledAt > 0 {
		cancelledAt := time.Unix(stripeSubscription.CanceledAt, 0)
		sub.CancelledAt = &cancelledAt
	}

	if err := s.subRepo.UpdateSubscription(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Record event
	eventData, err := json.Marshal(stripeSubscription)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription data: %w", err)
	}
	if err := s.subRepo.CreateSubscriptionEvent(ctx, &model.SubscriptionEvent{
		SubscriptionID: sub.ID,
		StripeEventID:  event.ID,
		EventType:      string(event.Type),
		EventData:      string(eventData),
	}); err != nil {
		return fmt.Errorf("failed to record subscription event: %w", err)
	}

	return nil
}

// handleSubscriptionDeleted handles subscription cancellation
func (s *SubscriptionService) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	var stripeSubscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSubscription); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	// Find existing subscription
	sub, err := s.subRepo.GetSubscriptionByStripeID(ctx, stripeSubscription.ID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}
	if sub == nil {
		return fmt.Errorf("subscription not found: %s", stripeSubscription.ID)
	}

	// Update subscription status
	sub.Status = model.StatusCanceled
	if stripeSubscription.CanceledAt > 0 {
		cancelledAt := time.Unix(stripeSubscription.CanceledAt, 0)
		sub.CancelledAt = &cancelledAt
	}

	if err := s.subRepo.UpdateSubscription(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Record event
	eventData, err := json.Marshal(stripeSubscription)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription data: %w", err)
	}
	if err := s.subRepo.CreateSubscriptionEvent(ctx, &model.SubscriptionEvent{
		SubscriptionID: sub.ID,
		StripeEventID:  event.ID,
		EventType:      string(event.Type),
		EventData:      string(eventData),
	}); err != nil {
		return fmt.Errorf("failed to record subscription event: %w", err)
	}

	// Get user details for email
	users, err := s.userRepo.Find(repository.FindUserArgs{ID: &sub.UserID})
	if err != nil || len(users) == 0 {
		// Log error but don't fail the webhook
		fmt.Printf("Failed to get user for email notification: %v\n", err)
		return nil
	}
	user := &users[0]

	// Send subscription cancelled email
	emailData := mail.SubscriptionEmailData{
		PlanName:         string(sub.Plan),
		CancellationDate: sub.CurrentPeriodEnd, // Service ends at period end
		SubscriptionID:   stripeSubscription.ID,
	}

	if err := mail.SendSubscriptionCancelledEmail(user, emailData); err != nil {
		// Log error but don't fail the webhook
		fmt.Printf("Failed to send subscription cancelled email: %v\n", err)
	}

	return nil
}

// handleInvoicePaymentSucceeded handles successful payments
func (s *SubscriptionService) handleInvoicePaymentSucceeded(ctx context.Context, event stripe.Event) error {
	var stripeInvoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &stripeInvoice); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	// Skip if this is the first payment (subscription creation)
	if stripeInvoice.BillingReason == stripe.InvoiceBillingReasonSubscriptionCreate {
		return nil
	}

	// Get subscription
	sub, err := s.subRepo.GetSubscriptionByStripeID(ctx, stripeInvoice.ID)
	if err != nil || sub == nil {
		fmt.Printf("Failed to get subscription for invoice: %v\n", err)
		return nil
	}

	// Get user details
	users, err := s.userRepo.Find(repository.FindUserArgs{ID: &sub.UserID})
	if err != nil || len(users) == 0 {
		fmt.Printf("Failed to get user for email notification: %v\n", err)
		return nil
	}
	user := &users[0]

	// Send renewal confirmation email
	emailData := mail.SubscriptionEmailData{
		PlanName:        string(sub.Plan),
		AmountCents:     int(stripeInvoice.AmountPaid),
		Currency:        string(stripeInvoice.Currency),
		NextBillingDate: sub.CurrentPeriodEnd,
		SubscriptionID:  stripeInvoice.ID,
	}

	if err := mail.SendSubscriptionRenewedEmail(user, emailData); err != nil {
		fmt.Printf("Failed to send subscription renewed email: %v\n", err)
	}

	return nil
}

// handleInvoicePaymentFailed handles failed payments
func (s *SubscriptionService) handleInvoicePaymentFailed(ctx context.Context, event stripe.Event) error {
	var stripeInvoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &stripeInvoice); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	// Get subscription
	sub, err := s.subRepo.GetSubscriptionByStripeID(ctx, stripeInvoice.ID)
	if err != nil || sub == nil {
		fmt.Printf("Failed to get subscription for invoice: %v\n", err)
		return nil
	}

	// Get user details
	users, err := s.userRepo.Find(repository.FindUserArgs{ID: &sub.UserID})
	if err != nil || len(users) == 0 {
		fmt.Printf("Failed to get user for email notification: %v\n", err)
		return nil
	}
	user := &users[0]

	// Determine next retry date (Stripe typically retries in 3-7 days)
	nextRetryDate := time.Now().AddDate(0, 0, 3)

	// Send payment failed email
	emailData := mail.SubscriptionEmailData{
		PlanName:       string(sub.Plan),
		AmountCents:    int(stripeInvoice.AmountDue),
		Currency:       string(stripeInvoice.Currency),
		NextRetryDate:  nextRetryDate,
		SubscriptionID: stripeInvoice.ID,
		PaymentError:   "Payment method declined", // Generic error message
	}

	if err := mail.SendPaymentFailedEmail(user, emailData); err != nil {
		fmt.Printf("Failed to send payment failed email: %v\n", err)
	}

	return nil
}

// getOrCreateStripeCustomer gets existing or creates new Stripe customer
func (s *SubscriptionService) getOrCreateStripeCustomer(ctx context.Context, userID int64, email string) (string, error) {

	users, err := s.userRepo.Find(repository.FindUserArgs{ID: &userID})
	if err != nil || len(users) == 0 {
		return "", fmt.Errorf("no user for id %d", userID)
	}
	user := &users[0]
	if user.StripeCustomerID != nil {
		return *user.StripeCustomerID, nil
	}

	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Metadata: map[string]string{
			"user_id": fmt.Sprintf("%d", userID),
		},
	}

	stripeCustomer, err := customer.New(params)
	if err != nil {
		return "", err
	}

	// Update user with Stripe customer ID
	if err := s.subRepo.UpdateUserStripeCustomerID(ctx, userID, stripeCustomer.ID); err != nil {
		return "", fmt.Errorf("failed to update user stripe customer ID: %w", err)
	}

	return stripeCustomer.ID, nil
}

// getPriceIDForPlan returns the Stripe price ID for a plan
func (s *SubscriptionService) getPriceIDForPlan(plan model.SubscriptionPlan) string {
	// These should be configured as environment variables
	// For now, using placeholders
	switch plan {
	case model.PlanMonthly:
		return os.Getenv("STRIPE_PRICE_ID_MONTHLY")
	case model.PlanAnnual:
		return os.Getenv("STRIPE_PRICE_ID_ANNUAL")
	default:
		return ""
	}
}

// getPlanFromPriceID returns the plan for a Stripe price ID
func (s *SubscriptionService) getPlanFromPriceID(priceID string) model.SubscriptionPlan {
	monthlyPriceID := os.Getenv("STRIPE_PRICE_ID_MONTHLY")
	annualPriceID := os.Getenv("STRIPE_PRICE_ID_ANNUAL")

	switch priceID {
	case monthlyPriceID:
		return model.PlanMonthly
	case annualPriceID:
		return model.PlanAnnual
	default:
		return model.PlanFree
	}
}

// CancelSubscription cancels a user's subscription
func (s *SubscriptionService) CancelSubscription(ctx context.Context, userID int64) error {
	// Get active subscription
	sub, err := s.subRepo.GetActiveSubscriptionForUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get active subscription: %w", err)
	}
	if sub == nil {
		return fmt.Errorf("no active subscription found")
	}

	// Cancel at period end in Stripe
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}

	_, err = subscription.Update(sub.StripeSubscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to cancel stripe subscription: %w", err)
	}

	return nil
}

// CheckSubscriptionLimits checks if user can perform action based on subscription
func (s *SubscriptionService) CheckSubscriptionLimits(ctx context.Context, userID int64, action string) error {
	// Get user's active subscription
	sub, err := s.subRepo.GetActiveSubscriptionForUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// If no subscription or free plan, check limits
	if sub == nil || sub.Plan == model.PlanFree {
		switch action {
		case "create_wordlist":
			count, err := s.subRepo.CountUserWordlists(ctx, userID)
			if err != nil {
				return err
			}
			if count >= model.FreeWordlistLimit {
				return fmt.Errorf("free plan limit reached: maximum %d wordlist allowed", model.FreeWordlistLimit)
			}
		case "add_word":
			// This would need wordlist ID to check word count
			// Implement as needed
		}
	}

	return nil
}

// CountWordsInWordlist counts the number of words in a wordlist
func (s *SubscriptionService) CountWordsInWordlist(ctx context.Context, wordlistID int64) (int, error) {
	return s.subRepo.CountWordsInWordlist(ctx, wordlistID)
}
