package http

import (
	"os"
	"strings"

	"decorebator.com/internal/common"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

// SentryMiddlewares returns a slice of Sentry-related middlewares
// It includes both the sentrygin middleware and our request context capture middleware
// This handles request context globally and static routes via URL pattern detection
func SentryMiddlewares() []gin.HandlerFunc {
	// Skip if not in production or Sentry not configured
	if os.Getenv("ENV") != productionEnv || os.Getenv("SENTRY_DSN") == "" {
		return []gin.HandlerFunc{}
	}

	return []gin.HandlerFunc{
		// First: sentrygin middleware to capture errors and create Sentry hub
		sentrygin.New(sentrygin.Options{
			Repanic:         true,
			WaitForDelivery: false,
		}),
		// Second: our request context middleware to capture request info and static user context
		sentryRequestContextMiddleware(),
	}
}

// sentryRequestContextMiddleware captures request context and static user context
// This runs globally and handles static routes via URL pattern detection
func sentryRequestContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Sentry hub from context (set by sentrygin middleware)
		hub := sentrygin.GetHubFromContext(c)
		if hub == nil {
			// If no Sentry hub available, just continue
			c.Next()
			return
		}

		// Always capture request context
		reqCtx := common.CreateRequestContextFromGin(c)

		// Set request context in Sentry scope immediately
		hub.WithScope(func(scope *sentry.Scope) {
			scope.SetContext("request", map[string]interface{}{
				"url":          reqCtx.URL,
				"method":       reqCtx.Method,
				"path":         reqCtx.Path,
				"query_params": reqCtx.QueryParams,
				"user_agent":   reqCtx.UserAgent,
				"remote_ip":    reqCtx.RemoteIP,
			})
		})

		// Handle static routes only (JWT routes handled by separate middleware after auth)
		authType := detectAuthType(c)
		if authType == "static" {
			userCtx := common.CreateUserContextFromGin(c, authType)
			if userCtx != nil && userCtx.ID != "" {
				hub.WithScope(func(scope *sentry.Scope) {
					scope.SetUser(sentry.User{
						ID:    userCtx.ID,
						Email: userCtx.Email,
					})

					// Set additional user tags
					scope.SetTag("auth_type", userCtx.AuthType)
					if userCtx.SubscriptionPlan != "" {
						scope.SetTag("subscription_plan", userCtx.SubscriptionPlan)
					}
				})

				// Add user context to request context for logging
				enhancedCtx := common.SetUserContext(c.Request.Context(), userCtx)
				c.Request = c.Request.WithContext(enhancedCtx)
			}
		}

		// Create enhanced context with request info for downstream handlers
		enhancedCtx := common.SetRequestContext(c.Request.Context(), reqCtx)
		c.Request = c.Request.WithContext(enhancedCtx)

		c.Next()
	}
}

// SentryUserContextMiddleware captures user context for JWT authenticated routes
// This runs after the Authenticate middleware, so c.Get("userID") is available
func SentryUserContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Sentry hub from context (set by sentrygin middleware)
		hub := sentrygin.GetHubFromContext(c)
		if hub == nil {
			// If no Sentry hub available, just continue
			c.Next()
			return
		}

		// Detect JWT authentication (should be available after Authenticate middleware)
		authType := detectAuthType(c)
		if authType == "jwt" {
			userCtx := common.CreateUserContextFromGin(c, authType)
			if userCtx != nil && userCtx.ID != "" {
				hub.WithScope(func(scope *sentry.Scope) {
					scope.SetUser(sentry.User{
						ID:    userCtx.ID,
						Email: userCtx.Email,
					})

					// Set additional user tags
					scope.SetTag("auth_type", userCtx.AuthType)
					if userCtx.SubscriptionPlan != "" {
						scope.SetTag("subscription_plan", userCtx.SubscriptionPlan)
					}
				})

				// Add user context to request context for logging
				enhancedCtx := common.SetUserContext(c.Request.Context(), userCtx)
				c.Request = c.Request.WithContext(enhancedCtx)
			}
		}

		c.Next()
	}
}

// detectAuthType automatically determines the authentication type based on available context
func detectAuthType(c *gin.Context) string {
	// Check if JWT user info is available (after JWT authentication middleware)
	if _, exists := c.Get("userID"); exists {
		return "jwt"
	}

	// Check if we're on a static auth route (after static authentication middleware)
	if strings.HasPrefix(c.FullPath(), "/static/") {
		return "static"
	}

	// No authentication context available
	return ""
}
