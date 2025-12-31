package http

import (
	"net/http"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/repository"
	"github.com/gin-gonic/gin"
)

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
	var input registerPushTokenInput
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Platform != "ios" && input.Platform != "android" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid platform"})
		return
	}
	if _, err := time.LoadLocation(input.Timezone); err != nil {
		common.Logger.Warn("invalid timezone for push token registration",
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
		common.Logger.Error("failed to register push token", "error", err, "userID", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register token"})
		return
	}

	c.Status(http.StatusNoContent)
}

type unregisterPushTokenInput struct {
	ExpoPushToken string `json:"expoPushToken" binding:"required"`
}

func (h *PushNotificationRoutes) Unregister(c *gin.Context) {
	var input unregisterPushTokenInput
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt64("userID")
	if h.pushRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "push repository unavailable"})
		return
	}

	if err := h.pushRepo.Deactivate(c.Request.Context(), userID, input.ExpoPushToken); err != nil {
		common.Logger.Error("failed to deactivate push token", "error", err, "userID", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unregister token"})
		return
	}

	c.Status(http.StatusNoContent)
}
