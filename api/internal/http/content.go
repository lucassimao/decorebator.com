package http

import (
	"net/http"

	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

type ContentRoutes struct{}

// GetContentGuidelines returns the content moderation guidelines for users
func (h *ContentRoutes) GetContentGuidelines(c *gin.Context) {
	contentFilter := service.NewContentFilterService()
	guidelines := contentFilter.GetContentGuidelines()

	response := gin.H{
		"guidelines":  guidelines,
		"description": "Please follow these guidelines when creating wordlists and adding vocabulary words to ensure content is appropriate for all users.",
	}

	c.JSON(http.StatusOK, response)
}

// ValidateContent allows users to validate content before submitting
func (h *ContentRoutes) ValidateContent(c *gin.Context) {
	type ValidationRequest struct {
		Type     string `json:"type" binding:"required,oneof=word wordlist_name description"`
		Content  string `json:"content" binding:"required"`
		Language string `json:"language,omitempty"`
	}

	var request ValidationRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contentFilter := service.NewContentFilterService()

	// Default to English if no language provided
	if request.Language == "" {
		request.Language = "en"
	}

	var result service.ContentFilterResult

	switch request.Type {
	case "word":
		result = contentFilter.ValidateWord(request.Content)
	case "wordlist_name":
		result = contentFilter.ValidateWordlistName(request.Content)
	case "description":
		result = contentFilter.ValidateDescription(request.Content)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content type"})
		return
	}

	response := gin.H{
		"isAppropriate": result.IsAppropriate,
		"type":          request.Type,
		"content":       request.Content,
	}

	if !result.IsAppropriate {
		response["reason"] = result.Reason
		response["severity"] = result.Severity
		if len(result.FlaggedWords) > 0 {
			response["flaggedWords"] = result.FlaggedWords
		}
	}

	c.JSON(http.StatusOK, response)
}
