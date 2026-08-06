package common

import (
	"os"

	"github.com/getsentry/sentry-go"
)

// InitSentry initializes the Sentry SDK with the configured DSN.
// This function should be called early in the application lifecycle,
// before any logging occurs.
func InitSentry() error {
	sentryDsn := os.Getenv("SENTRY_DSN")

	// In production, Sentry DSN is required
	if os.Getenv("ENV") == ProductionEnv && sentryDsn == "" {
		panic("SENTRY_DSN not found in production environment")
	}

	// Skip initialization if no DSN is provided (development/testing)
	if sentryDsn == "" {
		Logger.Info("Sentry DSN not provided, skipping Sentry initialization")
		return nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:            sentryDsn,
		Debug:          false, // Set to false in production to reduce overhead
		SendDefaultPII: false,
		EnableLogs:     os.Getenv("ENV") == ProductionEnv,
		// Performance optimizations
		SampleRate:       1.0, // Capture 100% of errors
		TracesSampleRate: 0.1, // Only capture 10% of performance traces
		// Buffering settings
		MaxBreadcrumbs: 50,
		// The SDK handles async sending internally
		AttachStacktrace: true,
		Environment:      os.Getenv("ENV"),
		BeforeSendLog: func(log *sentry.Log) *sentry.Log {
			if log == nil {
				return nil
			}
			if log.Level == sentry.LogLevelTrace {
				return nil
			}
			return scrubSentryLog(log)
		},
		BeforeSend: scrubSentryEvent,
	})

	if err != nil {
		Logger.Error("Sentry initialization failed", "error", err)
		return err
	}

	Logger.Info("Sentry initialized successfully", "environment", os.Getenv("ENV"))
	return nil
}
