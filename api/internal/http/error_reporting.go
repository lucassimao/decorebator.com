package http

import (
	"net/http"

	service "decorebator.com/internal/service"

	"github.com/gin-gonic/gin"
)

type ErrorReportRoutes struct {
	errorReportService *service.ErrorReportService
}

func NewErrorReportRoutes(errorReportService *service.ErrorReportService) *ErrorReportRoutes {
	return &ErrorReportRoutes{
		errorReportService: errorReportService,
	}
}

type ErrorReportRequest struct {
	WordID       int64                   `json:"wordId"`
	DefinitionID *int64                  `json:"definitionId"`
	ErrorType    service.ErrorReportType `json:"errorType"`
	QuizDetails  *service.QuizDetails    `json:"quizDetails,omitempty"`
}

func (h *ErrorReportRoutes) Create(c *gin.Context) {

	var input ErrorReportRequest

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.GetInt64("userID")

	// Call service with individual parameters
	err := h.errorReportService.ReportError(
		c.Request.Context(),
		input.ErrorType,
		input.WordID,
		input.DefinitionID,
		userId,
		input.QuizDetails,
	)

	if err != nil {
		// Handle cooldown errors specifically
		if cooldownErr, ok := err.(service.CooldownError); ok {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":         cooldownErr.Message,
				"cooldownUntil": cooldownErr.CooldownUntil.Unix(),
				"retryAfter":    int(cooldownErr.RetryAfter.Seconds()),
				"windowType":    "cooldown", // Indicate this is a cooldown, not a rate limit
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Please try again later."})
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}

}
