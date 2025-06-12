package http

import (
	"net/http"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

// RateLimitErrorReports middleware checks rate limits for error reporting
func RateLimitErrorReports() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userAny, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}
		user, ok := userAny.(*model.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type"})
			c.Abort()
			return
		}

		// Get database connection
		db, err := common.GetDBConnection()
		if err != nil {
			common.Logger.Error("Failed to get database connection", "error", err)
			// Don't block on database failures, allow the request
			c.Next()
			return
		}

		// Create service
		rateLimitService := service.NewErrorReportRateLimitService(db)

		// Check rate limits
		err = rateLimitService.CheckRateLimit(c.Request.Context(), user)
		if err != nil {
			if rateLimitErr, ok := err.(service.RateLimitError); ok {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":      rateLimitErr.Message,
					"retryAfter": int(rateLimitErr.RetryAfter.Seconds()),
					"limit":      rateLimitErr.Limit,
					"remaining":  rateLimitErr.Remaining,
					"windowType": rateLimitErr.WindowType,
				})
				c.Abort()
				return
			}
			// Other errors - log but don't block
			common.Logger.Error("Rate limit check failed", "error", err)
			c.Next()
			return
		}

		c.Next()
	}
}