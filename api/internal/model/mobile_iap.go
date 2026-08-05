package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type IAPBillingPeriod string

const (
	IAPBillingPeriodMonthly IAPBillingPeriod = "monthly"
	IAPBillingPeriodAnnual  IAPBillingPeriod = "annual"
)

func (p IAPBillingPeriod) Valid() bool {
	return p == IAPBillingPeriodMonthly || p == IAPBillingPeriodAnnual
}

// MobileIAPProduct is server-owned catalog metadata. Localized title, price,
// currency, and offer details come from StoreKit/Google Play for this product ID.
type MobileIAPProduct struct {
	Store         EntitlementStore `json:"store"`
	ProductID     string           `json:"productId"`
	Entitlement   string           `json:"entitlement"`
	BillingPeriod IAPBillingPeriod `json:"billingPeriod"`
}

func (p MobileIAPProduct) Validate() error {
	if !p.Store.Valid() {
		return fmt.Errorf("unsupported product store %q", p.Store)
	}
	if strings.TrimSpace(p.ProductID) == "" {
		return fmt.Errorf("product ID is required")
	}
	if p.Entitlement != "premium" {
		return fmt.Errorf("unsupported product entitlement %q", p.Entitlement)
	}
	if !p.BillingPeriod.Valid() {
		return fmt.Errorf("unsupported billing period %q", p.BillingPeriod)
	}
	return nil
}

type MobileIAPEntitlement struct {
	Store            EntitlementStore  `json:"store"`
	ProductID        string            `json:"productId"`
	Entitlement      string            `json:"entitlement"`
	Status           EntitlementStatus `json:"status"`
	PeriodEnd        *time.Time        `json:"periodEnd,omitempty"`
	AutoRenewEnabled bool              `json:"autoRenewEnabled"`
	CanceledAt       *time.Time        `json:"canceledAt,omitempty"`
	RevokedAt        *time.Time        `json:"revokedAt,omitempty"`
}

func (e MobileIAPEntitlement) Validate() error {
	if !e.Store.Valid() || strings.TrimSpace(e.ProductID) == "" || e.Entitlement != "premium" {
		return fmt.Errorf("mobile entitlement identity is invalid")
	}
	if !e.Status.Valid() {
		return fmt.Errorf("mobile entitlement status is invalid")
	}
	if e.Status == EntitlementStatusPending {
		return fmt.Errorf("pending purchase must not be encoded as current entitlement")
	}
	if (e.Status == EntitlementStatusActive || e.Status == EntitlementStatusGracePeriod) && e.PeriodEnd == nil {
		return fmt.Errorf("period end is required for entitled status")
	}
	if e.Status == EntitlementStatusRevoked && e.RevokedAt == nil {
		return fmt.Errorf("revoked at is required for revoked status")
	}
	if e.Status != EntitlementStatusRevoked && e.RevokedAt != nil {
		return fmt.Errorf("revoked status is required when revoked at is set")
	}
	return nil
}

// MobileIAPPurchaseContext exposes only the server-issued identifier required
// to launch the selected store's purchase UI.
type MobileIAPPurchaseContext struct {
	Store                     EntitlementStore `json:"store"`
	AppleAppAccountToken      *uuid.UUID       `json:"appleAppAccountToken,omitempty"`
	GoogleObfuscatedAccountID *string          `json:"googleObfuscatedAccountId,omitempty"`
}

func (c MobileIAPPurchaseContext) Validate() error {
	switch c.Store {
	case EntitlementStoreApple:
		if c.AppleAppAccountToken == nil || *c.AppleAppAccountToken == uuid.Nil || c.GoogleObfuscatedAccountID != nil {
			return fmt.Errorf("apple purchase context is invalid")
		}
	case EntitlementStoreGoogle:
		if c.AppleAppAccountToken != nil || c.GoogleObfuscatedAccountID == nil {
			return fmt.Errorf("google purchase context is invalid")
		}
		if err := validateOpaqueStoreIdentifier("google obfuscated account ID", *c.GoogleObfuscatedAccountID, GoogleObfuscatedAccountIDMaxLength); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported purchase context store %q", c.Store)
	}
	return nil
}

type MobileIAPPending struct {
	Store     EntitlementStore `json:"store"`
	ProductID string           `json:"productId"`
	StartedAt time.Time        `json:"startedAt"`
}

func (p MobileIAPPending) Validate() error {
	if !p.Store.Valid() || strings.TrimSpace(p.ProductID) == "" || p.StartedAt.IsZero() {
		return fmt.Errorf("pending purchase is invalid")
	}
	return nil
}

type MobileIAPError struct {
	Code       EntitlementResultCode `json:"code"`
	MessageKey string                `json:"messageKey"`
	Retryable  bool                  `json:"retryable"`
}

