package service

import (
	"testing"
	"time"

	"decorebator.com/internal/model"
	"github.com/stretchr/testify/require"
)

func TestValidateRealtimeTelemetryEnforcesCollectionAndNumericBounds(t *testing.T) {
	t.Parallel()
	base := &model.RealtimeChatTelemetry{
		UserID:      1,
		SessionUUID: "01234567-89ab-cdef-0123-456789abcdef",
		StartedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndedAt:     time.Date(2026, 1, 1, 0, 0, 1, 0, time.UTC),
		DurationMs:  1000,
		DeviceInfo:  map[string]any{"isDevice": true},
	}
	require.NoError(t, ValidateRealtimeTelemetry(base))

	tooManyTurns := *base
	tooManyTurns.TotalTurns = MaxRealtimeTelemetryTurns + 1
	require.Error(t, ValidateRealtimeTelemetry(&tooManyTurns))

	negativeTokens := *base
	negativeTokens.TotalAudioTokensIn = -1
	require.Error(t, ValidateRealtimeTelemetry(&negativeTokens))

	invalidSession := *base
	invalidSession.SessionUUID = "not-a-uuid"
	require.Error(t, ValidateRealtimeTelemetry(&invalidSession))

	nestedDeviceInfo := *base
	nestedDeviceInfo.DeviceInfo = map[string]any{"nested": map[string]any{}}
	require.Error(t, ValidateRealtimeTelemetry(&nestedDeviceInfo))

	badRecordedAt := *base
	badRecordedAt.RateLimits = []model.RateLimitSnapshot{{
		Name:         "requests",
		Limit:        1,
		Remaining:    0,
		ResetSeconds: 1,
		RecordedAt:   "2201-01-01T00:00:00Z",
	}}
	require.Error(t, ValidateRealtimeTelemetry(&badRecordedAt))
}
