package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PublicQuizRoutes struct {
	repo          *repository.PublicQuizRepository
	wordlistSvc   *service.WordlistService
	definitionSvc *service.DefinitionService
	db            *pgxpool.Pool
	redis         *redis.Client
	jobs          service.JobService
	quizSvc       *service.PublicQuizService
}

func NewPublicQuizRoutes(repo *repository.PublicQuizRepository, wordlistSvc *service.WordlistService, definitionSvc *service.DefinitionService, db *pgxpool.Pool, redis *redis.Client, jobs service.JobService) *PublicQuizRoutes {
	return &PublicQuizRoutes{repo: repo, wordlistSvc: wordlistSvc, definitionSvc: definitionSvc, db: db, redis: redis, jobs: jobs, quizSvc: service.NewPublicQuizService(repo, definitionSvc, wordlistSvc, jobs)}
}

// Publish creates a new public quiz for a given wordlist. Returns only the slug.
func (h *PublicQuizRoutes) Publish(c *gin.Context) {
	wordlistID, err := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wordlist id"})
		return
	}

	var input model.PublishQuizDTO
	if bindErr := c.BindJSON(&input); bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": bindErr.Error()})
		return
	}
	userID := c.GetInt64("userID")
	slug, err := h.quizSvc.PublishPublicQuiz(c.Request.Context(), wordlistID, userID, input)
	if err != nil {
		var notFound common.NotFoundError
		if errors.As(err, &notFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if be, ok := err.(common.BusinessError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": be.Error()})
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"slug": slug})
}

// GetBySlug returns public quiz metadata by slug (unauthenticated)
func (h *PublicQuizRoutes) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing slug"})
		return
	}
	quiz, err := h.quizSvc.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		var notFound common.NotFoundError
		if errors.As(err, &notFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, quiz)
}

// ListActive returns a list of active public quizzes for sitemap and discovery (unauthenticated)
// Response is intentionally lightweight
func (h *PublicQuizRoutes) ListActive(c *gin.Context) {
	// Optional limit query param; default to 5000 (sitemap chunk size)
	limit := 5000
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 10000 {
			limit = n
		}
	}

	onlyActive := true
	list, err := h.repo.Find(c.Request.Context(), repository.FindPublicQuizArgs{OnlyActive: &onlyActive, Limit: &limit})
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	type item struct {
		Slug        string  `json:"slug"`
		Title       string  `json:"title"`
		Description *string `json:"description,omitempty"`
		Preview     *string `json:"previewImageUrl,omitempty"`
		UpdatedAt   any     `json:"updatedAt,omitempty"`
		PublishedAt any     `json:"publishedAt,omitempty"`
	}
	out := make([]item, 0, len(list))
	for _, q := range list {
		var updated any
		if q.UpdatedAt.Status == 1 {
			updated = q.UpdatedAt.Time.UTC().Format(time.RFC3339)
		}
		var published any
		if q.PublishedAt.Status == 1 {
			published = q.PublishedAt.Time.UTC().Format(time.RFC3339)
		}
		out = append(out, item{
			Slug:        q.Slug,
			Title:       q.Title,
			Description: q.Description,
			Preview:     q.PreviewImageURL,
			UpdatedAt:   updated,
			PublishedAt: published,
		})
	}
	c.JSON(http.StatusOK, gin.H{"quizzes": out})
}

// GetQuestionsBySlug returns a spaced-repetition-like sequence of questions for a public quiz
// Minimal, unauthenticated: generated from the creator's wordlist definitions
func (h *PublicQuizRoutes) GetQuestionsBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing slug"})
		return
	}

	questions, err := h.quizSvc.BuildQuestionsForSlug(c.Request.Context(), slug, 100)
	if err != nil {
		var notFound common.NotFoundError
		if errors.As(err, &notFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"questions": questions})
}

// RecordAttempt records final aggregated performance for a public quiz (tries/right)
// MVP: Accepts counters and increments aggregated stats; no per-player data
type recordAttemptInput struct {
	Tried           int     `json:"tried"`
	Correct         int     `json:"correct"`
	Name            *string `json:"name,omitempty"`
	Email           *string `json:"email,omitempty"`
	DurationSeconds *int    `json:"durationSeconds,omitempty"`
}

func (h *PublicQuizRoutes) RecordAttempt(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing slug"})
		return
	}
	var input recordAttemptInput
	if err := c.BindJSON(&input); err != nil || input.Tried < 0 || input.Correct < 0 || input.Correct > input.Tried {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	var name *string
	if input.Name != nil && *input.Name != "" {
		name = input.Name
	}
	var email *string
	if input.Email != nil && *input.Email != "" {
		email = input.Email
	}
	if err := h.quizSvc.RecordAttempt(c.Request.Context(), slug, input.Tried, input.Correct, name, email, input.DurationSeconds); err != nil {
		var notFound common.NotFoundError
		if errors.As(err, &notFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetLeaderboardBySlug returns top 10 players for a public quiz ordered by score desc then time asc
func (h *PublicQuizRoutes) GetLeaderboardBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing slug"})
		return
	}
	top, err := h.quizSvc.GetLeaderboardTop(c.Request.Context(), slug, 10)
	if err != nil {
		var notFound common.NotFoundError
		if errors.As(err, &notFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"top": top})
}

// Unpublish deactivates the active public quiz for a user's wordlist (MVP)
func (h *PublicQuizRoutes) Unpublish(c *gin.Context) {
	wordlistID, err := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wordlist id"})
		return
	}

	userID := c.GetInt64("userID")
	if err := h.quizSvc.UnpublishPublicQuiz(c.Request.Context(), wordlistID, userID); err != nil {
		var notFound common.NotFoundError
		if errors.As(err, &notFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusNoContent)
}
