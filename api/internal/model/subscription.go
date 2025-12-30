package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// SubscriptionPlan represents the subscription tier
type SubscriptionPlan string

const (
	PlanFree    SubscriptionPlan = "free"
	PlanMonthly SubscriptionPlan = "monthly"
	PlanAnnual  SubscriptionPlan = "annual"
)

// Scan implements the sql.Scanner interface
func (p *SubscriptionPlan) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		*p = SubscriptionPlan(v)
	case []byte:
		*p = SubscriptionPlan(v)
	default:
		return fmt.Errorf("cannot scan type %T into SubscriptionPlan", value)
	}
	return nil
}

// Value implements the driver.Valuer interface
func (p SubscriptionPlan) Value() (driver.Value, error) {
	return string(p), nil
}

// SubscriptionStatus represents the subscription state
type SubscriptionStatus string

const (
	StatusActive   SubscriptionStatus = "active"
	StatusCanceled SubscriptionStatus = "cancelled"
	StatusPastDue  SubscriptionStatus = "past_due"
	StatusUnpaid   SubscriptionStatus = "unpaid"
)

// Scan implements the sql.Scanner interface
func (s *SubscriptionStatus) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		*s = SubscriptionStatus(v)
	case []byte:
		*s = SubscriptionStatus(v)
	default:
		return fmt.Errorf("cannot scan type %T into SubscriptionStatus", value)
	}
	return nil
}

// Value implements the driver.Valuer interface
func (s SubscriptionStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// SubscriptionProvider represents the payment provider
type SubscriptionProvider string

const (
	ProviderStripe     SubscriptionProvider = "stripe"
	ProviderRevenueCat SubscriptionProvider = "revenuecat"
)

// Scan implements the sql.Scanner interface
func (p *SubscriptionProvider) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		*p = SubscriptionProvider(v)
	case []byte:
		*p = SubscriptionProvider(v)
	default:
		return fmt.Errorf("cannot scan type %T into SubscriptionProvider", value)
	}
	return nil
}

// Value implements the driver.Valuer interface
func (p SubscriptionProvider) Value() (driver.Value, error) {
	return string(p), nil
}

// PlatformType represents the platform
type PlatformType string

const (
	PlatformIOS     PlatformType = "ios"
	PlatformAndroid PlatformType = "android"
	PlatformWeb     PlatformType = "web"
)

// Scan implements the sql.Scanner interface
func (p *PlatformType) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		*p = PlatformType(v)
	case []byte:
		*p = PlatformType(v)
	default:
		return fmt.Errorf("cannot scan type %T into PlatformType", value)
	}
	return nil
}

// Value implements the driver.Valuer interface
func (p PlatformType) Value() (driver.Value, error) {
	return string(p), nil
}

// Subscription represents a user's subscription
type Subscription struct {
	ID                       int64                `json:"id"`
	UserID                   int64                `json:"userId"`
	Provider                 SubscriptionProvider `json:"provider"`
	StripeSubscriptionID     *string              `json:"stripeSubscriptionId,omitempty"`
	StripeCustomerID         *string              `json:"stripeCustomerId,omitempty"`
	RevenueCatSubscriptionID *string              `json:"revenuecatSubscriptionId,omitempty"`
	AppStoreProductID        *string              `json:"appStoreProductId,omitempty"`
	Platform                 *PlatformType        `json:"platform,omitempty"`
	Plan                     SubscriptionPlan     `json:"plan"`
	Status                   SubscriptionStatus   `json:"status"`
	CurrentPeriodStart       time.Time            `json:"currentPeriodStart"`
	CurrentPeriodEnd         time.Time            `json:"currentPeriodEnd"`
	CancelAtPeriodEnd        bool                 `json:"cancelAtPeriodEnd"`
	CanceledAt               *time.Time           `json:"canceledAt,omitempty"`
	TrialEnd                 *time.Time           `json:"trialEnd,omitempty"`
	AmountCents              int                  `json:"amountCents"`
	Currency                 string               `json:"currency"`
	CreatedAt                time.Time            `json:"createdAt"`
	UpdatedAt                time.Time            `json:"updatedAt"`
}

// SubscriptionEvent represents a webhook event from any payment provider
type SubscriptionEvent struct {
	ID              int64                `json:"id"`
	SubscriptionID  int64                `json:"subscriptionId"`
	ExternalEventID string               `json:"externalEventId"` // Provider-specific event ID
	Provider        SubscriptionProvider `json:"provider"`        // stripe, revenuecat, etc.
	EventType       string               `json:"eventType"`
	EventData       string               `json:"eventData"` // JSON string
	ProcessedAt     time.Time            `json:"processedAt"`
}

// Pricing configuration
var SubscriptionPrices = map[SubscriptionPlan]struct {
	AmountCents int
	Interval    string
}{
	PlanMonthly: {
		AmountCents: 699, // $6.99
		Interval:    "month",
	},
	PlanAnnual: {
		AmountCents: 6990, // $69.90
		Interval:    "year",
	},
}

// Subscription limits
const (
	FreeWordlistLimit = 1
	FreeWordsPerList  = 10
)

// Grace period configuration
const (
	GracePeriodDays = 3 // Number of days to allow access after payment failure
)

// UserAction represents actions that require subscription limit checking
type UserAction int

const (
	UserActionCreateWordlist UserAction = iota
	UserActionAddWord
	UserActionChatSession
)

// IsActive returns true if the subscription is in a valid state
// Includes active subscriptions and past_due subscriptions within grace period
func (s *Subscription) IsActive() bool {
	if s.Status == StatusActive {
		return true
	}

	// Allow access during grace period for past_due subscriptions
	if s.Status == StatusPastDue {
		// Grace period starts from the subscription end date
		gracePeriodEnd := s.CurrentPeriodEnd.Add(GracePeriodDays * 24 * time.Hour)
		return time.Now().Before(gracePeriodEnd)
	}

	return false
}
