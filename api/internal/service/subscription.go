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
	"github.com/stripe/stripe-go/v82/webhook"
)

type SubscriptionService struct {
	db                *pgxpool.Pool
	subRepo           *repository.SubscriptionRepository
	userRepo          *repository.UserRepository
	mailService       *mail.MailService
	revenueCatService RevenueCatService
	stripeKey         string
	webhookSecret     string
	successURL        string
	cancelURL         string
	effectiveAccess   *EffectiveAccessService
}

func (s *SubscriptionService) SetEffectiveAccess(service *EffectiveAccessService) {
	s.effectiveAccess = service
}

func NewSubscriptionService(db *pgxpool.Pool, mailService *mail.MailService, revenueCatService RevenueCatService) *SubscriptionService {
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
		db:                db,
		subRepo:           repository.NewSubscriptionRepository(db),
		userRepo:          &repository.UserRepository{Db: db},
		mailService:       mailService,
		revenueCatService: revenueCatService,
		stripeKey:         stripeKey,
		webhookSecret:     webhookSecret,
		successURL:        successURL,
		cancelURL:         cancelURL,
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
		Customer:            stripe.String(stripeCustomerID),
		AllowPromotionCodes: stripe.Bool(true),
		Mode:                stripe.String(string(stripe.CheckoutSessionModeSubscription)),
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
	var event stripe.Event

	// In test mode with test webhook secret, bypass signature verification
	if s.webhookSecret == "test-stripe-webhook-secret" && signature == "test_signature" {
		// Parse the event directly without signature verification
		if err := json.Unmarshal(payload, &event); err != nil {
			return fmt.Errorf("failed to unmarshal event: %w", err)
		}
	} else {
		// Verify webhook signature
		var err error
		event, err = webhook.ConstructEvent(payload, signature, s.webhookSecret)
		if err != nil {
			return fmt.Errorf("webhook signature verification failed: %w", err)
		}
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
	case "customer.subscription.created":
		return s.handleSubscriptionCreated(ctx, event)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(ctx, event)
	case "invoice.payment_failed":
		return s.handleInvoicePaymentFailed(ctx, event)
	default:
		// Log unhandled event types
		common.Logger.Warn("Unhandled webhook event type", "event_type", event.Type)
	}

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

	// Create subscription record
	sub := &model.Subscription{
		UserID:               userID,
		Provider:             model.ProviderStripe,
		StripeSubscriptionID: &stripeSubscription.ID,
		StripeCustomerID:     &stripeSubscription.Customer.ID,
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

	// Prepare event data
	eventData, err := json.Marshal(stripeSubscription)
	if err != nil {
		common.Logger.Error("Failed to marshal subscription data", "error", err)
		return err
	} else {
		// Create subscription with event in transaction
		subscriptionEvent := &model.SubscriptionEvent{
			ExternalEventID: event.ID,
			Provider:        model.ProviderStripe,
			EventType:       string(event.Type),
			EventData:       string(eventData),
		}

		subID, createErr := s.subRepo.CreateSubscription(ctx, sub, *subscriptionEvent)
		if createErr != nil {
			return fmt.Errorf("failed to create subscription with event: %w", createErr)
		}
		sub.ID = subID
	}

	// Get user details for email
	users, err := s.userRepo.Find(ctx, repository.FindUserArgs{ID: &userID})
	if err != nil || len(users) == 0 {
		// Log error but don't fail the webhook
		common.Logger.Error("Failed to get user for email notification", "error", err, "user_id", userID)
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

	if err := s.mailService.SendSubscriptionActivatedEmail(user, emailData); err != nil {
		// Log error but don't fail the webhook
		common.CaptureError(ctx, err, "Failed to send subscription activated email",
			"user_id", userID,
			"plan", plan)
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
		sub.CanceledAt = &cancelledAt
	}

	// Prepare event data
	eventData, err := json.Marshal(stripeSubscription)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription data: %w", err)
	}

	// Update subscription with event in transaction
	subscriptionEvent := &model.SubscriptionEvent{
		SubscriptionID:  sub.ID,
		ExternalEventID: event.ID,
		Provider:        model.ProviderStripe,
		EventType:       string(event.Type),
		EventData:       string(eventData),
	}

	if updateErr := s.subRepo.UpdateSubscription(ctx, sub, *subscriptionEvent); updateErr != nil {
		return fmt.Errorf("failed to update subscription with event: %w", updateErr)
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
		sub.CanceledAt = &cancelledAt
	}

	// Prepare event data
	eventData, err := json.Marshal(stripeSubscription)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription data: %w", err)
	}

	// Update subscription with event in transaction
	subscriptionEvent := &model.SubscriptionEvent{
		SubscriptionID:  sub.ID,
		ExternalEventID: event.ID,
		Provider:        model.ProviderStripe,
		EventType:       string(event.Type),
		EventData:       string(eventData),
	}

	if updateErr := s.subRepo.UpdateSubscription(ctx, sub, *subscriptionEvent); updateErr != nil {
		return fmt.Errorf("failed to update subscription with event: %w", updateErr)
	}

	// Get user details for email
	users, err := s.userRepo.Find(ctx, repository.FindUserArgs{ID: &sub.UserID})
	if err != nil || len(users) == 0 {
		// Log error but don't fail the webhook
		common.Logger.Error("Failed to get user for email notification", "error", err, "user_id", sub.UserID)
		return nil
	}
	user := &users[0]

	// Send subscription cancelled email
	emailData := mail.SubscriptionEmailData{
		PlanName:         string(sub.Plan),
		CancellationDate: sub.CurrentPeriodEnd, // Service ends at period end
		SubscriptionID:   stripeSubscription.ID,
	}

	if err := s.mailService.SendSubscriptionCancelledEmail(user, emailData); err != nil {
		// Log error but don't fail the webhook
		common.CaptureError(ctx, err, "Failed to send subscription canceled email",
			"user_id", sub.UserID,
			"email", user.Email)
	}

	return nil
}

