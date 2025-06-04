package http

import (
	"net/http"
	"strconv"

	"decorebator.com/internal/common"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

// RegisterAnalyticsRoutes registers all analytics endpoints
func RegisterAnalyticsRoutes(r *gin.RouterGroup) {
	analytics := r.Group("/analytics")

	analytics.GET("/wordlists/:id/mastery", getWordMastery)
	analytics.GET("/wordlists/:id/progress", getLearningProgress)
	analytics.GET("/wordlists/:id/distribution", getBoxDistribution)
	analytics.GET("/quiz-performance", getQuizTypePerformance)
	analytics.GET("/dashboard", getDashboardStats)
}

// getWordMastery returns word mastery statistics for a wordlist
func getWordMastery(c *gin.Context) {
	wordlistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}

	userID := c.GetInt64("user_id")

	// Verify wordlist ownership
	wordlist, err := service.GetWordlistById(wordlistID, userID)
	if err != nil || wordlist == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wordlist not found"})
		return
	}

	analyticsService, err := service.NewAnalyticsService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize analytics"})
		return
	}

	stats, err := analyticsService.GetWordMastery(c.Request.Context(), userID, wordlistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch mastery stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wordlist_id": wordlistID,
		"stats":       stats,
	})
}

// getLearningProgress returns daily learning progress
func getLearningProgress(c *gin.Context) {
	wordlistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}

	// Get days parameter (default 30)
	days := 30
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}

	userID := c.GetInt64("user_id")

	// Verify wordlist ownership
	wordlist, err := service.GetWordlistById(wordlistID, userID)
	if err != nil || wordlist == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wordlist not found"})
		return
	}

	analyticsService, err := service.NewAnalyticsService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize analytics"})
		return
	}

	progress, err := analyticsService.GetLearningProgress(c.Request.Context(), userID, wordlistID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch progress"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wordlist_id": wordlistID,
		"days":        days,
		"progress":    progress,
	})
}

// getBoxDistribution returns historical box distribution
func getBoxDistribution(c *gin.Context) {
	wordlistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}

	// Get days parameter (default 30)
	days := 30
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}

	userID := c.GetInt64("user_id")

	// Verify wordlist ownership
	wordlist, err := service.GetWordlistById(wordlistID, userID)
	if err != nil || wordlist == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wordlist not found"})
		return
	}

	analyticsService, err := service.NewAnalyticsService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize analytics"})
		return
	}

	// Update current snapshot
	err = analyticsService.UpdateBoxDistribution(c.Request.Context(), userID, wordlistID)
	if err != nil {
		// Log but don't fail
		common.Logger.Error("Failed to update box distribution", "error", err)
	}

	distribution, err := analyticsService.GetBoxDistributionHistory(c.Request.Context(), userID, wordlistID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch distribution"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wordlist_id":  wordlistID,
		"days":         days,
		"distribution": distribution,
	})
}

// getQuizTypePerformance returns performance statistics by quiz type
func getQuizTypePerformance(c *gin.Context) {
	userID := c.GetInt64("user_id")

	analyticsService, err := service.NewAnalyticsService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize analytics"})
		return
	}

	performance, err := analyticsService.GetQuizTypePerformance(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch performance stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"quiz_performance": performance,
	})
}

// getDashboardStats returns overall dashboard statistics
func getDashboardStats(c *gin.Context) {
	userID := c.GetInt64("user_id")

	analyticsService, err := service.NewAnalyticsService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize analytics"})
		return
	}

	// Pass along request‐scoped context
	ctx := c.Request.Context()

	stats, err := analyticsService.GetDashboardStats(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dashboard stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}
