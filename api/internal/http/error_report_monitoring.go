package http

import (
	"net/http"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

// GetErrorReportStats returns error reporting statistics (admin only)
func GetErrorReportStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This endpoint should be protected by admin authentication
		// For now, we'll implement basic stats retrieval

		db, err := common.GetDBConnection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
			return
		}

		// Create service
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
func GetUserErrorReportStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userAny, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		user := userAny.(*model.User)

		// Get database connection
		db, err := common.GetDBConnection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
			return
		}

		// Create service
		rateLimitService := service.NewErrorReportRateLimitService(db)

		// Get rate limit status
		status, err := rateLimitService.GetRateLimitStatus(c.Request.Context(), user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rate limit status"})
			return
		}

		// Get active cooldowns
		cooldowns, err := rateLimitService.GetUserActiveCooldowns(c.Request.Context(), user.ID)
		if err != nil {
			// Don't fail the whole request if cooldowns can't be fetched
			cooldowns = []map[string]interface{}{}
		}

		response := gin.H{
			"rateLimits":      status,
			"activeCooldowns": cooldowns,
		}

		c.JSON(http.StatusOK, response)
	}
}