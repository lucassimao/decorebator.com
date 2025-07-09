package common

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

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
	if os.Getenv("ENV") == ProductionEnv && r.Level >= slog.LevelError {
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

		// Capture current stack trace for better debugging
		stack := debug.Stack()
		stacktrace := parseGoStackTrace(stack)

		// Add caller information to extra context
		if callerFrame := createCallerFrame(3); callerFrame != nil { // Skip: Handle -> slog internals -> caller
			if event.Extra == nil {
				event.Extra = make(map[string]interface{})
			}
			event.Extra["caller_file"] = callerFrame.Filename
			event.Extra["caller_line"] = callerFrame.Lineno
			event.Extra["caller_function"] = callerFrame.Function
			event.Extra["caller_package"] = callerFrame.Package
		}

		// Add error if present
		if err, ok := attrs["error"].(error); ok {
			event.Exception = []sentry.Exception{
				{
					Type:       fmt.Sprintf("%T", err),
					Value:      err.Error(),
					Stacktrace: stacktrace,
				},
			}
		} else {
			// Even without an explicit error, create an exception with stack trace
			// This helps pinpoint where the error log was generated
			event.Exception = []sentry.Exception{
				{
					Type:       "LogError",
					Value:      r.Message,
					Stacktrace: stacktrace,
				},
			}
		}

		// Add request context from Go context if available
		if reqCtx := GetRequestContext(ctx); reqCtx != nil {
			event.Request = &sentry.Request{
				URL:    reqCtx.URL,
				Method: reqCtx.Method,
				Headers: map[string]string{
					"User-Agent": reqCtx.UserAgent,
				},
				QueryString: reqCtx.QueryParams,
			}
			// Also add as context for additional detail
			event.Contexts["request"] = map[string]interface{}{
				"url":          reqCtx.URL,
				"method":       reqCtx.Method,
				"path":         reqCtx.Path,
				"query_params": reqCtx.QueryParams,
				"user_agent":   reqCtx.UserAgent,
				"remote_ip":    reqCtx.RemoteIP,
			}
		}

		// Add user context from Go context if available
		if userCtx := GetUserContext(ctx); userCtx != nil && userCtx.ID != "" {
			event.User = sentry.User{
				ID:    userCtx.ID,
				Email: userCtx.Email,
			}
			// Add user tags
			if event.Tags == nil {
				event.Tags = make(map[string]string)
			}
			event.Tags["auth_type"] = userCtx.AuthType
			if userCtx.SubscriptionPlan != "" {
				event.Tags["subscription_plan"] = userCtx.SubscriptionPlan
			}
		}

		// Fallback: Add user context from log attributes if not in context
		if event.User.ID == "" {
			if userID, ok := attrs["userId"].(int64); ok {
				event.User = sentry.User{
					ID: fmt.Sprintf("%d", userID),
				}
			} else if userID, ok := attrs["user_id"].(int64); ok {
				event.User = sentry.User{
					ID: fmt.Sprintf("%d", userID),
				}
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

// CaptureError logs an error using the application's standard logger.
// It is a logging-first approach. If the environment is set to production,
// the SentryHandler will intercept the log entry and send it to Sentry as an event.
// This is best used for notable errors that should be recorded in application logs
// and also appear in Sentry.
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

	// Log the error (SentryHandler will pick up context automatically)
	Logger.ErrorContext(ctx, message, attrs...)
}

// CaptureException sends an exception directly to the Sentry SDK, bypassing the
// application's standard logger. This is a direct reporting mechanism intended for
// critical, unexpected errors (e.g., in a panic recovery middleware).
// It provides Sentry with a more detailed report, including a full stack trace,
// which allows for better grouping and analysis of critical failures.
func CaptureException(ctx context.Context, err error, attrs map[string]interface{}) {
	if os.Getenv("ENV") != ProductionEnv || err == nil {
		return
	}

	hub := sentry.CurrentHub().Clone()
	hub.WithScope(func(scope *sentry.Scope) {
		// Set request context from enhanced context
		if reqCtx := GetRequestContext(ctx); reqCtx != nil {
			scope.SetContext("request", map[string]interface{}{
				"url":          reqCtx.URL,
				"method":       reqCtx.Method,
				"path":         reqCtx.Path,
				"query_params": reqCtx.QueryParams,
				"user_agent":   reqCtx.UserAgent,
				"remote_ip":    reqCtx.RemoteIP,
			})
		}

		// Set user context from enhanced context
		if userCtx := GetUserContext(ctx); userCtx != nil && userCtx.ID != "" {
			scope.SetUser(sentry.User{
				ID:    userCtx.ID,
				Email: userCtx.Email,
			})
			scope.SetTag("auth_type", userCtx.AuthType)
			if userCtx.SubscriptionPlan != "" {
				scope.SetTag("subscription_plan", userCtx.SubscriptionPlan)
			}
		}

		// Fallback: Set context from legacy format
		if ctx != nil {
			if reqCtx, ok := ctx.Value("request").(map[string]interface{}); ok {
				scope.SetContext("request", reqCtx)
			}
		}

		// Set extra context
		for k, v := range attrs {
			scope.SetExtra(k, v)
		}

		// Add caller information
		if callerFrame := createCallerFrame(1); callerFrame != nil { // Skip: CaptureException -> caller
			scope.SetExtra("caller_file", callerFrame.Filename)
			scope.SetExtra("caller_line", callerFrame.Lineno)
			scope.SetExtra("caller_function", callerFrame.Function)
			scope.SetExtra("caller_package", callerFrame.Package)
		}

		// Capture the exception with stack trace
		hub.CaptureException(err)
	})
}
