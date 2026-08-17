package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxRealtimeTelemetryBodyBytes = 64 << 10

const (
	maxRealtimeTelemetryTopLevelFields  = 18
	maxRealtimeTelemetryTimestampBytes  = 64
	maxRealtimeTelemetryTurnBytes       = 512
	maxRealtimeTelemetryRateEventBytes  = 1024
	maxRealtimeTelemetryDeviceInfoBytes = 8 << 10
)

type realtimeTelemetryRequest struct {
	SessionID           string                    `json:"sessionId"`
	ConversationID      *string                   `json:"conversationId"`
	WordlistID          *int64                    `json:"wordlistId"`
	LanguageCode        *string                   `json:"languageCode"`
	Model               *string                   `json:"model"`
	StartedAt           string                    `json:"startedAt"`
	EndedAt             string                    `json:"endedAt"`
	DurationMs          int                       `json:"durationMs"`
	TotalAudioTokensIn  int                       `json:"totalAudioTokensIn"`
	TotalAudioTokensOut int                       `json:"totalAudioTokensOut"`
	TotalTextTokensIn   int                       `json:"totalTextTokensIn"`
	TotalTextTokensOut  int                       `json:"totalTextTokensOut"`
	TotalTurns          int                       `json:"totalTurns"`
	Turns               []model.RealtimeChatTurn  `json:"turns"`
	RateLimitEvents     []model.RateLimitSnapshot `json:"rateLimitEvents"`
	ErrorCount          int                       `json:"errorCount"`
	ClientVersion       *string                   `json:"clientVersion"`
	DeviceInfo          map[string]any            `json:"deviceInfo"`
}

type realtimeTelemetryStore interface {
	StoreRealtimeTelemetry(context.Context, *model.RealtimeChatTelemetry) error
}

type realtimeWordlistOwner interface {
	GetWordlistByID(context.Context, int64, int64) (*service.Wordlist, error)
}

func RegisterRealtimeTelemetryRoutes(r *gin.RouterGroup, wordlistService realtimeWordlistOwner, telemetryService realtimeTelemetryStore) {
	r.POST("/telemetry/realtime-chat", recordRealtimeTelemetry(wordlistService, telemetryService))
}

func recordRealtimeTelemetry(wordlistService realtimeWordlistOwner, telemetryService realtimeTelemetryStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		req, err := decodeRealtimeTelemetryRequest(c)
		if err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Telemetry payload is too large"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid telemetry payload"})
			return
		}

		sessionUUID, err := uuid.Parse(req.SessionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sessionId"})
			return
		}
		startedAt, err := time.Parse(time.RFC3339, req.StartedAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid startedAt timestamp"})
			return
		}
		endedAt, err := time.Parse(time.RFC3339, req.EndedAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid endedAt timestamp"})
			return
		}

		telemetry := &model.RealtimeChatTelemetry{
			UserID:              c.GetInt64("userID"),
			WordlistID:          req.WordlistID,
			SessionUUID:         sessionUUID.String(),
			ConversationID:      req.ConversationID,
			LanguageCode:        req.LanguageCode,
			Model:               req.Model,
			StartedAt:           startedAt,
			EndedAt:             endedAt,
			DurationMs:          req.DurationMs,
			TotalAudioTokensIn:  req.TotalAudioTokensIn,
			TotalAudioTokensOut: req.TotalAudioTokensOut,
			TotalTextTokensIn:   req.TotalTextTokensIn,
			TotalTextTokensOut:  req.TotalTextTokensOut,
			TotalTurns:          req.TotalTurns,
			Turns:               req.Turns,
			RateLimits:          req.RateLimitEvents,
			ErrorCount:          req.ErrorCount,
			ClientVersion:       req.ClientVersion,
			DeviceInfo:          req.DeviceInfo,
		}
		if err := service.ValidateRealtimeTelemetry(telemetry); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid telemetry payload"})
			return
		}
		if telemetry.WordlistID != nil {
			if wordlistService == nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Telemetry unavailable"})
				return
			}
			wordlist, ownershipErr := wordlistService.GetWordlistByID(c.Request.Context(), *telemetry.WordlistID, telemetry.UserID)
			if ownershipErr != nil {
				if isNotFound(ownershipErr) {
					respondNotFound(c)
					return
				}
				common.Logger.ErrorContext(c.Request.Context(), "failed to verify telemetry wordlist ownership", "error", ownershipErr, "userID", telemetry.UserID)
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Telemetry unavailable"})
				return
			}
			if wordlist == nil {
				respondNotFound(c)
				return
			}
		}
		if telemetryService == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Telemetry unavailable"})
			return
		}
		if err := telemetryService.StoreRealtimeTelemetry(c.Request.Context(), telemetry); err != nil {
			common.Logger.ErrorContext(c.Request.Context(), "failed to persist realtime telemetry", "error", err)
		}
		c.Status(http.StatusAccepted)
	}
}

