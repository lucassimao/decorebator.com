package http

import (
	"errors"
	"net/http"
	"strconv"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

type WordlistInput struct {
	Name                string                     `json:"name" binding:"required"`
	Description         string                     `json:"description"`
	LanguageCode        string                     `json:"languageCode" binding:"required"`
	PronunciationSystem *model.PronunciationSystem `json:"pronunciationSystem,omitempty"`
}

type WordlistsRoutes struct{}
type Wordlist = service.Wordlist

func (h *WordlistsRoutes) GetAll(c *gin.Context) {
	var userId int64 = c.GetInt64("userID")
	wordlists, err := service.GetUserWordlistsWithWordStats(userId)
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

	var userId int64 = c.GetInt64("userID")
	saved, err := service.SaveWordlist(&Wordlist{
		Name:                input.Name,
		Description:         input.Description,
		UserID:              userId,
		LanguageCode:        input.LanguageCode,
		PronunciationSystem: pronunciationSystem,
	})

	if err != nil {
		switch err.(type) {
		case common.BusinessError:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			common.Logger.Error("failed to create wordlist", "error", err, "userId", userId)
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

	var userId int64 = c.GetInt64("userID")

	wordlist, err := service.GetWordlistById(id, userId)
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
	var userId int64 = c.GetInt64("userID")

	_, err := service.DeleteWordlist(id, userId)
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

	var userId int64 = c.GetInt64("userID")
	err := service.UpdateWordlist(&Wordlist{ID: id, Name: input.Name, Description: input.Description, LanguageCode: input.LanguageCode, UserID: userId})
	if err != nil {
		if errors.Is(err, common.NotFoundError{}) {
			c.Status(http.StatusNotFound)
		} else {
			switch err.(type) {
			case common.BusinessError:
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			default:
				common.Logger.Error("failed to update wordlist", "error", err, "userId", userId)
				c.Status(http.StatusInternalServerError)
			}
		}
		return
	}
	c.Status(http.StatusNoContent)
}
