package http

import (
	"errors"
	"net"
	"net/http"
	"strconv"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RateLimitAuthSource(limiter *service.AuthRateLimiter, operation service.AuthLimitOperation) gin.HandlerFunc {
	return func(c *gin.Context) {
		source := canonicalAuthSource(c.ClientIP())
		if !applyAuthLimit(c, limiter, operation, service.AuthLimitSource, source) {
			return
		}
		c.Next()
	}
}

func canonicalAuthSource(source string) string {
	if parsed := net.ParseIP(source); parsed != nil {
		return parsed.String()
	}
	return source
}

func applyAuthLimit(
	c *gin.Context,
	limiter *service.AuthRateLimiter,
	operation service.AuthLimitOperation,
	dimension service.AuthLimitDimension,
	value string,
) bool {
	decision, err := limiter.Check(c.Request.Context(), operation, dimension, value)
	if err != nil {
		common.Logger.ErrorContext(c.Request.Context(), "auth limiter unavailable",
			"operation", operation, "dimension", dimension, "error", err)
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Authentication temporarily unavailable"})
		return false
	}
	if decision.Allowed {
		return true
	}
	retryAfter := service.RetryAfterSeconds(decision.RetryAfter)
	c.Header("Retry-After", retryAfter)
	common.Logger.WarnContext(c.Request.Context(), "auth rate limit reached",
		"operation", operation, "dimension", dimension, "retry_after_seconds", retryAfter)
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
	return false
}

// RateLimitErrorReports middleware checks rate limits for error reporting
func RateLimitErrorReports(db *pgxpool.Pool) gin.HandlerFunc {
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user type"})
			c.Abort()
			return
		}

		// Create service with injected database
		rateLimitService := service.NewErrorReportRateLimitService(db)

		// Check rate limits
		err := rateLimitService.CheckRateLimit(c.Request.Context(), user)
		if err != nil {
			var rateLimitErr service.RateLimitError
			if errors.As(err, &rateLimitErr) {
				writeErrorReportRateLimit(c, rateLimitErr)
				return
			}
			common.Logger.ErrorContext(c.Request.Context(), "error report quota unavailable", "error", err)
			writeErrorReportQuotaUnavailable(c)
			return
		}

		c.Next()
	}
}

func writeErrorReportRateLimit(c *gin.Context, rateLimitErr service.RateLimitError) {
	retryAfter := service.RetryAfterSeconds(rateLimitErr.RetryAfter)
	retryAfterSeconds, err := strconv.Atoi(retryAfter)
	if err != nil {
		retryAfterSeconds = 1
	}
	c.Header("Retry-After", retryAfter)
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"error":      rateLimitErr.Message,
		"retryAfter": retryAfterSeconds,
		"limit":      rateLimitErr.Limit,
		"remaining":  rateLimitErr.Remaining,
		"windowType": rateLimitErr.WindowType,
	})
}

func writeErrorReportQuotaUnavailable(c *gin.Context) {
	c.Header("Retry-After", "1")
	c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
		"error":      "Error reporting temporarily unavailable",
		"retryAfter": 1,
	})
}
