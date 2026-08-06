package common

import (
	"testing"

	"github.com/getsentry/sentry-go"
)

func TestScrubSentryEventRemovesDirectAndNestedPII(t *testing.T) {
	event := &sentry.Event{
		User: sentry.User{
			ID:        "42",
			Email:     "person@example.com",
			IPAddress: "203.0.113.7",
			Username:  "person",
		},
		Request: &sentry.Request{
			URL:         "https://api.example.test/users?reset_token=secret",
			QueryString: "reset_token=secret",
			Cookies:     "session=secret",
			Data:        "email=person@example.com",
			Env:         map[string]string{"REMOTE_ADDR": "203.0.113.7"},
			Headers: map[string]string{
				"Authorization": "Bearer secret",
				"User-Agent":    "test",
			},
		},
		Extra: map[string]any{
			"email": "person@example.com",
			"safe":  "kept",
			"nested": map[string]any{
				"purchaseToken": "secret",
				"attempt":       2,
			},
		},
		Contexts: map[string]map[string]any{
			"request": {"remote_ip": "203.0.113.7", "path": "/users"},
		},
	}

	got := scrubSentryEvent(event, nil)
	if got.User.ID != "42" || got.User.Email != "" || got.User.IPAddress != "" || got.User.Username != "" {
		t.Fatalf("user was not reduced to ID only: %#v", got.User)
	}
	if got.Request.URL != "https://api.example.test/users" || got.Request.QueryString != "" || got.Request.Cookies != "" || got.Request.Data != "" || got.Request.Env != nil {
		t.Fatalf("request PII was not removed: %#v", got.Request)
	}
	if _, ok := got.Request.Headers["Authorization"]; ok || got.Request.Headers["User-Agent"] != "test" {
		t.Fatalf("request headers were not selectively scrubbed: %#v", got.Request.Headers)
	}
	if _, ok := got.Extra["email"]; ok || got.Extra["safe"] != "kept" {
		t.Fatalf("extra fields were not selectively scrubbed: %#v", got.Extra)
	}
	nested := got.Extra["nested"].(map[string]any)
	if _, ok := nested["purchaseToken"]; ok || nested["attempt"] != 2 {
		t.Fatalf("nested fields were not selectively scrubbed: %#v", nested)
	}
	if _, ok := got.Contexts["request"]["remote_ip"]; ok || got.Contexts["request"]["path"] != "/users" {
		t.Fatalf("contexts were not selectively scrubbed: %#v", got.Contexts)
	}
}

func TestScrubSentryLogRemovesSensitiveAttributes(t *testing.T) {
	log := &sentry.Log{Attributes: map[string]sentry.Attribute{
		"email":         {Value: "person@example.com", Type: "string"},
		"purchaseToken": {Value: "secret", Type: "string"},
		"attempt":       {Value: int64(2), Type: "integer"},
	}}
	got := scrubSentryLog(log)
	if _, ok := got.Attributes["email"]; ok {
		t.Fatal("email attribute was not scrubbed")
	}
	if _, ok := got.Attributes["purchaseToken"]; ok {
		t.Fatal("token attribute was not scrubbed")
	}
	if got.Attributes["attempt"].Value != int64(2) {
		t.Fatalf("safe attribute was changed: %#v", got.Attributes)
	}
}
