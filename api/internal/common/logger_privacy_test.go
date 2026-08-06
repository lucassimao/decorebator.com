package common

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestPrivacyHandlerRemovesSensitiveRecordAndContextAttributes(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(NewPrivacyHandler(slog.NewJSONHandler(&output, nil))).With(
		"email", "person@example.com",
		"request_id", "request-42",
	)

	logger.Info(
		"login performance metrics",
		"purchaseToken", "raw-provider-token",
		"attempt", 2,
		"db_query_ms", int64(3),
		slog.Group("identity",
			"password", "not-for-logs",
			"user_id", int64(7),
		),
	)
	logger.WithGroup("authorization").Info("sensitive group", "error", "group-secret")

	logged := output.String()
	for _, forbidden := range []string{
		"email", "person@example.com", "purchaseToken", "raw-provider-token",
		"password", "not-for-logs", "authorization", "group-secret",
	} {
		if strings.Contains(logged, forbidden) {
			t.Fatalf("sensitive log content %q was not removed: %s", forbidden, logged)
		}
	}
	for _, required := range []string{"request_id", "request-42", "attempt", "user_id", "db_query_ms"} {
		if !strings.Contains(logged, required) {
			t.Fatalf("safe log content %q was removed: %s", required, logged)
		}
	}
}
