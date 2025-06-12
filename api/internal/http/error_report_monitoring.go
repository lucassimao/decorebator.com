package http

import (
	"net/http"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

// GetUserErrorReportStatus returns the user's current error report status including rate limits
func GetUserErrorReportStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userAny, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		user, ok := userAny.(*model.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type"})
			return
		}

		// Get database connection
		db, err := common.GetDBConnection()
		if err != nil {
			common.Logger.Error("Failed to get database connection", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
			return
		}

		// Create service
		rateLimitService := service.NewErrorReportRateLimitService(db)

		// Get status
		status, err := rateLimitService.GetRateLimitStatus(c.Request.Context(), user)
		if err != nil {
			common.Logger.Error("Failed to get user error report status", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get status"})
			return
		}

		c.JSON(http.StatusOK, status)
	}
}

// GetErrorReportStats returns analytics about error reports for admin use
func GetErrorReportStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userAny, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		user, ok := userAny.(*model.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type"})
			return
		}

		// Get database connection
		db, err := common.GetDBConnection()
		if err != nil {
			common.Logger.Error("Failed to get database connection", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
			return
		}

		// Create service
		rateLimitService := service.NewErrorReportRateLimitService(db)

		// Get stats
		stats, err := rateLimitService.GetErrorReportStats(c.Request.Context(), 24)
		if err != nil {
			common.Logger.Error("Failed to get error report stats", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"stats":  stats,
			"userID": user.ID,
		})
	}
}