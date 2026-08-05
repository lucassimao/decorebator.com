package model

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const GoogleObfuscatedAccountIDMaxLength = 64

// AppleAccountIdentity is a server-owned, per-user identifier created before
// purchase. API handlers must use a dedicated response DTO when sending the
// token to an authenticated client; domain identity records are never exposed.
type AppleAccountIdentity struct {
	UserID          int64     `json:"-"`
	AppAccountToken uuid.UUID `json:"-"`
}

func (i AppleAccountIdentity) Validate() error {
	if i.UserID <= 0 {
		return fmt.Errorf("user ID must be positive")
	}
	if i.AppAccountToken == uuid.Nil {
		return fmt.Errorf("Apple app account token is required")
	}

	return nil
}

// ApplePurchaseIdentity is created only from a successfully verified Apple
// signed transaction. UserID must be resolved through AppleAccountIdentity;
// it must never be copied from a client request.
type ApplePurchaseIdentity struct {
	UserID                int64            `json:"-"`
	AppAccountToken       uuid.UUID        `json:"-"`
	OriginalTransactionID string           `json:"-"`
	Environment           StoreEnvironment `json:"-"`
	LastVerifiedAt        time.Time        `json:"-"`
}

func (i ApplePurchaseIdentity) Validate() error {
	if i.UserID <= 0 {
		return fmt.Errorf("user ID must be positive")
	}
	if i.AppAccountToken == uuid.Nil {
		return fmt.Errorf("Apple app account token is required")
	}
	if err := validateOpaqueStoreIdentifier("Apple original transaction ID", i.OriginalTransactionID, 0); err != nil {
		return err
	}
	if !i.Environment.Valid() {
		return fmt.Errorf("unsupported store environment %q", i.Environment)
	}
	if i.LastVerifiedAt.IsZero() {
		return fmt.Errorf("last verified time is required")
	}

	return nil
}

// MatchesAccount proves that the verified transaction carries the exact
// server-issued Apple account token for the resolved user.
func (i ApplePurchaseIdentity) MatchesAccount(account AppleAccountIdentity) bool {
	if i.Validate() != nil || account.Validate() != nil {
		return false
	}

	return i.UserID == account.UserID && i.AppAccountToken == account.AppAccountToken
}

// GoogleAccountIdentity is a server-generated, PII-free account identifier
// passed to BillingFlowParams. It must be produced from server-side entropy or
// a one-way keyed derivation, never from clear-text email or another client ID.
type GoogleAccountIdentity struct {
	UserID                      int64  `json:"-"`
	ObfuscatedExternalAccountID string `json:"-"`
}

func (i GoogleAccountIdentity) Validate() error {
	if i.UserID <= 0 {
		return fmt.Errorf("user ID must be positive")
	}

	return validateOpaqueStoreIdentifier(
		"Google obfuscated external account ID",
		i.ObfuscatedExternalAccountID,
		GoogleObfuscatedAccountIDMaxLength,
	)
}

// GooglePurchaseIdentity is created only after the backend verifies the token
// with Google Play. PurchaseToken is sensitive and must never be logged or
// serialized into an API response.
type GooglePurchaseIdentity struct {
	UserID                      int64            `json:"-"`
	ObfuscatedExternalAccountID string           `json:"-"`
	PurchaseToken               string           `json:"-"`
	Environment                 StoreEnvironment `json:"-"`
	LastVerifiedAt              time.Time        `json:"-"`
}

func (i GooglePurchaseIdentity) Validate() error {
	if i.UserID <= 0 {
		return fmt.Errorf("user ID must be positive")
	}
	if err := validateOpaqueStoreIdentifier(
		"Google obfuscated external account ID",
		i.ObfuscatedExternalAccountID,
		GoogleObfuscatedAccountIDMaxLength,
	); err != nil {
		return err
	}
	if err := validateOpaqueStoreIdentifier("Google purchase token", i.PurchaseToken, 0); err != nil {
		return err
	}
	if !i.Environment.Valid() {
		return fmt.Errorf("unsupported store environment %q", i.Environment)
	}
	if i.LastVerifiedAt.IsZero() {
		return fmt.Errorf("last verified time is required")
	}

	return nil
}

// MatchesAccount proves that Google returned the exact server-issued account
// identifier for the resolved user after purchase-token verification.
func (i GooglePurchaseIdentity) MatchesAccount(account GoogleAccountIdentity) bool {
	if i.Validate() != nil || account.Validate() != nil {
		return false
	}

	return i.UserID == account.UserID &&
		i.ObfuscatedExternalAccountID == account.ObfuscatedExternalAccountID
}

func validateOpaqueStoreIdentifier(label, value string, maxLength int) error {
	if value == "" {
		return fmt.Errorf("%s is required", label)
	}
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("%s must not contain surrounding whitespace", label)
	}
	if maxLength > 0 && utf8.RuneCountInString(value) > maxLength {
		return fmt.Errorf("%s exceeds %d characters", label, maxLength)
	}

	return nil
}
