package service

import (
	"context"
	"fmt"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
)

type AnalyticsService struct {
	repo *repository.AnalyticsRepository
}

func NewAnalyticsService() (*AnalyticsService, error) {
	db, err := common.GetDBConnection()
	if err != nil {
		return nil, err
	}
	repo := repository.NewAnalyticsRepository(db)
	return &AnalyticsService{repo: repo}, nil
}

// QuizResult contains the data needed to track quiz performance
type QuizResult = common.QuizResult

// TrackQuizPerformance records the result of a quiz attempt
func (as *AnalyticsService) TrackQuizPerformance(ctx context.Context, result QuizResult, tx pgx.Tx) error {
	// 1. Record individual quiz performance
	err := as.recordQuizPerformance(ctx, result, tx)
	if err != nil {
		return fmt.Errorf("failed to record quiz performance: %w", err)
	}

	// 2. Update word mastery
	err = as.updateWordMastery(ctx, result, tx)
	if err != nil {
		return fmt.Errorf("failed to update word mastery: %w", err)
	}

	// 3. Update daily learning progress
	err = as.updateLearningProgress(ctx, result, tx)
	if err != nil {
		return fmt.Errorf("failed to update learning progress: %w", err)
	}

	// 4. Update quiz type analytics
	err = as.updateQuizTypeAnalytics(ctx, result, tx)
	if err != nil {
		return fmt.Errorf("failed to update quiz type analytics: %w", err)
	}

	return nil
}

func (as *AnalyticsService) recordQuizPerformance(ctx context.Context, result QuizResult, tx pgx.Tx) error {
	return as.repo.RecordQuizPerformance(ctx, tx,
		result.UserID, result.WordlistID, result.WordID, result.DefinitionID,
		result.LeitnerSystemTrackingID, string(result.QuizType), result.BoxID,
		result.IsCorrect, result.ResponseTimeMs,
	)
}

func (as *AnalyticsService) updateWordMastery(ctx context.Context, result QuizResult, tx pgx.Tx) error {
	return as.repo.UpsertWordMastery(ctx, tx, result.UserID, result.WordID, result.BoxID, result.IsCorrect)
}

func (as *AnalyticsService) updateLearningProgress(ctx context.Context, result QuizResult, tx pgx.Tx) error {
	// Use UTC to ensure consistent date handling across timezones
	today := time.Now().UTC().Format("2006-01-02")
	return as.repo.UpsertLearningProgress(ctx, tx, result.UserID, result.WordlistID, today, result.IsCorrect, result.ResponseTimeMs)
}

func (as *AnalyticsService) updateQuizTypeAnalytics(ctx context.Context, result QuizResult, tx pgx.Tx) error {
	return as.repo.UpsertQuizTypeAnalytics(ctx, tx, result.UserID, string(result.QuizType), result.IsCorrect, result.ResponseTimeMs)
}

// Analytics Query Methods

// Type aliases to maintain compatibility
type WordMasteryStats = repository.WordMasteryStats

// GetWordMastery retrieves mastery stats for all words in a wordlist
func (as *AnalyticsService) GetWordMastery(ctx context.Context, userID, wordlistID int64) ([]WordMasteryStats, error) {
	return as.repo.GetWordMastery(ctx, userID, wordlistID)
}

type QuizTypePerformance = repository.QuizTypePerformance

// GetQuizTypePerformance retrieves performance stats by quiz type for a specific wordlist
func (as *AnalyticsService) GetQuizTypePerformance(ctx context.Context, userID int64, wordlistID int64) ([]QuizTypePerformance, error) {
	return as.repo.GetQuizTypePerformance(ctx, userID, wordlistID)
}

type LearningProgressStats = repository.LearningProgressStats

// GetLearningProgress retrieves daily learning progress
func (as *AnalyticsService) GetLearningProgress(ctx context.Context, userID, wordlistID int64, days int) ([]LearningProgressStats, error) {
	return as.repo.GetLearningProgress(ctx, userID, wordlistID, days)
}

// UpdateBoxDistribution takes a snapshot of current box distribution
func (as *AnalyticsService) UpdateBoxDistribution(ctx context.Context, userID, wordlistID int64) error {
	return as.repo.UpsertBoxDistribution(ctx, userID, wordlistID)
}

// GetBoxDistributionHistory retrieves historical box distribution
func (as *AnalyticsService) GetBoxDistributionHistory(ctx context.Context, userID, wordlistID int64, days int) ([]map[string]interface{}, error) {
	return as.repo.GetBoxDistributionHistory(ctx, userID, wordlistID, days)
}

// DashboardStats holds all the pieces of data we need for the dashboard.
type DashboardStats struct {
	TotalWords        int      `json:"totalWords"`
	WordsMastered     int      `json:"wordsMastered"`
	AverageMastery    *float64 `json:"averageMastery"` // could be nil if no rows
	BestStreak        *int     `json:"bestStreak"`     // could be nil if no data
	WordsStudiedToday int      `json:"wordsStudiedToday"`
	QuizzesToday      int      `json:"quizzesToday"`
	AccuracyToday     float64  `json:"accuracyToday"`
	CurrentStreak     int      `json:"currentStreak"`
}

// GetDashboardStats fetches and returns all pieces of data for a given user.
func (svc *AnalyticsService) GetDashboardStats(ctx context.Context, userID int64) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// 1) Total mastery summary
	if err := svc.fetchTotalMasteryStats(ctx, userID, stats); err != nil {
		return nil, fmt.Errorf("fetchTotalMasteryStats: %w", err)
	}

	// 2) Today's activity summary
	if err := svc.fetchTodayStats(ctx, userID, stats); err != nil {
		return nil, fmt.Errorf("fetchTodayStats: %w", err)
	}

	// 3) Current streak
	streak, err := svc.fetchCurrentStreak(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetchCurrentStreak: %w", err)
	}
	stats.CurrentStreak = streak

	return stats, nil
}

func (svc *AnalyticsService) fetchTotalMasteryStats(ctx context.Context, userID int64, stats *DashboardStats) error {
	totalWords, wordsMastered, avgMastery, bestStreak, err := svc.repo.GetTotalMasteryStats(ctx, userID)
	if err != nil {
		return err
	}

	stats.TotalWords = totalWords
	stats.WordsMastered = wordsMastered
	stats.AverageMastery = avgMastery
	stats.BestStreak = bestStreak
	return nil
}

func (svc *AnalyticsService) fetchTodayStats(ctx context.Context, userID int64, stats *DashboardStats) error {
	wordsStudiedToday, quizzesToday, accuracyToday, err := svc.repo.GetTodayStats(ctx, userID)
	if err != nil {
		return err
	}

	stats.WordsStudiedToday = wordsStudiedToday
	stats.QuizzesToday = quizzesToday
	stats.AccuracyToday = accuracyToday
	return nil
}

func (svc *AnalyticsService) fetchCurrentStreak(ctx context.Context, userID int64) (int, error) {
	return svc.repo.GetCurrentStreak(ctx, userID)
}

// Type alias for BoxDistribution
type BoxDistribution = repository.BoxDistribution

// GetCurrentBoxDistribution retrieves the current distribution of words across Leitner boxes
func (as *AnalyticsService) GetCurrentBoxDistribution(ctx context.Context, userID, wordlistID int64) (*BoxDistribution, error) {
	return as.repo.GetCurrentBoxDistribution(ctx, userID, wordlistID)
}