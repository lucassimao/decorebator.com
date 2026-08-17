package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetErrorReportStats returns error reporting statistics (admin only)
func GetErrorReportStats(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This endpoint should be protected by admin authentication
		// For now, we'll implement basic stats retrieval

		// Create service with injected database
		rateLimitService := service.NewErrorReportRateLimitService(db)

		// Get stats for last 24 hours by default
		stats, err := rateLimitService.GetErrorReportStats(c.Request.Context(), 24)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statistics"})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

// GetUserErrorReportStatus returns the current rate limit status for a user
func GetUserErrorReportStatus(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userAny, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		user, ok := userAny.(*model.User)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user type"})
			return
		}

		// Create service with injected database
		rateLimitService := service.NewErrorReportRateLimitService(db)
		limit, err := parsePageLimit(c)
		if err != nil {
			writeInvalidPage(c, err)
			return
		}
		cursor, err := parseCooldownCursor(c.Query("cursor"))
		if err != nil {
			writeInvalidPage(c, err)
			return
		}

		// Get rate limit status
		status, err := rateLimitService.GetRateLimitStatus(c.Request.Context(), user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rate limit status"})
			return
		}

		// Get active cooldowns
		cooldowns, next, err := rateLimitService.GetUserActiveCooldowns(c.Request.Context(), user.ID, limit, cursor)
		if err != nil {
			// Don't fail the whole request if cooldowns can't be fetched
			cooldowns = []map[string]interface{}{}
			next = nil
		}
		c.Header(nextCursorHeader, "")
		if next != nil {
			c.Header(nextCursorHeader, encodeCooldownCursor(*next))
		}

		response := gin.H{
			"rateLimits":      status,
			"activeCooldowns": cooldowns,
		}

		c.JSON(http.StatusOK, response)
	}
}

const maxCooldownCursorBytes = 512

type cooldownCursorPayload struct {
	CooldownUntil time.Time `json:"cooldownUntil"`
	WordID        int64     `json:"wordId"`
	DefinitionID  int64     `json:"definitionId"`
	ErrorType     string    `json:"errorType"`
}

func parseCooldownCursor(raw string) (*service.ErrorReportCooldownCursor, error) {
	if raw == "" {
		return nil, nil
	}
	if len(raw) > maxCooldownCursorBytes {
		return nil, fmt.Errorf("cursor is too long")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || len(decoded) == 0 || len(decoded) > maxCooldownCursorBytes {
		return nil, fmt.Errorf("cursor is invalid")
	}
	decoder := json.NewDecoder(bytes.NewReader(decoded))
	decoder.DisallowUnknownFields()
	var payload cooldownCursorPayload
	if err := decoder.Decode(&payload); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return nil, fmt.Errorf("cursor is invalid")
	}
	if payload.CooldownUntil.IsZero() || payload.WordID < 0 || payload.DefinitionID < 0 || payload.ErrorType == "" || len(payload.ErrorType) > 50 {
		return nil, fmt.Errorf("cursor is invalid")
	}
	return &service.ErrorReportCooldownCursor{
		CooldownUntil: payload.CooldownUntil,
		WordID:        payload.WordID,
		DefinitionID:  payload.DefinitionID,
		ErrorType:     payload.ErrorType,
	}, nil
}

func encodeCooldownCursor(cursor service.ErrorReportCooldownCursor) string {
	payload, err := json.Marshal(cooldownCursorPayload{
		CooldownUntil: cursor.CooldownUntil,
		WordID:        cursor.WordID,
		DefinitionID:  cursor.DefinitionID,
		ErrorType:     cursor.ErrorType,
	})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(payload)
}
