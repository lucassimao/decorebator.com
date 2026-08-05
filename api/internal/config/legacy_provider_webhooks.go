package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type LegacyStripeConfig struct {
	APIKey, WebhookSecret, SuccessURL, CancelURL, MonthlyPriceID, AnnualPriceID string
}

type LegacyProviderConfig struct {
	Enabled                        bool
	Stripe                         LegacyStripeConfig
	RevenueCatAPIKey               string
	RevenueCatWebhookAuthorization string
}

// LoadLegacyProviderConfig controls the complete legacy Stripe and RevenueCat
// purchase surface. The switch defaults on outside production; enabled
// environments must provide the complete legacy contract explicitly.
func LoadLegacyProviderConfig() (LegacyProviderConfig, error) {
	return LoadLegacyProviderConfigFrom(os.LookupEnv)
}

func LoadLegacyProviderConfigFrom(lookup EnvironmentLookup) (LegacyProviderConfig, error) {
	if lookup == nil {
		return LegacyProviderConfig{}, fmt.Errorf("legacy provider environment lookup is required")
	}
	raw, exists := lookup("LEGACY_PROVIDER_WEBHOOKS_ENABLED")
	if !exists || strings.TrimSpace(raw) == "" {
		if environment, ok := lookup("ENV"); ok && strings.EqualFold(strings.TrimSpace(environment), "production") {
			return LegacyProviderConfig{}, fmt.Errorf("LEGACY_PROVIDER_WEBHOOKS_ENABLED must be explicitly configured in production")
		}
		raw = "true"
	}
	enabled, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return LegacyProviderConfig{}, fmt.Errorf("LEGACY_PROVIDER_WEBHOOKS_ENABLED must be a boolean")
	}
	result := LegacyProviderConfig{Enabled: enabled}
	if !enabled {
		return result, nil
	}
	values := []struct {
		name        string
		destination *string
	}{
		{"STRIPE_API_KEY", &result.Stripe.APIKey},
		{"STRIPE_WEBHOOK_SECRET", &result.Stripe.WebhookSecret},
		{"STRIPE_SUCCESS_URL", &result.Stripe.SuccessURL},
		{"STRIPE_CANCEL_URL", &result.Stripe.CancelURL},
		{"STRIPE_PRICE_ID_MONTHLY", &result.Stripe.MonthlyPriceID},
		{"STRIPE_PRICE_ID_ANNUAL", &result.Stripe.AnnualPriceID},
		{"REVENUECAT_API_KEY", &result.RevenueCatAPIKey},
		{"REVENUECAT_WEBHOOK_AUTHORIZATION", &result.RevenueCatWebhookAuthorization},
	}
	for _, required := range values {
		value, ok := lookup(required.name)
		if !ok || strings.TrimSpace(value) == "" {
			return LegacyProviderConfig{}, fmt.Errorf("%s is required when legacy provider surface is enabled", required.name)
		}
		*required.destination = strings.TrimSpace(value)
	}
	return result, nil
}
