package http

import (
	"net/http"

	"decorebator.com/internal/model"
	service "decorebator.com/internal/service"

	"github.com/gin-gonic/gin"
)

type ErrorReportRoutes struct{}

type ErrorReportInput struct {
	Quiz      model.Quiz              `json:"quiz"`
	ErrorType service.ErrorReportType `json:"errorType"`
}

func (h *ErrorReportRoutes) Create(c *gin.Context) {

	var input ErrorReportInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.GetInt64("userID")
	err := service.ReportError(input.ErrorType, input.Quiz, userId, c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Please try again later."})
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}

}
