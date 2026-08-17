package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

type WordInput struct {
	Name    string `json:"name" binding:"required"`
	Notes   string `json:"notes"`
	Learned bool   `json:"learned"`
}

// WordUpdateInput contains only fields a user may change. Pointer fields retain
// the difference between an omitted value and an explicit zero value.
type WordUpdateInput struct {
	Name       *string `json:"name"`
	Notes      *string `json:"notes"`
	Learned    *bool   `json:"learned"`
	WordlistID *int64  `json:"wordlistId"`
}

type WordRoutes struct {
	wordService       *service.WordService
	wordlistService   *service.WordlistService
	definitionService *service.DefinitionService
}

type Word = service.Word

const (
	maxDefinitionBatchWordIDs      = repository.MaxDefinitionBatchWordIDs
	maxDefinitionBatchIDsQuerySize = 1024
	maxDefinitionBatchCursorBytes  = 4096
	definitionContinuationHeader   = "X-Definitions-Continuation"
)

func NewWordRoutes(wordService *service.WordService, wordlistService *service.WordlistService, definitionService *service.DefinitionService) *WordRoutes {
	return &WordRoutes{
		wordService:       wordService,
		wordlistService:   wordlistService,
		definitionService: definitionService,
	}
}

func (h *WordRoutes) GetAll(c *gin.Context) {
	page, pageErr := parsePageRequest(c)
	if pageErr != nil {
		writeInvalidPage(c, pageErr)
		return
	}
	wordlistID, err := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	if err != nil || wordlistID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}
	userID := c.GetInt64("userID")
	if _, err = h.wordlistService.GetWordlistByID(c.Request.Context(), wordlistID, userID); err != nil {
		if isNotFound(err) {
			respondNotFound(c)
		} else {
			common.Logger.ErrorContext(c.Request.Context(), "failed to verify wordlist ownership", "error", err, "userID", userID, "wordlistID", wordlistID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get user words"})
		}
		return
	}

	// Parse optional query parameter for filtering words with definitions
	onlyWithDefinitions := c.Query("onlyWithDefinitions") == "true"

	words, err := h.wordService.GetWordByWordlist(c.Request.Context(), wordlistID, userID, onlyWithDefinitions, page.Limit, page.Cursor)
	if err != nil {
		common.Logger.ErrorContext(c.Request.Context(), "failed to get words", "error", err, "userID", userID, "wordlistID", wordlistID, "onlyWithDefinitions", onlyWithDefinitions)
		c.String(http.StatusInternalServerError, "Could not get user words")
		return
	}
	c.IndentedJSON(http.StatusOK, pageItems(c, page, words, func(word Word) int64 { return word.ID }))
}

