package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	MaxRealtimeTelemetryTurns           = 100
	MaxRealtimeTelemetryRateLimitEvents = 64
	MaxRealtimeTelemetryDurationMs      = 4 * 60 * 60 * 1000
	MaxRealtimeTelemetryTokenCount      = 100_000_000
	MaxRealtimeTelemetryErrorCount      = 1_000
	maxTelemetryConversationIDLength    = 256
	maxTelemetryLanguageCodeLength      = 16
	maxTelemetryModelLength             = 128
	maxTelemetryClientVersionLength     = 64
	maxTelemetryDeviceInfoFields        = 8
	maxTelemetryDeviceInfoKeyLength     = 32
	maxTelemetryDeviceInfoStringLength  = 128
	maxTelemetryRateLimitNameLength     = 64
	maxTelemetryRecordedAtLength        = 64
	maxTelemetryRateLimitResetSeconds   = 24 * 60 * 60
)

type RealtimeTelemetryService struct {
	repo *repository.RealtimeTelemetryRepository
}

func NewRealtimeTelemetryService(db *pgxpool.Pool) *RealtimeTelemetryService {
	return &RealtimeTelemetryService{
		repo: repository.NewRealtimeTelemetryRepository(db),
	}
}

func (s *RealtimeTelemetryService) StoreRealtimeTelemetry(ctx context.Context, telemetry *model.RealtimeChatTelemetry) error {
	if telemetry == nil {
		return nil
	}
	if err := ValidateRealtimeTelemetry(telemetry); err != nil {
		return err
	}
	return s.repo.InsertRealtimeChatTelemetry(ctx, telemetry)
}

// ValidateRealtimeTelemetry keeps direct callers from bypassing the HTTP
// contract. Telemetry is diagnostic data, never a source of billing truth.
func ValidateRealtimeTelemetry(telemetry *model.RealtimeChatTelemetry) error { //nolint:gocyclo // The fields form one external contract.
	if telemetry == nil {
		return nil
	}
	if telemetry.UserID <= 0 {
		return fmt.Errorf("telemetry user ID must be positive")
	}
	if telemetry.WordlistID != nil && *telemetry.WordlistID <= 0 {
		return fmt.Errorf("telemetry wordlist ID must be positive")
	}
	if telemetry.SessionUUID == "" || len(telemetry.SessionUUID) > 36 {
		return fmt.Errorf("telemetry session ID is invalid")
	}
	if _, err := uuid.Parse(telemetry.SessionUUID); err != nil {
		return fmt.Errorf("telemetry session ID is invalid")
	}
	if telemetry.StartedAt.IsZero() || telemetry.EndedAt.IsZero() || telemetry.EndedAt.Before(telemetry.StartedAt) {
		return fmt.Errorf("telemetry timestamps are invalid")
	}
	if err := validateTelemetryOptionalString("conversation ID", telemetry.ConversationID, maxTelemetryConversationIDLength); err != nil {
		return err
	}
	if err := validateTelemetryOptionalString("language code", telemetry.LanguageCode, maxTelemetryLanguageCodeLength); err != nil {
		return err
	}
	if err := validateTelemetryOptionalString("model", telemetry.Model, maxTelemetryModelLength); err != nil {
		return err
	}
	if err := validateTelemetryOptionalString("client version", telemetry.ClientVersion, maxTelemetryClientVersionLength); err != nil {
		return err
	}
	if telemetry.DurationMs < 0 || telemetry.DurationMs > MaxRealtimeTelemetryDurationMs {
		return fmt.Errorf("telemetry duration is out of range")
	}
	for _, value := range []int{
		telemetry.TotalAudioTokensIn,
		telemetry.TotalAudioTokensOut,
		telemetry.TotalTextTokensIn,
		telemetry.TotalTextTokensOut,
	} {
		if value < 0 || value > MaxRealtimeTelemetryTokenCount {
			return fmt.Errorf("telemetry token count is out of range")
		}
	}
	if telemetry.TotalTurns < 0 || telemetry.TotalTurns > MaxRealtimeTelemetryTurns || telemetry.TotalTurns != len(telemetry.Turns) {
		return fmt.Errorf("telemetry turn count is invalid")
	}
	if len(telemetry.Turns) > MaxRealtimeTelemetryTurns || len(telemetry.RateLimits) > MaxRealtimeTelemetryRateLimitEvents {
		return fmt.Errorf("telemetry collection exceeds its limit")
	}
	if telemetry.ErrorCount < 0 || telemetry.ErrorCount > MaxRealtimeTelemetryErrorCount {
		return fmt.Errorf("telemetry error count is out of range")
	}
	for _, turn := range telemetry.Turns {
		if err := validateTelemetryTurn(turn); err != nil {
			return err
		}
	}
	for _, rateLimit := range telemetry.RateLimits {
		if err := validateTelemetryRateLimit(rateLimit); err != nil {
			return err
		}
	}
	return validateTelemetryDeviceInfo(telemetry.DeviceInfo)
}

