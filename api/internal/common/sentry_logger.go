package common

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/getsentry/sentry-go"
)

// SentryHandler wraps slog.Handler to send errors to Sentry
// It sends events asynchronously to avoid blocking the main request flow
type SentryHandler struct {
	handler slog.Handler
}

// NewSentryHandler creates a new SentryHandler
func NewSentryHandler(h slog.Handler) *SentryHandler {
	return &SentryHandler{handler: h}
}

// Enabled returns whether the handler is enabled for the given level
func (h *SentryHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle processes a log record and sends errors to Sentry
func (h *SentryHandler) Handle(ctx context.Context, r slog.Record) error {
	// Always pass through to the wrapped handler
	if err := h.handler.Handle(ctx, r); err != nil {
		return err
	}

	// Only send errors and above to Sentry in production
	if os.Getenv("ENV") == productionEnv && r.Level >= slog.LevelError {
		// Extract attributes
		attrs := make(map[string]interface{})
		r.Attrs(func(a slog.Attr) bool {
			attrs[a.Key] = a.Value.Any()
			return true
		})

		// Create Sentry event
		event := sentry.NewEvent()
		event.Level = sentry.LevelError
		event.Message = r.Message
		event.Extra = attrs

		// Add error if present
		if err, ok := attrs["error"].(error); ok {
			event.Exception = []sentry.Exception{
				{
					Type:  fmt.Sprintf("%T", err),
					Value: err.Error(),
				},
			}
		}

		// Add user context if available
		if userID, ok := attrs["userId"].(int64); ok {
			event.User = sentry.User{
				ID: fmt.Sprintf("%d", userID),
			}
		} else if userID, ok := attrs["user_id"].(int64); ok {
			event.User = sentry.User{
				ID: fmt.Sprintf("%d", userID),
			}
		}

		// Send to Sentry asynchronously to avoid blocking
		// The Sentry SDK handles buffering and retries internally
		go func() {
			hub := sentry.CurrentHub().Clone()
			hub.CaptureEvent(event)
		}()
	}

	return nil
}

// WithAttrs returns a handler with the given attributes
func (h *SentryHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SentryHandler{handler: h.handler.WithAttrs(attrs)}
}

// WithGroup returns a handler with the given group
func (h *SentryHandler) WithGroup(name string) slog.Handler {
	return &SentryHandler{handler: h.handler.WithGroup(name)}
}

// CaptureError is a helper function to log an error and send it to Sentry
func CaptureError(ctx context.Context, err error, message string, attrs ...any) {
	// Add error to attributes if not already present
	hasError := false
	for i := 0; i < len(attrs); i += 2 {
		if key, ok := attrs[i].(string); ok && key == "error" {
			hasError = true
			break
		}
	}
	if !hasError {
		attrs = append(attrs, "error", err)
	}

	// Log the error
	Logger.ErrorContext(ctx, message, attrs...)
}

// CaptureException sends an exception directly to Sentry with additional context
func CaptureException(ctx context.Context, err error, attrs map[string]interface{}) {
	if os.Getenv("ENV") != productionEnv || err == nil {
		return
	}

	hub := sentry.CurrentHub().Clone()
	hub.WithScope(func(scope *sentry.Scope) {
		// Set context
		if ctx != nil {
			if reqCtx, ok := ctx.Value("request").(map[string]interface{}); ok {
				scope.SetContext("request", reqCtx)
			}
		}

		// Set extra context
		for k, v := range attrs {
			scope.SetExtra(k, v)
		}

		// Capture the exception
		hub.CaptureException(err)
	})
}
