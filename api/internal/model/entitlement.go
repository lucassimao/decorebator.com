package model

import (
	"fmt"
	"strings"
	"time"
)

// EntitlementStore identifies the first-party store that verified a purchase.
type EntitlementStore string

const (
	EntitlementStoreApple  EntitlementStore = "apple"
	EntitlementStoreGoogle EntitlementStore = "google"
)

func (s EntitlementStore) Valid() bool {
	return s == EntitlementStoreApple || s == EntitlementStoreGoogle
}

// StoreEnvironment separates test purchases from production entitlements.
// Google Play test purchases map to sandbox in this store-neutral vocabulary.
type StoreEnvironment string

const (
	StoreEnvironmentSandbox    StoreEnvironment = "sandbox"
	StoreEnvironmentProduction StoreEnvironment = "production"
)

func (e StoreEnvironment) Valid() bool {
	return e == StoreEnvironmentSandbox || e == StoreEnvironmentProduction
}

// EntitlementStatus describes access state without copying provider-specific
// status strings. Cancellation is represented separately because a canceled
// subscription may remain entitled through its paid period.
type EntitlementStatus string

const (
	EntitlementStatusPending     EntitlementStatus = "pending"
	EntitlementStatusActive      EntitlementStatus = "active"
	EntitlementStatusGracePeriod EntitlementStatus = "grace_period"
	EntitlementStatusOnHold      EntitlementStatus = "on_hold"
	EntitlementStatusPaused      EntitlementStatus = "paused"
	EntitlementStatusExpired     EntitlementStatus = "expired"
	EntitlementStatusRevoked     EntitlementStatus = "revoked"
)

func (s EntitlementStatus) Valid() bool {
	switch s {
	case EntitlementStatusPending,
		EntitlementStatusActive,
		EntitlementStatusGracePeriod,
		EntitlementStatusOnHold,
		EntitlementStatusPaused,
		EntitlementStatusExpired,
		EntitlementStatusRevoked:
		return true
	default:
		return false
	}
}

// StoreEntitlement is the canonical server-owned access record. Store-specific
// transaction and account bindings are defined separately from this lifecycle
// model so clients cannot grant access by constructing this object.
type StoreEntitlement struct {
	ID               int64             `json:"id"`
	UserID           int64             `json:"userId"`
	Store            EntitlementStore  `json:"store"`
	ProductID        string            `json:"productId"`
	Entitlement      string            `json:"entitlement"`
	Status           EntitlementStatus `json:"status"`
	PeriodStart      *time.Time        `json:"periodStart,omitempty"`
	PeriodEnd        *time.Time        `json:"periodEnd,omitempty"`
	AutoRenewEnabled bool              `json:"autoRenewEnabled"`
	CanceledAt       *time.Time        `json:"canceledAt,omitempty"`
	RevokedAt        *time.Time        `json:"revokedAt,omitempty"`
	RevocationReason *string           `json:"revocationReason,omitempty"`
	Environment      StoreEnvironment  `json:"environment"`
	LastVerifiedAt   time.Time         `json:"lastVerifiedAt"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`
}

func (e StoreEntitlement) Validate() error {
	if e.UserID <= 0 {
		return fmt.Errorf("user ID must be positive")
	}
	if !e.Store.Valid() {
		return fmt.Errorf("unsupported entitlement store %q", e.Store)
	}
	if strings.TrimSpace(e.ProductID) == "" {
		return fmt.Errorf("product ID is required")
	}
	if strings.TrimSpace(e.Entitlement) == "" {
		return fmt.Errorf("entitlement is required")
	}
	if !e.Status.Valid() {
		return fmt.Errorf("unsupported entitlement status %q", e.Status)
	}
	if !e.Environment.Valid() {
		return fmt.Errorf("unsupported store environment %q", e.Environment)
	}
	if e.LastVerifiedAt.IsZero() {
		return fmt.Errorf("last verified time is required")
	}
	if e.PeriodStart != nil && e.PeriodEnd != nil && e.PeriodEnd.Before(*e.PeriodStart) {
		return fmt.Errorf("period end cannot be before period start")
	}
	if (e.Status == EntitlementStatusActive || e.Status == EntitlementStatusGracePeriod) && e.PeriodEnd == nil {
		return fmt.Errorf("period end is required for status %q", e.Status)
	}
	if e.Status == EntitlementStatusRevoked && e.RevokedAt == nil {
		return fmt.Errorf("revoked at is required for revoked status")
	}
	if e.Status != EntitlementStatusRevoked && e.RevokedAt != nil {
		return fmt.Errorf("revoked status is required when revoked at is set")
	}

	return nil
}

// GrantsAccess is deliberately conservative: only active and grace-period
// records with an unexpired verified period grant premium access.
func (e StoreEntitlement) GrantsAccess(now time.Time) bool {
	if e.Status != EntitlementStatusActive && e.Status != EntitlementStatusGracePeriod {
		return false
	}
	return e.PeriodEnd != nil && now.Before(*e.PeriodEnd)
}
