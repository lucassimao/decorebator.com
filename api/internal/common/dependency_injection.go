package common

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

type ctxKey struct{}

var digContextKey = &ctxKey{}

// Gin Middleware to add the dig container in the http request
func InjectDigContainer(container *dig.Container) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store the dig container in Gin's request context
		ctx := SetDigContainerInContext(c.Request.Context(), container)
		c.Request = c.Request.WithContext(ctx)

		// Call the next handler
		c.Next()
	}
}

// stores a dig container instance in any context
func SetDigContainerInContext(ctx context.Context, container *dig.Container) context.Context {
	return context.WithValue(ctx, digContextKey, container)
}

func GetDigContainerFromContext(ctx context.Context) (*dig.Container, bool) {
	container, ok := ctx.Value(digContextKey).(*dig.Container)
	return container, ok
}
