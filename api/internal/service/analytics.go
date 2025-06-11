package service

import (
	"context"
	"fmt"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
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
	// Use errgroup to run all analytics operations concurrently
	g, ctx := errgroup.WithContext(ctx)
	
	// 1. Record individual quiz performance
	g.Go(func() error {
		err := as.recordQuizPerformance(ctx, result, tx)
		if err != nil {
			return fmt.Errorf("failed to record quiz performance: %w", err)
		}
		return nil
	})

	// 2. Update word mastery
	g.Go(func() error {
		err := as.updateWordMastery(ctx, result, tx)
		if err != nil {
			return fmt.Errorf("failed to update word mastery: %w", err)
		}
		return nil
	})

	// 3. Update daily learning progress
	g.Go(func() error {
		err := as.updateLearningProgress(ctx, result, tx)
		if err != nil {
			return fmt.Errorf("failed to update learning progress: %w", err)
		}
		return nil
	})

	// 4. Update quiz type analytics
	g.Go(func() error {
		err := as.updateQuizTypeAnalytics(ctx, result, tx)
		if err != nil {
			return fmt.Errorf("failed to update quiz type analytics: %w", err)
		}
		return nil
	})

	// Wait for all operations to complete
	return g.Wait()
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


// Type alias for BoxDistribution
type BoxDistribution = repository.BoxDistribution

// GetCurrentBoxDistribution retrieves the current distribution of words across Leitner boxes
func (as *AnalyticsService) GetCurrentBoxDistribution(ctx context.Context, userID, wordlistID int64) (*BoxDistribution, error) {
	return as.repo.GetCurrentBoxDistribution(ctx, userID, wordlistID)
}

// PracticeTimeStats represents daily practice time statistics
type PracticeTimeStats = repository.PracticeTimeStats

// GetPracticeTime retrieves daily practice time for the last N days
func (as *AnalyticsService) GetPracticeTime(ctx context.Context, userID, wordlistID int64, days int) ([]PracticeTimeStats, error) {
	return as.repo.GetPracticeTime(ctx, userID, wordlistID, days)
}

// WordlistDashboardStats holds dashboard statistics for a specific wordlist
type WordlistDashboardStats struct {
	TotalWords        int      `json:"totalWords"`
	WordsMastered     int      `json:"wordsMastered"`
	AverageMastery    *float64 `json:"averageMastery"` // could be nil if no rows
	BestStreak        *int     `json:"bestStreak"`     // could be nil if no data
	WordsStudiedToday int      `json:"wordsStudiedToday"`
	QuizzesToday      int      `json:"quizzesToday"`
	AccuracyToday     float64  `json:"accuracyToday"`
	CurrentStreak     int      `json:"currentStreak"`
}

// GetWordlistDashboardStats fetches and returns dashboard statistics for a specific wordlist
func (svc *AnalyticsService) GetWordlistDashboardStats(ctx context.Context, userID, wordlistID int64) (*WordlistDashboardStats, error) {
	stats := &WordlistDashboardStats{}
	
	// Use errgroup to run all queries concurrently
	g, ctx := errgroup.WithContext(ctx)
	
	// 1) Wordlist mastery summary
	g.Go(func() error {
		return svc.fetchWordlistMasteryStats(ctx, userID, wordlistID, stats)
	})
	
	// 2) Today's activity summary for this wordlist
	g.Go(func() error {
		return svc.fetchWordlistTodayStats(ctx, userID, wordlistID, stats)
	})
	
	// 3) Current streak for this wordlist
	var streak int
	g.Go(func() error {
		var err error
		streak, err = svc.fetchWordlistCurrentStreak(ctx, userID, wordlistID)
		return err
	})
	
	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("failed to fetch wordlist dashboard stats: %w", err)
	}
	
	stats.CurrentStreak = streak
	return stats, nil
}

func (svc *AnalyticsService) fetchWordlistMasteryStats(ctx context.Context, userID, wordlistID int64, stats *WordlistDashboardStats) error {
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

func (svc *AnalyticsService) fetchWordlistTodayStats(ctx context.Context, userID, wordlistID int64, stats *WordlistDashboardStats) error {
	wordsStudiedToday, quizzesToday, accuracyToday, err := svc.repo.GetWordlistTodayStats(ctx, userID, wordlistID)
	if err != nil {
		return err
	}

	stats.WordsStudiedToday = wordsStudiedToday
	stats.QuizzesToday = quizzesToday
	stats.AccuracyToday = accuracyToday
	return nil
}

func (svc *AnalyticsService) fetchWordlistCurrentStreak(ctx context.Context, userID, wordlistID int64) (int, error) {
	return svc.repo.GetWordlistCurrentStreak(ctx, userID, wordlistID)
}