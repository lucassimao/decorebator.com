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

type revenueCatService struct {
	db         *pgxpool.Pool
	subRepo    *repository.SubscriptionRepository
	userRepo   *repository.UserRepository
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

func NewRevenueCatService(db *pgxpool.Pool) RevenueCatService {
	apiKey := os.Getenv("REVENUECAT_API_KEY")
	if apiKey == "" {
		common.Logger.Warn("REVENUECAT_API_KEY not set - RevenueCat support disabled")
	}

	baseURL := os.Getenv("REVENUECAT_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.revenuecat.com/v1"
	}

	return &revenueCatService{
		db:       db,
		subRepo:  repository.NewSubscriptionRepository(db),
		userRepo: &repository.UserRepository{Db: db},
		apiKey:   apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
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
func (s *revenueCatService) GetCustomerInfo(ctx context.Context, appUserID string) (*CustomerInfo, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("RevenueCat API key not configured")
	}

	url := fmt.Sprintf("%s/subscribers/%s", s.baseURL, appUserID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
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
func (s *revenueCatService) CreateOrUpdateSubscriptionFromRevenueCat(ctx context.Context, userID int64, customerInfo *CustomerInfo, platform model.PlatformType) error {
	// Check if user has premium entitlement
	entitlement, hasPremium := customerInfo.Subscriber.Entitlements[EntitlementPremium]
	if !hasPremium {
		// No active subscription
		return nil
	}

	// Determine plan based on product ID
	plan := s.GetPlanFromProductID(entitlement.ProductIdentifier)
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
func (s *revenueCatService) HandleWebhook(ctx context.Context, payload []byte) error {
	var webhook model.RevenueCatWebhook
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

	// Find user by RevenueCat customer ID
	users, err := s.userRepo.Find(repository.FindUserArgs{RevenueCatCustomerID: &webhook.Event.AppUserID})
	if err != nil || len(users) == 0 {
		common.Logger.Warn("User not found for RevenueCat customer",
			"app_user_id", webhook.Event.AppUserID,
			"error", err,
		)
		// Store event without user ID for debugging
		if err := s.storeRevenueCatEvent(ctx, &webhook.Event, nil); err != nil {
			return fmt.Errorf("failed to store event: %w", err)
		}
		return nil
	}
	user := &users[0]

	// Process event based on type
	if err := s.processRevenueCatEvent(ctx, webhook.Event, user.ID); err != nil {
		return fmt.Errorf("failed to process event: %w", err)
	}

	// Store the event
	if err := s.storeRevenueCatEvent(ctx, &webhook.Event, &user.ID); err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	return nil
}

func (s *revenueCatService) GetPlanFromProductID(productID string) model.SubscriptionPlan {
	switch productID {
	case ProductMonthlyIOS, ProductMonthlyAndroid:
		return model.PlanMonthly
	case ProductAnnualIOS, ProductAnnualAndroid:
		return model.PlanAnnual
	default:
		return model.PlanFree
	}
}

func (s *revenueCatService) getPlatformFromStore(store string) model.PlatformType {
	switch store {
	case "app_store", "APP_STORE":
		return model.PlatformIOS
	case "play_store", "PLAY_STORE":
		return model.PlatformAndroid
	default:
		return model.PlatformWeb
	}
}

func (s *revenueCatService) checkEventExists(ctx context.Context, eventID string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM revenuecat_events WHERE event_id = $1)",
		eventID,
	).Scan(&exists)
	return exists, err
}

func (s *revenueCatService) storeRevenueCatEvent(ctx context.Context, event *model.RevenueCatEvent, userID *int64) error {
	query := `
		INSERT INTO revenuecat_events 
		(user_id, event_id, event_type, app_user_id, product_id, entitlement_id, event_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (event_id) DO NOTHING
	`

	var entitlementID *string
	if len(event.EntitlementIDs) > 0 {
		entitlementID = &event.EntitlementIDs[0]
	}

	eventData, _ := json.Marshal(event)

	_, err := s.db.Exec(ctx, query,
		userID,
		event.ID,
		event.Type,
		event.AppUserID,
		event.ProductID,
		entitlementID,
		eventData,
	)

	return err
}

// LinkUserToRevenueCat links a user to their RevenueCat customer ID
func (s *revenueCatService) LinkUserToRevenueCat(ctx context.Context, userID int64, appUserID string) error {
	query := `
		UPDATE users 
		SET revenuecat_customer_id = $1 
		WHERE id = $2
	`

	_, err := s.db.Exec(ctx, query, appUserID, userID)
	return err
}

// RestorePurchases checks RevenueCat for any active subscriptions and updates local state
func (s *revenueCatService) RestorePurchases(ctx context.Context, userID int64, appUserID string, platform model.PlatformType) error {
	// Link user to RevenueCat customer
	if err := s.LinkUserToRevenueCat(ctx, userID, appUserID); err != nil {
		return fmt.Errorf("failed to link user: %w", err)
	}

	// In test mode, skip API calls and just return success
	if os.Getenv("ENV") == "test" || os.Getenv("TEST_MODE") == "true" {
		return nil
	}

	// Fetch customer info
	customerInfo, err := s.GetCustomerInfo(ctx, appUserID)
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
func (s *revenueCatService) processRevenueCatEvent(ctx context.Context, event model.RevenueCatEvent, userID int64) error {
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
func (s *revenueCatService) processSubscriptionEvent(ctx context.Context, event model.RevenueCatEvent, userID int64) error {
	platform := s.getPlatformFromStore(event.Store)

	// Fetch latest customer info and update subscription
	customerInfo, err := s.GetCustomerInfo(ctx, event.AppUserID)
	if err != nil {
		return fmt.Errorf("failed to get customer info: %w", err)
	}

	if err := s.CreateOrUpdateSubscriptionFromRevenueCat(ctx, userID, customerInfo, platform); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

// processCancellationEvent handles subscription cancellation events
func (s *revenueCatService) processCancellationEvent(ctx context.Context, event model.RevenueCatEvent, userID int64) error {
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
func (s *revenueCatService) processBillingIssueEvent(ctx context.Context, userID int64) error {
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
