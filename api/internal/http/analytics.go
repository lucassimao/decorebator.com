package http

import (
	"net/http"
	"strconv"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
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

// Update the quiz handler to include analytics
func saveQuizResultWithAnalytics(c *gin.Context) {
	var req struct {
		LeitnerSystemTrackingID int64  `json:"leitner_system_tracking_id" binding:"required"`
		Success                 bool   `json:"success"`
		ResponseTimeMs          int    `json:"response_time_ms"`
		QuizType                string `json:"quiz_type" binding:"required"`
		WordID                  int64  `json:"word_id" binding:"required"`
		DefinitionID            int64  `json:"definition_id" binding:"required"`
		WordlistID              int64  `json:"wordlist_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt64("user_id")

	// Create quiz result for analytics
	quizResult := service.QuizResult{
		UserID:                  userID,
		WordlistID:              req.WordlistID,
		WordID:                  req.WordID,
		DefinitionID:            req.DefinitionID,
		LeitnerSystemTrackingID: req.LeitnerSystemTrackingID,
		QuizType:                model.QuizType(req.QuizType),
		IsCorrect:               req.Success,
		ResponseTimeMs:          req.ResponseTimeMs,
	}

	strategy := service.LeitnerSystemStrategy{}
	err := strategy.SaveQuizResultWithAnalytics(
		req.LeitnerSystemTrackingID,
		req.Success,
		quizResult,
		nil,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save quiz result"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
