package i18n

import (
	"embed"
	"encoding/json"
	"strings"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed notifications/*.json
var notificationsFS embed.FS

var (
	bundle         *i18n.Bundle
	localizerCache sync.Map
)

func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	entries, err := notificationsFS.ReadDir("notifications")
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := notificationsFS.ReadFile("notifications/" + entry.Name())
		if err != nil {
			panic(err)
		}
		if _, err := bundle.ParseMessageFileBytes(data, entry.Name()); err != nil {
			panic(err)
		}
	}
}

func LocalizerForLocale(locale string) *i18n.Localizer {
	normalized := normalizeLocale(locale)
	cacheKey := normalized
	if cacheKey == "" {
		cacheKey = "en"
	}
	if value, ok := localizerCache.Load(cacheKey); ok {
		localizer, ok := value.(*i18n.Localizer)
		if ok {
			return localizer
		}
	}

	var tags []string
	if normalized != "" {
		tags = append(tags, normalized)
		if base := baseLocale(normalized); base != "" && base != normalized {
			tags = append(tags, base)
		}
	}
	tags = append(tags, "en")

	localizer := i18n.NewLocalizer(bundle, tags...)
	localizerCache.Store(cacheKey, localizer)
	return localizer
}

func normalizeLocale(locale string) string {
	locale = strings.TrimSpace(strings.ToLower(locale))
	if locale == "" {
		return ""
	}
	return strings.ReplaceAll(locale, "_", "-")
}

func baseLocale(locale string) string {
	if locale == "" {
		return ""
	}
	if idx := strings.Index(locale, "-"); idx > 0 {
		return locale[:idx]
	}
	return locale
}
