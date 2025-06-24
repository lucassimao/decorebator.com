package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/mail"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

// AntiBurstMiddleware checks for burst patterns and implements progressive blocking
func AntiBurstMiddleware(endpoint string) gin.HandlerFunc {
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

		// Get Redis client
		redisClient, err := common.GetRedisClient()
		if err != nil {
			common.Logger.Error("Failed to get Redis client", "error", err)
			// Don't block on Redis failures, allow the request
			c.Next()
			return
		}

		// Create burst detector
		detector := service.NewBurstDetector(redisClient)

		// First check if user is blocked
		if detector.IsUserBlocked(c.Request.Context(), user.ID) {
			// Get the actual block expiry time
			blockedUntil := detector.GetBlockExpiry(c.Request.Context(), user.ID)
			if blockedUntil == nil {
				// Fallback if we can't get expiry time
				defaultExpiry := time.Now().Add(24 * time.Hour)
				blockedUntil = &defaultExpiry
			}

			c.JSON(http.StatusForbidden, gin.H{
				"error":           "Your account has been temporarily suspended due to unusual activity. It will be automatically restored after 24 hours.",
				"error_code":      "ACCOUNT_SUSPENDED_BURST",
				"suspended_until": blockedUntil.Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		// Check for burst
		isBurst, violations, err := detector.CheckAndTrackBurst(c.Request.Context(), user.ID, endpoint)
		if err != nil {
			common.Logger.Error("Failed to check burst", "error", err)
			// Don't block on error, allow the request
			c.Next()
			return
		}

		if isBurst {
			if violations >= 2 {
				// Second violation today - block for the day
				if err := detector.BlockUserFor24Hours(c.Request.Context(), user.ID); err != nil {
					common.Logger.Error("Failed to block user", "error", err)
				}

				// Get the actual block expiry time after blocking
				blockedUntil := detector.GetBlockExpiry(c.Request.Context(), user.ID)
				if blockedUntil == nil {
					// Fallback - should not happen, but just in case
					defaultExpiry := time.Now().Add(24 * time.Hour)
					blockedUntil = &defaultExpiry
				}

				// Send alerts asynchronously
				go func() {
					activityType := strings.Replace(endpoint, "_", " ", -1)
					if err := mail.SendBurstBlockedEmail(user.ID, user.Email, user.FirstName, activityType, violations, *blockedUntil); err != nil {
						common.Logger.Error("Failed to send burst block email", "error", err, "user_id", user.ID)
					}
					logToSentry(user, endpoint, violations, *blockedUntil)
				}()

				c.JSON(http.StatusForbidden, gin.H{
					"error":           "Your account has been temporarily suspended due to repeated unusual activity. It will be automatically restored after 24 hours.",
					"error_code":      "ACCOUNT_SUSPENDED_BURST",
					"suspended_until": blockedUntil.Format(time.RFC3339),
				})
			} else {
				// First violation - warning
				common.Logger.Warn("Burst detected - first violation",
					"user_id", user.ID,
					"email", user.Email,
					"endpoint", endpoint,
				)

				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "Unusual activity detected. Please slow down. Continued abuse may result in temporary suspension.",
					"error_code":  "BURST_WARNING",
					"retry_after": 60,
				})
			}
			c.Abort()
			return
		}

		c.Next()
	}
}

// logToSentry logs burst abuse incidents to Sentry for monitoring
func logToSentry(user *model.User, endpoint string, violations int, blockedUntil time.Time) {
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelWarning)
		scope.SetTag("alert_type", "burst_abuse")
		scope.SetTag("endpoint", endpoint)
		scope.SetUser(sentry.User{
			ID:       strconv.FormatInt(user.ID, 10),
			Email:    user.Email,
			Username: user.FirstName + " " + user.LastName,
		})
		scope.SetContext("burst_details", map[string]interface{}{
			"endpoint":         endpoint,
			"violations_today": violations,
			"action":           "24h_block",
			"blocked_until":    blockedUntil.Format(time.RFC3339),
			"block_duration":   "24 hours",
		})

		message := fmt.Sprintf("User blocked for burst abuse - %s (ID: %d)", user.Email, user.ID)
		sentry.CaptureMessage(message)
	})
}
