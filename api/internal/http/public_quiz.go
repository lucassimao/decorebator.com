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

	// If a quiz exists for this wordlist/user, reactivate and update it (slug stays immutable). Otherwise create.
	limit := 1
	existingList, findErr := h.repo.Find(c.Request.Context(), repository.FindPublicQuizArgs{
		WordlistID: &wordlistID,
		CreatorID:  &userID,
		Limit:      &limit,
	})
	if findErr != nil {
		common.Logger.Error("failed to check existing public quiz", "error", findErr)
		c.Status(http.StatusInternalServerError)
		return
	}
	if len(existingList) > 0 {
		existing := existingList[0]
		var desc string
		if input.Description != nil {
			desc = *input.Description
		}
		if updateErr := h.repo.ReactivateAndUpdatePublicQuiz(c.Request.Context(), existing.ID, input.Title, desc, input.Difficulty, input.TimeLimitMinutes); updateErr != nil {
			common.Logger.Error("failed to reactivate/update public quiz", "error", updateErr)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"slug": existing.Slug})
		return
	}

	// Create brand new with generated slug
	slug, err := h.repo.GenerateUniqueSlug(c.Request.Context(), input.Title)
	if err != nil {
		common.Logger.Error("failed to generate slug", "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}
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
	c.JSON(http.StatusCreated, gin.H{"slug": pq.Slug})
}

// GetBySlug returns public quiz metadata by slug (unauthenticated)
func (h *PublicQuizRoutes) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing slug"})
		return
	}

	onlyActive := true
	limit := 1
	res, err := h.repo.Find(c.Request.Context(), repository.FindPublicQuizArgs{Slug: &slug, OnlyActive: &onlyActive, Limit: &limit})
	if err != nil {
		common.Logger.Error("failed to get public quiz by slug", "error", err, "slug", slug)
		c.Status(http.StatusInternalServerError)
		return
	}
	if len(res) == 0 {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, res[0])
}

// Unpublish deactivates the active public quiz for a user's wordlist (MVP)
func (h *PublicQuizRoutes) Unpublish(c *gin.Context) {
	wordlistID, err := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wordlist id"})
		return
	}

	userID := c.GetInt64("userID")

	// Verify ownership
	wl, err := h.wordlistSvc.GetWordlistByID(c.Request.Context(), wordlistID, userID)
	if err != nil {
		var notFound common.NotFoundError
		if errors.As(err, &notFound) {
			c.Status(http.StatusNotFound)
		} else {
			common.Logger.Error("failed to get wordlist for unpublish", "error", err, "wordlistId", wordlistID, "userID", userID)
			c.Status(http.StatusInternalServerError)
		}
		return
	}
	if wl == nil {
		c.Status(http.StatusNotFound)
		return
	}

	onlyActive := true
	limit := 1
	list, err := h.repo.Find(c.Request.Context(), repository.FindPublicQuizArgs{WordlistID: &wordlistID, CreatorID: &userID, OnlyActive: &onlyActive, Limit: &limit})
	if err != nil {
		common.Logger.Error("failed to get active public quiz for owner", "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if len(list) == 0 {
		// Nothing to unpublish
		c.Status(http.StatusNoContent)
		return
	}
	pq := list[0]

	if err := h.repo.DeactivatePublicQuiz(c.Request.Context(), pq.ID); err != nil {
		common.Logger.Error("failed to deactivate public quiz", "error", err, "publicQuizId", pq.ID)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}
