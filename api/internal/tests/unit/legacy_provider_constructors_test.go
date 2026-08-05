package unit

import (
	"testing"

	"decorebator.com/internal/config"
	"decorebator.com/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLegacyProviderConstructorsNeverPanicOnMissingConfiguration(t *testing.T) {
	assert.NotPanics(t, func() {
		subscriptions := service.NewSubscriptionService(nil, nil, nil, config.LegacyStripeConfig{})
		require.NotNil(t, subscriptions)
	})

	assert.NotPanics(t, func() {
		client, err := service.NewRevenueCatAPIClient("  ")
		require.Nil(t, client)
		require.ErrorContains(t, err, "REVENUECAT_API_KEY")
	})
}
