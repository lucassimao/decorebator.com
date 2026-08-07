package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

const productionEnv = "production"

func Authenticate(sessions *service.AuthSessionService, users *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		const BearerSchema = "Bearer "
		authorization := c.GetHeader("Authorization")
		if authorization == "" {
			authorization, _ = c.Cookie("Authorization")
		}
		if authorization == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization missing"})
			return
		}
		tokenString := strings.TrimPrefix(authorization, BearerSchema)
		identity, err := sessions.ValidateAccess(c.Request.Context(), tokenString)
		if err != nil {
			if errors.Is(err, service.ErrInvalidAccessToken) || errors.Is(err, repository.ErrSessionExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token validation error"})
			} else {
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Authentication temporarily unavailable"})
			}
			return
		}
		currentUsers, err := users.Find(c.Request.Context(), repository.FindUserArgs{ID: &identity.UserID})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Authentication temporarily unavailable"})
			return
		}
		if len(currentUsers) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		user := currentUsers[0]
		user.PasswordHash = ""
		user.StripeCustomerID = nil
		c.Set("userID", identity.UserID)
		c.Set("sessionID", identity.SessionID)
		c.Set("user", &user)
		c.Next()
	}
}

// ResolveEffectiveSubscription replaces only the request-scoped legacy plan
// projection. It never mutates the JWT or users table.
func ResolveEffectiveSubscription(access *service.EffectiveAccessService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if access == nil {
			c.Next()
			return
		}
		userAny, exists := c.Get("user")
		user, ok := userAny.(*model.User)
		if !exists || !ok || user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user context"})
			return
		}
		plan, err := access.Plan(c.Request.Context(), user.ID, user.SubscriptionPlan)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "Subscription access is temporarily unavailable",
			})
			return
		}
		user.SubscriptionPlan = plan
		c.Set("user", user)
		c.Next()
	}
}

func AuthenticateStatic(c *gin.Context) {
	authorization := c.GetHeader("Authorization")

	if authorization == "" || authorization != os.Getenv("STATIC_AUTHENTICATION") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Wrong credentials"})
		return
	}

	c.Next()
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow only a single, valid origin per request. Multiple values are invalid per CORS spec
		// and may be coalesced by proxies (e.g., Cloudflare) into a comma-separated list, causing failures.
		allowedOrigins := map[string]struct{}{
			"https://decorebator.com":     {},
			"https://www.decorebator.com": {},
		}

		origin := c.Request.Header.Get("Origin")
		if os.Getenv("ENV") == productionEnv {
			if origin != "" {
				if _, ok := allowedOrigins[origin]; ok {
					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
					c.Writer.Header().Set("Vary", "Origin")
				}
			}
		} else {
			if origin != "" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Set("Vary", "Origin")
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Type, Content-Length, Authorization, Cookie")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Cookie")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create timeout context and replace request context
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Check if request timed out during processing
		if ctx.Err() == context.DeadlineExceeded {
			// Only respond if no response was already written
			if !c.Writer.Written() {
				common.Logger.ErrorContext(ctx, "request timed out",
					"path", c.FullPath(),
					"method", c.Request.Method,
					"timeout", timeout)

				c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
					"error": "Request timeout - please try again",
				})
			}
		}
	}
}

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				var err error
				switch v := rec.(type) {
				case string:
					err = errors.New(v)
				case error:
					err = v
				default:
					err = fmt.Errorf("unknown panic: %v", v)
				}

				stackTrace := string(debug.Stack())
				origErr := errors.Unwrap(err)

				attrs := []any{
					slog.Any("error", err),
					slog.Any("original_error", origErr),
					slog.String("path", c.FullPath()),
					slog.String("method", c.Request.Method),
					slog.String("url", c.Request.URL.Path),
				}

				// optionally include userID
				if val, exists := c.Get("userID"); exists {
					if userID, ok := val.(int64); ok {
						attrs = append(attrs, slog.Int64("userId", userID))
					}
				}

				if os.Getenv("ENV") == productionEnv {
					attrs = append(attrs, slog.Any("stack_trace", stackTrace))
				} else {
					// In development, include stack trace in the log
					attrs = append(attrs, slog.Any("stack_trace", stackTrace))
				}

				// Capture the error with Sentry if available
				if hub := sentrygin.GetHubFromContext(c); hub != nil {
					hub.WithScope(func(scope *sentry.Scope) {
						// Set error context
						scope.SetTag("error_type", "panic")
						scope.SetContext("request", map[string]interface{}{
							"url":    c.Request.URL.Path,
							"method": c.Request.Method,
							"path":   c.FullPath(),
						})

						// Add user context if available
						if val, exists := c.Get("userID"); exists {
							if userID, ok := val.(int64); ok {
								scope.SetUser(sentry.User{
									ID: fmt.Sprintf("%d", userID),
								})
							}
						}

						// Capture the exception
						hub.CaptureException(err)
					})
				}

				common.Logger.ErrorContext(c.Request.Context(), "Recovered from panic", attrs...)

				// Return a 500 error in JSON format
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "We could not process your request at this time.",
				})
			}
		}()

		// Continue to the next handler
		c.Next()
	}
}

// CheckSubscriptionLimits middleware checks if user has permission based on subscription
func CheckSubscriptionLimits(subService *service.SubscriptionService, action model.UserAction) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userAny, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			c.Abort()
			return
		}
		user, ok := userAny.(*model.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
			return
		}

		// Prepare options for subscription check
		options := &service.SubscriptionCheckOptions{
			HasOptimisticSubscription: c.Query("hasOptimisticSubscription") == "true",
		}

		// Get wordlist ID for add word operations
		if action == model.UserActionAddWord {
			wordlistIDStr := c.Param("wordlistId")
			if wordlistIDStr == "" {
				wordlistIDStr = c.Query("wordlistId")
			}
			if wordlistIDStr != "" {
				wordlistID, err := strconv.ParseInt(wordlistIDStr, 10, 64)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
					c.Abort()
					return
				}
				options.WordlistID = &wordlistID
			}
		}

		// Use the service to check subscription limits (includes RevenueCat verification)
		if err := subService.CheckSubscriptionLimits(c.Request.Context(), user.ID, action, options); err != nil {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error": err.Error(),
				"code":  "SUBSCRIPTION_LIMIT_REACHED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
