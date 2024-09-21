package common

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

type ctxKey struct{}

var digContextKey = &ctxKey{}

// Middleware to add the dig container to the context
func InjectDigContainer(container *dig.Container) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store the dig container in Gin's context
		ctx := context.WithValue(c.Request.Context(), digContextKey, container)
		c.Request = c.Request.WithContext(ctx)

		// Call the next handler
		c.Next()
	}
}

// Retrieve the dig container from the Gin context
func GetDigContainerFromContext(ctx context.Context) (*dig.Container, bool) {
	container, ok := ctx.Value(digContextKey).(*dig.Container)
	return container, ok
}
