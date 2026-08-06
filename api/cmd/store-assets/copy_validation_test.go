package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateStoreCopyRejectsMissingEmptyAndUnsafeClaims(t *testing.T) {
	slots := []SlotConfig{{Key: "dashboard"}, {Key: "premium"}}
	valid := CopyFile{
		"dashboard": {Headline: "Your study shelf", Subhead: "See every wordlist"},
		"premium":   {Headline: "Unlimited learning", Subhead: "Plans shown in the app"},
	}
	require.NoError(t, validateStoreCopy(valid, slots))

	tests := []struct {
		name string
		copy CopyFile
		want string
	}{
		{name: "missing slot", copy: CopyFile{"dashboard": valid["dashboard"]}, want: "missing copy for slot premium"},
		{name: "empty headline", copy: CopyFile{"dashboard": {Subhead: "Body"}, "premium": valid["premium"]}, want: "dashboard headline is empty"},
		{name: "legacy provider", copy: CopyFile{"dashboard": valid["dashboard"], "premium": {Headline: "Stripe checkout", Subhead: "Safe"}}, want: "legacy payment provider"},
		{name: "hardcoded price", copy: CopyFile{"dashboard": valid["dashboard"], "premium": {Headline: "Premium", Subhead: "$6.99 monthly"}}, want: "hard-coded price"},
		{name: "hardcoded price with suffix", copy: CopyFile{"dashboard": valid["dashboard"], "premium": {Headline: "Premium", Subhead: "Apenas 6,99 € por mês"}}, want: "hard-coded price"},
		{name: "hardcoded ISO price", copy: CopyFile{"dashboard": valid["dashboard"], "premium": {Headline: "Premium", Subhead: "Only 6.99 USD monthly"}}, want: "hard-coded price"},
		{name: "unsupported offer", copy: CopyFile{"dashboard": valid["dashboard"], "premium": {Headline: "Premium", Subhead: "7-day free trial"}}, want: "unsupported promotional claim"},
		{name: "unsupported localized offer", copy: CopyFile{"dashboard": valid["dashboard"], "premium": {Headline: "Premium", Subhead: "Teste grátis de 7 dias"}}, want: "unsupported promotional claim"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.ErrorContains(t, validateStoreCopy(test.copy, slots), test.want)
		})
	}
}

func TestCheckedInStoreCopyPassesComplianceValidation(t *testing.T) {
	configBytes, err := os.ReadFile("config.json")
	require.NoError(t, err)

	cfg, err := parseConfig(configBytes)
	require.NoError(t, err)

	foundFlashcardsSlot := false
	for _, slot := range cfg.Slots {
		if slot.Key == "flashcards" {
			foundFlashcardsSlot = true
			assert.NotContains(t, slot.Keywords, "swipe")
			assert.Contains(t, slot.Keywords, "tap")
		}
	}
	require.True(t, foundFlashcardsSlot)

	for _, locale := range cfg.Locales {
		t.Run(locale, func(t *testing.T) {
			contents, readErr := os.ReadFile(filepath.Join("store-copy", locale+".json"))
			require.NoError(t, readErr)

			copyFile, parseErr := parseCopyFile(contents)
			require.NoError(t, parseErr)
			require.NoError(t, validateStoreCopy(copyFile, cfg.Slots))
		})
	}
}
