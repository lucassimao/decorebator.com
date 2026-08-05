package unit

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplePurchaseIdentityRequiresServerAccountBinding(t *testing.T) {
	now := time.Now().UTC()
	account := model.AppleAccountIdentity{
		UserID:          42,
		AppAccountToken: uuid.New(),
	}
	purchase := model.ApplePurchaseIdentity{
		UserID:                account.UserID,
		AppAccountToken:       account.AppAccountToken,
		OriginalTransactionID: "2000000123456789",
		Environment:           model.StoreEnvironmentSandbox,
		LastVerifiedAt:        now,
	}

	require.NoError(t, account.Validate())
	require.NoError(t, purchase.Validate())
	assert.True(t, purchase.MatchesAccount(account))

	wrongToken := account
	wrongToken.AppAccountToken = uuid.New()
	assert.False(t, purchase.MatchesAccount(wrongToken))

	wrongUser := account
	wrongUser.UserID++
	assert.False(t, purchase.MatchesAccount(wrongUser))

	malformedPurchase := purchase
	malformedPurchase.OriginalTransactionID = ""
	assert.False(t, malformedPurchase.MatchesAccount(account))
}

func TestAppleIdentityRejectsIncompleteOrMalformedBindings(t *testing.T) {
	now := time.Now().UTC()
	validToken := uuid.New()

	accountTests := []struct {
		name    string
		account model.AppleAccountIdentity
	}{
		{"missing user", model.AppleAccountIdentity{AppAccountToken: validToken}},
		{"missing app account token", model.AppleAccountIdentity{UserID: 42}},
	}

	for _, tt := range accountTests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, tt.account.Validate())
		})
	}

	purchaseTests := []struct {
		name   string
		mutate func(*model.ApplePurchaseIdentity)
	}{
		{"missing user", func(p *model.ApplePurchaseIdentity) { p.UserID = 0 }},
		{"missing app account token", func(p *model.ApplePurchaseIdentity) { p.AppAccountToken = uuid.Nil }},
		{"missing original transaction", func(p *model.ApplePurchaseIdentity) { p.OriginalTransactionID = "" }},
		{"padded original transaction", func(p *model.ApplePurchaseIdentity) { p.OriginalTransactionID = " 2000000123456789 " }},
		{"unknown environment", func(p *model.ApplePurchaseIdentity) { p.Environment = "staging" }},
		{"missing verification time", func(p *model.ApplePurchaseIdentity) { p.LastVerifiedAt = time.Time{} }},
	}

	for _, tt := range purchaseTests {
		t.Run(tt.name, func(t *testing.T) {
			purchase := model.ApplePurchaseIdentity{
				UserID:                42,
				AppAccountToken:       validToken,
				OriginalTransactionID: "2000000123456789",
				Environment:           model.StoreEnvironmentProduction,
				LastVerifiedAt:        now,
			}
			tt.mutate(&purchase)
			require.Error(t, purchase.Validate())
		})
	}
}

func TestGooglePurchaseIdentityRequiresServerAccountBinding(t *testing.T) {
	now := time.Now().UTC()
	account := model.GoogleAccountIdentity{
		UserID:                      42,
		ObfuscatedExternalAccountID: "acct_KMt6uZ7EXW1V8x5Q3jNRzg",
	}
	purchase := model.GooglePurchaseIdentity{
		UserID:                      account.UserID,
		ObfuscatedExternalAccountID: account.ObfuscatedExternalAccountID,
		PurchaseToken:               "verified-google-purchase-token",
		Environment:                 model.StoreEnvironmentSandbox,
		LastVerifiedAt:              now,
	}

	require.NoError(t, account.Validate())
	require.NoError(t, purchase.Validate())
	assert.True(t, purchase.MatchesAccount(account))

	wrongAccountID := account
	wrongAccountID.ObfuscatedExternalAccountID = "acct_other"
	assert.False(t, purchase.MatchesAccount(wrongAccountID))

	wrongUser := account
	wrongUser.UserID++
	assert.False(t, purchase.MatchesAccount(wrongUser))

	malformedAccount := account
	malformedAccount.ObfuscatedExternalAccountID = ""
	assert.False(t, purchase.MatchesAccount(malformedAccount))
}

func TestGoogleIdentityRejectsIncompleteOrMalformedBindings(t *testing.T) {
	now := time.Now().UTC()
	validAccountID := "acct_KMt6uZ7EXW1V8x5Q3jNRzg"

	accountTests := []struct {
		name    string
		account model.GoogleAccountIdentity
	}{
		{"missing user", model.GoogleAccountIdentity{ObfuscatedExternalAccountID: validAccountID}},
		{"missing obfuscated account", model.GoogleAccountIdentity{UserID: 42}},
		{"padded obfuscated account", model.GoogleAccountIdentity{UserID: 42, ObfuscatedExternalAccountID: " " + validAccountID}},
		{"oversized obfuscated account", model.GoogleAccountIdentity{UserID: 42, ObfuscatedExternalAccountID: strings.Repeat("a", 65)}},
	}

	for _, tt := range accountTests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, tt.account.Validate())
		})
	}

	purchaseTests := []struct {
		name   string
		mutate func(*model.GooglePurchaseIdentity)
	}{
		{"missing user", func(p *model.GooglePurchaseIdentity) { p.UserID = 0 }},
		{"missing obfuscated account", func(p *model.GooglePurchaseIdentity) { p.ObfuscatedExternalAccountID = "" }},
		{"missing purchase token", func(p *model.GooglePurchaseIdentity) { p.PurchaseToken = "" }},
		{"padded purchase token", func(p *model.GooglePurchaseIdentity) { p.PurchaseToken = " token " }},
		{"unknown environment", func(p *model.GooglePurchaseIdentity) { p.Environment = "staging" }},
		{"missing verification time", func(p *model.GooglePurchaseIdentity) { p.LastVerifiedAt = time.Time{} }},
	}

	for _, tt := range purchaseTests {
		t.Run(tt.name, func(t *testing.T) {
			purchase := model.GooglePurchaseIdentity{
				UserID:                      42,
				ObfuscatedExternalAccountID: validAccountID,
				PurchaseToken:               "verified-google-purchase-token",
				Environment:                 model.StoreEnvironmentProduction,
				LastVerifiedAt:              now,
			}
			tt.mutate(&purchase)
			require.Error(t, purchase.Validate())
		})
	}
}

func TestPurchaseIdentityJSONDoesNotExposeStoreIdentifiers(t *testing.T) {
	now := time.Now().UTC()
	apple := model.ApplePurchaseIdentity{
		UserID:                42,
		AppAccountToken:       uuid.New(),
		OriginalTransactionID: "apple-secret-id",
		Environment:           model.StoreEnvironmentProduction,
		LastVerifiedAt:        now,
	}
	google := model.GooglePurchaseIdentity{
		UserID:                      42,
		ObfuscatedExternalAccountID: "google-account-id",
		PurchaseToken:               "google-secret-token",
		Environment:                 model.StoreEnvironmentProduction,
		LastVerifiedAt:              now,
	}

	for _, identity := range []any{
		model.AppleAccountIdentity{UserID: apple.UserID, AppAccountToken: apple.AppAccountToken},
		apple,
		model.GoogleAccountIdentity{
			UserID:                      google.UserID,
			ObfuscatedExternalAccountID: google.ObfuscatedExternalAccountID,
		},
		google,
	} {
		encoded, err := json.Marshal(identity)
		require.NoError(t, err)
		assert.JSONEq(t, `{}`, string(encoded))
	}
}
