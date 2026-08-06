package http

import (
	"errors"
	"net/http"

	"decorebator.com/internal/common"
	"github.com/gin-gonic/gin"
)

func isNotFound(err error) bool {
	var notFoundErr common.NotFoundError
	return errors.As(err, &notFoundErr)
}

func respondNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
}
