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

	analytics.GET("/wordlists/:id/stats", getWordlistStats)
	analytics.GET("/wordlists/:id/word-mastery", getWordlistWordMastery)
	analytics.GET("/wordlists/:id/learning-progress", getWordlistLearningProgress)
	analytics.GET("/wordlists/:id/practice-time", getWordlistPracticeTime)
	analytics.GET("/wordlists/:id/box-distribution-history", getWordlistBoxDistributionHistory)
	analytics.GET("/wordlists/:id/current-box-distribution", getWordlistCurrentBoxDistribution)
	analytics.GET("/wordlists/:id/quiz-type-performance", getWordlistQuizTypePerformance)
	analytics.GET("/progress-summary", getProgressSummary)
}

// getWordlistWordMastery returns word mastery statistics for a wordlist
func getWordlistWordMastery(c *gin.Context) {
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
	wordlist, err := service.GetWordlistById(wordlistID, userID)
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

	// Ensure stats is never nil - return empty array if no data
	if stats == nil {
		stats = []model.WordMasteryStats{}
	}

	c.JSON(http.StatusOK, gin.H{
		"wordlist_id": wordlistID,
		"stats":       stats,
	})
}

// getWordlistLearningProgress returns daily learning progress
func getWordlistLearningProgress(c *gin.Context) {
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

	userID := c.GetInt64("userID")

	// Verify wordlist ownership
	wordlist, err := service.GetWordlistById(wordlistID, userID)
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

	progress, err := analyticsService.LearningProgress(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch progress"})
		return
	}

	// Ensure progress is never nil - return empty array if no data
	if progress == nil {
		progress = []model.LearningProgressStats{}
	}

	c.JSON(http.StatusOK, gin.H{
		"wordlist_id": wordlistID,
		"days":        days,
		"progress":    progress,
	})
}

// getWordlistBoxDistributionHistory returns historical box distribution
func getWordlistBoxDistributionHistory(c *gin.Context) {
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

	userID := c.GetInt64("userID")

	// Verify wordlist ownership
	wordlist, err := service.GetWordlistById(wordlistID, userID)
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

	distribution, err := analyticsService.BoxDistributionHistory(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch distribution"})
		return
	}

	// Ensure distribution is never nil - return empty array if no data
	if distribution == nil {
		distribution = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"wordlistId":   wordlistID,
		"days":         days,
		"distribution": distribution,
	})
}

// getWordlistQuizTypePerformance returns performance statistics by quiz type for a specific wordlist
func getWordlistQuizTypePerformance(c *gin.Context) {
	wordlistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}

	userID := c.GetInt64("userID")

	// Verify wordlist ownership
	wordlist, err := service.GetWordlistById(wordlistID, userID)
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

	performance, err := analyticsService.QuizTypePerformance(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch performance stats"})
		return
	}

	// Ensure performance is never nil - return empty array if no data
	if performance == nil {
		performance = []model.QuizTypePerformance{}
	}

	c.JSON(http.StatusOK, gin.H{
		"wordlistId":      wordlistID,
		"quizPerformance": performance,
	})
}

// getWordlistCurrentBoxDistribution returns current distribution of words across Leitner boxes
func getWordlistCurrentBoxDistribution(c *gin.Context) {
	wordlistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}

	userID := c.GetInt64("userID")

	// Verify wordlist ownership
	wordlist, err := service.GetWordlistById(wordlistID, userID)
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

	distribution, err := analyticsService.CurrentBoxDistribution(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch distribution"})
		return
	}

	// Ensure distribution is never nil - return empty distribution if no data
	if distribution == nil {
		distribution = &model.BoxDistribution{
			Box1Count: 0, Box2Count: 0, Box3Count: 0, Box4Count: 0,
			Box5Count: 0, Box6Count: 0, Box7Count: 0, TotalWords: 0,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"wordlistId":   wordlistID,
		"distribution": distribution,
		"totalWords":   distribution.TotalWords,
	})
}

// getWordlistPracticeTime returns daily practice time statistics for a wordlist
func getWordlistPracticeTime(c *gin.Context) {
	wordlistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}

	// Get days parameter (default 7)
	days := 7
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 30 {
			days = parsed
		}
	}

	userID := c.GetInt64("userID")

	// Verify wordlist ownership
	wordlist, err := service.GetWordlistById(wordlistID, userID)
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

	practiceTime, err := analyticsService.PracticeTime(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch practice time"})
		return
	}

	// Ensure practiceTime is never nil - return empty array if no data
	if practiceTime == nil {
		practiceTime = []model.PracticeTimeStats{}
	}

	c.JSON(http.StatusOK, gin.H{
		"wordlistId":   wordlistID,
		"days":         days,
		"practiceTime": practiceTime,
	})
}

// getWordlistStats returns statistics for a specific wordlist
func getWordlistStats(c *gin.Context) {
	wordlistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}

	userID := c.GetInt64("userID")

	// Verify wordlist ownership
	wordlist, err := service.GetWordlistById(wordlistID, userID)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch wordlist stats"})
		return
	}

	// Ensure stats is never nil - return safe defaults if no data
	if stats == nil {
		stats = &model.WordlistStats{
			TotalWords: 0, WordsMastered: 0, AverageMastery: nil,
			BestStreak: nil, WordsStudiedToday: 0, QuizzesToday: 0,
			AccuracyToday: nil, CurrentStreak: 0,
		}
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

	// Ensure summary is never nil and has valid wordlists array
	if summary == nil {
		summary = &model.ProgressSummaryResponse{Wordlists: []model.WordlistProgress{}}
	} else if summary.Wordlists == nil {
		summary.Wordlists = []model.WordlistProgress{}
	}

	c.JSON(http.StatusOK, summary)
}
