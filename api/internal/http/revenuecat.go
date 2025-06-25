package http

import (
	"io"
	"net/http"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

// HandleRevenueCatWebhook handles RevenueCat webhook events
func HandleRevenueCatWebhook(rcService *service.RevenueCatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the request body
		payload, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		// Get authorization header for webhook verification
		authHeader := c.GetHeader("Authorization")

		// Handle the webhook
		if err := rcService.HandleWebhook(c.Request.Context(), payload, authHeader); err != nil {
			// Log the error but return 200 to RevenueCat to avoid retries
			common.Logger.Error("Failed to handle RevenueCat webhook", "error", err)
			// Return 200 to acknowledge receipt and prevent retries
			c.JSON(http.StatusOK, gin.H{"received": true})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

// RestorePurchases handles purchase restoration from RevenueCat
func RestorePurchases(rcService *service.RevenueCatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userAny, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			return
		}
		user := userAny.(*model.User)

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
