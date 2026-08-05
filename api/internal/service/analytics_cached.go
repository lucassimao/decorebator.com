package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

// CachedAnalyticsService wraps the analytics service with Redis caching
type CachedAnalyticsService struct {
	repo       *repository.AnalyticsRepository
	cache      *redis.Client
	userRepo   *repository.UserRepository
	userID     int64
	wordlistID int64
	cacheTTL   time.Duration
}

// NewCachedAnalyticsService creates a new cached analytics service
func NewCachedAnalyticsService(db *pgxpool.Pool, cache *redis.Client, cfg AnalyticsConfig) (*CachedAnalyticsService, error) {
	return &CachedAnalyticsService{
		repo:       repository.NewAnalyticsRepository(db),
		cache:      cache,
		userRepo:   &repository.UserRepository{Db: db},
		userID:     cfg.UserID,
		wordlistID: cfg.WordlistID,
		cacheTTL:   cfg.CacheTTL,
	}, nil
}

// Stats gets stats with caching
func (s *CachedAnalyticsService) Stats(ctx context.Context) (*model.WordlistStats, error) {
	cacheKey := fmt.Sprintf("analytics:stats:%d:%d", s.userID, s.wordlistID)

	// Try cache first
	cached, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var stats model.WordlistStats
		if err := json.Unmarshal([]byte(cached), &stats); err == nil {
			return &stats, nil
		}
	}

	// Compute from database
	stats := &model.WordlistStats{}

	// Use errgroup to run queries concurrently
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		totalWords, wordsMastered, avgMastery, bestStreak, err := s.repo.GetWordlistMasteryStats(ctx, s.userID, s.wordlistID)
		if err != nil {
			return err
		}
		stats.TotalWords = totalWords
		stats.WordsMastered = wordsMastered
		stats.AverageMastery = avgMastery
		stats.BestStreak = bestStreak
		return nil
	})

	g.Go(func() error {
		wordsStudiedToday, quizzesToday, accuracyToday, err := s.repo.GetWordlistTodayStats(ctx, s.userID, s.wordlistID)
		if err != nil {
			return err
		}
		stats.WordsStudiedToday = wordsStudiedToday
		stats.QuizzesToday = quizzesToday
		stats.AccuracyToday = &accuracyToday
		return nil
	})

	g.Go(func() error {
		streak, err := s.repo.GetWordlistCurrentStreak(ctx, s.userID, s.wordlistID)
		stats.CurrentStreak = streak
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Cache the result
	if data, err := json.Marshal(stats); err == nil {
		s.cache.Set(ctx, cacheKey, data, s.cacheTTL)
	}

	return stats, nil
}

// ProgressSummary gets progress summary with caching
func (s *CachedAnalyticsService) ProgressSummary(ctx context.Context) (*model.ProgressSummaryResponse, error) {
	cacheKey := fmt.Sprintf("analytics:progress-summary:v2:%d", s.userID)

	// Try cache first
	cached, cacheErr := s.cache.Get(ctx, cacheKey).Result()
	var summary *model.ProgressSummaryResponse
	if cacheErr == nil {
		var cachedSummary model.ProgressSummaryResponse
		if unmarshalErr := json.Unmarshal([]byte(cached), &cachedSummary); unmarshalErr == nil {
			summary = &cachedSummary
		}
	}

	if summary == nil {
		progress, progressErr := s.repo.GetAllWordlistsProgress(ctx, s.userID)
		if progressErr != nil {
			return nil, progressErr
		}

		summary = &model.ProgressSummaryResponse{Wordlists: progress}
		// Cache only slowly changing aggregates. DueCount remains zero in this
		// serialized value and is refreshed below on every request.
		if data, err := json.Marshal(summary); err == nil {
			s.cache.Set(ctx, cacheKey, data, s.cacheTTL)
		}
	}

	dueCounts, err := s.repo.GetDueCounts(ctx, s.userID)
	if err != nil {
		return nil, err
	}
	applyDueCounts(summary.Wordlists, dueCounts)

	return summary, nil
}

// WordMastery gets word mastery with caching
func (s *CachedAnalyticsService) WordMastery(ctx context.Context) ([]model.WordMasteryStats, error) {
	cacheKey := fmt.Sprintf("analytics:mastery:%d:%d", s.userID, s.wordlistID)

	// Try cache first
	cached, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var stats []model.WordMasteryStats
		if err := json.Unmarshal([]byte(cached), &stats); err == nil {
			return stats, nil
		}
	}

	// Compute from database
	stats, err := s.repo.GetWordMastery(ctx, s.userID, s.wordlistID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if data, err := json.Marshal(stats); err == nil {
		s.cache.Set(ctx, cacheKey, data, s.cacheTTL)
	}

	return stats, nil
}

// LearningProgress gets learning progress with caching
func (s *CachedAnalyticsService) LearningProgress(ctx context.Context, days int) ([]model.LearningProgressStats, error) {
	cacheKey := fmt.Sprintf("analytics:progress:%d:%d:%d", s.userID, s.wordlistID, days)

	// Try cache first
	cached, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var progress []model.LearningProgressStats
		if err := json.Unmarshal([]byte(cached), &progress); err == nil {
			return progress, nil
		}
	}

	// Compute from database
	progress, err := s.repo.GetLearningProgress(ctx, s.userID, s.wordlistID, days)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if data, err := json.Marshal(progress); err == nil {
		s.cache.Set(ctx, cacheKey, data, s.cacheTTL)
	}

	return progress, nil
}

