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
	event := &model.RevenueCatEvent{
		EventID:   webhook.Event.ID,
		EventType: webhook.Event.Type,
		AppUserID: webhook.Event.AppUserID,
		ProductID: &webhook.Event.ProductID,
		EventData: string(eventData),
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
	switch webhook.Event.Type {
	case EventInitialPurchase, EventRenewal, EventUncancellation:
		// Fetch latest customer info and update subscription
		platform := s.getPlatformFromStore(webhook.Event.Store)
		customerInfo, err := s.GetCustomerInfo(ctx, webhook.Event.AppUserID, platform)
		if err != nil {
			return fmt.Errorf("failed to get customer info: %w", err)
		}

		if err := s.CreateOrUpdateSubscriptionFromRevenueCat(ctx, user.ID, customerInfo, platform); err != nil {
			return fmt.Errorf("failed to update subscription: %w", err)
		}

	case EventCancellation, EventExpiration:
		// Mark subscription as cancelled
		sub, err := s.subRepo.GetActiveSubscriptionForUser(ctx, user.ID)
		if err != nil {
			return fmt.Errorf("failed to get subscription: %w", err)
		}

		if sub != nil && sub.Provider == model.ProviderRevenueCat {
			sub.Status = model.StatusCanceled
			now := time.Now()
			sub.CancelledAt = &now

			if webhook.Event.Type == EventCancellation {
				sub.CancelAtPeriodEnd = true
			}

			if err := s.subRepo.UpdateSubscription(ctx, sub); err != nil {
				return fmt.Errorf("failed to update subscription: %w", err)
			}
		}

	case EventBillingIssue:
		// Mark subscription as past due
		sub, err := s.subRepo.GetActiveSubscriptionForUser(ctx, user.ID)
		if err != nil {
			return fmt.Errorf("failed to get subscription: %w", err)
		}

		if sub != nil && sub.Provider == model.ProviderRevenueCat {
			sub.Status = model.StatusPastDue
			if err := s.subRepo.UpdateSubscription(ctx, sub); err != nil {
				return fmt.Errorf("failed to update subscription: %w", err)
			}
		}
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
	ID                 string  `json:"id"`
	Type               string  `json:"type"`
	AppID              string  `json:"app_id"`
	AppUserID          string  `json:"app_user_id"`
	ProductID          string  `json:"product_id"`
	EntitlementID      string  `json:"entitlement_id"`
	PeriodType         string  `json:"period_type"`
	PurchasedAtMS      int64   `json:"purchased_at_ms"`
	ExpirationAtMS     int64   `json:"expiration_at_ms"`
	Environment        string  `json:"environment"`
	Store              string  `json:"store"`
	IsFamilyShare      bool    `json:"is_family_share"`
	Currency           string  `json:"currency"`
	Price              float64 `json:"price"`
	TakehomePercentage float64 `json:"takehome_percentage"`
	EventTimestampMS   int64   `json:"event_timestamp_ms"`
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
	case "app_store":
		return model.PlatformIOS
	case "play_store":
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

// RestorePurchases checks RevenueCat for any active subscriptions and updates local state
func (s *RevenueCatService) RestorePurchases(ctx context.Context, userID int64, appUserID string, platform model.PlatformType) error {
	// Link user to RevenueCat customer
	if err := s.LinkUserToRevenueCat(ctx, userID, appUserID); err != nil {
		return fmt.Errorf("failed to link user: %w", err)
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
