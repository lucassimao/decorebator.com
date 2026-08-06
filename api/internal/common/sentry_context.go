package common

import (
	"context"
	"fmt"

	"decorebator.com/internal/model"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

// Context keys for storing request and user information
type contextKey string

const (
	RequestContextKey contextKey = "sentry_request_context"
	UserContextKey    contextKey = "sentry_user_context"
)

// RequestContext contains request-level information for Sentry
type RequestContext struct {
	URL       string `json:"url"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	UserAgent string `json:"user_agent"`
}

// UserContext contains user-level information for Sentry
type UserContext struct {
	ID               string `json:"id"`
	SubscriptionPlan string `json:"subscription_plan"`
	AuthType         string `json:"auth_type"` // "jwt" or "static"
}

// SetRequestContext stores request context in the Go context
func SetRequestContext(ctx context.Context, reqCtx *RequestContext) context.Context {
	return context.WithValue(ctx, RequestContextKey, reqCtx)
}

// GetRequestContext retrieves request context from the Go context
func GetRequestContext(ctx context.Context) *RequestContext {
	if reqCtx, ok := ctx.Value(RequestContextKey).(*RequestContext); ok {
		return reqCtx
	}
	return nil
}

// SetUserContext stores user context in the Go context
func SetUserContext(ctx context.Context, userCtx *UserContext) context.Context {
	return context.WithValue(ctx, UserContextKey, userCtx)
}

// GetUserContext retrieves user context from the Go context
func GetUserContext(ctx context.Context) *UserContext {
	if userCtx, ok := ctx.Value(UserContextKey).(*UserContext); ok {
		return userCtx
	}
	return nil
}

// CreateRequestContextFromGin creates RequestContext from Gin context
func CreateRequestContextFromGin(c *gin.Context) *RequestContext {
	return &RequestContext{
		URL:       c.Request.URL.Path,
		Method:    c.Request.Method,
		Path:      c.FullPath(),
		UserAgent: c.Request.UserAgent(),
	}
}

// CreateUserContextFromGin creates UserContext from Gin context
func CreateUserContextFromGin(c *gin.Context, authType string) *UserContext {
	userCtx := &UserContext{
		AuthType: authType,
	}

	// Handle JWT authentication
	if authType == "jwt" {
		if userID, exists := c.Get("userID"); exists {
			if id, ok := userID.(int64); ok {
				userCtx.ID = fmt.Sprintf("%d", id)
			}
		}

		if user, exists := c.Get("user"); exists {
			if u, ok := user.(*model.User); ok {
				userCtx.ID = fmt.Sprintf("%d", u.ID)
				userCtx.SubscriptionPlan = string(u.SubscriptionPlan)
			}
		}
	} else if authType == "static" {
		// Static authentication - mark as admin
		userCtx.ID = "admin"
		userCtx.SubscriptionPlan = "admin"
	}

	return userCtx
}

// SetSentryScope sets both request and user context in Sentry scope
func SetSentryScope(hub *sentry.Hub, reqCtx *RequestContext, userCtx *UserContext) {
	hub.WithScope(func(scope *sentry.Scope) {
		// Set request context
		if reqCtx != nil {
			scope.SetContext("request", map[string]interface{}{
				"url":        reqCtx.URL,
				"method":     reqCtx.Method,
				"path":       reqCtx.Path,
				"user_agent": reqCtx.UserAgent,
			})
		}

		// Set user context
		if userCtx != nil && userCtx.ID != "" {
			scope.SetUser(sentry.User{ID: userCtx.ID})

			// Set additional user tags
			scope.SetTag("auth_type", userCtx.AuthType)
			if userCtx.SubscriptionPlan != "" {
				scope.SetTag("subscription_plan", userCtx.SubscriptionPlan)
			}
		}
	})
}

// CreateContextFromGin creates a Go context with both request and user context from Gin
func CreateContextFromGin(c *gin.Context, authType string) context.Context {
	ctx := c.Request.Context()

	// Add request context
	reqCtx := CreateRequestContextFromGin(c)
	ctx = SetRequestContext(ctx, reqCtx)

	// Add user context if authentication is present
	if authType != "" {
		userCtx := CreateUserContextFromGin(c, authType)
		ctx = SetUserContext(ctx, userCtx)
	}

	return ctx
}