// QuizTypePerformance gets quiz performance with caching
func (s *CachedAnalyticsService) QuizTypePerformance(ctx context.Context) ([]model.QuizTypePerformance, error) {
	cacheKey := fmt.Sprintf("analytics:quiz-performance:%d:%d", s.userID, s.wordlistID)

	// Try cache first
	cached, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var performance []model.QuizTypePerformance
		if err := json.Unmarshal([]byte(cached), &performance); err == nil {
			return performance, nil
		}
	}

	// Compute from database
	performance, err := s.repo.GetQuizTypePerformance(ctx, s.userID, s.wordlistID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if data, err := json.Marshal(performance); err == nil {
		s.cache.Set(ctx, cacheKey, data, s.cacheTTL)
	}

	return performance, nil
}

// CurrentBoxDistribution gets current box distribution with caching
func (s *CachedAnalyticsService) CurrentBoxDistribution(ctx context.Context) (*model.BoxDistribution, error) {
	cacheKey := fmt.Sprintf("analytics:box-distribution:%d:%d", s.userID, s.wordlistID)

	// Try cache first
	cached, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var distribution model.BoxDistribution
		if err := json.Unmarshal([]byte(cached), &distribution); err == nil {
			return &distribution, nil
		}
	}

	// Compute from database
	distribution, err := s.repo.GetCurrentBoxDistribution(ctx, s.userID, s.wordlistID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if data, err := json.Marshal(distribution); err == nil {
		s.cache.Set(ctx, cacheKey, data, s.cacheTTL)
	}

	return distribution, nil
}

// BoxDistributionHistory gets historical box distribution with caching
func (s *CachedAnalyticsService) BoxDistributionHistory(ctx context.Context, days int) ([]map[string]interface{}, error) {
	cacheKey := fmt.Sprintf("analytics:historical-box-distribution:%d:%d:%d", s.userID, s.wordlistID, days)

	// Try cache first
	cached, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var distribution []map[string]interface{}
		if err := json.Unmarshal([]byte(cached), &distribution); err == nil {
			return distribution, nil
		}
	}

	// Compute from database
	distribution, err := s.repo.GetBoxDistributionHistory(ctx, s.userID, s.wordlistID, days)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if data, err := json.Marshal(distribution); err == nil {
		s.cache.Set(ctx, cacheKey, data, s.cacheTTL)
	}

	return distribution, nil
}

