package http

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/dgrijalva/jwt-go"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

func Authenticate(c *gin.Context) {
	const BearerSchema = "Bearer "
	authorization, err := c.Cookie("Authorization")

	if err == http.ErrNoCookie {
		// fallback to header
		authorization = c.GetHeader("Authorization")
	}

	if authorization == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization missing"})
		return
	}

	tokenString := strings.TrimPrefix(authorization, BearerSchema)

	if tokenString == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token not found"})
		return
	}

	token, err := jwt.ParseWithClaims(tokenString, &service.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_KEY")), nil
	})

	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token validation error"})
		return
	}

	claims, ok := token.Claims.(*service.Claims)

	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Set userID for backward compatibility
	c.Set("userID", userID)

	// Also set user object with subscription info
	user := &model.User{
		ID:               userID,
		Email:            claims.Email,
		SubscriptionPlan: claims.SubscriptionPlan,
	}
	c.Set("user", user)

	c.Next()
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
		if os.Getenv("ENV") == "common.EnvProduction" {
			c.Writer.Header().Add("Access-Control-Allow-Origin", "https://decorebator.com")
			c.Writer.Header().Add("Access-Control-Allow-Origin", "https://api.decorebator.com")
		} else {
			origin := c.Request.Header.Get("Origin")
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
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
					slog.String("url", c.Request.URL.String()),
				}

				// optionally include userID
				if val, exists := c.Get("userID"); exists {
					if userID, ok := val.(int64); ok {
						attrs = append(attrs, slog.Int64("userId", userID))
					}
				}

				if os.Getenv("ENV") == "common.EnvProduction" {
					attrs = append(attrs, slog.Any("stack_trace", stackTrace))
				} else {
					fmt.Println(stackTrace)
				}

				// Capture the error with Sentry if available
				if hub := sentrygin.GetHubFromContext(c); hub != nil {
					hub.WithScope(func(scope *sentry.Scope) {
						// Set error context
						scope.SetTag("error_type", "panic")
						scope.SetContext("request", map[string]interface{}{
							"url":    c.Request.URL.String(),
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

				common.Logger.Error("Recovered from panic", attrs...)

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
