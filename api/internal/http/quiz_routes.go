package http

import (
	"net/http"
	"strconv"

	"decorebator.com/internal/api"
	"decorebator.com/internal/common"
	"github.com/gin-gonic/gin"
)

type QuizRoutes struct{}

// Using the LeitnerSystemAlgorithm as the default strategy. Should be replaced by a factory method based on user preferences.
var strategy api.SpacedRepetitionStrategy = api.LeitnerSystemStrategy{}

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
	Success bool  `json:"success"`
	Id      int64 `json:"id"`
}

func (h *QuizRoutes) Save(c *gin.Context) {

	var input SaveInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var err = strategy.SaveQuizResult(input.Id, input.Success)

	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusOK)
}
