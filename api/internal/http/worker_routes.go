package http

import (
	"net/http"
	"strconv"

	"decorebator.com/internal/api"
	"github.com/gin-gonic/gin"
)

type GenerateNewImageInput struct {
	Prompt string `json:"prompt"`
}

type WorkerRoutes struct{}

func (h *WorkerRoutes) GenerateNewImage(c *gin.Context) {
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

	jobId, err := api.TriggerImageGenerator(definitionId, input.Prompt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "definitionId": definitionId})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": jobId})
}
