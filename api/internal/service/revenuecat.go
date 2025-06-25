package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RevenueCat API endpoints
const (
	revenueCatAPIBaseURL = "https://api.revenuecat.com/v1"
)

// RevenueCat webhook event types
const (
	EventInitialPurchase     = "INITIAL_PURCHASE"
	EventRenewal             = "RENEWAL"
	EventCancellation        = "CANCELLATION"
	EventUncancellation      = "UNCANCELLATION"
	EventNonRenewingPurchase = "NON_RENEWING_PURCHASE"
	EventExpiration          = "EXPIRATION"
	EventProductChange       = "PRODUCT_CHANGE"
	EventBillingIssue        = "BILLING_ISSUE"
	EventSubscriberAlias     = "SUBSCRIBER_ALIAS"
)

// RevenueCat entitlement IDs (configured in RevenueCat dashboard)
const (
	EntitlementPremium = "premium"
)

// RevenueCat product IDs (must match App Store/Google Play product IDs)
const (
	ProductMonthlyIOS     = "com.decorebator.premium.monthly"
	ProductAnnualIOS      = "com.decorebator.premium.annual"
	ProductMonthlyAndroid = "premium_monthly"
	ProductAnnualAndroid  = "premium_annual"
)

type RevenueCatService struct {
	db            *pgxpool.Pool
	subRepo       *repository.SubscriptionRepository
	userRepo      *repository.UserRepository
	apiKeyIOS     string
	apiKeyAndroid string
	webhookSecret string
	httpClient    *http.Client
}