// PracticeTime gets practice time with caching
func (s *CachedAnalyticsService) PracticeTime(ctx context.Context, days int) ([]model.PracticeTimeStats, error) {
	cacheKey := fmt.Sprintf("analytics:practice-time:%d:%d:%d", s.userID, s.wordlistID, days)

	// Try cache first
	cached, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var practiceTime []model.PracticeTimeStats
		if err := json.Unmarshal([]byte(cached), &practiceTime); err == nil {
			return practiceTime, nil
		}
	}

	// Compute from database
	practiceTime, err := s.repo.GetPracticeTime(ctx, s.userID, s.wordlistID, days)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if data, err := json.Marshal(practiceTime); err == nil {
		s.cache.Set(ctx, cacheKey, data, s.cacheTTL)
	}

	return practiceTime, nil
}

// TrackQuiz delegates to the underlying repository (no caching for writes)
func (s *CachedAnalyticsService) TrackQuiz(ctx context.Context, result common.QuizResult, tx pgx.Tx) error {
	// Create a non-cached analytics service for tracking
	analyticsService := &AnalyticsService{repo: s.repo, userID: s.userID, wordlistID: s.wordlistID}
	return analyticsService.TrackQuiz(ctx, result, tx)
}

// UpdateBoxDistribution updates box distribution snapshot and invalidates related caches
func (s *CachedAnalyticsService) UpdateBoxDistribution(ctx context.Context, userID, wordlistID int64) error {
	// Update the box distribution snapshot
	err := s.repo.UpsertBoxDistribution(ctx, userID, wordlistID)
	if err != nil {
		return err
	}

	// Invalidate historical box distribution cache for this wordlist
	if s.cache != nil {
		keys := []string{}
		// Use specific key deletion for historical distribution with different day ranges
		for days := 7; days <= 365; days += 7 {
			keys = append(keys, fmt.Sprintf("analytics:historical-box-distribution:%d:%d:%d", userID, wordlistID, days))
		}
		// Also add common day ranges that might be cached
		commonDays := []int{1, 3, 7, 14, 30, 60, 90, 180, 365}
		for _, days := range commonDays {
			keys = append(keys, fmt.Sprintf("analytics:historical-box-distribution:%d:%d:%d", userID, wordlistID, days))
		}
		if len(keys) > 0 {
			s.cache.Del(ctx, keys...)
		}
	}

	return nil
}

// InvalidateWordlistAnalytics invalidates analytics caches for a specific wordlist
func (s *CachedAnalyticsService) InvalidateWordlistAnalytics(ctx context.Context, userID, wordlistID int64) error {
	if s.cache == nil {
		return nil // No cache available, nothing to invalidate
	}

	keys := []string{
		fmt.Sprintf("analytics:stats:%d:%d", userID, wordlistID),
		fmt.Sprintf("analytics:mastery:%d:%d", userID, wordlistID),
		fmt.Sprintf("analytics:progress:%d:%d", userID, wordlistID),
		fmt.Sprintf("analytics:quiz-performance:%d:%d", userID, wordlistID),
		fmt.Sprintf("analytics:box-distribution:%d:%d", userID, wordlistID),
		fmt.Sprintf("analytics:practice-time:%d:%d", userID, wordlistID),
		fmt.Sprintf("analytics:progress-summary:%d", userID),
		fmt.Sprintf("analytics:progress-summary:v2:%d", userID),
	}

	// Also invalidate historical box distribution caches with common day ranges
	for days := 7; days <= 365; days += 7 {
		keys = append(keys, fmt.Sprintf("analytics:historical-box-distribution:%d:%d:%d", userID, wordlistID, days))
	}

	if len(keys) > 0 {
		return s.cache.Del(ctx, keys...).Err()
	}

	return nil
}

// SetCache sets the Redis client for the service
// This is useful when creating the service for cache invalidation only
func (s *CachedAnalyticsService) SetCache(cache *redis.Client) {
	s.cache = cache
}
