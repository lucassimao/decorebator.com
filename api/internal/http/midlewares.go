package http

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/config"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

const productionEnv = "production"

func Authenticate(
	sessions *service.AuthSessionService,
	users *repository.UserRepository,
	policy config.HTTPSecurityConfig,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		const BearerSchema = "Bearer "
		authorization := c.GetHeader("Authorization")
		if authorization == "" {
			authorization, _ = c.Cookie(authenticationCookieName(policy))
		}
		if authorization == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization missing"})
			return
		}
		tokenString := strings.TrimPrefix(authorization, BearerSchema)
		identity, err := sessions.ValidateAccess(c.Request.Context(), tokenString)
		if err != nil {
			if c.Request.Context().Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout - please try again"})
				return
			}
			if errors.Is(err, service.ErrInvalidAccessToken) || errors.Is(err, repository.ErrSessionExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token validation error"})
			} else {
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Authentication temporarily unavailable"})
			}
			return
		}
		currentUsers, err := users.Find(c.Request.Context(), repository.FindUserArgs{ID: &identity.UserID})
		if err != nil {
			if c.Request.Context().Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout - please try again"})
				return
			}
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
			if c.Request.Context().Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout - please try again"})
				return
			}
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

func AuthenticateStatic(expected string) gin.HandlerFunc {
	expectedHash := sha256.Sum256([]byte(expected))
	return func(c *gin.Context) {
		providedHash := sha256.Sum256([]byte(c.GetHeader("Authorization")))
		if expected == "" || subtle.ConstantTimeCompare(providedHash[:], expectedHash[:]) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Wrong credentials"})
			return
		}
		c.Next()
	}
}

func CORSMiddleware(policy config.HTTPSecurityConfig) gin.HandlerFunc {
	allowedOrigins := make(map[string]struct{}, len(policy.AllowedOrigins))
	for _, origin := range policy.AllowedOrigins {
		allowedOrigins[origin] = struct{}{}
	}
	return func(c *gin.Context) {
		c.Writer.Header().Set("Vary", "Origin")
		originHeaders := c.Request.Header.Values("Origin")
		if len(originHeaders) > 1 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Malformed Origin header"})
			return
		}
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			if _, allowed := allowedOrigins[origin]; !allowed {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Origin not allowed"})
				return
			}
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Type, Content-Length, Retry-After, X-Next-Cursor, X-Definitions-Continuation")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func SecurityHeaders(production bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		headers := c.Writer.Header()
		headers.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'")
		headers.Set("Cache-Control", "no-store")
		headers.Set("X-Frame-Options", "DENY")
		headers.Set("X-Content-Type-Options", "nosniff")
		headers.Set("Referrer-Policy", "no-referrer")
		headers.Set("Permissions-Policy", "accelerometer=(), camera=(), geolocation=(), gyroscope=(), microphone=(), payment=(), usb=()")
		if production {
			headers.Set("Strict-Transport-Security", "max-age=31536000")
		}
		c.Next()
	}
}

func ExpireLegacyParentDomainCookies(policy config.HTTPSecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if policy.LegacyCookieDomain != "" && policy.LegacyCookieDomain != policy.CookieDomain {
			c.SetSameSite(http.SameSiteStrictMode)
			c.SetCookie("Authorization", "", -1, "/", policy.LegacyCookieDomain, policy.SecureCookies, true)
			c.SetCookie("RefreshToken", "", -1, "/", policy.LegacyCookieDomain, policy.SecureCookies, true)
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

				attrs = append(attrs, slog.Any("stack_trace", stackTrace))

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
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
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
