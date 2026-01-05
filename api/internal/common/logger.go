package common

import (
	"log/slog"
	"os"
)

var options = &slog.HandlerOptions{Level: slog.LevelDebug}
var handler slog.Handler
var Logger *slog.Logger

func init() {
	// Create base handler based on environment
	if os.Getenv("ENV") == ProductionEnv {
		handler = slog.NewJSONHandler(os.Stdout, options)
	} else {
		handler = slog.NewTextHandler(os.Stdout, options)
	}

	// Wrap with Sentry handler if Sentry is configured
	if os.Getenv("SENTRY_DSN") != "" {
		sentryHandler := NewSentryHandler(handler)
		handlers := []slog.Handler{sentryHandler}
		if os.Getenv("ENV") == ProductionEnv {
			handlers = append(handlers, NewSentryLogsHandler(slog.LevelDebug))
		}
		handler = NewMultiHandler(handlers...)
	}

	Logger = slog.New(handler)
}
