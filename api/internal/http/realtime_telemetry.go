package http

import (
	"net/http"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type realtimeTelemetryRequest struct {
	SessionID           string                    `json:"sessionId" binding:"required"`
	ConversationID      *string                   `json:"conversationId"`
	WordlistID          *int64                    `json:"wordlistId"`
	LanguageCode        *string                   `json:"languageCode"`
	Model               *string                   `json:"model"`
	StartedAt           string                    `json:"startedAt" binding:"required"`
	EndedAt             string                    `json:"endedAt" binding:"required"`
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

func RegisterRealtimeTelemetryRoutes(r *gin.RouterGroup, telemetryService *service.RealtimeTelemetryService) {
	r.POST("/telemetry/realtime-chat", recordRealtimeTelemetry(telemetryService))
}

func recordRealtimeTelemetry(telemetryService *service.RealtimeTelemetryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req realtimeTelemetryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
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

		if err := telemetryService.StoreRealtimeTelemetry(c.Request.Context(), telemetry); err != nil {
			common.Logger.Error("failed to persist realtime telemetry", "error", err)
			c.Status(http.StatusAccepted)
			return
		}

		c.Status(http.StatusAccepted)
	}
}
