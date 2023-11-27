package common

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ValidateIDParamMidleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id != "" {
			if _, err := strconv.Atoi(id); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid ID. ID must be an integer.",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
