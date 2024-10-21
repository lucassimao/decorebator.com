package http

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"decorebator.com/internal/common"
	"github.com/dgrijalva/jwt-go"
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

	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_KEY")), nil
	})

	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token vaidation error"})
		return
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)

	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.Set("userID", userID)

	c.Next()
}

func AuthenticateStatic(c *gin.Context) {

	authorization := c.GetHeader("Authorization")

	if authorization == "" || authorization != common.Env.StaticAuthentication {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Wrong credentials"})
		return
	}

	c.Next()
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var origin string

		if os.Getenv("ENV") == "production" {
			origin = "https://decorebator.com"
		} else {
			origin = c.Request.Header.Get("Origin")
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
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
