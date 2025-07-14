package http

import (
	"io"
	"net/http"
	"os"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

// HandleRevenueCatWebhook handles RevenueCat webhook events
func HandleRevenueCatWebhook(jobService service.JobService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Verify Authorization header
		authHeader := c.GetHeader("Authorization")
		expectedToken, exists := os.LookupEnv("REVENUECAT_WEBHOOK_AUTHORIZATION")

		if !exists {
			common.Logger.Error("REVENUECAT_WEBHOOK_AUTHORIZATION is not set")
			panic("REVENUECAT_WEBHOOK_AUTHORIZATION env is missing")
		}

		if authHeader != expectedToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token"})
			return
		}

		// 2. Read the request body
		payload, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		// Enqueue a job to process the webhook
		_, err = jobService.TriggerRevenueCatWebhookWorker("webhook", payload)

		if err != nil {
			common.Logger.Error("failed to enqueue revenuecat webhook job", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue webhook job"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

// RestorePurchases handles purchase restoration from RevenueCat
func RestorePurchases(rcService service.RevenueCatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userAny, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			return
		}
		user, ok := userAny.(*model.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
			return
		}

		// Parse request
		var req struct {
			AppUserID string             `json:"appUserId" binding:"required"`
			Platform  model.PlatformType `json:"platform" binding:"required,oneof=ios android"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Restore purchases
		if err := rcService.RestorePurchases(c.Request.Context(), user.ID, req.AppUserID, req.Platform); err != nil {
			common.Logger.Error("Failed to restore purchases", "error", err, "user_id", user.ID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore purchases"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Purchases restored successfully"})
	}
}