func NewRevenueCatService(db *pgxpool.Pool) *RevenueCatService {
	// Get API keys from environment
	apiKeyIOS := os.Getenv("REVENUECAT_API_KEY_IOS")
	if apiKeyIOS == "" {
		common.Logger.Warn("REVENUECAT_API_KEY_IOS not set - RevenueCat iOS support disabled")
	}

	apiKeyAndroid := os.Getenv("REVENUECAT_API_KEY_ANDROID")
	if apiKeyAndroid == "" {
		common.Logger.Warn("REVENUECAT_API_KEY_ANDROID not set - RevenueCat Android support disabled")
	}

	webhookSecret := os.Getenv("REVENUECAT_WEBHOOK_SECRET")
	if webhookSecret == "" {
		common.Logger.Warn("REVENUECAT_WEBHOOK_SECRET not set - webhook verification disabled")
	}

	return &RevenueCatService{
		db:            db,
		subRepo:       repository.NewSubscriptionRepository(db),
		userRepo:      &repository.UserRepository{Db: db},
		apiKeyIOS:     apiKeyIOS,
		apiKeyAndroid: apiKeyAndroid,
		webhookSecret: webhookSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CustomerInfo represents RevenueCat customer information
type CustomerInfo struct {
	RequestDate   string     `json:"request_date"`
	RequestDateMS int64      `json:"request_date_ms"`
	Subscriber    Subscriber `json:"subscriber"`
}

type Subscriber struct {
	Entitlements      map[string]Entitlement  `json:"entitlements"`
	FirstSeen         string                  `json:"first_seen"`
	LastSeen          string                  `json:"last_seen"`
	OriginalAppUserID string                  `json:"original_app_user_id"`
	Subscriptions     map[string]Subscription `json:"subscriptions"`
}

type Entitlement struct {
	ExpiresDate       *string `json:"expires_date"`
	ProductIdentifier string  `json:"product_identifier"`
	PurchaseDate      string  `json:"purchase_date"`
}

type Subscription struct {
	BillingIssuesDetectedAt *string `json:"billing_issues_detected_at"`
	ExpiresDate             string  `json:"expires_date"`
	GracePeriodExpiresDate  *string `json:"grace_period_expires_date"`
	IsSandbox               bool    `json:"is_sandbox"`
	OriginalPurchaseDate    string  `json:"original_purchase_date"`
	PeriodType              string  `json:"period_type"`
	PurchaseDate            string  `json:"purchase_date"`
	RefundedAt              *string `json:"refunded_at"`
	Store                   string  `json:"store"`
	UnsubscribeDetectedAt   *string `json:"unsubscribe_detected_at"`
	AutoResumeDate          *string `json:"auto_resume_date"`
}

// GetCustomerInfo fetches customer info from RevenueCat
func (s *RevenueCatService) GetCustomerInfo(ctx context.Context, appUserID string, platform model.PlatformType) (*CustomerInfo, error) {
	// Select API key based on platform
	apiKey := s.getAPIKey(platform)
	if apiKey == "" {
		return nil, fmt.Errorf("RevenueCat API key not configured for platform: %s", platform)
	}

	url := fmt.Sprintf("%s/subscribers/%s", revenueCatAPIBaseURL, appUserID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch customer info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RevenueCat API error: %d - %s", resp.StatusCode, string(body))
	}

	var customerInfo CustomerInfo
	if err := json.NewDecoder(resp.Body).Decode(&customerInfo); err != nil {
		return nil, fmt.Errorf("failed to decode customer info: %w", err)
	}

	return &customerInfo, nil
}

// CreateOrUpdateSubscriptionFromRevenueCat creates or updates a subscription based on RevenueCat data
func (s *RevenueCatService) CreateOrUpdateSubscriptionFromRevenueCat(ctx context.Context, userID int64, customerInfo *CustomerInfo, platform model.PlatformType) error {
	// Check if user has premium entitlement
	entitlement, hasPremium := customerInfo.Subscriber.Entitlements[EntitlementPremium]
	if !hasPremium {
		// No active subscription
		return nil
	}

	// Determine plan based on product ID
	plan := s.getPlanFromProductID(entitlement.ProductIdentifier)
	if plan == model.PlanFree {
		return fmt.Errorf("unknown product ID: %s", entitlement.ProductIdentifier)
	}

	// Parse dates
	expiresDate, err := time.Parse(time.RFC3339, *entitlement.ExpiresDate)
	if err != nil {
		return fmt.Errorf("failed to parse expires date: %w", err)
	}

	purchaseDate, err := time.Parse(time.RFC3339, entitlement.PurchaseDate)
	if err != nil {
		return fmt.Errorf("failed to parse purchase date: %w", err)
	}

	// Determine subscription status
	status := model.StatusActive
	if time.Now().After(expiresDate) {
		status = model.StatusCanceled
	}

	// Check for existing subscription
	existing, err := s.subRepo.GetActiveSubscriptionForUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get existing subscription: %w", err)
	}

	if existing != nil && existing.Provider == model.ProviderRevenueCat {
		// Update existing RevenueCat subscription
		existing.Plan = plan
		existing.Status = status
		existing.CurrentPeriodEnd = expiresDate
		existing.Platform = &platform
		existing.AppStoreProductID = &entitlement.ProductIdentifier

		if err := s.subRepo.UpdateSubscription(ctx, existing); err != nil {
			return fmt.Errorf("failed to update subscription: %w", err)
		}
	} else {
		// Create new subscription
		sub := &model.Subscription{
			UserID:                   userID,
			Provider:                 model.ProviderRevenueCat,
			RevenueCatSubscriptionID: &customerInfo.Subscriber.OriginalAppUserID,
			AppStoreProductID:        &entitlement.ProductIdentifier,
			Platform:                 &platform,
			Plan:                     plan,
			Status:                   status,
			CurrentPeriodStart:       purchaseDate,
			CurrentPeriodEnd:         expiresDate,
			AmountCents:              model.SubscriptionPrices[plan].AmountCents,
			Currency:                 "USD",
		}

		if _, err := s.subRepo.CreateSubscription(ctx, sub); err != nil {
			return fmt.Errorf("failed to create subscription: %w", err)
		}
	}

	return nil
}

// HandleWebhook processes RevenueCat webhook events
func (s *RevenueCatService) HandleWebhook(ctx context.Context, payload []byte, authHeader string) error {
	// Verify webhook if secret is configured
	if s.webhookSecret != "" && authHeader != s.webhookSecret {
		return fmt.Errorf("invalid webhook authorization")
	}

	var webhook WebhookPayload
	if err := json.Unmarshal(payload, &webhook); err != nil {
		return fmt.Errorf("failed to unmarshal webhook: %w", err)
	}

	// Log webhook event
	common.Logger.Info("RevenueCat webhook received",
		"event_type", webhook.Event.Type,
		"app_user_id", webhook.Event.AppUserID,
		"product_id", webhook.Event.ProductID,
		"entitlement_ids", webhook.Event.EntitlementIDs,
	)

	// Check if we've already processed this event
	exists, err := s.checkEventExists(ctx, webhook.Event.ID)
	if err != nil {
		return fmt.Errorf("failed to check event existence: %w", err)
	}
	if exists {
		common.Logger.Info("RevenueCat event already processed", "event_id", webhook.Event.ID)
		return nil
	}

	// Store the event
	eventData, _ := json.Marshal(webhook)

	// Get first entitlement ID if available
	var entitlementID *string
	if len(webhook.Event.EntitlementIDs) > 0 {
		entitlementID = &webhook.Event.EntitlementIDs[0]
	}

	event := &model.RevenueCatEvent{
		EventID:       webhook.Event.ID,
		EventType:     webhook.Event.Type,
		AppUserID:     webhook.Event.AppUserID,
		ProductID:     &webhook.Event.ProductID,
		EntitlementID: entitlementID,
		EventData:     string(eventData),
	}

	// Find user by RevenueCat customer ID
	user, err := s.userRepo.GetUserByRevenueCatCustomerID(ctx, webhook.Event.AppUserID)
	if err != nil {
		common.Logger.Warn("User not found for RevenueCat customer",
			"app_user_id", webhook.Event.AppUserID,
			"error", err,
		)
		// Store event without user ID for debugging
		if err := s.storeRevenueCatEvent(ctx, event); err != nil {
			return fmt.Errorf("failed to store event: %w", err)
		}
		return nil
	}

	event.UserID = &user.ID

	// Process event based on type
	if err := s.processRevenueCatEvent(ctx, webhook.Event, user.ID); err != nil {
		return fmt.Errorf("failed to process event: %w", err)
	}

	// Store the event
	if err := s.storeRevenueCatEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	return nil
}

// WebhookPayload represents the RevenueCat webhook structure
type WebhookPayload struct {
	APIVersion string       `json:"api_version"`
	Event      WebhookEvent `json:"event"`
}

type WebhookEvent struct {
	ID                    string   `json:"id"`
	Type                  string   `json:"type"`
	AppID                 string   `json:"app_id"`
	AppUserID             string   `json:"app_user_id"`
	OriginalAppUserID     string   `json:"original_app_user_id,omitempty"`
	ProductID             string   `json:"product_id"`
	EntitlementIDs        []string `json:"entitlement_ids,omitempty"`
	PeriodType            string   `json:"period_type,omitempty"`
	PurchasedAtMS         int64    `json:"purchased_at_ms,omitempty"`
	ExpirationAtMS        int64    `json:"expiration_at_ms,omitempty"`
	Environment           string   `json:"environment"`
	Store                 string   `json:"store"`
	Currency              string   `json:"currency,omitempty"`
	Price                 float64  `json:"price,omitempty"`
	TransactionID         string   `json:"transaction_id,omitempty"`
	OriginalTransactionID string   `json:"original_transaction_id,omitempty"`
	CountryCode           string   `json:"country_code,omitempty"`
	EventTimestampMS      int64    `json:"event_timestamp_ms"`
	Aliases               []string `json:"aliases,omitempty"`
}

// Helper methods

func (s *RevenueCatService) getAPIKey(platform model.PlatformType) string {
	switch platform {
	case model.PlatformIOS:
		return s.apiKeyIOS
	case model.PlatformAndroid:
		return s.apiKeyAndroid
	default:
		return ""
	}
}

func (s *RevenueCatService) getPlanFromProductID(productID string) model.SubscriptionPlan {
	switch productID {
	case ProductMonthlyIOS, ProductMonthlyAndroid:
		return model.PlanMonthly
	case ProductAnnualIOS, ProductAnnualAndroid:
		return model.PlanAnnual
	default:
		return model.PlanFree
	}
}

func (s *RevenueCatService) getPlatformFromStore(store string) model.PlatformType {
	switch store {
	case "app_store", "APP_STORE":
		return model.PlatformIOS
	case "play_store", "PLAY_STORE":
		return model.PlatformAndroid
	default:
		return model.PlatformWeb
	}
}

func (s *RevenueCatService) checkEventExists(ctx context.Context, eventID string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM revenuecat_events WHERE event_id = $1)",
		eventID,
	).Scan(&exists)
	return exists, err
}

