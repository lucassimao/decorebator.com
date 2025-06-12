package service

import (
	"context"
	"fmt"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
)

// Removed unused singleton pattern variables

type AnalyticsService struct {
	repo       *repository.AnalyticsRepository
	userID     int64
	wordlistID int64
}

// AnalyticsServiceInterface defines the methods that both regular and cached analytics services must implement
type AnalyticsServiceInterface interface {
	Stats(ctx context.Context) (*model.WordlistStats, error)
	ProgressSummary(ctx context.Context) (*model.ProgressSummaryResponse, error)
	WordMastery(ctx context.Context) ([]model.WordMasteryStats, error)
	Progress(ctx context.Context, days int) ([]model.LearningProgressStats, error)
	QuizPerformance(ctx context.Context) ([]model.QuizTypePerformance, error)
	BoxDistribution(ctx context.Context) (*model.BoxDistribution, error)
	BoxHistory(ctx context.Context, days int) ([]map[string]interface{}, error)
	PracticeTime(ctx context.Context, days int) ([]model.PracticeTimeStats, error)
	TrackQuiz(ctx context.Context, result QuizResult, tx pgx.Tx) error
	UpdateBoxDistribution(ctx context.Context, userID, wordlistID int64) error
}

// AnalyticsConfig holds configuration for creating analytics service
type AnalyticsConfig struct {
	UserID     int64
	WordlistID int64
	UseCache   bool
	CacheTTL   time.Duration
}

// NewAnalyticsService creates an analytics service instance with the given configuration
func NewAnalyticsService(cfg AnalyticsConfig) (AnalyticsServiceInterface, error) {
	db, err := common.GetDBConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	repo := repository.NewAnalyticsRepository(db)
	baseService := &AnalyticsService{
		repo:       repo,
		userID:     cfg.UserID,
		wordlistID: cfg.WordlistID,
	}

	// If caching is disabled, return base service
	if !cfg.UseCache {
		return baseService, nil
	}

	// Try to get Redis client for caching
	cache, err := common.GetRedisClient()
	if err != nil {
		// Redis is optional - log error but continue with non-cached service
		common.Logger.Warn("Redis not available, analytics will use database only", "error", err)
		return baseService, nil
	}

	// Use cached service with TTL
	return NewCachedAnalyticsService(db, cache, cfg)
}

// QuizResult contains the data needed to track quiz performance
type QuizResult = common.QuizResult

// TrackQuizPerformance records the result of a quiz attempt
func (as *AnalyticsService) TrackQuiz(ctx context.Context, result QuizResult, tx pgx.Tx) error {
	// 1. Record individual quiz performance

	err := as.repo.RecordQuizPerformance(ctx, tx,
		result.UserID, result.WordlistID, result.WordID, result.DefinitionID,
		result.LeitnerSystemTrackingID, string(result.QuizType), result.BoxID,
		result.IsCorrect, result.ResponseTimeMs,
	)

	if err != nil {
		return fmt.Errorf("failed to record quiz performance: %w", err)
	}

	// 2. Update word mastery
	err = as.repo.UpsertWordMastery(ctx, tx, result.UserID, result.WordID, result.BoxID, result.IsCorrect)
	if err != nil {
		return fmt.Errorf("failed to update word mastery: %w", err)
	}

	// 3. Update daily learning progress
	// Use UTC to ensure consistent date handling across timezones
	today := time.Now().UTC().Format("2006-01-02")
	err = as.repo.UpsertLearningProgress(ctx, tx, result.UserID, result.WordlistID, today, result.IsCorrect, result.ResponseTimeMs)
	if err != nil {
		return fmt.Errorf("failed to update learning progress: %w", err)
	}

	// 4. Update quiz type analytics
	err = as.repo.UpsertQuizTypeAnalytics(ctx, tx, result.UserID, string(result.QuizType), result.IsCorrect, result.ResponseTimeMs)
	if err != nil {
		return fmt.Errorf("failed to update quiz type analytics: %w", err)
	}

	return nil
}

// GetWordMastery retrieves mastery stats for all words in a wordlist
func (as *AnalyticsService) WordMastery(ctx context.Context) ([]model.WordMasteryStats, error) {
	return as.repo.GetWordMastery(ctx, as.userID, as.wordlistID)
}

