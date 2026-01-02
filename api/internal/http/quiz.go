package http

import (
	"errors"
	"net/http"
	"strconv"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/gin-gonic/gin"
)

type QuizRoutes struct {
	strategy common.SpacedRepetitionStrategy
	userRepo *repository.UserRepository
}

// NewQuizRoutes creates a new QuizRoutes instance with dependency injection
func NewQuizRoutes(strategy common.SpacedRepetitionStrategy, userRepo *repository.UserRepository) *QuizRoutes {
	return &QuizRoutes{strategy: strategy, userRepo: userRepo}
}

func (h *QuizRoutes) Create(c *gin.Context) {

	wordlistID, _ := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	userId := c.GetInt64("userID")

	challenge, err := h.strategy.CreateQuiz(c.Request.Context(), wordlistID, userId)

	if err != nil {
		var quizUnavailable common.QuizUnavailableError
		if errors.As(err, &quizUnavailable) {
			c.JSON(http.StatusConflict, gin.H{
				"error":  quizUnavailable.Message,
				"status": quizUnavailable.Reason,
			})
			return
		}

		common.Logger.Error("failed to create quiz", "error", err, "wordlistID", wordlistID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, challenge)
}

type SaveInput struct {
	WordlistID              int64  `json:"wordlistID" binding:"required"`
	WordID                  int64  `json:"wordID" binding:"required"`
	DefinitionID            int64  `json:"definitionID" binding:"required"`
	LeitnerSystemTrackingID int64  `json:"leitnerSystemTrackingID" binding:"required"`
	QuizType                string `json:"quizType" binding:"required"`
	IsCorrect               bool   `json:"isCorrect"`
	ResponseTimeMs          int    `json:"responseTimeMs" binding:"required"`
}

// Save if the users answered correctly or not
func (h *QuizRoutes) Save(c *gin.Context) {

	var input SaveInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.GetInt64("userID")

	// Get user from context to check if premium
	isPremium := false
	if user, exists := c.Get("user"); exists {
		if userObj, ok := user.(*model.User); ok {
			isPremium = userObj.SubscriptionPlan != model.PlanFree
		}
	}

	var err = h.strategy.SaveQuizResult(c.Request.Context(), common.QuizResult{
		WordlistID:              input.WordlistID,
		WordID:                  input.WordID,
		DefinitionID:            input.DefinitionID,
		LeitnerSystemTrackingID: input.LeitnerSystemTrackingID,
		QuizType:                model.QuizType(input.QuizType),
		IsCorrect:               input.IsCorrect,
		UserID:                  userId,
		ResponseTimeMs:          input.ResponseTimeMs,
	}, isPremium, nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if h.userRepo != nil {
		if updateErr := h.userRepo.UpdateLastPracticeAt(c.Request.Context(), userId); updateErr != nil {
			common.Logger.Error("failed to update last practice timestamp", "error", updateErr, "userID", userId)
		}
	}

	c.Status(http.StatusNoContent)
}