func validateTelemetryOptionalString(name string, value *string, limit int) error {
	if value == nil {
		return nil
	}
	if len(*value) == 0 || len(*value) > limit {
		return fmt.Errorf("telemetry %s is invalid", name)
	}
	return nil
}

func validateTelemetryTurn(turn model.RealtimeChatTurn) error {
	for _, value := range []*int{turn.UserSpeechMs, turn.AssistantLatencyMs} {
		if value != nil && (*value < 0 || *value > MaxRealtimeTelemetryDurationMs) {
			return fmt.Errorf("telemetry turn duration is out of range")
		}
	}
	if turn.Usage == nil {
		return nil
	}
	for _, value := range []int{
		turn.Usage.InputTokens,
		turn.Usage.OutputTokens,
		turn.Usage.AudioInputTokens,
		turn.Usage.AudioOutputTokens,
		turn.Usage.TextInputTokens,
		turn.Usage.TextOutputTokens,
	} {
		if value < 0 || value > MaxRealtimeTelemetryTokenCount {
			return fmt.Errorf("telemetry turn token count is out of range")
		}
	}
	return nil
}

func validateTelemetryRateLimit(rateLimit model.RateLimitSnapshot) error {
	if len(rateLimit.Name) == 0 || len(rateLimit.Name) > maxTelemetryRateLimitNameLength ||
		rateLimit.Limit < 0 || rateLimit.Limit > MaxRealtimeTelemetryTokenCount ||
		rateLimit.Remaining < 0 || rateLimit.Remaining > rateLimit.Limit ||
		math.IsNaN(rateLimit.ResetSeconds) || math.IsInf(rateLimit.ResetSeconds, 0) ||
		rateLimit.ResetSeconds < 0 || rateLimit.ResetSeconds > maxTelemetryRateLimitResetSeconds {
		return fmt.Errorf("telemetry rate limit is invalid")
	}
	if len(rateLimit.RecordedAt) == 0 || len(rateLimit.RecordedAt) > maxTelemetryRecordedAtLength {
		return fmt.Errorf("telemetry rate limit timestamp is invalid")
	}
	recordedAt, err := time.Parse(time.RFC3339, rateLimit.RecordedAt)
	if err != nil || recordedAt.Year() < 2000 || recordedAt.Year() > 2100 {
		return fmt.Errorf("telemetry rate limit timestamp is invalid")
	}
	return nil
}

func validateTelemetryDeviceInfo(deviceInfo map[string]any) error {
	if len(deviceInfo) > maxTelemetryDeviceInfoFields {
		return fmt.Errorf("telemetry device information has too many fields")
	}
	for key, value := range deviceInfo {
		if len(key) == 0 || len(key) > maxTelemetryDeviceInfoKeyLength {
			return fmt.Errorf("telemetry device information key is invalid")
		}
		switch typed := value.(type) {
		case nil, bool:
		case string:
			if len(typed) > maxTelemetryDeviceInfoStringLength {
				return fmt.Errorf("telemetry device information value is too long")
			}
		case float64:
			if math.IsNaN(typed) || math.IsInf(typed, 0) || math.Abs(typed) > MaxRealtimeTelemetryTokenCount {
				return fmt.Errorf("telemetry device information number is out of range")
			}
		default:
			return fmt.Errorf("telemetry device information value is invalid")
		}
	}
	return nil
}
