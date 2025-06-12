package http

import (
	"net/http"
	"strconv"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

// getUserFromContext extracts the user object from the gin context
func getUserFromContext(c *gin.Context) (*model.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	userObj, ok := user.(*model.User)
	return userObj, ok
}

// setupAnalyticsHandler is a helper function to reduce duplication in analytics handlers
func setupAnalyticsHandler(c *gin.Context, defaultDays, maxDays int) (wordlistID int64, days int, analyticsService service.AnalyticsServiceInterface, ok bool) {
	var err error
	wordlistID, err = strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return 0, 0, nil, false
	}

	// Get days parameter
	days = defaultDays
	if d := c.Query("days"); d != "" {
		if parsed, parseErr := strconv.Atoi(d); parseErr == nil && parsed > 0 && parsed <= maxDays {
			days = parsed
		}
	}

	userID := c.GetInt64("userID")

	// Verify wordlist ownership
	wordlist, err := service.GetWordlistByID(wordlistID, userID)
	if err != nil || wordlist == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wordlist not found"})
		return 0, 0, nil, false
	}

	userObj, userOk := getUserFromContext(c)
	if !userOk {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return 0, 0, nil, false
	}

	analyticsService, err = service.NewAnalyticsService(service.AnalyticsConfig{
		UserID:     userID,
		WordlistID: wordlistID,
		UseCache:   true,
		CacheTTL:   getCacheTTL(userObj.SubscriptionPlan),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize analytics"})
		return 0, 0, nil, false
	}

	return wordlistID, days, analyticsService, true
}

// getCacheTTL returns the cache TTL based on user's subscription
func getCacheTTL(subscriptionPlan model.SubscriptionPlan) time.Duration {
	if subscriptionPlan == model.PlanFree {
		return 1 * time.Hour
	}
	return 1 * time.Minute
}

// RegisterAnalyticsRoutes registers all analytics endpoints
func RegisterAnalyticsRoutes(r *gin.RouterGroup) {
	analytics := r.Group("/analytics")

	analytics.GET("/wordlists/:id/mastery", getWordMastery)
	analytics.GET("/wordlists/:id/progress", getLearningProgress)
	analytics.GET("/wordlists/:id/distribution", getBoxDistributionHistory)
	analytics.GET("/wordlists/:id/current-distribution", getCurrentBoxDistribution)
	analytics.GET("/wordlists/:id/quiz-performance", getQuizTypePerformance)
	analytics.GET("/wordlists/:id/practice-time", getPracticeTime)
	analytics.GET("/wordlists/:id/overview", getWordlistOverviewStats)
	analytics.GET("/progress-summary", getProgressSummary)
}

// getWordMastery returns word mastery statistics for a wordlist
func getWordMastery(c *gin.Context) {
	wordlistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}

	userID := c.GetInt64("userID")
	userObj, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	// Verify wordlist ownership
	wordlist, err := service.GetWordlistByID(wordlistID, userID)
	if err != nil || wordlist == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wordlist not found"})
		return
	}

	analyticsService, err := service.NewAnalyticsService(service.AnalyticsConfig{
		UserID:     userID,
		WordlistID: wordlistID,
		UseCache:   true,
		CacheTTL:   getCacheTTL(userObj.SubscriptionPlan),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize analytics"})
		return
	}

	stats, err := analyticsService.WordMastery(c.Request.Context())
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
		if parsed, parseErr := strconv.Atoi(d); parseErr == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}

	userID := c.GetInt64("userID")

	// Verify wordlist ownership
	wordlist, err := service.GetWordlistByID(wordlistID, userID)
	if err != nil || wordlist == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wordlist not found"})
		return
	}

	userObj, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	analyticsService, err := service.NewAnalyticsService(service.AnalyticsConfig{
		UserID:     userID,
		WordlistID: wordlistID,
		UseCache:   true,
		CacheTTL:   getCacheTTL(userObj.SubscriptionPlan),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize analytics"})
		return
	}

	progress, err := analyticsService.Progress(c.Request.Context(), days)
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

// getBoxDistributionHistory returns historical box distribution
func getBoxDistributionHistory(c *gin.Context) {
	wordlistID, days, analyticsService, ok := setupAnalyticsHandler(c, 30, 365)
	if !ok {
		return
	}

	distribution, err := analyticsService.BoxHistory(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch distribution"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wordlistId":   wordlistID,
		"days":         days,
		"distribution": distribution,
	})
}

// getQuizTypePerformance returns performance statistics by quiz type for a specific wordlist
func getQuizTypePerformance(c *gin.Context) {
	wordlistID, _, analyticsService, ok := setupAnalyticsHandler(c, 0, 0)
	if !ok {
		return
	}

	performance, err := analyticsService.QuizPerformance(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch performance stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wordlistId":      wordlistID,
		"quizPerformance": performance,
	})
}

// getCurrentBoxDistribution returns current distribution of words across Leitner boxes
func getCurrentBoxDistribution(c *gin.Context) {
	wordlistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}

	userID := c.GetInt64("userID")

	// Verify wordlist ownership
	wordlist, err := service.GetWordlistByID(wordlistID, userID)
	if err != nil || wordlist == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wordlist not found"})
		return
	}

	userObj, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	analyticsService, err := service.NewAnalyticsService(service.AnalyticsConfig{
		UserID:     userID,
		WordlistID: wordlistID,
		UseCache:   true,
		CacheTTL:   getCacheTTL(userObj.SubscriptionPlan),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize analytics"})
		return
	}

	distribution, err := analyticsService.BoxDistribution(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch distribution"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wordlistId":   wordlistID,
		"distribution": distribution,
		"totalWords":   distribution.TotalWords,
	})
}

// getPracticeTime returns daily practice time statistics for a wordlist
func getPracticeTime(c *gin.Context) {
	wordlistID, days, analyticsService, ok := setupAnalyticsHandler(c, 7, 30)
	if !ok {
		return
	}

	practiceTime, err := analyticsService.PracticeTime(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch practice time"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wordlistId":   wordlistID,
		"days":         days,
		"practiceTime": practiceTime,
	})
}

// getWordlistOverviewStats returns dashboard statistics for a specific wordlist
func getWordlistOverviewStats(c *gin.Context) {
	wordlistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}

	userID := c.GetInt64("userID")

	// Verify wordlist ownership
	wordlist, err := service.GetWordlistByID(wordlistID, userID)
	if err != nil || wordlist == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wordlist not found"})
		return
	}

	userObj, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	analyticsService, err := service.NewAnalyticsService(service.AnalyticsConfig{
		UserID:     userID,
		WordlistID: wordlistID,
		UseCache:   true,
		CacheTTL:   getCacheTTL(userObj.SubscriptionPlan),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize analytics"})
		return
	}

	stats, err := analyticsService.Stats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch wordlist dashboard stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wordlistId": wordlistID,
		"stats":      stats,
	})
}

// getProgressSummary returns progress summary for all user's wordlists
func getProgressSummary(c *gin.Context) {
	userID := c.GetInt64("userID")

	userObj, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	analyticsService, err := service.NewAnalyticsService(service.AnalyticsConfig{
		UserID:     userID,
		WordlistID: 0, // Not needed for progress summary
		UseCache:   true,
		CacheTTL:   getCacheTTL(userObj.SubscriptionPlan),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize analytics"})
		return
	}

	summary, err := analyticsService.ProgressSummary(c.Request.Context())
	if err != nil {
		common.Logger.Error("Failed to fetch progress summary", "error", err, "userID", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch progress summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}
