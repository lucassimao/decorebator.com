package http

import (
	"net/http"
	"strconv"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

type QuizRoutes struct{}

// Using the LeitnerSystemAlgorithm as the default strategy. Should be replaced by a factory method based on user preferences.
var strategy common.SpacedRepetitionStrategy = service.LeitnerSystemStrategy{}

func (h *QuizRoutes) Create(c *gin.Context) {
	wordlistID, _ := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	userId := c.GetInt64("userID")

	challenge, err := strategy.CreateQuiz(wordlistID, userId)

	if err != nil {
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

	var err = strategy.SaveQuizResult(common.QuizResult{
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

	c.Status(http.StatusNoContent)
}
