package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// LoadLegacyProviderWebhooksEnabled controls only the legacy Stripe and
// RevenueCat webhook ingress. It defaults on outside production so existing
// development and test environments keep their current behavior until the
// IAP-only cutover explicitly disables it.
func LoadLegacyProviderWebhooksEnabled() (bool, error) {
	return LoadLegacyProviderWebhooksEnabledFrom(os.LookupEnv)
}

func LoadLegacyProviderWebhooksEnabledFrom(lookup EnvironmentLookup) (bool, error) {
	if lookup == nil {
		return false, fmt.Errorf("legacy provider webhook environment lookup is required")
	}
	raw, exists := lookup("LEGACY_PROVIDER_WEBHOOKS_ENABLED")
	if !exists || strings.TrimSpace(raw) == "" {
		if environment, ok := lookup("ENV"); ok && strings.EqualFold(strings.TrimSpace(environment), "production") {
			return false, fmt.Errorf("LEGACY_PROVIDER_WEBHOOKS_ENABLED must be explicitly configured in production")
		}
		return true, nil
	}
	enabled, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return false, fmt.Errorf("LEGACY_PROVIDER_WEBHOOKS_ENABLED must be a boolean")
	}
	return enabled, nil
}
