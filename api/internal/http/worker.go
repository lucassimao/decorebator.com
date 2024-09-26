package http

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"decorebator.com/internal/common"
	"decorebator.com/internal/workers"
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

	var container, ok = common.GetDigContainerFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no dig context"})
		return
	}
	var jobId int64

	err = container.Invoke(func(srv common.ContentGenerationService) error {
		jobId, err = srv.GenerateImage(definitionId, input.Prompt)
		return err
	})

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

	var container, ok = common.GetDigContainerFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no dig context"})
		return
	}

	var jobId int64

	err = container.Invoke(func(srv common.ContentGenerationService) error {
		jobId, err = srv.TextToSpeech(wordId)
		return err
	})

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

	riverClient, err := workers.GetRiverClient()
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
