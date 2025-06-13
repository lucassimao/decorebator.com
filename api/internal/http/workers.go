package http

import (
	"context"
	"net/http"
	"strconv"

	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

type WorkerRoutes struct{}

func (h *WorkerRoutes) GenerateNewImage(c *gin.Context) {
	definitionId, err := strconv.ParseInt(c.Param("definitionId"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid definition"})
		return
	}

	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	jobID, err := service.TriggerGenerateImageWorker(definitionId, userID, nil, nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "definitionId": definitionId})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": jobID})
}

func (h *WorkerRoutes) GenerateNewAudio(c *gin.Context) {
	wordId, err := strconv.ParseInt(c.Param("wordId"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word"})
		return
	}

	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	jobID, err := service.TriggerTextToSpeechWorker(wordId, userID, nil, nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "wordId": wordId})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": jobID})
}

func (h *WorkerRoutes) GenerateNewDefinition(c *gin.Context) {
	wordId, err := strconv.ParseInt(c.Param("wordId"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word"})
		return
	}

	// Admin context - pass nil userId to bypass validation
	service.DeleteWordDefinitions(wordId, nil)
	jobID, err := service.TriggerFetchDefinitionWorker(wordId, nil, nil, nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "wordId": wordId})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": jobID})
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
