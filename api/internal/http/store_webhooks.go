package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"sync"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	appleWebhookMaxBodyBytes = 300 * 1024
	googleRTDNMaxBodyBytes   = 64 * 1024
)

type AppleNotificationReceiver interface {
	Ingest(context.Context, string) (repository.ProviderEventInboxResult, error)
}

type GoogleRTDNReceiver interface {
	Ingest(context.Context, string, []byte) (repository.ProviderEventInboxResult, error)
}

type StoreWebhookObservation struct {
	Provider        string                            `json:"provider"`
	Disposition     string                            `json:"disposition"`
	Outcome         model.EntitlementOperationOutcome `json:"outcome,omitempty"`
	ResultCode      model.EntitlementResultCode       `json:"resultCode,omitempty"`
	Duplicate       bool                              `json:"duplicate"`
	DeliveryAttempt int                               `json:"deliveryAttempt,omitempty"`
}

type StoreWebhookObserver interface {
	Observe(StoreWebhookObservation)
}

type StoreWebhookMetrics struct {
	mu     sync.RWMutex
	counts map[StoreWebhookObservation]uint64
}

type StoreWebhookMetricEntry struct {
	StoreWebhookObservation
	Count uint64 `json:"count"`
}

func NewStoreWebhookMetrics() *StoreWebhookMetrics {
	return &StoreWebhookMetrics{counts: make(map[StoreWebhookObservation]uint64)}
}

func (m *StoreWebhookMetrics) Observe(observation StoreWebhookObservation) {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.counts[observation]++
	m.mu.Unlock()
}

func (m *StoreWebhookMetrics) Snapshot() map[StoreWebhookObservation]uint64 {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[StoreWebhookObservation]uint64, len(m.counts))
	for observation, count := range m.counts {
		result[observation] = count
	}
	return result
}

func (m *StoreWebhookMetrics) Entries() []StoreWebhookMetricEntry {
	snapshot := m.Snapshot()
	entries := make([]StoreWebhookMetricEntry, 0, len(snapshot))
	for observation, count := range snapshot {
		entries = append(entries, StoreWebhookMetricEntry{StoreWebhookObservation: observation, Count: count})
	}
	sort.Slice(entries, func(left, right int) bool {
		if entries[left].Provider != entries[right].Provider {
			return entries[left].Provider < entries[right].Provider
		}
		if entries[left].Disposition != entries[right].Disposition {
			return entries[left].Disposition < entries[right].Disposition
		}
		if entries[left].ResultCode != entries[right].ResultCode {
			return entries[left].ResultCode < entries[right].ResultCode
		}
		if entries[left].Outcome != entries[right].Outcome {
			return entries[left].Outcome < entries[right].Outcome
		}
		if entries[left].DeliveryAttempt != entries[right].DeliveryAttempt {
			return entries[left].DeliveryAttempt < entries[right].DeliveryAttempt
		}
		return !entries[left].Duplicate && entries[right].Duplicate
	})
	return entries
}

func GetStoreWebhookMetrics(metrics *StoreWebhookMetrics) gin.HandlerFunc {
	if metrics == nil {
		panic("store webhook metrics are required")
	}
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"metrics": metrics.Entries()})
	}
}