func decodeRealtimeTelemetryRequest(c *gin.Context) (realtimeTelemetryRequest, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRealtimeTelemetryBodyBytes)
	decoder := json.NewDecoder(c.Request.Body)
	var request realtimeTelemetryRequest
	first, err := decoder.Token()
	if err != nil {
		return realtimeTelemetryRequest{}, err
	}
	if delimiter, ok := first.(json.Delim); !ok || delimiter != '{' {
		return realtimeTelemetryRequest{}, errors.New("telemetry must be an object")
	}
	seen := make(map[string]struct{}, maxRealtimeTelemetryTopLevelFields)
	for decoder.More() {
		nameToken, tokenErr := decoder.Token()
		if tokenErr != nil {
			return realtimeTelemetryRequest{}, tokenErr
		}
		name, ok := nameToken.(string)
		if !ok {
			return realtimeTelemetryRequest{}, errors.New("telemetry field name is invalid")
		}
		if _, exists := seen[name]; exists || len(seen) >= maxRealtimeTelemetryTopLevelFields {
			return realtimeTelemetryRequest{}, errors.New("telemetry fields are invalid")
		}
		seen[name] = struct{}{}
		if err := decodeRealtimeTelemetryField(decoder, &request, name); err != nil {
			return realtimeTelemetryRequest{}, err
		}
	}
	if closing, closeErr := decoder.Token(); closeErr != nil {
		return realtimeTelemetryRequest{}, closeErr
	} else if delimiter, ok := closing.(json.Delim); !ok || delimiter != '}' {
		return realtimeTelemetryRequest{}, errors.New("telemetry object is invalid")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return realtimeTelemetryRequest{}, errors.New("multiple telemetry JSON values")
		}
		return realtimeTelemetryRequest{}, err
	}
	return request, nil
}

