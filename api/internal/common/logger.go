package common

import (
	"log/slog"
	"os"
)

const productionEnv = "production"

var options = &slog.HandlerOptions{Level: slog.LevelDebug}
var handler slog.Handler
var Logger *slog.Logger

func init() {
	// Create base handler based on environment
	if os.Getenv("ENV") == productionEnv {
		handler = slog.NewJSONHandler(os.Stdout, options)
	} else {
		handler = slog.NewTextHandler(os.Stdout, options)
	}

	// Wrap with Sentry handler if Sentry is configured
	if os.Getenv("SENTRY_DSN") != "" {
		handler = NewSentryHandler(handler)
	}

	Logger = slog.New(handler)
}
