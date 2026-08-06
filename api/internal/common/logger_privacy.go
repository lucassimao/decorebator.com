package common

import (
	"context"
	"log/slog"
	"strings"
)

// PrivacyHandler removes sensitive structured attributes before they reach any
// local, Sentry, or future fan-out log sink. Call sites should still avoid
// attaching private values; this is the final defense against regressions.
type PrivacyHandler struct {
	next      slog.Handler
	dropAttrs bool
}

func NewPrivacyHandler(next slog.Handler) *PrivacyHandler {
	return &PrivacyHandler{next: next}
}

func (h *PrivacyHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *PrivacyHandler) Handle(ctx context.Context, record slog.Record) error {
	clean := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	if !h.dropAttrs {
		record.Attrs(func(attr slog.Attr) bool {
			if filtered, ok := privacySafeLogAttr(attr); ok {
				clean.AddAttrs(filtered)
			}
			return true
		})
	}
	return h.next.Handle(ctx, clean)
}

func (h *PrivacyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if h.dropAttrs {
		return h
	}
	return &PrivacyHandler{next: h.next.WithAttrs(privacySafeLogAttrs(attrs))}
}

func (h *PrivacyHandler) WithGroup(name string) slog.Handler {
	if h.dropAttrs || sensitiveLogKey(name) {
		return &PrivacyHandler{next: h.next, dropAttrs: true}
	}
	return &PrivacyHandler{next: h.next.WithGroup(name)}
}

func privacySafeLogAttrs(attrs []slog.Attr) []slog.Attr {
	clean := make([]slog.Attr, 0, len(attrs))
	for _, attr := range attrs {
		if filtered, ok := privacySafeLogAttr(attr); ok {
			clean = append(clean, filtered)
		}
	}
	return clean
}

func privacySafeLogAttr(attr slog.Attr) (slog.Attr, bool) {
	if attr.Equal(slog.Attr{}) || sensitiveLogKey(attr.Key) {
		return slog.Attr{}, false
	}
	attr.Value = attr.Value.Resolve()
	if attr.Value.Kind() != slog.KindGroup {
		return attr, true
	}
	children := privacySafeLogAttrs(attr.Value.Group())
	if len(children) == 0 {
		return slog.Attr{}, false
	}
	attr.Value = slog.GroupValue(children...)
	return attr, true
}

func sensitiveLogKey(key string) bool {
	normalized := normalizedTelemetryKey(key)
	for _, exact := range []string{"query", "query_string", "raw_query", "url_query", "search_query"} {
		if normalized == exact {
			return true
		}
	}
	for _, fragment := range []string{
		"email", "authorization", "cookie", "password", "token", "secret",
		"remote_ip", "ip_address", "x_forwarded_for", "x_real_ip",
	} {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}
	return false
}
