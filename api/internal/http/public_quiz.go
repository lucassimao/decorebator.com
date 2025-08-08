package http

import (
	"errors"
	"net/http"
	"strconv"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

type PublicQuizRoutes struct {
	repo        *repository.PublicQuizRepository
	wordlistSvc *service.WordlistService
}

func NewPublicQuizRoutes(repo *repository.PublicQuizRepository, wordlistSvc *service.WordlistService) *PublicQuizRoutes {
	return &PublicQuizRoutes{repo: repo, wordlistSvc: wordlistSvc}
}

// Publish creates a new public quiz for a given wordlist. Returns only the slug.
func (h *PublicQuizRoutes) Publish(c *gin.Context) {
	wordlistID, err := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wordlist id"})
		return
	}

	var input model.PublicQuizSettings
	if bindErr := c.BindJSON(&input); bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": bindErr.Error()})
		return
	}
	if validateErr := input.Validate(); validateErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validateErr.Error()})
		return
	}

	userID := c.GetInt64("userID")

	// Ensure the wordlist belongs to the user
	wl, err := h.wordlistSvc.GetWordlistByID(c.Request.Context(), wordlistID, userID)
	if err != nil {
		var notFound common.NotFoundError
		if errors.As(err, &notFound) {
			c.Status(http.StatusNotFound)
		} else {
			common.Logger.Error("failed to get wordlist for publish", "error", err, "wordlistId", wordlistID, "userID", userID)
			c.Status(http.StatusInternalServerError)
		}
		return
	}
	if wl == nil {
		c.Status(http.StatusNotFound)
		return
	}

	// Generate unique slug
	slug, err := h.repo.GenerateUniqueSlug(c.Request.Context(), input.Title)
	if err != nil {
		common.Logger.Error("failed to generate slug", "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Create public quiz
	pq := &model.PublicQuiz{
		Slug:             slug,
		WordlistID:       wordlistID,
		CreatorID:        userID,
		Title:            input.Title,
		Description:      input.Description,
		Difficulty:       input.Difficulty,
		TimeLimitMinutes: input.TimeLimitMinutes,
		IsActive:         true,
	}

	if err := h.repo.CreatePublicQuiz(c.Request.Context(), pq); err != nil {
		common.Logger.Error("failed to create public quiz", "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Return only the slug per MVP
	c.JSON(http.StatusCreated, gin.H{"slug": pq.Slug})
}

// GetBySlug returns public quiz metadata by slug (unauthenticated)
func (h *PublicQuizRoutes) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing slug"})
		return
	}

	pq, err := h.repo.GetPublicQuizBySlug(c.Request.Context(), slug)
	if err != nil {
		common.Logger.Error("failed to get public quiz by slug", "error", err, "slug", slug)
		c.Status(http.StatusInternalServerError)
		return
	}
	if pq == nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, pq)
}
