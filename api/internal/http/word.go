package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

type WordInput struct {
	Name    string `json:"name" binding:"required"`
	Notes   string `json:"notes"`
	Learned bool   `json:"learned"`
}

type WordRoutes struct {
	wordService       *service.WordService
	definitionService *service.DefinitionService
}

type Word = service.Word

func NewWordRoutes(wordService *service.WordService, definitionService *service.DefinitionService) *WordRoutes {
	return &WordRoutes{
		wordService:       wordService,
		definitionService: definitionService,
	}
}

func (h *WordRoutes) GetAll(c *gin.Context) {
	wordlistID, _ := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	userID := c.GetInt64("userID")

	// Parse optional query parameter for filtering words with definitions
	onlyWithDefinitions := c.Query("onlyWithDefinitions") == "true"

	words, err := h.wordService.GetWordByWordlist(c.Request.Context(), wordlistID, userID, onlyWithDefinitions)
	if err != nil {
		common.Logger.ErrorContext(c.Request.Context(), "failed to get words", "error", err, "userID", userID, "wordlistID", wordlistID, "onlyWithDefinitions", onlyWithDefinitions)
		c.String(http.StatusInternalServerError, "Could not get user words")
		return
	}
	c.IndentedJSON(http.StatusOK, words)
}

func (h *WordRoutes) Create(ctx *gin.Context) {
	var wordlistID, _ = strconv.ParseInt(ctx.Param("wordlistId"), 10, 64)
	var userID = ctx.GetInt64("userID")
	var input WordInput

	if err := ctx.BindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var saved, err = h.wordService.SaveWord(ctx.Request.Context(), &Word{Name: input.Name, UserID: userID, WordlistID: wordlistID, Notes: input.Notes})
	var logger = common.Logger.With("word", input.Name, "userID", userID, "endpoint", ctx.Request.URL.Path)
	if err != nil {
		switch err.(type) {
		case common.BusinessError:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			logger.ErrorContext(ctx.Request.Context(), "failed to create word", "error", err)
			ctx.Status(http.StatusInternalServerError)
		}
	} else {
		ctx.JSON(http.StatusCreated, saved)
	}
}

func (h *WordRoutes) Delete(c *gin.Context) {
	userID := c.GetInt64("userID")
	id, _ := strconv.ParseInt(c.Param("wordId"), 10, 64)

	_, err := h.wordService.DeleteWord(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, &common.NotFoundError{}) {
			c.String(http.StatusNotFound, err.Error())
		} else {
			c.String(http.StatusInternalServerError, "Couldn't delete wordlist #%d", id)
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *WordRoutes) Update(c *gin.Context) {
	var input WordInput

	id, _ := strconv.ParseInt(c.Param("wordId"), 10, 64)
	wordlistID, _ := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	userID := c.GetInt64("userID")

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.wordService.UpdateWord(c.Request.Context(), &Word{ID: id, Name: input.Name, UserID: userID, Learned: input.Learned, WordlistID: wordlistID}, nil)
	if err != nil {
		if errors.Is(err, common.NotFoundError{}) {
			c.String(http.StatusNotFound, err.Error())
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update word"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *WordRoutes) GetDefinitions(c *gin.Context) {
	userID := c.GetInt64("userID")
	wordlistID, err := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	if err != nil || wordlistID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}

	wordID, err := strconv.ParseInt(c.Param("wordId"), 10, 64)
	if err != nil || wordID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word ID"})
		return
	}

	definitions, err := h.definitionService.GetDefinitionsByWordID(c.Request.Context(), wordlistID, wordID, userID)
	if err != nil {
		common.Logger.ErrorContext(c.Request.Context(), "failed to get definitions", "error", err, "userID", userID, "wordId", wordID)
		c.String(http.StatusInternalServerError, "Could not get word definitions")
		return
	}

	c.IndentedJSON(http.StatusOK, definitions)
}

// GetDefinitionsBatch returns definitions for multiple word IDs in one request
// GET /wordlists/:wordlistId/words/definitions?ids=1,2,3
func (h *WordRoutes) GetDefinitionsBatch(c *gin.Context) {
	userID := c.GetInt64("userID")
	wordlistID, err := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}

	idsParam := c.Query("ids")
	if idsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing ids query parameter"})
		return
	}

	// Parse and validate IDs
	parts := strings.Split(idsParam, ",")
	wordIDs := make([]int64, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		id, convErr := strconv.ParseInt(trimmed, 10, 64)
		if convErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id in ids parameter"})
			return
		}
		wordIDs = append(wordIDs, id)
	}

	if len(wordIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid ids provided"})
		return
	}

	// Fetch batched definitions and names (token) scoped by wordlist and user
	results, err := h.definitionService.GetDefinitionsByWordIDs(c.Request.Context(), wordlistID, userID, wordIDs)
	if err != nil {
		common.Logger.ErrorContext(c.Request.Context(), "failed to get batched definitions", "error", err, "userID", userID, "wordlistID", wordlistID, "ids", idsParam)
		c.String(http.StatusInternalServerError, "Could not get definitions")
		return
	}

	c.IndentedJSON(http.StatusOK, results)
}