func (h *WordRoutes) Create(ctx *gin.Context) {
	wordlistID, parseErr := strconv.ParseInt(ctx.Param("wordlistId"), 10, 64)
	if parseErr != nil || wordlistID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}
	var userID = ctx.GetInt64("userID")
	var input WordInput

	if err := ctx.BindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var saved, err = h.wordService.SaveWord(ctx.Request.Context(), &Word{Name: input.Name, UserID: userID, WordlistID: wordlistID, Notes: input.Notes})
	var logger = common.Logger.With("word", input.Name, "userID", userID, "endpoint", ctx.Request.URL.Path)
	if err != nil {
		var businessErr common.BusinessError
		switch {
		case isNotFound(err):
			respondNotFound(ctx)
		case errors.As(err, &businessErr):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": businessErr.Error()})
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
	wordlistID, err := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	if err != nil || wordlistID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}
	id, err := strconv.ParseInt(c.Param("wordId"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word ID"})
		return
	}

	_, err = h.wordService.DeleteWord(c.Request.Context(), id, wordlistID, userID)
	if err != nil {
		if isNotFound(err) {
			respondNotFound(c)
		} else {
			common.Logger.ErrorContext(c.Request.Context(), "failed to delete word", "error", err, "userID", userID, "wordlistID", wordlistID, "wordID", id)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete word"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *WordRoutes) Update(c *gin.Context) {
	var input WordUpdateInput

	id, err := strconv.ParseInt(c.Param("wordId"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word ID"})
		return
	}
	wordlistID, err := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	if err != nil || wordlistID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}
	userID := c.GetInt64("userID")

	if err = c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.WordlistID != nil && *input.WordlistID != wordlistID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wordlistId must match the route"})
		return
	}

	err = h.wordService.UpdateOwnedWord(c.Request.Context(), service.OwnedWordUpdate{
		ID:               id,
		OwnerID:          userID,
		TargetWordlistID: wordlistID,
		Name:             input.Name,
		Notes:            input.Notes,
		Learned:          input.Learned,
	})
	if err != nil {
		var notFoundErr common.NotFoundError
		var businessErr common.BusinessError
		switch {
		case errors.As(err, &notFoundErr):
			respondNotFound(c)
		case errors.As(err, &businessErr):
			c.JSON(http.StatusBadRequest, gin.H{"error": businessErr.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update word"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *WordRoutes) GetDefinitions(c *gin.Context) {
	page, pageErr := parsePageRequest(c)
	if pageErr != nil {
		writeInvalidPage(c, pageErr)
		return
	}
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
	word, err := h.wordService.GetOwnedWordByID(c.Request.Context(), wordID, userID)
	if err != nil || word.WordlistID != wordlistID {
		if err == nil || isNotFound(err) {
			respondNotFound(c)
		} else {
			common.Logger.ErrorContext(c.Request.Context(), "failed to verify word ownership", "error", err, "userID", userID, "wordId", wordID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get word definitions"})
		}
		return
	}

	definitions, err := h.definitionService.GetDefinitionsByWordID(c.Request.Context(), wordlistID, wordID, userID, page.Limit, page.Cursor)
	if err != nil {
		common.Logger.ErrorContext(c.Request.Context(), "failed to get definitions", "error", err, "userID", userID, "wordId", wordID)
		c.String(http.StatusInternalServerError, "Could not get word definitions")
		return
	}

	c.IndentedJSON(http.StatusOK, pageItems(c, page, definitions, func(definition *model.Definition) int64 { return definition.ID }))
}

// GetDefinitionsBatch returns definitions for multiple word IDs in one request
// GET /wordlists/:wordlistId/words/definitions?ids=1,2,3
func (h *WordRoutes) GetDefinitionsBatch(c *gin.Context) {
	userID := c.GetInt64("userID")
	wordlistID, err := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	if err != nil || wordlistID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}

	idsParam := c.Query("ids")
	if idsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing ids query parameter"})
		return
	}

	wordIDs, parseErr := parseDefinitionBatchWordIDs(idsParam)
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseErr.Error()})
		return
	}
	definitionCursors, cursorErr := parseDefinitionBatchCursors(c.Query("definitionCursors"), wordIDs)
	if cursorErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": cursorErr.Error()})
		return
	}
	if _, err = h.wordlistService.GetWordlistByID(c.Request.Context(), wordlistID, userID); err != nil {
		if isNotFound(err) {
			respondNotFound(c)
		} else {
			common.Logger.ErrorContext(c.Request.Context(), "failed to verify wordlist ownership", "error", err, "userID", userID, "wordlistID", wordlistID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get definitions"})
		}
		return
	}

	// Fetch batched definitions and names (token) scoped by wordlist and user
	page, err := h.definitionService.GetDefinitionsByWordIDs(c.Request.Context(), wordlistID, userID, wordIDs, definitionCursors)
	if err != nil {
		common.Logger.ErrorContext(c.Request.Context(), "failed to get batched definitions", "error", err, "userID", userID, "wordlistID", wordlistID, "ids", idsParam)
		c.String(http.StatusInternalServerError, "Could not get definitions")
		return
	}

	if continuation := encodeDefinitionBatchCursors(page.NextCursors); continuation != "" {
		c.Header(definitionContinuationHeader, continuation)
	}
	c.IndentedJSON(http.StatusOK, page.Results)
}

func parseDefinitionBatchWordIDs(idsParam string) ([]int64, error) {
	if len(idsParam) == 0 {
		return nil, errors.New("No valid ids provided")
	}
	if len(idsParam) > maxDefinitionBatchIDsQuerySize {
		return nil, fmt.Errorf("ids parameter is too large")
	}
	parts := strings.Split(idsParam, ",")
	if len(parts) > maxDefinitionBatchWordIDs {
		return nil, fmt.Errorf("at most %d ids are allowed", maxDefinitionBatchWordIDs)
	}
	wordIDs := make([]int64, 0, len(parts))
	seen := make(map[int64]struct{}, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			return nil, errors.New("Invalid id in ids parameter")
		}
		id, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil || id <= 0 {
			return nil, errors.New("Invalid id in ids parameter")
		}
		if _, duplicate := seen[id]; duplicate {
			return nil, errors.New("Duplicate id in ids parameter")
		}
		seen[id] = struct{}{}
		wordIDs = append(wordIDs, id)
	}
	if len(wordIDs) == 0 {
		return nil, errors.New("No valid ids provided")
	}
	return wordIDs, nil
}

func parseDefinitionBatchCursors(raw string, wordIDs []int64) (map[int64]int64, error) {
	cursors := make(map[int64]int64)
	if raw == "" {
		return cursors, nil
	}
	if len(raw) > maxDefinitionBatchCursorBytes {
		return nil, errors.New("definition cursors are too large")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || len(decoded) == 0 || len(decoded) > maxDefinitionBatchCursorBytes {
		return nil, errors.New("definition cursors are invalid")
	}
	decoder := json.NewDecoder(bytes.NewReader(decoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cursors); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return nil, errors.New("definition cursors are invalid")
	}
	requested := make(map[int64]struct{}, len(wordIDs))
	for _, wordID := range wordIDs {
		requested[wordID] = struct{}{}
	}
	if len(cursors) > len(requested) {
		return nil, errors.New("definition cursors are invalid")
	}
	for wordID, definitionID := range cursors {
		if wordID <= 0 || definitionID < 0 {
			return nil, errors.New("definition cursors are invalid")
		}
		if _, found := requested[wordID]; !found {
			return nil, errors.New("definition cursors do not match requested words")
		}
	}
	return cursors, nil
}

func encodeDefinitionBatchCursors(cursors map[int64]int64) string {
	if len(cursors) == 0 {
		return ""
	}
	encoded, err := json.Marshal(cursors)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(encoded)
}
