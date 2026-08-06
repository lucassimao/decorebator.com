package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const readinessDatabaseTimeout = time.Second

type databaseHealthChecker interface {
	Ping(context.Context) error
}

// RegisterHealthRoutes exposes public process and dependency health probes.
// Keep these routes outside authentication and ordinary request timeouts so
// orchestrators can distinguish a live process from one ready for traffic.
func RegisterHealthRoutes(router *gin.Engine, database databaseHealthChecker) {
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/readyz", func(c *gin.Context) {
		if database == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), readinessDatabaseTimeout)
		defer cancel()

		if err := database.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
}
