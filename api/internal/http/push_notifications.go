package http

import (
	"net/http"
	"regexp"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/repository"
	"github.com/gin-gonic/gin"
)

var expoPushTokenPattern = regexp.MustCompile(`^Expo(?:nent)?PushToken\[[A-Za-z0-9_-]{8,180}\]$`)

type PushNotificationRoutes struct {
	pushRepo *repository.PushTokenRepository
}

func NewPushNotificationRoutes(pushRepo *repository.PushTokenRepository) *PushNotificationRoutes {
	return &PushNotificationRoutes{pushRepo: pushRepo}
}

type registerPushTokenInput struct {
	ExpoPushToken string  `json:"expoPushToken" binding:"required"`
	Platform      string  `json:"platform" binding:"required"`
	DeviceID      *string `json:"deviceId,omitempty"`
	Timezone      string  `json:"timezone" binding:"required"`
	Locale        *string `json:"locale,omitempty"`
}

func (h *PushNotificationRoutes) Register(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 512)
	var input registerPushTokenInput
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Platform != "ios" && input.Platform != "android" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid platform"})
		return
	}
	if !expoPushTokenPattern.MatchString(input.ExpoPushToken) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid push token"})
		return
	}
	if _, err := time.LoadLocation(input.Timezone); err != nil {
		common.Logger.WarnContext(c.Request.Context(), "invalid timezone for push token registration",
			"timezone", input.Timezone,
			"userID", c.GetInt64("userID"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid timezone"})
		return
	}

	userID := c.GetInt64("userID")
	if h.pushRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "push repository unavailable"})
		return
	}

	err := h.pushRepo.Upsert(c.Request.Context(), repository.UpsertPushTokenInput{
		UserID:    userID,
		ExpoToken: input.ExpoPushToken,
		Platform:  input.Platform,
		DeviceID:  input.DeviceID,
		Timezone:  input.Timezone,
		Locale:    input.Locale,
	})
	if err != nil {
		common.Logger.ErrorContext(c.Request.Context(), "failed to register push token", "error", err, "userID", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register token"})
		return
	}

	c.Status(http.StatusNoContent)
}

type unregisterPushTokenInput struct {
	ExpoPushToken string `json:"expoPushToken" binding:"required"`
}

func (h *PushNotificationRoutes) Unregister(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 512)
	var input unregisterPushTokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid push token request"})
		return
	}
	if !expoPushTokenPattern.MatchString(input.ExpoPushToken) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid push token"})
		return
	}

	if h.pushRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "push repository unavailable"})
		return
	}

	// Unregistration is intentionally a possession-based public operation: an
	// expired local session must still be able to stop notifications for the
	// exact opaque device token it holds. Registration remains authenticated.
	if err := h.pushRepo.DeactivateByTokens(c.Request.Context(), []string{input.ExpoPushToken}); err != nil {
		common.Logger.ErrorContext(c.Request.Context(), "failed to deactivate push token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unregister token"})
		return
	}

	c.Status(http.StatusNoContent)
}
