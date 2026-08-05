package unit

import (
	"encoding/json"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validMobileIAPResponse(now time.Time) model.MobileIAPResponse {
	end := now.Add(30 * 24 * time.Hour)
	accountToken := uuid.MustParse("db22924d-8f4b-4a7e-845e-3304f319210e")
	return model.MobileIAPResponse{
		Products: []model.MobileIAPProduct{{
			Store: model.EntitlementStoreApple, ProductID: "com.decorebator.premium.monthly",
			Entitlement: "premium", BillingPeriod: model.IAPBillingPeriodMonthly,
		}},
		PurchaseContext: &model.MobileIAPPurchaseContext{
			Store: model.EntitlementStoreApple, AppleAppAccountToken: &accountToken,
		},
		CurrentEntitlement: &model.MobileIAPEntitlement{
			Store: model.EntitlementStoreApple, ProductID: "com.decorebator.premium.monthly",
			Entitlement: "premium", Status: model.EntitlementStatusActive, PeriodEnd: &end,
			AutoRenewEnabled: true,
		},
		ServerTime: now,
	}
}

func TestMobileIAPPurchaseContextRequiresExactlyOneStoreIdentifier(t *testing.T) {
	appleToken := uuid.New()
	googleID := "acct_KMt6uZ7EXW1V8x5Q3jNRzg"
	require.NoError(t, (model.MobileIAPPurchaseContext{Store: model.EntitlementStoreApple, AppleAppAccountToken: &appleToken}).Validate())
	require.NoError(t, (model.MobileIAPPurchaseContext{Store: model.EntitlementStoreGoogle, GoogleObfuscatedAccountID: &googleID}).Validate())
	require.Error(t, (model.MobileIAPPurchaseContext{Store: model.EntitlementStoreApple, AppleAppAccountToken: &appleToken, GoogleObfuscatedAccountID: &googleID}).Validate())
	require.Error(t, (model.MobileIAPPurchaseContext{Store: model.EntitlementStoreGoogle}).Validate())
	nilToken := uuid.Nil
	require.Error(t, (model.MobileIAPPurchaseContext{Store: model.EntitlementStoreApple, AppleAppAccountToken: &nilToken}).Validate())
	require.Error(t, (model.MobileIAPPurchaseContext{Store: "other"}).Validate())
}

func TestMobileIAPEntitlementRejectsInvalidLifecycleViews(t *testing.T) {
	now := time.Now().UTC()
	end := now.Add(time.Hour)
	valid := model.MobileIAPEntitlement{
		Store: model.EntitlementStoreApple, ProductID: "premium_monthly", Entitlement: "premium",
		Status: model.EntitlementStatusActive, PeriodEnd: &end,
	}
	mutations := []func(*model.MobileIAPEntitlement){
		func(value *model.MobileIAPEntitlement) { value.Store = "other" },
		func(value *model.MobileIAPEntitlement) { value.Status = "other" },
		func(value *model.MobileIAPEntitlement) { value.PeriodEnd = nil },
		func(value *model.MobileIAPEntitlement) {
			value.Status = model.EntitlementStatusRevoked
			value.PeriodEnd = nil
		},
		func(value *model.MobileIAPEntitlement) { value.RevokedAt = &now },
	}
	for _, mutate := range mutations {
		value := valid
		mutate(&value)
		require.Error(t, value.Validate())
	}

	revoked := valid
	revoked.Status, revoked.PeriodEnd, revoked.RevokedAt = model.EntitlementStatusRevoked, nil, &now
	require.NoError(t, revoked.Validate())
}

func TestMobileIAPResponseValidatesCatalogAndEntitlement(t *testing.T) {
	response := validMobileIAPResponse(time.Now().UTC())
	require.NoError(t, response.Validate())

	encoded, err := json.Marshal(response)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "userId")
	assert.NotContains(t, string(encoded), "purchaseToken")
	assert.NotContains(t, string(encoded), "originalTransaction")
	assert.Contains(t, string(encoded), `"pending":null`)
	assert.Contains(t, string(encoded), `"restore":null`)
}

func TestMobileIAPResponseRejectsProductContextStoreMismatch(t *testing.T) {
	response := validMobileIAPResponse(time.Now().UTC())
	response.Products[0].Store = model.EntitlementStoreGoogle
	require.ErrorContains(t, response.Validate(), "does not match")
}

func TestMobileIAPResponseRequiresOneStoreAndContextForAvailableProducts(t *testing.T) {
	response := validMobileIAPResponse(time.Now().UTC())
	response.PurchaseContext = nil
	require.ErrorContains(t, response.Validate(), "purchase context")

	response = validMobileIAPResponse(time.Now().UTC())
	response.Products = append(response.Products, model.MobileIAPProduct{
		Store: model.EntitlementStoreGoogle, ProductID: "premium_annual",
		Entitlement: "premium", BillingPeriod: model.IAPBillingPeriodAnnual,
	})
	require.ErrorContains(t, response.Validate(), "one store")
}

func TestMobileIAPResponseRequiresNonNilProductsAndUniqueStoreProducts(t *testing.T) {
	now := time.Now().UTC()
	response := validMobileIAPResponse(now)
	response.Products = nil
	require.ErrorContains(t, response.Validate(), "products")

	response = validMobileIAPResponse(now)
	response.Products = append(response.Products, response.Products[0])
	require.ErrorContains(t, response.Validate(), "duplicate")
}