// handleInvoicePaymentFailed handles failed payments
func (s *SubscriptionService) handleInvoicePaymentFailed(ctx context.Context, event stripe.Event) error {
	var stripeInvoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &stripeInvoice); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	// Get subscription ID from invoice line items
	if stripeInvoice.Lines == nil || len(stripeInvoice.Lines.Data) == 0 {
		// Log and return if no line items (could be one-time payment)
		common.Logger.Warn("Invoice payment failed without line items",
			"invoice_id", stripeInvoice.ID,
			"customer", stripeInvoice.Customer)
		return nil
	}

	// Get subscription from first line item
	lineItem := stripeInvoice.Lines.Data[0]
	if lineItem.Subscription == nil {
		// Log and return if no subscription (could be one-time payment)
		common.Logger.Warn("Invoice payment failed without subscription",
			"invoice_id", stripeInvoice.ID,
			"customer", stripeInvoice.Customer)
		return nil
	}

	subscriptionID := lineItem.Subscription.ID

	// Get the existing subscription from our database
	existingSubscription, err := s.subRepo.GetSubscriptionByStripeID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}
	if existingSubscription == nil {
		return fmt.Errorf("subscription not found for stripe ID: %s", subscriptionID)
	}

	// Update subscription status to past_due
	existingSubscription.Status = model.StatusPastDue

	// Prepare event data
	eventData, err := json.Marshal(stripeInvoice)
	if err != nil {
		return fmt.Errorf("failed to marshal invoice data: %w", err)
	}

	// Create event for payment failure
	subscriptionEvent := model.SubscriptionEvent{
		SubscriptionID:  existingSubscription.ID,
		ExternalEventID: event.ID,
		Provider:        model.ProviderStripe,
		EventType:       string(event.Type),
		EventData:       string(eventData),
	}

	// Save the updated subscription
	if err := s.subRepo.UpdateSubscription(ctx, existingSubscription, subscriptionEvent); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Send payment failed email notification
	if err := s.sendPaymentFailedEmail(ctx, existingSubscription.UserID, subscriptionID, stripeInvoice); err != nil {
		// Log error but don't fail the webhook processing
		common.Logger.Error("Failed to send payment failed email",
			"error", err,
			"user_id", existingSubscription.UserID)
	}

	// Log and capture error for monitoring
	common.CaptureError(ctx, fmt.Errorf("payment failed for subscription"), "Invoice payment failed",
		"invoice_id", stripeInvoice.ID,
		"subscription_id", subscriptionID,
		"user_id", existingSubscription.UserID,
		"amount_due", stripeInvoice.AmountDue,
		"currency", stripeInvoice.Currency,
		"attempt_count", stripeInvoice.AttemptCount,
		"next_payment_attempt", stripeInvoice.NextPaymentAttempt)

	return nil
}

