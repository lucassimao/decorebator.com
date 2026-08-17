package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserJSONNeverExposesCredentialOrProviderIdentifiers(t *testing.T) {
	t.Parallel()
	stripeCustomerID := "cus_sensitive"
	encoded, err := json.Marshal(User{
		ID:               1,
		Email:            "user@example.test",
		PasswordHash:     "bcrypt-sensitive",
		StripeCustomerID: &stripeCustomerID,
	})
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "bcrypt-sensitive")
	require.NotContains(t, string(encoded), "cus_sensitive")
	require.NotContains(t, string(encoded), "passwordHash")
	require.NotContains(t, string(encoded), "stripeCustomerId")
}

func TestSubscriptionJSONNeverExposesProviderIdentifiers(t *testing.T) {
	t.Parallel()
	stripeSubscriptionID := "sub_sensitive"
	stripeCustomerID := "cus_sensitive"
	revenueCatSubscriptionID := "rc_sensitive"
	encoded, err := json.Marshal(Subscription{
		StripeSubscriptionID:     &stripeSubscriptionID,
		StripeCustomerID:         &stripeCustomerID,
		RevenueCatSubscriptionID: &revenueCatSubscriptionID,
	})
	require.NoError(t, err)
	for _, sensitiveValue := range []string{"sub_sensitive", "cus_sensitive", "rc_sensitive", "stripeSubscriptionId", "stripeCustomerId", "revenuecatSubscriptionId"} {
		require.NotContains(t, string(encoded), sensitiveValue)
	}
}