func decodeRealtimeTelemetryField(decoder *json.Decoder, request *realtimeTelemetryRequest, name string) error {
	switch name {
	case "sessionId":
		value, err := decodeBoundedJSONString(decoder, 36, false)
		request.SessionID = value
		return err
	case "conversationId":
		value, err := decodeOptionalBoundedJSONString(decoder, 256)
		request.ConversationID = value
		return err
	case "wordlistId":
		return decoder.Decode(&request.WordlistID)
	case "languageCode":
		value, err := decodeOptionalBoundedJSONString(decoder, 16)
		request.LanguageCode = value
		return err
	case "model":
		value, err := decodeOptionalBoundedJSONString(decoder, 128)
		request.Model = value
		return err
	case "startedAt":
		value, err := decodeBoundedJSONString(decoder, maxRealtimeTelemetryTimestampBytes, false)
		request.StartedAt = value
		return err
	case "endedAt":
		value, err := decodeBoundedJSONString(decoder, maxRealtimeTelemetryTimestampBytes, false)
		request.EndedAt = value
		return err
	case "durationMs":
		return decoder.Decode(&request.DurationMs)
	case "totalAudioTokensIn":
		return decoder.Decode(&request.TotalAudioTokensIn)
	case "totalAudioTokensOut":
		return decoder.Decode(&request.TotalAudioTokensOut)
	case "totalTextTokensIn":
		return decoder.Decode(&request.TotalTextTokensIn)
	case "totalTextTokensOut":
		return decoder.Decode(&request.TotalTextTokensOut)
	case "totalTurns":
		return decoder.Decode(&request.TotalTurns)
	case "turns":
		turns, err := decodeBoundedTelemetryArray[model.RealtimeChatTurn](decoder, service.MaxRealtimeTelemetryTurns, maxRealtimeTelemetryTurnBytes)
		request.Turns = turns
		return err
	case "rateLimitEvents":
		rateLimits, err := decodeBoundedTelemetryArray[model.RateLimitSnapshot](decoder, service.MaxRealtimeTelemetryRateLimitEvents, maxRealtimeTelemetryRateEventBytes)
		request.RateLimitEvents = rateLimits
		return err
	case "errorCount":
		return decoder.Decode(&request.ErrorCount)
	case "clientVersion":
		value, err := decodeOptionalBoundedJSONString(decoder, 64)
		request.ClientVersion = value
		return err
	case "deviceInfo":
		value, err := decodeBoundedTelemetryValue[map[string]any](decoder, maxRealtimeTelemetryDeviceInfoBytes)
		request.DeviceInfo = value
		return err
	default:
		return errors.New("telemetry field is unknown")
	}
}

func decodeBoundedJSONString(decoder *json.Decoder, limit int, nullable bool) (string, error) {
	raw, err := decodeBoundedRawValue(decoder, 2+limit*6)
	if err != nil {
		return "", err
	}
	if nullable && string(raw) == "null" {
		return "", nil
	}
	var value string
	if err := strictUnmarshalTelemetryJSON(raw, &value); err != nil || len(value) > limit {
		return "", errors.New("telemetry string is invalid")
	}
	return value, nil
}

func decodeOptionalBoundedJSONString(decoder *json.Decoder, limit int) (*string, error) {
	raw, err := decodeBoundedRawValue(decoder, 2+limit*6)
	if err != nil {
		return nil, err
	}
	if string(raw) == "null" {
		return nil, nil
	}
	var value string
	if err := strictUnmarshalTelemetryJSON(raw, &value); err != nil || len(value) > limit {
		return nil, errors.New("telemetry string is invalid")
	}
	return &value, nil
}

func decodeBoundedTelemetryArray[T any](decoder *json.Decoder, limit int, itemBytes int) ([]T, error) {
	opening, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delimiter, ok := opening.(json.Delim); !ok || delimiter != '[' {
		return nil, errors.New("telemetry collection is invalid")
	}
	items := make([]T, 0, limit)
	for decoder.More() {
		if len(items) == limit {
			return nil, errors.New("telemetry collection exceeds its limit")
		}
		var item T
		raw, rawErr := decodeBoundedRawValue(decoder, itemBytes)
		if rawErr != nil || strictUnmarshalTelemetryJSON(raw, &item) != nil {
			return nil, errors.New("telemetry collection item is invalid")
		}
		items = append(items, item)
	}
	closing, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delimiter, ok := closing.(json.Delim); !ok || delimiter != ']' {
		return nil, errors.New("telemetry collection is invalid")
	}
	return items, nil
}

func decodeBoundedTelemetryValue[T any](decoder *json.Decoder, limit int) (T, error) {
	var value T
	raw, err := decodeBoundedRawValue(decoder, limit)
	if err != nil || strictUnmarshalTelemetryJSON(raw, &value) != nil {
		return value, errors.New("telemetry value is invalid")
	}
	return value, nil
}

func decodeBoundedRawValue(decoder *json.Decoder, limit int) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return nil, err
	}
	if len(raw) == 0 || len(raw) > limit {
		return nil, errors.New("telemetry value exceeds its limit")
	}
	return raw, nil
}

func strictUnmarshalTelemetryJSON(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return func() error {
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			return errors.New("trailing JSON value")
		}
		return nil
	}()
}
