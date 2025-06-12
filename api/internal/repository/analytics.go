package repository

import (
	"context"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository/analytics"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// QuizPerformanceRepositoryInterface defines quiz performance operations
type QuizPerformanceRepositoryInterface interface {
	RecordQuizPerformance(ctx context.Context, tx pgx.Tx, userID, wordlistID, wordID, definitionID, leitnerSystemTrackingID int64, quizType string, boxID int64, isCorrect bool, responseTimeMs int) error
	GetQuizTypePerformance(ctx context.Context, userID int64, wordlistID int64) ([]model.QuizTypePerformance, error)
	UpsertQuizTypeAnalytics(ctx context.Context, tx pgx.Tx, userID int64, quizType string, isCorrect bool, responseTimeMs int) error
}

// WordMasteryRepositoryInterface defines word mastery operations
type WordMasteryRepositoryInterface interface {
	UpsertWordMastery(ctx context.Context, tx pgx.Tx, userID, wordID int64, boxID int64, isCorrect bool) error
	GetWordMastery(ctx context.Context, userID, wordlistID int64) ([]model.WordMasteryStats, error)
	GetWordlistMasteryStats(ctx context.Context, userID, wordlistID int64) (int, int, *float64, *int, error)
}

// LearningProgressRepositoryInterface defines learning progress operations
type LearningProgressRepositoryInterface interface {
	UpsertLearningProgress(ctx context.Context, tx pgx.Tx, userID, wordlistID int64, date string, isCorrect bool, responseTimeMs int) error
	GetLearningProgress(ctx context.Context, userID, wordlistID int64, days int) ([]model.LearningProgressStats, error)
	GetWordlistTodayStats(ctx context.Context, userID, wordlistID int64) (int, int, float64, error)
	GetWordlistCurrentStreak(ctx context.Context, userID, wordlistID int64) (int, error)
}

// BoxDistributionRepositoryInterface defines box distribution operations
type BoxDistributionRepositoryInterface interface {
	UpsertBoxDistribution(ctx context.Context, userID, wordlistID int64) error
	GetCurrentBoxDistribution(ctx context.Context, userID, wordlistID int64) (*model.BoxDistribution, error)
	GetBoxDistributionHistory(ctx context.Context, userID, wordlistID int64, days int) ([]map[string]interface{}, error)
}

// DashboardStatsRepositoryInterface defines dashboard statistics operations
type DashboardStatsRepositoryInterface interface {
	GetPracticeTime(ctx context.Context, userID, wordlistID int64, days int) ([]model.PracticeTimeStats, error)
}

// BatchProgressRepositoryInterface defines batch progress operations
type BatchProgressRepositoryInterface interface {
	GetAllWordlistsProgress(ctx context.Context, userID int64) ([]model.WordlistProgress, error)
}

// AnalyticsRepositoryInterface combines all analytics operations
type AnalyticsRepositoryInterface interface {
	QuizPerformanceRepositoryInterface
	WordMasteryRepositoryInterface
	LearningProgressRepositoryInterface
	BoxDistributionRepositoryInterface
	DashboardStatsRepositoryInterface
	BatchProgressRepositoryInterface
}

// AnalyticsRepository is a composite repository that implements all analytics interfaces
type AnalyticsRepository struct {
	QuizPerformance  QuizPerformanceRepositoryInterface
	WordMastery      WordMasteryRepositoryInterface
	LearningProgress LearningProgressRepositoryInterface
	BoxDistribution  BoxDistributionRepositoryInterface
	DashboardStats   DashboardStatsRepositoryInterface
	BatchProgress    BatchProgressRepositoryInterface
}

// NewAnalyticsRepository creates a new composite analytics repository
func NewAnalyticsRepository(db *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{
		QuizPerformance:  analytics.NewQuizPerformanceRepository(db),
		WordMastery:      analytics.NewWordMasteryRepository(db),
		LearningProgress: analytics.NewLearningProgressRepository(db),
		BoxDistribution:  analytics.NewBoxDistributionRepository(db),
		DashboardStats:   analytics.NewDashboardStatsRepository(db),
		BatchProgress:    analytics.NewBatchProgressRepository(db),
	}
}

// Quiz Performance Operations - delegate to sub-repository
func (r *AnalyticsRepository) RecordQuizPerformance(ctx context.Context, tx pgx.Tx, userID, wordlistID, wordID, definitionID, leitnerSystemTrackingID int64, quizType string, boxID int64, isCorrect bool, responseTimeMs int) error {
	return r.QuizPerformance.RecordQuizPerformance(ctx, tx, userID, wordlistID, wordID, definitionID, leitnerSystemTrackingID, quizType, boxID, isCorrect, responseTimeMs)
}

func (r *AnalyticsRepository) GetQuizTypePerformance(ctx context.Context, userID int64, wordlistID int64) ([]model.QuizTypePerformance, error) {
	return r.QuizPerformance.GetQuizTypePerformance(ctx, userID, wordlistID)
}

func (r *AnalyticsRepository) UpsertQuizTypeAnalytics(ctx context.Context, tx pgx.Tx, userID int64, quizType string, isCorrect bool, responseTimeMs int) error {
	return r.QuizPerformance.UpsertQuizTypeAnalytics(ctx, tx, userID, quizType, isCorrect, responseTimeMs)
}

// Word Mastery Operations - delegate to sub-repository
func (r *AnalyticsRepository) UpsertWordMastery(ctx context.Context, tx pgx.Tx, userID, wordID int64, boxID int64, isCorrect bool) error {
	return r.WordMastery.UpsertWordMastery(ctx, tx, userID, wordID, boxID, isCorrect)
}

func (r *AnalyticsRepository) GetWordMastery(ctx context.Context, userID, wordlistID int64) ([]model.WordMasteryStats, error) {
	return r.WordMastery.GetWordMastery(ctx, userID, wordlistID)
}

func (r *AnalyticsRepository) GetWordlistMasteryStats(ctx context.Context, userID, wordlistID int64) (int, int, *float64, *int, error) {
	return r.WordMastery.GetWordlistMasteryStats(ctx, userID, wordlistID)
}

// Learning Progress Operations - delegate to sub-repository
func (r *AnalyticsRepository) UpsertLearningProgress(ctx context.Context, tx pgx.Tx, userID, wordlistID int64, date string, isCorrect bool, responseTimeMs int) error {
	return r.LearningProgress.UpsertLearningProgress(ctx, tx, userID, wordlistID, date, isCorrect, responseTimeMs)
}

func (r *AnalyticsRepository) GetLearningProgress(ctx context.Context, userID, wordlistID int64, days int) ([]model.LearningProgressStats, error) {
	return r.LearningProgress.GetLearningProgress(ctx, userID, wordlistID, days)
}

func (r *AnalyticsRepository) GetWordlistTodayStats(ctx context.Context, userID, wordlistID int64) (int, int, float64, error) {
	return r.LearningProgress.GetWordlistTodayStats(ctx, userID, wordlistID)
}

func (r *AnalyticsRepository) GetWordlistCurrentStreak(ctx context.Context, userID, wordlistID int64) (int, error) {
	return r.LearningProgress.GetWordlistCurrentStreak(ctx, userID, wordlistID)
}

// Box Distribution Operations - delegate to sub-repository
func (r *AnalyticsRepository) UpsertBoxDistribution(ctx context.Context, userID, wordlistID int64) error {
	return r.BoxDistribution.UpsertBoxDistribution(ctx, userID, wordlistID)
}

func (r *AnalyticsRepository) GetCurrentBoxDistribution(ctx context.Context, userID, wordlistID int64) (*model.BoxDistribution, error) {
	return r.BoxDistribution.GetCurrentBoxDistribution(ctx, userID, wordlistID)
}

func (r *AnalyticsRepository) GetBoxDistributionHistory(ctx context.Context, userID, wordlistID int64, days int) ([]map[string]interface{}, error) {
	return r.BoxDistribution.GetBoxDistributionHistory(ctx, userID, wordlistID, days)
}

// Dashboard Stats Operations - delegate to sub-repository
func (r *AnalyticsRepository) GetPracticeTime(ctx context.Context, userID, wordlistID int64, days int) ([]model.PracticeTimeStats, error) {
	return r.DashboardStats.GetPracticeTime(ctx, userID, wordlistID, days)
}

// Batch Progress Operations - delegate to sub-repository
func (r *AnalyticsRepository) GetAllWordlistsProgress(ctx context.Context, userID int64) ([]model.WordlistProgress, error) {
	return r.BatchProgress.GetAllWordlistsProgress(ctx, userID)
}
