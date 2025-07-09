package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/pgtype"
)

type WordlistInput struct {
	Name                string                     `json:"name" binding:"required"`
	Description         string                     `json:"description"`
	LanguageCode        string                     `json:"languageCode" binding:"required"`
	PronunciationSystem *model.PronunciationSystem `json:"pronunciationSystem,omitempty"`
}

type WordlistsRoutes struct {
	wordlistService *service.WordlistService
	wordService     *service.WordService
}
type Wordlist = service.Wordlist

func NewWordlistsRoutes(wordlistService *service.WordlistService, wordService *service.WordService) *WordlistsRoutes {
	return &WordlistsRoutes{
		wordlistService: wordlistService,
		wordService:     wordService,
	}
}

func (h *WordlistsRoutes) GetAll(c *gin.Context) {
	var userID int64 = c.GetInt64("userID")
	wordlists, err := h.wordlistService.GetUserWordlistsWithWordStats(userID)
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, wordlists)
}

func (h *WordlistsRoutes) Create(c *gin.Context) {
	var input WordlistInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determine pronunciation system
	var pronunciationSystem model.PronunciationSystem
	if input.PronunciationSystem != nil {
		// Validate that the pronunciation system is supported for this language
		supportedSystems := model.GetSupportedPronunciationSystems(input.LanguageCode)
		isSupported := false
		for _, system := range supportedSystems {
			if system == *input.PronunciationSystem {
				isSupported = true
				break
			}
		}
		if !isSupported {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Pronunciation system not supported for this language"})
			return
		}
		pronunciationSystem = *input.PronunciationSystem
	} else {
		// Use default pronunciation system for the language
		pronunciationSystem = model.GetDefaultPronunciationSystem(input.LanguageCode)
	}

	var userID int64 = c.GetInt64("userID")
	saved, err := h.wordlistService.SaveWordlist(&Wordlist{
		Name:                input.Name,
		Description:         input.Description,
		UserID:              userID,
		LanguageCode:        input.LanguageCode,
		PronunciationSystem: pronunciationSystem,
	})

	if err != nil {
		switch err.(type) {
		case common.BusinessError:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			common.Logger.Error("failed to create wordlist", "error", err, "userID", userID)
			c.Status(http.StatusInternalServerError)
		}
	} else {
		c.JSON(http.StatusCreated, saved)
	}
}

// GetPronunciationSystems returns the supported pronunciation systems for a language
func (h *WordlistsRoutes) GetPronunciationSystems(c *gin.Context) {
	languageCode := c.Query("languageCode")
	if languageCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "languageCode query parameter is required"})
		return
	}

	supportedSystems := model.GetSupportedPronunciationSystems(languageCode)
	defaultSystem := model.GetDefaultPronunciationSystem(languageCode)
	canChange := model.CanChangePronunciationSystem(languageCode)

	c.JSON(http.StatusOK, gin.H{
		"supportedSystems": supportedSystems,
		"defaultSystem":    defaultSystem,
		"canChange":        canChange,
	})
}

func (h *WordlistsRoutes) GetById(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	var userID int64 = c.GetInt64("userID")

	wordlist, err := h.wordlistService.GetWordlistByID(id, userID)
	if err != nil {
		if errors.Is(err, &common.NotFoundError{}) {
			c.Status(http.StatusNotFound)
		} else {
			panic(err)
		}
		return
	}
	c.JSON(http.StatusOK, wordlist)
}

func (h *WordlistsRoutes) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	var userID int64 = c.GetInt64("userID")

	_, err := h.wordlistService.DeleteWordlist(id, userID)
	if err != nil {
		if errors.Is(err, &common.NotFoundError{}) {
			c.Status(http.StatusNotFound)
		} else {
			panic(err)
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *WordlistsRoutes) Update(c *gin.Context) {
	var input WordlistInput

	id, _ := strconv.ParseInt(c.Param("wordlistId"), 10, 64)

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID int64 = c.GetInt64("userID")
	err := h.wordlistService.UpdateWordlist(&Wordlist{ID: id, Name: input.Name, Description: input.Description, LanguageCode: input.LanguageCode, UserID: userID})
	if err != nil {
		if errors.Is(err, common.NotFoundError{}) {
			c.Status(http.StatusNotFound)
		} else {
			switch err.(type) {
			case common.BusinessError:
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			default:
				common.Logger.Error("failed to update wordlist", "error", err, "userID", userID)
				c.Status(http.StatusInternalServerError)
			}
		}
		return
	}
	c.Status(http.StatusNoContent)
}

// GetProcessingStatus returns the processing status of words in a wordlist
func (h *WordlistsRoutes) GetProcessingStatus(c *gin.Context) {
	wordlistIDStr := c.Param("wordlistId")
	wordlistID, err := strconv.ParseInt(wordlistIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}
	userID := c.GetInt64("userID")

	// First check if the wordlist belongs to the user
	wordlist, err := h.wordlistService.GetWordlistByID(wordlistID, userID)
	if err != nil {
		var notFoundErr common.NotFoundError
		if errors.As(err, &notFoundErr) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Wordlist not found"})
		} else {
			common.Logger.Error("failed to get wordlist", "wordlistId", wordlistID, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get wordlist"})
		}
		return
	}

	// Get words from wordlist using word service
	words, err := h.wordService.GetWordByWordlist(wordlist.ID, userID, false)

	if err != nil {
		common.Logger.Error("failed to get words processing status", "wordlistId", wordlistID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get processing status"})
		return
	}

	// Build response with only relevant processing info
	type ProcessingInfo struct {
		ID                    int64  `json:"id"`
		Name                  string `json:"name"`
		ProcessingStatus      string `json:"processingStatus"`
		ProcessingError       string `json:"processingError,omitempty"`
		ProcessingStartedAt   string `json:"processingStartedAt,omitempty"`
		ProcessingCompletedAt string `json:"processingCompletedAt,omitempty"`
	}

	var processingInfos = []ProcessingInfo{}

	for _, word := range words {
		info := ProcessingInfo{
			ID:               word.ID,
			Name:             word.Name,
			ProcessingStatus: word.ProcessingStatus,
		}

		if word.ProcessingError != "" {
			info.ProcessingError = word.ProcessingError
		}

		// Format timestamps
		if word.ProcessingStartedAt.Status == pgtype.Present {
			info.ProcessingStartedAt = word.ProcessingStartedAt.Time.Format(time.RFC3339)
		}
		if word.ProcessingCompletedAt.Status == pgtype.Present {
			info.ProcessingCompletedAt = word.ProcessingCompletedAt.Time.Format(time.RFC3339)
		}

		processingInfos = append(processingInfos, info)
	}

	c.JSON(http.StatusOK, gin.H{
		"words": processingInfos,
		"summary": gin.H{
			"total":      len(words),
			"pending":    countByStatus(words, "pending"),
			"processing": countByStatus(words, "processing"),
			"completed":  countByStatus(words, "completed"),
			"failed":     countByStatus(words, "failed"),
		},
	})
}

func countByStatus(words []model.Word, status string) int {
	count := 0
	for _, w := range words {
		if w.ProcessingStatus == status {
			count++
		}
	}
	return count
}