func (e MobileIAPError) Validate() error {
	if !strings.HasPrefix(e.MessageKey, "iap.") || strings.ContainsAny(e.MessageKey, " \t\n") {
		return fmt.Errorf("iap error message key is invalid")
	}
	switch e.Code {
	case EntitlementResultRetryableProvider:
		if !e.Retryable {
			return fmt.Errorf("retryable provider error must be retryable")
		}
	case EntitlementResultInvalidPurchase, EntitlementResultAccountMismatch, EntitlementResultUnknownProduct:
		if e.Retryable {
			return fmt.Errorf("permanent IAP error must not be retryable")
		}
	default:
		return fmt.Errorf("unsupported mobile IAP error code %q", e.Code)
	}
	return nil
}

type IAPRestoreStatus string

const (
	IAPRestoreRestored    IAPRestoreStatus = "restored"
	IAPRestoreNoPurchases IAPRestoreStatus = "no_purchases"
	IAPRestorePartial     IAPRestoreStatus = "partial"
	IAPRestorePending     IAPRestoreStatus = "pending"
	IAPRestoreFailed      IAPRestoreStatus = "failed"
)

type IAPRestoreEvidenceStatus string

const (
	IAPRestoreEvidenceRestored  IAPRestoreEvidenceStatus = "restored"
	IAPRestoreEvidenceUnchanged IAPRestoreEvidenceStatus = "unchanged"
	IAPRestoreEvidenceRejected  IAPRestoreEvidenceStatus = "rejected"
	IAPRestoreEvidenceRetryable IAPRestoreEvidenceStatus = "retryable"
)

// MobileIAPRestoreEvidenceResult deliberately exposes only the request index;
// raw transaction IDs and purchase tokens must never be echoed to clients.
type MobileIAPRestoreEvidenceResult struct {
	Index  int                      `json:"index"`
	Status IAPRestoreEvidenceStatus `json:"status"`
	Error  *MobileIAPError          `json:"error,omitempty"`
}

func (r MobileIAPRestoreEvidenceResult) Validate() error {
	if r.Index < 0 {
		return fmt.Errorf("restore evidence index must be non-negative")
	}
	switch r.Status {
	case IAPRestoreEvidenceRestored, IAPRestoreEvidenceUnchanged:
		if r.Error != nil {
			return fmt.Errorf("successful restore evidence cannot include an error")
		}
	case IAPRestoreEvidenceRejected:
		if r.Error == nil || r.Error.Retryable {
			return fmt.Errorf("rejected restore evidence requires a permanent error")
		}
	case IAPRestoreEvidenceRetryable:
		if r.Error == nil || !r.Error.Retryable {
			return fmt.Errorf("retryable restore evidence requires a retryable error")
		}
	default:
		return fmt.Errorf("unsupported restore evidence status %q", r.Status)
	}
	if r.Error != nil {
		return r.Error.Validate()
	}
	return nil
}

type MobileIAPRestoreResult struct {
	Status             IAPRestoreStatus                 `json:"status"`
	RestoredPurchases  int                              `json:"restoredPurchases"`
	UnchangedPurchases int                              `json:"unchangedPurchases"`
	RejectedPurchases  int                              `json:"rejectedPurchases"`
	RetryablePurchases int                              `json:"retryablePurchases"`
	Outcomes           []MobileIAPRestoreEvidenceResult `json:"outcomes"`
	Error              *MobileIAPError                  `json:"error,omitempty"`
}

func (r MobileIAPRestoreResult) Validate() error {
	if r.RestoredPurchases < 0 || r.UnchangedPurchases < 0 || r.RejectedPurchases < 0 || r.RetryablePurchases < 0 {
		return fmt.Errorf("restore counts must be non-negative")
	}
	if err := r.validateOutcomes(); err != nil {
		return err
	}
	return r.validateStatus()
}

func (r MobileIAPRestoreResult) validateOutcomes() error {
	counts := map[IAPRestoreEvidenceStatus]int{}
	seen := make(map[int]struct{}, len(r.Outcomes))
	for _, outcome := range r.Outcomes {
		if err := outcome.Validate(); err != nil {
			return err
		}
		if _, exists := seen[outcome.Index]; exists {
			return fmt.Errorf("duplicate restore evidence index")
		}
		seen[outcome.Index] = struct{}{}
		counts[outcome.Status]++
	}
	if counts[IAPRestoreEvidenceRestored] > r.RestoredPurchases ||
		counts[IAPRestoreEvidenceUnchanged] > r.UnchangedPurchases ||
		counts[IAPRestoreEvidenceRejected] > r.RejectedPurchases ||
		counts[IAPRestoreEvidenceRetryable] > r.RetryablePurchases {
		return fmt.Errorf("restore per-evidence outcomes exceed aggregate counts")
	}
	return nil
}

