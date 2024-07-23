package workers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GenerateNewImageInput struct {
	Prompt string `json:"prompt"`
}

type WorkerHandlers struct{}

var Handlers = &WorkerHandlers{}

func (h *WorkerHandlers) GenerateNewImage(c *gin.Context) {
	definitionId, err := strconv.ParseInt(c.Param("definitionId"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid definition"})
		return
	}

	var input GenerateNewImageInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobId, err := TriggerImageGenerator(definitionId, input.Prompt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "definitionId": definitionId})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": jobId})
}
