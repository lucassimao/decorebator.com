package http

import (
	"net/http"

	"decorebator.com/internal/api"
	"decorebator.com/internal/model"

	"github.com/gin-gonic/gin"
)

type ErrorReportRoutes struct{}

type ErrorReportInput struct {
	Quiz      model.Quiz    `json:"quiz"`
	ErrorType api.ErrorType `json:"errorType"`
}

func (h *ErrorReportRoutes) Create(c *gin.Context) {

	var input ErrorReportInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.GetInt64("userID")
	err := api.SaveErrorReport(input.ErrorType, input.Quiz, userId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Please try again later."})
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}

}
