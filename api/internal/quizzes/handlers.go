package quizzes

import (
	"net/http"
	"strconv"

	sr "decorebator.com/internal/quizzes/spacedrepetition"
	"github.com/gin-gonic/gin"
)

type QuizHandlers struct{}

var Handlers = &QuizHandlers{}

var strategy sr.SpacedRepetionStrategy = sr.LeitnerSystemAlgorithm{}

func (h *QuizHandlers) Create(c *gin.Context) {

	wordlistID, _ := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	userId := c.GetInt64("userID")

	challenge, err := strategy.CreateChallenge(wordlistID, userId)

	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, challenge)
}

type SaveInput struct {
	Success      bool  `json:"success"`
	DefinitionID int64 `json:"definitionId"`
}

func (h *QuizHandlers) Save(c *gin.Context) {

	var input SaveInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var err = strategy.SaveChallengeResult(input.DefinitionID, input.Success)

	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusOK)
}
