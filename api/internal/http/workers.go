package http

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"decorebator.com/internal/service"
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

	// body is optional
	if err := c.ShouldBind(&input); err != nil && err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobId, err := service.TriggerGenerateImageWorker(definitionId, input.Prompt, nil, nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "definitionId": definitionId})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": jobId})
}

func (h *WorkerRoutes) GenerateNewAudio(c *gin.Context) {
	wordId, err := strconv.ParseInt(c.Param("wordId"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word"})
		return
	}

	jobId, err := service.TriggerTextToSpeechWorker(wordId, nil, nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "wordId": wordId})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": jobId})
}

func (h *WorkerRoutes) GenerateNewDefinition(c *gin.Context) {
	wordId, err := strconv.ParseInt(c.Param("wordId"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word"})
		return
	}

	jobId, err := service.TriggerFetchDefinitionWorker(wordId, nil, nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "wordId": wordId})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": jobId})
}

func (h *WorkerRoutes) TriggerJob(c *gin.Context) {
	jobId, err := strconv.ParseInt(c.Param("jobId"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job"})
		return
	}

	riverClient, err := service.GetRiverClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "jobId": jobId})
		return
	}

	_, err = riverClient.JobRetry(context.Background(), jobId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "jobId": jobId})
		return
	}

	c.Status(http.StatusOK)
}