// GetQuizTypePerformance retrieves performance stats by quiz type for a specific wordlist
func (as *AnalyticsService) QuizPerformance(ctx context.Context) ([]model.QuizTypePerformance, error) {
	return as.repo.GetQuizTypePerformance(ctx, as.userID, as.wordlistID)
}

// GetLearningProgress retrieves daily learning progress
func (as *AnalyticsService) Progress(ctx context.Context, days int) ([]model.LearningProgressStats, error) {
	return as.repo.GetLearningProgress(ctx, as.userID, as.wordlistID, days)
}

// UpdateBoxDistribution takes a snapshot of current box distribution
func (as *AnalyticsService) UpdateBoxDistribution(ctx context.Context, userID, wordlistID int64) error {
	return as.repo.UpsertBoxDistribution(ctx, userID, wordlistID)
}

// GetHistoricalBoxDistribution retrieves historical box distribution
func (as *AnalyticsService) BoxHistory(ctx context.Context, days int) ([]map[string]interface{}, error) {
	return as.repo.GetBoxDistributionHistory(ctx, as.userID, as.wordlistID, days)
}

// GetCurrentBoxDistribution retrieves the current distribution of words across Leitner boxes
func (as *AnalyticsService) BoxDistribution(ctx context.Context) (*model.BoxDistribution, error) {
	return as.repo.GetCurrentBoxDistribution(ctx, as.userID, as.wordlistID)
}

// GetPracticeTime retrieves daily practice time for the last N days
func (as *AnalyticsService) PracticeTime(ctx context.Context, days int) ([]model.PracticeTimeStats, error) {
	return as.repo.GetPracticeTime(ctx, as.userID, as.wordlistID, days)
}

// Stats fetches and returns dashboard statistics for the wordlist
func (svc *AnalyticsService) Stats(ctx context.Context) (*model.WordlistStats, error) {
	stats := &model.WordlistStats{}

	// Use errgroup to run all queries concurrently
	g, ctx := errgroup.WithContext(ctx)

	// 1) Wordlist mastery summary
	g.Go(func() error {
		return svc.fetchWordlistMasteryStats(ctx, svc.userID, svc.wordlistID, stats)
	})

	// 2) Today's activity summary for this wordlist
	g.Go(func() error {
		return svc.fetchWordlistTodayStats(ctx, svc.userID, svc.wordlistID, stats)
	})

	// 3) Current streak for this wordlist
	var streak int
	g.Go(func() error {
		var err error
		streak, err = svc.fetchWordlistCurrentStreak(ctx, svc.userID, svc.wordlistID)
		return err
	})

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("failed to fetch wordlist stats: %w", err)
	}

	stats.CurrentStreak = streak
	return stats, nil
}

func (svc *AnalyticsService) fetchWordlistMasteryStats(ctx context.Context, userID, wordlistID int64, stats *model.WordlistStats) error {
	totalWords, wordsMastered, avgMastery, bestStreak, err := svc.repo.GetWordlistMasteryStats(ctx, userID, wordlistID)
	if err != nil {
		return err
	}

	stats.TotalWords = totalWords
	stats.WordsMastered = wordsMastered
	stats.AverageMastery = avgMastery
	stats.BestStreak = bestStreak
	return nil
}

func (svc *AnalyticsService) fetchWordlistTodayStats(ctx context.Context, userID, wordlistID int64, stats *model.WordlistStats) error {
	wordsStudiedToday, quizzesToday, accuracyToday, err := svc.repo.GetWordlistTodayStats(ctx, userID, wordlistID)
	if err != nil {
		return err
	}

	stats.WordsStudiedToday = wordsStudiedToday
	stats.QuizzesToday = quizzesToday
	stats.AccuracyToday = &accuracyToday
	return nil
}

func (svc *AnalyticsService) fetchWordlistCurrentStreak(ctx context.Context, userID, wordlistID int64) (int, error) {
	return svc.repo.GetWordlistCurrentStreak(ctx, userID, wordlistID)
}

// GetAllWordlistsProgress retrieves progress summary for all user's wordlists
func (as *AnalyticsService) ProgressSummary(ctx context.Context) (*model.ProgressSummaryResponse, error) {
	progress, err := as.repo.GetAllWordlistsProgress(ctx, as.userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wordlists progress: %w", err)
	}

	return &model.ProgressSummaryResponse{
		Wordlists: progress,
	}, nil
}