func (s *RevenueCatService) storeRevenueCatEvent(ctx context.Context, event *model.RevenueCatEvent) error {
	query := `
		INSERT INTO revenuecat_events 
		(user_id, event_id, event_type, app_user_id, product_id, entitlement_id, event_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (event_id) DO NOTHING
	`

	_, err := s.db.Exec(ctx, query,
		event.UserID,
		event.EventID,
		event.EventType,
		event.AppUserID,
		event.ProductID,
		event.EntitlementID,
		event.EventData,
	)

	return err
}

// LinkUserToRevenueCat links a user to their RevenueCat customer ID
func (s *RevenueCatService) LinkUserToRevenueCat(ctx context.Context, userID int64, appUserID string) error {
	query := `
		UPDATE users 
		SET revenuecat_customer_id = $1 
		WHERE id = $2
	`

	_, err := s.db.Exec(ctx, query, appUserID, userID)
	return err
}

// createSubscriptionFromWebhook creates a subscription directly from webhook data (test mode only)
func (s *RevenueCatService) createSubscriptionFromWebhook(ctx context.Context, userID int64, event WebhookEvent, platform model.PlatformType) error {
	// Determine plan from product ID
	plan := s.getPlanFromProductID(event.ProductID)
	if plan == model.PlanFree {
		return fmt.Errorf("unknown product ID: %s", event.ProductID)
	}

	// Convert timestamps to time.Time
	purchaseDate := time.UnixMilli(event.PurchasedAtMS)
	expirationDate := time.UnixMilli(event.ExpirationAtMS)

	// Check for existing subscription
	existing, err := s.subRepo.GetActiveSubscriptionForUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get existing subscription: %w", err)
	}

	if existing != nil && existing.Provider == model.ProviderRevenueCat {
		// Update existing RevenueCat subscription
		existing.Plan = plan
		existing.Status = model.StatusActive
		existing.CurrentPeriodStart = purchaseDate
		existing.CurrentPeriodEnd = expirationDate
		existing.Platform = &platform
		existing.AppStoreProductID = &event.ProductID

		if err := s.subRepo.UpdateSubscription(ctx, existing); err != nil {
			return fmt.Errorf("failed to update subscription: %w", err)
		}
	} else {
		// Create new subscription
		sub := &model.Subscription{
			UserID:                   userID,
			Provider:                 model.ProviderRevenueCat,
			RevenueCatSubscriptionID: &event.AppUserID,
			AppStoreProductID:        &event.ProductID,
			Platform:                 &platform,
			Plan:                     plan,
			Status:                   model.StatusActive,
			CurrentPeriodStart:       purchaseDate,
			CurrentPeriodEnd:         expirationDate,
			AmountCents:              model.SubscriptionPrices[plan].AmountCents,
			Currency:                 event.Currency,
		}

		if _, err := s.subRepo.CreateSubscription(ctx, sub); err != nil {
			return fmt.Errorf("failed to create subscription: %w", err)
		}
	}

	return nil
}

