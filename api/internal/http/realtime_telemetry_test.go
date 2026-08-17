package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type telemetryStoreRecorder struct {
	calls int
}

func (s *telemetryStoreRecorder) StoreRealtimeTelemetry(_ context.Context, _ *model.RealtimeChatTelemetry) error {
	s.calls++
	return nil
}

type telemetryWordlistOwner struct {
	err error
}

func (o telemetryWordlistOwner) GetWordlistByID(_ context.Context, _ int64, _ int64) (*service.Wordlist, error) {
	return nil, o.err
}

func TestRealtimeTelemetryRejectsOversizedOrMalformedPayloadBeforeStorage(t *testing.T) {
	t.Parallel()
	for _, payload := range []string{
		strings.Repeat("x", maxRealtimeTelemetryBodyBytes+1),
		`{"sessionId":"01234567-89ab-cdef-0123-456789abcdef","startedAt":"2026-01-01T00:00:00Z","endedAt":"2026-01-01T00:00:01Z","durationMs":1000,"totalTurns":0,"turns":[],"rateLimitEvents":[],"deviceInfo":{"nested":{}}}`,
	} {
		t.Run("invalid telemetry", func(t *testing.T) {
			store := &telemetryStoreRecorder{}
			response := performRealtimeTelemetryRequest(recordRealtimeTelemetry(nil, store), payload)
			require.Contains(t, []int{http.StatusBadRequest, http.StatusRequestEntityTooLarge}, response.Code)
			require.Zero(t, store.calls)
		})
	}
}

func TestRealtimeTelemetryRejectsOversizedCollectionsAndRateTimestampsBeforeStorage(t *testing.T) {
	t.Parallel()
	turns := make([]string, service.MaxRealtimeTelemetryTurns+1)
	for index := range turns {
		turns[index] = `{}`
	}
	tooManyTurns := `{"sessionId":"01234567-89ab-cdef-0123-456789abcdef","startedAt":"2026-01-01T00:00:00Z","endedAt":"2026-01-01T00:00:01Z","durationMs":1000,"totalTurns":101,"turns":[` + strings.Join(turns, ",") + `],"rateLimitEvents":[],"deviceInfo":{}}`
	longRecordedAt := `{"sessionId":"01234567-89ab-cdef-0123-456789abcdef","startedAt":"2026-01-01T00:00:00Z","endedAt":"2026-01-01T00:00:01Z","durationMs":1000,"totalTurns":0,"turns":[],"rateLimitEvents":[{"name":"requests","limit":1,"remaining":0,"resetSeconds":1,"recordedAt":"` + strings.Repeat("x", maxRealtimeTelemetryRateEventBytes) + `"}],"deviceInfo":{}}`

	for _, payload := range []string{tooManyTurns, longRecordedAt} {
		store := &telemetryStoreRecorder{}
		response := performRealtimeTelemetryRequest(recordRealtimeTelemetry(nil, store), payload)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Zero(t, store.calls)
	}
}

func TestRealtimeTelemetryChecksWordlistOwnershipBeforeStorage(t *testing.T) {
	t.Parallel()
	store := &telemetryStoreRecorder{}
	handler := recordRealtimeTelemetry(telemetryWordlistOwner{err: common.NotFoundError{ID: 9, Entity: "Wordlist"}}, store)
	response := performRealtimeTelemetryRequest(handler, validRealtimeTelemetryPayload(`,"wordlistId":9`))
	require.Equal(t, http.StatusNotFound, response.Code)
	require.Zero(t, store.calls)
}

func TestRealtimeTelemetryStoresOnlyValidatedOwnedPayload(t *testing.T) {
	t.Parallel()
	store := &telemetryStoreRecorder{}
	response := performRealtimeTelemetryRequest(recordRealtimeTelemetry(nil, store), validRealtimeTelemetryPayload(""))
	require.Equal(t, http.StatusAccepted, response.Code)
	require.Equal(t, 1, store.calls)
}

func performRealtimeTelemetryRequest(handler gin.HandlerFunc, payload string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	router := gin.New()
	router.POST("/telemetry/realtime-chat", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler(c)
	})
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/telemetry/realtime-chat", strings.NewReader(payload)))
	return response
}

func validRealtimeTelemetryPayload(extra string) string {
	return `{"sessionId":"01234567-89ab-cdef-0123-456789abcdef","startedAt":"2026-01-01T00:00:00Z","endedAt":"2026-01-01T00:00:01Z","durationMs":1000,"totalAudioTokensIn":0,"totalAudioTokensOut":0,"totalTextTokensIn":0,"totalTextTokensOut":0,"totalTurns":0,"turns":[],"rateLimitEvents":[],"errorCount":0,"deviceInfo":{"isDevice":true}` + extra + `}`
}
