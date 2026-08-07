package http

import (
	"net/http"
	"strconv"

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
	definitionID, err := strconv.ParseInt(c.Param("definitionId"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid definition"})
		return
	}

	// Admin context - trigger image generation
	jobID, err := h.jobService.ScheduleImageJob(c.Request.Context(), definitionID, nil, nil, nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "definitionId": definitionID})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": jobID})
}

func (h *WorkerRoutes) GenerateNewAudio(c *gin.Context) {
	wordID, err := strconv.ParseInt(c.Param("wordId"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word"})
		return
	}

	// Admin context - trigger text to speech
	jobID, err := h.jobService.ScheduleAudioJob(c.Request.Context(), wordID, nil, nil, nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "wordId": wordID})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": jobID})
}

func (h *WorkerRoutes) GenerateNewDefinition(c *gin.Context) {
	wordID, err := strconv.ParseInt(c.Param("wordId"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word"})
		return
	}

	// Admin context - delete existing definitions and trigger new generation as
	// one transaction so either both durable changes commit or neither does.
	jobID, err := h.definitionService.ScheduleDefinitionRegeneration(c.Request.Context(), wordID, h.jobService)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "wordId": wordID})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": jobID})
}

func (h *WorkerRoutes) TriggerJob(c *gin.Context) {
	jobID, err := strconv.ParseInt(c.Param("jobId"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job"})
		return
	}

	err = h.jobService.RetryJob(c.Request.Context(), jobID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "jobId": jobID})
		return
	}

	c.Status(http.StatusOK)
}
