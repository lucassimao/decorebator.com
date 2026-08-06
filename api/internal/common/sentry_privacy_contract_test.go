package common

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSentrySourcesDoNotAttachDirectPII(t *testing.T) {
	files := []string{
		"sentry_init.go",
		"sentry_context.go",
		"sentry_logger.go",
		filepath.Join("..", "http", "sentry_context.go"),
		filepath.Join("..", "http", "midlewares.go"),
	}
	banned := []string{
		"SendDefaultPII: true",
		"Email: userCtx.Email",
		"QueryString: reqCtx.QueryParams",
		"reqCtx.QueryParams",
		"reqCtx.RemoteIP",
		"c.Request.URL.String()",
	}
	for _, name := range files {
		source, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, value := range banned {
			if strings.Contains(string(source), value) {
				t.Errorf("%s contains banned Sentry PII source %q", name, value)
			}
		}
	}
}