func (r MobileIAPRestoreResult) validateStatus() error {
	switch r.Status {
	case IAPRestoreNoPurchases:
		return r.validateNoPurchases()
	case IAPRestorePending:
		return r.validatePending()
	case IAPRestoreRestored:
		return r.validateRestored()
	case IAPRestorePartial:
		return r.validatePartial()
	case IAPRestoreFailed:
		return r.validateFailed()
	default:
		return fmt.Errorf("unsupported restore status %q", r.Status)
	}
}

func (r MobileIAPRestoreResult) validateNoPurchases() error {
	if r.RestoredPurchases != 0 || r.UnchangedPurchases != 0 || r.RejectedPurchases != 0 ||
		r.RetryablePurchases != 0 || len(r.Outcomes) != 0 || r.Error != nil {
		return fmt.Errorf("restore result fields conflict with status %q", r.Status)
	}
	return nil
}

func (r MobileIAPRestoreResult) validatePending() error {
	if r.RestoredPurchases != 0 || r.RetryablePurchases == 0 && len(r.Outcomes) > 0 || r.Error != nil {
		return fmt.Errorf("restore result fields conflict with status %q", r.Status)
	}
	return nil
}

func (r MobileIAPRestoreResult) validateRestored() error {
	if r.RestoredPurchases+r.UnchangedPurchases <= 0 || r.RejectedPurchases != 0 || r.RetryablePurchases != 0 || r.Error != nil {
		return fmt.Errorf("restored result requires a positive purchase count")
	}
	return nil
}

func (r MobileIAPRestoreResult) validatePartial() error {
	successes := r.RestoredPurchases + r.UnchangedPurchases
	failures := r.RejectedPurchases + r.RetryablePurchases
	mixedFailures := r.RejectedPurchases > 0 && r.RetryablePurchases > 0
	if r.Error != nil || failures == 0 || (successes == 0 && !mixedFailures) {
		return fmt.Errorf("partial restore requires mixed outcomes")
	}
	return nil
}

func (r MobileIAPRestoreResult) validateFailed() error {
	if r.RestoredPurchases != 0 || r.UnchangedPurchases != 0 || r.Error == nil {
		return fmt.Errorf("failed restore requires an error")
	}
	return r.Error.Validate()
}

type MobileIAPResponse struct {
	Products           []MobileIAPProduct        `json:"products"`
	PurchaseContext    *MobileIAPPurchaseContext `json:"purchaseContext"`
	CurrentEntitlement *MobileIAPEntitlement     `json:"currentEntitlement"`
	Pending            *MobileIAPPending         `json:"pending"`
	Error              *MobileIAPError           `json:"error"`
	Restore            *MobileIAPRestoreResult   `json:"restore"`
	ServerTime         time.Time                 `json:"serverTime"`
}

func (r MobileIAPResponse) Validate() error {
	if err := validateMobileIAPCatalog(r.Products, r.PurchaseContext); err != nil {
		return err
	}
	if err := validateMobileIAPState(r); err != nil {
		return err
	}
	if r.ServerTime.IsZero() {
		return fmt.Errorf("server time is required")
	}
	return nil
}

func validateMobileIAPCatalog(products []MobileIAPProduct, context *MobileIAPPurchaseContext) error {
	if products == nil {
		return fmt.Errorf("products must serialize as an array")
	}
	seen := make(map[string]struct{}, len(products))
	var productStore EntitlementStore
	for _, product := range products {
		if err := product.Validate(); err != nil {
			return err
		}
		key := string(product.Store) + "\x00" + product.ProductID
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate store product %q", product.ProductID)
		}
		seen[key] = struct{}{}
		if productStore == "" {
			productStore = product.Store
		} else if product.Store != productStore {
			return fmt.Errorf("products must belong to one store")
		}
	}
	if len(products) > 0 && context == nil {
		return fmt.Errorf("purchase context is required when products are available")
	}
	if context != nil {
		if err := context.Validate(); err != nil {
			return err
		}
		for _, product := range products {
			if product.Store != context.Store {
				return fmt.Errorf("product store does not match purchase context")
			}
		}
	}
	return nil
}

func validateMobileIAPState(r MobileIAPResponse) error {
	if r.CurrentEntitlement != nil {
		if err := r.CurrentEntitlement.Validate(); err != nil {
			return err
		}
	}
	if r.Pending != nil {
		if err := r.Pending.Validate(); err != nil {
			return err
		}
	}
	if r.Error != nil {
		if err := r.Error.Validate(); err != nil {
			return err
		}
		if !r.Error.Retryable && r.Pending != nil {
			return fmt.Errorf("permanent error cannot coexist with pending purchase")
		}
	}
	if r.Restore != nil {
		if err := r.Restore.Validate(); err != nil {
			return err
		}
	}
	return nil
}
