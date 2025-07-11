package service

import (
	"context"
	"encoding/json"
	"fmt"
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
	EntitlementPremium = "Premium"
)

// RevenueCat product IDs
const (
	ProductMonthlyIOS     = "com.decorebator.premium.monthly"
	ProductAnnualIOS      = "com.decorebator.premium.annual"
	ProductMonthlyAndroid = "prode6374c2df7"
	ProductAnnualAndroid  = "prodbdd5434ac7"
)

type revenueCatService struct {
	db        *pgxpool.Pool
	subRepo   *repository.SubscriptionRepository
	userRepo  *repository.UserRepository
	apiClient RevenueCatAPIClient
}

func NewRevenueCatService(db *pgxpool.Pool, apiClient RevenueCatAPIClient) RevenueCatService {
	return &revenueCatService{
		db:        db,
		subRepo:   repository.NewSubscriptionRepository(db),
		userRepo:  &repository.UserRepository{Db: db},
		apiClient: apiClient,
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

// getCustomerInfo fetches customer info from RevenueCat (internal use)
func (s *revenueCatService) getCustomerInfo(ctx context.Context, appUserID string) (*CustomerInfo, error) {
	return s.apiClient.GetCustomerInfo(ctx, appUserID)
}

// createOrUpdateSubscriptionFromRevenueCat creates or updates a subscription based on RevenueCat data
func (s *revenueCatService) createOrUpdateSubscriptionFromRevenueCat(ctx context.Context, userID int64, customerInfo *CustomerInfo, platform model.PlatformType, event *model.RevenueCatEvent) error {
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

	// Create subscription event if event data is provided
	var subscriptionEvent *model.SubscriptionEvent
	if event != nil {
		eventData, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event data: %w", err)
		}

		subscriptionEvent = &model.SubscriptionEvent{
			ExternalEventID: event.ID,
			Provider:        model.ProviderRevenueCat,
			EventType:       event.Type,
			EventData:       string(eventData),
		}
	}

	if existing != nil && existing.Provider == model.ProviderRevenueCat {
		// Update existing RevenueCat subscription
		existing.Plan = plan
		existing.Status = status
		existing.CurrentPeriodEnd = expiresDate
		existing.Platform = &platform
		existing.AppStoreProductID = &entitlement.ProductIdentifier

		// Create or use provided subscription event
		if subscriptionEvent == nil {
			// Create a manual restore event
			subscriptionEvent = &model.SubscriptionEvent{
				ExternalEventID: fmt.Sprintf("manual_restore_%d_%d", userID, time.Now().Unix()),
				Provider:        model.ProviderRevenueCat,
				EventType:       "manual_restore",
				EventData:       `{"type": "manual_restore", "source": "customer_info_sync"}`,
			}
		}
		subscriptionEvent.SubscriptionID = existing.ID

		if err := s.subRepo.UpdateSubscription(ctx, existing, *subscriptionEvent); err != nil {
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

		// Create or use provided subscription event
		if subscriptionEvent == nil {
			// Create a manual restore event for new subscription
			subscriptionEvent = &model.SubscriptionEvent{
				ExternalEventID: fmt.Sprintf("manual_restore_new_%d_%d", userID, time.Now().Unix()),
				Provider:        model.ProviderRevenueCat,
				EventType:       "manual_restore",
				EventData:       `{"type": "manual_restore", "source": "customer_info_sync", "action": "create"}`,
			}
		}

		if _, err := s.subRepo.CreateSubscription(ctx, sub, *subscriptionEvent); err != nil {
			return fmt.Errorf("failed to create subscription: %w", err)
		}
	}

	return nil
}

func (s *revenueCatService) getPlanFromProductID(productID string) model.SubscriptionPlan {
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

func (s *revenueCatService) CheckEventExists(ctx context.Context, eventID string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM revenuecat_events WHERE event_id = $1)",
		eventID,
	).Scan(&exists)
	return exists, err
}

func (s *revenueCatService) StoreRevenueCatEvent(ctx context.Context, event *model.RevenueCatEvent, userID *int64) error {
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

	// Fetch customer info
	customerInfo, err := s.getCustomerInfo(ctx, appUserID)
	if err != nil {
		return fmt.Errorf("failed to get customer info: %w", err)
	}

	// Update subscription based on customer info (no event for manual restore)
	if err := s.createOrUpdateSubscriptionFromRevenueCat(ctx, userID, customerInfo, platform, nil); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

// ProcessRevenueCatEvent processes different types of RevenueCat events
func (s *revenueCatService) ProcessRevenueCatEvent(ctx context.Context, event model.RevenueCatEvent, userID int64) error {
	switch event.Type {
	case EventInitialPurchase, EventRenewal, EventUncancellation:
		return s.processSubscriptionEvent(ctx, event, userID)
	case EventCancellation, EventExpiration:
		return s.processCancellationEvent(ctx, event, userID)
	case EventBillingIssue:
		return s.processBillingIssueEvent(ctx, event, userID)
	default:
		common.Logger.Info("Unhandled RevenueCat event type", "type", event.Type)
		return nil
	}
}

// processSubscriptionEvent handles subscription creation/renewal events
func (s *revenueCatService) processSubscriptionEvent(ctx context.Context, event model.RevenueCatEvent, userID int64) error {
	platform := s.getPlatformFromStore(event.Store)

	// Fetch latest customer info and update subscription
	customerInfo, err := s.getCustomerInfo(ctx, event.AppUserID)
	if err != nil {
		return fmt.Errorf("failed to get customer info: %w", err)
	}

	if err := s.createOrUpdateSubscriptionFromRevenueCat(ctx, userID, customerInfo, platform, &event); err != nil {
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
		sub.CanceledAt = &now

		if event.Type == EventCancellation {
			sub.CancelAtPeriodEnd = true
		}

		// Create subscription event
		eventData, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event data: %w", err)
		}

		subscriptionEvent := &model.SubscriptionEvent{
			SubscriptionID:  sub.ID,
			ExternalEventID: event.ID,
			Provider:        model.ProviderRevenueCat,
			EventType:       event.Type,
			EventData:       string(eventData),
		}

		if err := s.subRepo.UpdateSubscription(ctx, sub, *subscriptionEvent); err != nil {
			return fmt.Errorf("failed to update subscription: %w", err)
		}
	}

	return nil
}

// processBillingIssueEvent handles billing issue events
func (s *revenueCatService) processBillingIssueEvent(ctx context.Context, event model.RevenueCatEvent, userID int64) error {
	sub, err := s.subRepo.GetActiveSubscriptionForUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if sub != nil && sub.Provider == model.ProviderRevenueCat {
		sub.Status = model.StatusPastDue

		// Create subscription event
		eventData, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event data: %w", err)
		}

		subscriptionEvent := &model.SubscriptionEvent{
			SubscriptionID:  sub.ID,
			ExternalEventID: event.ID,
			Provider:        model.ProviderRevenueCat,
			EventType:       event.Type,
			EventData:       string(eventData),
		}

		if err := s.subRepo.UpdateSubscription(ctx, sub, *subscriptionEvent); err != nil {
			return fmt.Errorf("failed to update subscription: %w", err)
		}
	}

	return nil
}

// GetUserByRevenueCatCustomerID finds a user by their RevenueCat customer ID
func (s *revenueCatService) GetUserByRevenueCatCustomerID(ctx context.Context, appUserID string) (*model.User, error) {
	users, err := s.userRepo.Find(ctx, repository.FindUserArgs{RevenueCatCustomerID: &appUserID})
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	return &users[0], nil
}
