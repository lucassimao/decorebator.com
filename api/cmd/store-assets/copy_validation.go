package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	currencyMarker = `(?:US\$|R\$|CA\$|AU\$|[$€£¥₹]|\b(?:USD|EUR|GBP|JPY|BRL|CAD|AUD|INR)\b)`

	hardcodedPricePrefixPattern = regexp.MustCompile(`(?i)` + currencyMarker + `\s*\d`)
	hardcodedPriceSuffixPattern = regexp.MustCompile(`(?i)\d(?:[\d.,\s]*\d)?\s*` + currencyMarker)
	unsafeOfferPattern          = regexp.MustCompile(`(?i)(?:\b(?:free[ -]?trial|money[ -]?back|save\s+\d+\s*%|teste\s+(?:grátis|gratis|gratuit[oa])|economize\s+\d+\s*%|poupe\s+\d+\s*%|prueba\s+gratuit[oa]|ahorra\s+\d+\s*%|essai\s+gratuit|satisfait\s+ou\s+remboursé|economisez\s+\d+\s*%|kostenlose\s+testphase|geld-zurück|spare\s+\d+\s*%|prova\s+gratuita|rimborso|risparmia\s+\d+\s*%)\b|無料トライアル|返金保証|\d+\s*%オフ)`)
)

func parseConfig(content []byte) (*Config, error) {
	var cfg Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func parseCopyFile(content []byte) (CopyFile, error) {
	var copyFile CopyFile
	if err := json.Unmarshal(content, &copyFile); err != nil {
		return nil, err
	}
	return copyFile, nil
}

func validateStoreCopy(copyFile CopyFile, slots []SlotConfig) error {
	for _, slot := range slots {
		entry, exists := copyFile[slot.Key]
		if !exists {
			return fmt.Errorf("missing copy for slot %s", slot.Key)
		}
		if strings.TrimSpace(entry.Headline) == "" {
			return fmt.Errorf("%s headline is empty", slot.Key)
		}
		if strings.TrimSpace(entry.Subhead) == "" {
			return fmt.Errorf("%s subhead is empty", slot.Key)
		}
		combined := entry.Headline + "\n" + entry.Subhead
		lower := strings.ToLower(combined)
		if strings.Contains(lower, "stripe") || strings.Contains(lower, "revenuecat") {
			return fmt.Errorf("%s contains a legacy payment provider", slot.Key)
		}
		if hardcodedPricePrefixPattern.MatchString(combined) || hardcodedPriceSuffixPattern.MatchString(combined) {
			return fmt.Errorf("%s contains a hard-coded price", slot.Key)
		}
		if unsafeOfferPattern.MatchString(combined) {
			return fmt.Errorf("%s contains an unsupported promotional claim", slot.Key)
		}
	}
	return nil
}