func TestMobileIAPProductRejectsInvalidCatalogMetadata(t *testing.T) {
	valid := model.MobileIAPProduct{
		Store: model.EntitlementStoreGoogle, ProductID: "premium_monthly",
		Entitlement: "premium", BillingPeriod: model.IAPBillingPeriodMonthly,
	}
	mutations := []func(*model.MobileIAPProduct){
		func(product *model.MobileIAPProduct) { product.Store = "stripe" },
		func(product *model.MobileIAPProduct) { product.ProductID = "" },
		func(product *model.MobileIAPProduct) { product.Entitlement = "" },
		func(product *model.MobileIAPProduct) { product.BillingPeriod = "weekly" },
	}
	for _, mutate := range mutations {
		product := valid
		mutate(&product)
		require.Error(t, product.Validate())
	}
}

func TestMobileIAPResponseSupportsPendingAlongsideCurrentAccess(t *testing.T) {
	now := time.Now().UTC()
	response := validMobileIAPResponse(now)
	response.Pending = &model.MobileIAPPending{
		Store: model.EntitlementStoreApple, ProductID: "com.decorebator.premium.annual",
		StartedAt: now,
	}
	require.NoError(t, response.Validate())
	assert.Equal(t, model.EntitlementStatusActive, response.CurrentEntitlement.Status)
}

func TestMobileIAPPendingRejectsMissingFields(t *testing.T) {
	now := time.Now().UTC()
	for _, pending := range []model.MobileIAPPending{
		{Store: "other", ProductID: "product", StartedAt: now},
		{Store: model.EntitlementStoreApple, StartedAt: now},
		{Store: model.EntitlementStoreApple, ProductID: "product"},
	} {
		require.Error(t, pending.Validate())
	}
}

func TestMobileIAPErrorEnforcesSafeRetrySemantics(t *testing.T) {
	retryable := model.MobileIAPError{
		Code: model.EntitlementResultRetryableProvider, MessageKey: "iap.error.tryAgain", Retryable: true,
	}
	require.NoError(t, retryable.Validate())

	permanent := model.MobileIAPError{
		Code: model.EntitlementResultUnknownProduct, MessageKey: "iap.error.unavailable", Retryable: false,
	}
	require.NoError(t, permanent.Validate())

	retryable.Retryable = false
	require.ErrorContains(t, retryable.Validate(), "retryable")
	permanent.Retryable = true
	require.ErrorContains(t, permanent.Validate(), "retryable")
	require.Error(t, (model.MobileIAPError{Code: model.EntitlementResultUnknownProduct, MessageKey: "raw message"}).Validate())
	require.Error(t, (model.MobileIAPError{Code: model.EntitlementResultApplied, MessageKey: "iap.error.invalid"}).Validate())
}

func TestMobileIAPResponseRejectsPermanentErrorAlongsidePending(t *testing.T) {
	response := validMobileIAPResponse(time.Now().UTC())
	response.Pending = &model.MobileIAPPending{
		Store: model.EntitlementStoreApple, ProductID: "com.decorebator.premium.annual", StartedAt: response.ServerTime,
	}
	response.Error = &model.MobileIAPError{
		Code: model.EntitlementResultUnknownProduct, MessageKey: "iap.error.unavailable",
	}
	require.ErrorContains(t, response.Validate(), "permanent error")

	response.Error = &model.MobileIAPError{
		Code: model.EntitlementResultRetryableProvider, MessageKey: "iap.error.tryAgain", Retryable: true,
	}
	require.NoError(t, response.Validate())
}

func TestMobileIAPRestoreResultValidatesEveryTerminalState(t *testing.T) {
	for _, result := range []model.MobileIAPRestoreResult{
		{Status: model.IAPRestoreRestored, RestoredPurchases: 1},
		{Status: model.IAPRestoreNoPurchases},
		{Status: model.IAPRestorePending},
		{Status: model.IAPRestoreFailed, Error: &model.MobileIAPError{
			Code: model.EntitlementResultRetryableProvider, MessageKey: "iap.error.tryAgain", Retryable: true,
		}},
	} {
		require.NoError(t, result.Validate())
	}

	require.Error(t, (model.MobileIAPRestoreResult{Status: model.IAPRestoreRestored}).Validate())
	require.Error(t, (model.MobileIAPRestoreResult{Status: model.IAPRestoreFailed}).Validate())
	require.Error(t, (model.MobileIAPRestoreResult{Status: "unknown"}).Validate())
	require.Error(t, (model.MobileIAPRestoreResult{Status: model.IAPRestorePending, RestoredPurchases: 1}).Validate())
	require.Error(t, (model.MobileIAPRestoreResult{Status: model.IAPRestoreFailed, Error: &model.MobileIAPError{Code: model.EntitlementResultApplied, MessageKey: "iap.error.invalid"}}).Validate())
}

func TestMobileIAPResponseValidatesNestedStateAndServerTime(t *testing.T) {
	response := validMobileIAPResponse(time.Now().UTC())
	response.ServerTime = time.Time{}
	require.ErrorContains(t, response.Validate(), "server time")

	response = validMobileIAPResponse(time.Now().UTC())
	response.Pending = &model.MobileIAPPending{}
	require.Error(t, response.Validate())

	response = validMobileIAPResponse(time.Now().UTC())
	response.Restore = &model.MobileIAPRestoreResult{Status: "other"}
	require.Error(t, response.Validate())
}
