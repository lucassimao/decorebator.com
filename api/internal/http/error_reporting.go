package http

import (
	"errors"
	"net/http"
	"strconv"

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

	userID := c.GetInt64("userID")

	// Call service with individual parameters
	err := h.errorReportService.ReportError(
		c.Request.Context(),
		input.ErrorType,
		input.WordID,
		input.DefinitionID,
		userID,
		input.QuizDetails,
	)

	if err != nil {
		var rateLimitErr service.RateLimitError
		if errors.As(err, &rateLimitErr) {
			writeErrorReportRateLimit(c, rateLimitErr)
			return
		}
		var quotaUnavailableErr service.ErrorReportQuotaUnavailableError
		if errors.As(err, &quotaUnavailableErr) {
			writeErrorReportQuotaUnavailable(c)
			return
		}
		// Handle cooldown errors specifically
		var cooldownErr service.CooldownError
		if errors.As(err, &cooldownErr) {
			retryAfter := service.RetryAfterSeconds(cooldownErr.RetryAfter)
			retryAfterSeconds, conversionErr := strconv.Atoi(retryAfter)
			if conversionErr != nil {
				retryAfterSeconds = 1
			}
			c.Header("Retry-After", retryAfter)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":         cooldownErr.Message,
				"cooldownUntil": cooldownErr.CooldownUntil.Unix(),
				"retryAfter":    retryAfterSeconds,
				"windowType":    "cooldown", // Indicate this is a cooldown, not a rate limit
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Please try again later."})
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}
}
