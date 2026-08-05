package unit

import (
	"testing"
	"time"

	"decorebator.com/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validEntitlement(now time.Time) model.StoreEntitlement {
	start := now.Add(-24 * time.Hour)
	end := now.Add(29 * 24 * time.Hour)

	return model.StoreEntitlement{
		UserID:           42,
		Store:            model.EntitlementStoreApple,
		ProductID:        "com.decorebator.premium.monthly",
		Entitlement:      "premium",
		Status:           model.EntitlementStatusActive,
		PeriodStart:      &start,
		PeriodEnd:        &end,
		AutoRenewEnabled: true,
		Environment:      model.StoreEnvironmentSandbox,
		LastVerifiedAt:   now,
	}
}

func TestStoreEntitlementValidateAcceptsCanonicalStores(t *testing.T) {
	now := time.Now().UTC()

	for _, store := range []model.EntitlementStore{
		model.EntitlementStoreApple,
		model.EntitlementStoreGoogle,
	} {
		entitlement := validEntitlement(now)
		entitlement.Store = store
		require.NoError(t, entitlement.Validate())
	}
}

func TestStoreEntitlementValidateRejectsInvalidCanonicalFields(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name   string
		mutate func(*model.StoreEntitlement)
		want   string
	}{
		{"missing user", func(e *model.StoreEntitlement) { e.UserID = 0 }, "user ID"},
		{"unknown store", func(e *model.StoreEntitlement) { e.Store = "stripe" }, "store"},
		{"missing product", func(e *model.StoreEntitlement) { e.ProductID = "" }, "product ID"},
		{"missing entitlement", func(e *model.StoreEntitlement) { e.Entitlement = "" }, "entitlement"},
		{"unknown status", func(e *model.StoreEntitlement) { e.Status = "trialing" }, "status"},
		{"unknown environment", func(e *model.StoreEntitlement) { e.Environment = "staging" }, "environment"},
		{"missing verification time", func(e *model.StoreEntitlement) { e.LastVerifiedAt = time.Time{} }, "last verified"},
		{"missing active end", func(e *model.StoreEntitlement) { e.PeriodEnd = nil }, "period end"},
		{"missing grace-period end", func(e *model.StoreEntitlement) {
			e.Status = model.EntitlementStatusGracePeriod
			e.PeriodEnd = nil
		}, "period end"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entitlement := validEntitlement(now)
			tt.mutate(&entitlement)

			err := entitlement.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestStoreEntitlementValidateRejectsImpossiblePeriodsAndRevocation(t *testing.T) {
	now := time.Now().UTC()

	t.Run("period ends before it starts", func(t *testing.T) {
		entitlement := validEntitlement(now)
		end := entitlement.PeriodStart.Add(-time.Hour)
		entitlement.PeriodEnd = &end

		require.ErrorContains(t, entitlement.Validate(), "period end")
	})

	t.Run("revoked status requires timestamp", func(t *testing.T) {
		entitlement := validEntitlement(now)
		entitlement.Status = model.EntitlementStatusRevoked
		entitlement.AutoRenewEnabled = false
		entitlement.RevokedAt = nil

		require.ErrorContains(t, entitlement.Validate(), "revoked at")
	})

	t.Run("revocation timestamp requires revoked status", func(t *testing.T) {
		entitlement := validEntitlement(now)
		entitlement.RevokedAt = &now

		require.ErrorContains(t, entitlement.Validate(), "revoked status")
	})
}

func TestStoreEntitlementGrantsAccessUsesStatusAndPeriod(t *testing.T) {
	now := time.Now().UTC()

	active := validEntitlement(now)
	active.CanceledAt = &now
	assert.True(t, active.GrantsAccess(now), "cancellation does not end an active paid period")

	grace := validEntitlement(now)
	grace.Status = model.EntitlementStatusGracePeriod
	assert.True(t, grace.GrantsAccess(now))

	for _, status := range []model.EntitlementStatus{
		model.EntitlementStatusPending,
		model.EntitlementStatusOnHold,
		model.EntitlementStatusPaused,
		model.EntitlementStatusExpired,
		model.EntitlementStatusRevoked,
	} {
		entitlement := validEntitlement(now)
		entitlement.Status = status
		assert.False(t, entitlement.GrantsAccess(now), "status %s must not grant access", status)
	}

	expiredPeriod := validEntitlement(now)
	end := now.Add(-time.Second)
	expiredPeriod.PeriodEnd = &end
	assert.False(t, expiredPeriod.GrantsAccess(now))

	missingPeriodEnd := validEntitlement(now)
	missingPeriodEnd.PeriodEnd = nil
	assert.False(t, missingPeriodEnd.GrantsAccess(now), "missing period end fails closed")

	boundary := validEntitlement(now)
	boundary.PeriodEnd = &now
	assert.False(t, boundary.GrantsAccess(now), "period end is exclusive")
}
