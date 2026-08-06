package common

import (
	"net/url"
	"strings"

	"github.com/getsentry/sentry-go"
)

func scrubSentryEvent(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
	if event == nil {
		return nil
	}
	event.User = sentry.User{ID: event.User.ID}
	if event.Request != nil {
		if parsed, err := url.Parse(event.Request.URL); err == nil {
			parsed.RawQuery = ""
			parsed.ForceQuery = false
			event.Request.URL = parsed.String()
		}
		event.Request.QueryString = ""
		event.Request.Cookies = ""
		event.Request.Data = ""
		event.Request.Env = nil
		for key := range event.Request.Headers {
			if sensitiveSentryKey(key) {
				delete(event.Request.Headers, key)
			}
		}
	}
	scrubSentryMap(event.Extra)
	for _, values := range event.Contexts {
		scrubSentryMap(values)
	}
	return event
}

func scrubSentryLog(log *sentry.Log) *sentry.Log {
	if log == nil {
		return nil
	}
	for key := range log.Attributes {
		if sensitiveSentryKey(key) {
			delete(log.Attributes, key)
		}
	}
	return log
}

func scrubSentryMap(values map[string]any) {
	for key, value := range values {
		if sensitiveSentryKey(key) {
			delete(values, key)
			continue
		}
		switch nested := value.(type) {
		case map[string]any:
			scrubSentryMap(nested)
		case map[string]string:
			for nestedKey := range nested {
				if sensitiveSentryKey(nestedKey) {
					delete(nested, nestedKey)
				}
			}
		}
	}
}

func sensitiveSentryKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "-", "_"), " ", "_"))
	for _, fragment := range []string{
		"email", "authorization", "cookie", "password", "token", "secret",
		"query", "remote_ip", "ip_address", "x_forwarded_for", "x_real_ip",
	} {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}
	return false
}