// sendPaymentFailedEmail sends a notification email about failed payment
func (s *SubscriptionService) sendPaymentFailedEmail(ctx context.Context, userID int64, subscriptionID string, invoice stripe.Invoice) error {
	// Get user details
	users, err := s.userRepo.Find(ctx, repository.FindUserArgs{ID: &userID})
	if err != nil || len(users) == 0 {
		return fmt.Errorf("user not found: %d", userID)
	}
	user := &users[0]

	// Get the subscription from our database to get the plan name
	subscription, err := s.subRepo.GetSubscriptionByStripeID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription for email: %w", err)
	}
	if subscription == nil {
		return fmt.Errorf("subscription not found for email: %s", subscriptionID)
	}

	// Prepare email data
	emailData := mail.SubscriptionEmailData{
		PlanName:      string(subscription.Plan),
		AmountCents:   int(invoice.AmountDue),
		NextRetryDate: time.Unix(invoice.NextPaymentAttempt, 0),
	}

	// Send email using package-level mail function
	return s.mailService.SendPaymentFailedEmail(user, emailData)
}

// getOrCreateStripeCustomer gets existing or creates new Stripe customer
func (s *SubscriptionService) getOrCreateStripeCustomer(ctx context.Context, userID int64, email string) (string, error) {

	users, err := s.userRepo.Find(ctx, repository.FindUserArgs{ID: &userID})
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

// SubscriptionCheckOptions contains options for subscription limit checking
type SubscriptionCheckOptions struct {
	WordlistID                *int64
	HasOptimisticSubscription bool
}

// CheckSubscriptionLimits checks if user can perform action based on subscription
func (s *SubscriptionService) CheckSubscriptionLimits(ctx context.Context, userID int64, action model.UserAction, options *SubscriptionCheckOptions) error {
	// Get user's active subscription
	sub, err := s.subRepo.GetActiveSubscriptionForUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}
	if s.effectiveAccess != nil {
		legacyPlan := model.PlanFree
		if sub != nil {
			legacyPlan = sub.Plan
		}
		effectivePlan, accessErr := s.effectiveAccess.Plan(ctx, userID, legacyPlan)
		if accessErr != nil {
			return accessErr
		}
		if effectivePlan != model.PlanFree {
			return nil
		}
	}

	// If optimistic flag is present and user appears to be on free plan, verify with RevenueCat
	if options != nil && options.HasOptimisticSubscription && (sub == nil || sub.Plan == model.PlanFree) {
		hasPremium, err := s.revenueCatService.HasPremiumSubscription(ctx, strconv.FormatInt(userID, 10))
		if err != nil {
			common.Logger.Error("Failed to verify subscription with RevenueCat", "error", err, "user_id", userID)
			// Continue with local subscription check - don't fail on RevenueCat API errors
		} else if hasPremium {
			// User has premium subscription in RevenueCat but not locally - allow the operation
			common.Logger.Info("Allowing operation for user with RevenueCat premium subscription", "user_id", userID)
			return nil // Allow the operation
		}
	}

	// If no subscription or free plan, check limits
	if sub == nil || sub.Plan == model.PlanFree {
		switch action {
		case model.UserActionCreateWordlist:
			count, err := s.subRepo.CountUserWordlists(ctx, userID)
			if err != nil {
				return err
			}
			if count >= model.FreeWordlistLimit {
				return fmt.Errorf("free plan limit reached: maximum %d wordlist allowed", model.FreeWordlistLimit)
			}
		case model.UserActionAddWord:
			// Check word count if wordlist ID is provided
			if options != nil && options.WordlistID != nil {
				wordCount, err := s.subRepo.CountWordsInWordlist(ctx, *options.WordlistID)
				if err != nil {
					return err
				}
				if wordCount >= model.FreeWordsPerList {
					return fmt.Errorf("free plan limit reached: maximum %d words per wordlist", model.FreeWordsPerList)
				}
			}
		case model.UserActionChatSession:
			return fmt.Errorf("premium required to use chat feature")
		}
	}

	return nil
}

// CountWordsInWordlist counts the number of words in a wordlist
func (s *SubscriptionService) CountWordsInWordlist(ctx context.Context, wordlistID int64) (int, error) {
	return s.subRepo.CountWordsInWordlist(ctx, wordlistID)
}

// GetWebhookSecret returns the webhook secret for external use
func (s *SubscriptionService) GetWebhookSecret() string {
	return s.webhookSecret
}

// GetActiveSubscriptionForUser returns the active subscription for a user (includes grace period)
func (s *SubscriptionService) GetActiveSubscriptionForUser(ctx context.Context, userID int64) (*model.Subscription, error) {
	return s.subRepo.GetActiveSubscriptionForUser(ctx, userID)
}