// RestorePurchases checks RevenueCat for any active subscriptions and updates local state
func (s *RevenueCatService) RestorePurchases(ctx context.Context, userID int64, appUserID string, platform model.PlatformType) error {
	// Link user to RevenueCat customer
	if err := s.LinkUserToRevenueCat(ctx, userID, appUserID); err != nil {
		return fmt.Errorf("failed to link user: %w", err)
	}

	// In test mode, skip API calls and just return success
	if os.Getenv("ENV") == "test" || os.Getenv("TEST_MODE") == "true" {
		return nil
	}

	// Fetch customer info
	customerInfo, err := s.GetCustomerInfo(ctx, appUserID, platform)
	if err != nil {
		return fmt.Errorf("failed to get customer info: %w", err)
	}

	// Update subscription based on customer info
	if err := s.CreateOrUpdateSubscriptionFromRevenueCat(ctx, userID, customerInfo, platform); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

// processRevenueCatEvent processes different types of RevenueCat events
func (s *RevenueCatService) processRevenueCatEvent(ctx context.Context, event WebhookEvent, userID int64) error {
	switch event.Type {
	case EventInitialPurchase, EventRenewal, EventUncancellation:
		return s.processSubscriptionEvent(ctx, event, userID)
	case EventCancellation, EventExpiration:
		return s.processCancellationEvent(ctx, event, userID)
	case EventBillingIssue:
		return s.processBillingIssueEvent(ctx, userID)
	default:
		common.Logger.Info("Unhandled RevenueCat event type", "type", event.Type)
		return nil
	}
}

// processSubscriptionEvent handles subscription creation/renewal events
func (s *RevenueCatService) processSubscriptionEvent(ctx context.Context, event WebhookEvent, userID int64) error {
	platform := s.getPlatformFromStore(event.Store)

	// In test mode, create subscription directly from webhook data
	if os.Getenv("ENV") == "test" || os.Getenv("TEST_MODE") == "true" {
		if err := s.createSubscriptionFromWebhook(ctx, userID, event, platform); err != nil {
			return fmt.Errorf("failed to create subscription from webhook: %w", err)
		}
	} else {
		// Fetch latest customer info and update subscription
		customerInfo, err := s.GetCustomerInfo(ctx, event.AppUserID, platform)
		if err != nil {
			return fmt.Errorf("failed to get customer info: %w", err)
		}

		if err := s.CreateOrUpdateSubscriptionFromRevenueCat(ctx, userID, customerInfo, platform); err != nil {
			return fmt.Errorf("failed to update subscription: %w", err)
		}
	}

	return nil
}

// processCancellationEvent handles subscription cancellation events
func (s *RevenueCatService) processCancellationEvent(ctx context.Context, event WebhookEvent, userID int64) error {
	sub, err := s.subRepo.GetActiveSubscriptionForUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if sub != nil && sub.Provider == model.ProviderRevenueCat {
		sub.Status = model.StatusCanceled
		now := time.Now()
		sub.CancelledAt = &now

		if event.Type == EventCancellation {
			sub.CancelAtPeriodEnd = true
		}

		if err := s.subRepo.UpdateSubscription(ctx, sub); err != nil {
			return fmt.Errorf("failed to update subscription: %w", err)
		}
	}

	return nil
}

// processBillingIssueEvent handles billing issue events
func (s *RevenueCatService) processBillingIssueEvent(ctx context.Context, userID int64) error {
	sub, err := s.subRepo.GetActiveSubscriptionForUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if sub != nil && sub.Provider == model.ProviderRevenueCat {
		sub.Status = model.StatusPastDue
		if err := s.subRepo.UpdateSubscription(ctx, sub); err != nil {
			return fmt.Errorf("failed to update subscription: %w", err)
		}
	}

	return nil
}
