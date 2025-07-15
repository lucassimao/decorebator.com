package http

import (
	"net/http"
	"strconv"

	"decorebator.com/internal/common"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

type WorkerRoutes struct {
	definitionService *service.DefinitionService
	jobService        service.JobService
}

func NewWorkerRoutes(definitionService *service.DefinitionService, jobService service.JobService) *WorkerRoutes {
	return &WorkerRoutes{
		definitionService: definitionService,
		jobService:        jobService,
	}
}

func (h *WorkerRoutes) GenerateNewImage(c *gin.Context) {
	definitionId, err := strconv.ParseInt(c.Param("definitionId"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid definition"})
		return
	}

	// Admin context - trigger image generation
	jobID, err := h.jobService.ScheduleImageJob(c.Request.Context(), definitionId, nil, nil, nil)

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

	// Admin context - trigger text to speech
	jobID, err := h.jobService.ScheduleAudioJob(c.Request.Context(), wordId, nil, nil, nil)

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

	// Admin context - delete existing definitions and trigger new generation
	if deleteErr := h.definitionService.DeleteWordDefinitions(c.Request.Context(), wordId, nil); deleteErr != nil {
		common.Logger.Error("failed to delete word definitions", "wordId", wordId, "error", deleteErr)
	}
	jobID, err := h.jobService.ScheduleDefinitionJob(c.Request.Context(), wordId, nil, nil, nil)

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

	err = h.jobService.RetryJob(c.Request.Context(), jobId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "jobId": jobId})
		return
	}

	c.Status(http.StatusOK)
}