func HandleAppleStoreNotification(
	receiver AppleNotificationReceiver,
	observer StoreWebhookObserver,
) gin.HandlerFunc {
	if receiver == nil {
		panic("Apple notification receiver is required")
	}
	return func(c *gin.Context) {
		body, err := readBoundedWebhookBody(c, appleWebhookMaxBodyBytes)
		if err != nil {
			observeStoreWebhook(observer, "apple", "acknowledge", repository.ProviderEventInboxResult{}, 0)
			common.Logger.WarnContext(c.Request.Context(), "store webhook rejected",
				"provider", "apple", "disposition", "acknowledge", "reason", "invalid_body")
			c.Status(http.StatusNoContent)
			return
		}
		var envelope struct {
			SignedPayload string `json:"signedPayload"`
		}
		if err = decodeSingleJSON(body, &envelope); err != nil || envelope.SignedPayload == "" {
			observeStoreWebhook(observer, "apple", "acknowledge", repository.ProviderEventInboxResult{}, 0)
			common.Logger.WarnContext(c.Request.Context(), "store webhook rejected",
				"provider", "apple", "disposition", "acknowledge", "reason", "invalid_envelope")
			c.Status(http.StatusNoContent)
			return
		}
		result, ingestErr := receiver.Ingest(c.Request.Context(), envelope.SignedPayload)
		disposition := service.AppleNotificationDispositionOf(ingestErr)
		observeStoreWebhook(observer, "apple", string(disposition), result, 0)
		logStoreWebhookResult(c, "apple", string(disposition), result, 0, ingestErr)
		if disposition == service.AppleNotificationRetry {
			c.Status(http.StatusServiceUnavailable)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func HandleGooglePlayRTDN(
	receiver GoogleRTDNReceiver,
	observer StoreWebhookObserver,
) gin.HandlerFunc {
	if receiver == nil {
		panic("Google RTDN receiver is required")
	}
	return func(c *gin.Context) {
		body, err := readBoundedWebhookBody(c, googleRTDNMaxBodyBytes)
		if err != nil {
			observeStoreWebhook(observer, "google", "acknowledge", repository.ProviderEventInboxResult{}, 0)
			common.Logger.WarnContext(c.Request.Context(), "store webhook rejected",
				"provider", "google", "disposition", "acknowledge", "reason", "invalid_body")
			c.Status(http.StatusNoContent)
			return
		}
		result, ingestErr := receiver.Ingest(
			c.Request.Context(), c.GetHeader("Authorization"), body,
		)
		deliveryAttempt := googleDeliveryAttempt(body)
		disposition := service.GoogleRTDNDispositionOf(ingestErr)
		observeStoreWebhook(observer, "google", string(disposition), result, deliveryAttempt)
		logStoreWebhookResult(c, "google", string(disposition), result, deliveryAttempt, ingestErr)
		switch disposition {
		case service.GoogleRTDNUnauthorized:
			c.Status(http.StatusUnauthorized)
		case service.GoogleRTDNRetry:
			c.Status(http.StatusServiceUnavailable)
		default:
			c.Status(http.StatusNoContent)
		}
	}
}

func readBoundedWebhookBody(c *gin.Context, limit int64) ([]byte, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) == 0 {
		return nil, errors.New("webhook body is invalid")
	}
	return body, nil
}

func decodeSingleJSON(body []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("webhook body contains trailing data")
	}
	return nil
}

func observeStoreWebhook(
	observer StoreWebhookObserver,
	provider string,
	disposition string,
	result repository.ProviderEventInboxResult,
	deliveryAttempt int,
) {
	if observer == nil {
		return
	}
	observer.Observe(StoreWebhookObservation{
		Provider: provider, Disposition: disposition, Outcome: result.Outcome,
		ResultCode: result.Code, Duplicate: result.Duplicate, DeliveryAttempt: deliveryAttempt,
	})
}

func logStoreWebhookResult(
	c *gin.Context,
	provider string,
	disposition string,
	result repository.ProviderEventInboxResult,
	deliveryAttempt int,
	err error,
) {
	attributes := []any{
		"provider", provider, "disposition", disposition,
		"outcome", result.Outcome, "result_code", result.Code, "duplicate", result.Duplicate,
		"delivery_attempt", deliveryAttempt,
	}
	if err != nil {
		attributes = append(attributes, "error_kind", StoreWebhookErrorKind(err))
		common.Logger.WarnContext(c.Request.Context(), "store webhook delivery completed", attributes...)
		return
	}
	common.Logger.InfoContext(c.Request.Context(), "store webhook delivery completed", attributes...)
}

// StoreWebhookErrorKind deliberately exports only bounded classifications for
// logs and metrics. Raw provider/database errors can contain purchase evidence,
// credential paths, or upstream response data and must never be logged here.
func StoreWebhookErrorKind(err error) string {
	if err == nil {
		return "none"
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return "request_timeout"
	}
	if errors.Is(err, service.ErrInvalidAppleSignedData) {
		return "invalid_apple_signed_data"
	}
	if errors.Is(err, service.ErrInvalidGooglePubSubAuthentication) {
		return "invalid_google_authentication"
	}
	if errors.Is(err, service.ErrInvalidGoogleRTDNEnvelope) {
		return "invalid_google_envelope"
	}
	var appleVerification *service.ApplePurchaseVerificationError
	if errors.As(err, &appleVerification) {
		return storeVerificationFailureKind(appleVerification.Failure)
	}
	var googleVerification *service.GooglePurchaseVerificationError
	if errors.As(err, &googleVerification) {
		return storeVerificationFailureKind(googleVerification.Failure)
	}
	var googleToken *service.GoogleAccessTokenError
	if errors.As(err, &googleToken) {
		return "google_credentials"
	}
	var appleAPI *service.AppleAPIError
	if errors.As(err, &appleAPI) {
		return "apple_provider_api"
	}
	var googleAPI *service.GooglePlayAPIError
	if errors.As(err, &googleAPI) {
		return "google_provider_api"
	}
	return "internal_error"
}

func storeVerificationFailureKind(failure model.EntitlementOperationFailure) string {
	switch failure {
	case model.EntitlementFailureProviderUnavailable:
		return "provider_unavailable"
	case model.EntitlementFailureInvalidEvidence:
		return "invalid_evidence"
	case model.EntitlementFailureAccountMismatch:
		return "account_mismatch"
	case model.EntitlementFailureUnknownProduct:
		return "unknown_product"
	default:
		return "verification_failure"
	}
}

func googleDeliveryAttempt(body []byte) int {
	var envelope struct {
		DeliveryAttempt int `json:"deliveryAttempt"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil || envelope.DeliveryAttempt < 1 || envelope.DeliveryAttempt > 100 {
		return 0
	}
	return envelope.DeliveryAttempt
}
