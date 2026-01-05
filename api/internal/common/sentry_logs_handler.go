package common

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/getsentry/sentry-go/attribute"
)

// SentryLogsHandler emits slog records to Sentry Logs without affecting stdout logging.
// It is enabled only in production and only when Sentry is configured.
type SentryLogsHandler struct {
	minLevel slog.Level
	attrs    []slog.Attr
	groups   []string
}

// NewSentryLogsHandler creates a new SentryLogsHandler with the minimum level to emit.
func NewSentryLogsHandler(minLevel slog.Level) *SentryLogsHandler {
	return &SentryLogsHandler{minLevel: minLevel}
}

// Enabled returns whether the handler should process logs for the given level.
func (h *SentryLogsHandler) Enabled(_ context.Context, level slog.Level) bool {
	if os.Getenv("ENV") != ProductionEnv {
		return false
	}
	if os.Getenv("SENTRY_DSN") == "" {
		return false
	}
	if level < slog.LevelInfo {
		return false
	}
	return level >= h.minLevel
}

// Handle emits the log record to Sentry Logs.
func (h *SentryLogsHandler) Handle(ctx context.Context, r slog.Record) error {
	if !h.Enabled(ctx, r.Level) {
		return nil
	}

	logger := sentry.NewLogger(ctx)

	var attrs []attribute.Builder
	if len(h.attrs) > 0 {
		prefix := ""
		if len(h.groups) > 0 {
			prefix = strings.Join(h.groups, ".")
		}
		for _, a := range h.attrs {
			addSentryAttrs(&attrs, prefix, a)
		}
	}
	r.Attrs(func(a slog.Attr) bool {
		addSentryAttrs(&attrs, "", a)
		return true
	})
	if len(attrs) > 0 {
		logger.SetAttributes(attrs...)
	}

	switch {
	case r.Level >= slog.LevelError:
		logger.Error(ctx, r.Message)
	case r.Level >= slog.LevelWarn:
		logger.Warn(ctx, r.Message)
	case r.Level >= slog.LevelInfo:
		logger.Info(ctx, r.Message)
	default:
		logger.Debug(ctx, r.Message)
	}

	return nil
}

// WithAttrs returns a handler with the given attributes applied to every log.
func (h *SentryLogsHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	next = append(next, h.attrs...)
	next = append(next, attrs...)
	return &SentryLogsHandler{minLevel: h.minLevel, attrs: next, groups: h.groups}
}

// WithGroup returns a handler with the given group.
func (h *SentryLogsHandler) WithGroup(name string) slog.Handler {
	next := make([]string, 0, len(h.groups)+1)
	next = append(next, h.groups...)
	next = append(next, name)
	return &SentryLogsHandler{minLevel: h.minLevel, attrs: h.attrs, groups: next}
}

func addSentryAttrs(attrs *[]attribute.Builder, prefix string, a slog.Attr) {
	key := a.Key
	if prefix != "" {
		if key != "" {
			key = prefix + "." + key
		} else {
			key = prefix
		}
	}
	if key == "" && a.Value.Kind() != slog.KindGroup {
		return
	}

	value := a.Value.Resolve()
	switch value.Kind() {
	case slog.KindGroup:
		for _, groupAttr := range value.Group() {
			addSentryAttrs(attrs, key, groupAttr)
		}
	default:
		*attrs = append(*attrs, slogValueToAttributeBuilder(key, value))
	}
}

func slogValueToAttributeBuilder(key string, value slog.Value) attribute.Builder {
	switch value.Kind() {
	case slog.KindString:
		return attribute.String(key, value.String())
	case slog.KindBool:
		return attribute.Bool(key, value.Bool())
	case slog.KindInt64:
		return attribute.Int64(key, value.Int64())
	case slog.KindUint64:
		u := value.Uint64()
		if u <= math.MaxInt64 {
			return attribute.Int64(key, int64(u))
		}
		return attribute.String(key, fmt.Sprintf("%d", u))
	case slog.KindFloat64:
		return attribute.Float64(key, value.Float64())
	case slog.KindDuration:
		return attribute.String(key, value.Duration().String())
	case slog.KindTime:
		return attribute.String(key, value.Time().Format(time.RFC3339Nano))
	case slog.KindAny:
		return attribute.String(key, formatAny(value.Any()))
	default:
		return attribute.String(key, value.String())
	}
}

func formatAny(v any) string {
	switch typed := v.(type) {
	case error:
		return typed.Error()
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprint(v)
	}
}
