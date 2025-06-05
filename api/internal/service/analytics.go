package service

import (
	"context"
	"fmt"
	"time"

	"decorebator.com/internal/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsService struct {
	db *pgxpool.Pool
}

func NewAnalyticsService() (*AnalyticsService, error) {
	db, err := common.GetDBConnection()
	if err != nil {
		return nil, err
	}
	return &AnalyticsService{db: db}, nil
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
	query := `
		INSERT INTO quiz_performance (
			user_id, wordlist_id, word_id, definition_id, 
			leitner_system_tracking_id, quiz_type, box_id, 
			is_correct, response_time_ms
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := tx.Exec(ctx, query,
		result.UserID, result.WordlistID, result.WordID, result.DefinitionID,
		result.LeitnerSystemTrackingID, string(result.QuizType), result.BoxID,
		result.IsCorrect, result.ResponseTimeMs,
	)

	return err
}

func (as *AnalyticsService) updateWordMastery(ctx context.Context, result QuizResult, tx pgx.Tx) error {
	// Calculate new mastery level based on current box and performance
	query := `
		INSERT INTO word_mastery (
			user_id, word_id, mastery_level, total_attempts, 
			correct_attempts, streak_count, max_streak, last_seen_at
		) VALUES (
			$1, $2, 
			CASE WHEN $3 THEN $4::decimal / 7.0 ELSE 0 END, -- mastery based on box
			1, 
			CASE WHEN $3 THEN 1 ELSE 0 END,
			CASE WHEN $3 THEN 1 ELSE 0 END,
			CASE WHEN $3 THEN 1 ELSE 0 END,
			NOW()
		)
		ON CONFLICT (user_id, word_id) DO UPDATE SET
			mastery_level = CASE 
				WHEN $3 THEN 
					-- Weighted average: 70% current box progress, 30% historical accuracy
					(0.7 * ($4::decimal / 7.0) + 
					 0.3 * (word_mastery.correct_attempts::decimal / NULLIF(word_mastery.total_attempts, 0)))
				ELSE 
					-- On failure, reduce mastery but not below 0
					GREATEST(word_mastery.mastery_level - 0.1, 0)
			END,
			total_attempts = word_mastery.total_attempts + 1,
			correct_attempts = word_mastery.correct_attempts + CASE WHEN $3 THEN 1 ELSE 0 END,
			streak_count = CASE 
				WHEN $3 THEN word_mastery.streak_count + 1 
				ELSE 0 
			END,
			max_streak = GREATEST(
				word_mastery.max_streak, 
				CASE WHEN $3 THEN word_mastery.streak_count + 1 ELSE word_mastery.streak_count END
			),
			last_seen_at = NOW(),
			updated_at = NOW()
	`

	_, err := tx.Exec(ctx, query, result.UserID, result.WordID, result.IsCorrect, result.BoxID)
	return err
}

func (as *AnalyticsService) updateLearningProgress(ctx context.Context, result QuizResult, tx pgx.Tx) error {
	today := time.Now().Format("2006-01-02")

	query := `
		INSERT INTO learning_progress (
			user_id, wordlist_id, date, words_studied, 
			total_quiz_attempts, correct_attempts, average_response_time_ms
		) VALUES (
			$1, $2, $3::date, 1, 1, 
			CASE WHEN $4 THEN 1 ELSE 0 END, $5
		)
		ON CONFLICT (user_id, wordlist_id, date) DO UPDATE SET
			words_studied = learning_progress.words_studied + 1,
			total_quiz_attempts = learning_progress.total_quiz_attempts + 1,
			correct_attempts = learning_progress.correct_attempts + CASE WHEN $4 THEN 1 ELSE 0 END,
			average_response_time_ms = (
				(learning_progress.average_response_time_ms * learning_progress.total_quiz_attempts + $5) / 
				(learning_progress.total_quiz_attempts + 1)
			)
	`

	_, err := tx.Exec(ctx, query, result.UserID, result.WordlistID, today, result.IsCorrect, result.ResponseTimeMs)
	return err
}

func (as *AnalyticsService) updateQuizTypeAnalytics(ctx context.Context, result QuizResult, tx pgx.Tx) error {
	query := `
		INSERT INTO quiz_type_analytics (
			user_id, quiz_type, total_attempts, correct_attempts, average_response_time_ms
		) VALUES (
			$1, $2, 1, 
			CASE WHEN $3 THEN 1 ELSE 0 END, $4
		)
		ON CONFLICT (user_id, quiz_type) DO UPDATE SET
			total_attempts = quiz_type_analytics.total_attempts + 1,
			correct_attempts = quiz_type_analytics.correct_attempts + CASE WHEN $3 THEN 1 ELSE 0 END,
			average_response_time_ms = (
				(quiz_type_analytics.average_response_time_ms * quiz_type_analytics.total_attempts + $4) / 
				(quiz_type_analytics.total_attempts + 1)
			),
			last_updated = NOW()
	`

	_, err := tx.Exec(ctx, query, result.UserID, string(result.QuizType), result.IsCorrect, result.ResponseTimeMs)
	return err
}

// Analytics Query Methods

type WordMasteryStats struct {
	WordID       int64      `json:"word_id"`
	Word         string     `json:"word"`
	MasteryLevel float64    `json:"mastery_level"`
	Accuracy     float64    `json:"accuracy"`
	StreakCount  int        `json:"streak_count"`
	LastSeenAt   *time.Time `json:"last_seen_at"`
	HighestBox   int        `json:"highest_box"`
}

// GetWordMastery retrieves mastery stats for all words in a wordlist
func (as *AnalyticsService) GetWordMastery(ctx context.Context, userID, wordlistID int64) ([]WordMasteryStats, error) {
	query := `
		SELECT word_id, word, mastery_level, accuracy_rate, 
		       streak_count, last_seen_at, highest_box_reached
		FROM mv_word_mastery_current
		WHERE user_id = $1 AND wordlist_id = $2
		ORDER BY mastery_level DESC, last_seen_at DESC
	`

	rows, err := as.db.Query(ctx, query, userID, wordlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []WordMasteryStats
	for rows.Next() {
		var s WordMasteryStats
		err := rows.Scan(&s.WordID, &s.Word, &s.MasteryLevel,
			&s.Accuracy, &s.StreakCount, &s.LastSeenAt, &s.HighestBox)
		if err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, nil
}

type QuizTypePerformance struct {
	QuizType      string    `json:"quiz_type"`
	TotalAttempts int       `json:"total_attempts"`
	SuccessRate   float64   `json:"success_rate"`
	AvgResponseMs int       `json:"avg_response_ms"`
	LastUpdated   time.Time `json:"last_updated"`
}

// GetQuizTypePerformance retrieves performance stats by quiz type
func (as *AnalyticsService) GetQuizTypePerformance(ctx context.Context, userID int64) ([]QuizTypePerformance, error) {
	query := `
		SELECT quiz_type, total_attempts, success_rate, 
		       average_response_time_ms, last_updated
		FROM mv_quiz_type_performance
		WHERE user_id = $1
		ORDER BY success_rate DESC
	`

	rows, err := as.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var performances []QuizTypePerformance
	for rows.Next() {
		var p QuizTypePerformance
		err := rows.Scan(&p.QuizType, &p.TotalAttempts, &p.SuccessRate,
			&p.AvgResponseMs, &p.LastUpdated)
		if err != nil {
			return nil, err
		}
		performances = append(performances, p)
	}

	return performances, nil
}

type LearningProgressStats struct {
	Date             time.Time `json:"date"`
	WordsStudied     int       `json:"words_studied"`
	WordsMastered    int       `json:"words_mastered"`
	TotalAttempts    int       `json:"total_attempts"`
	AccuracyRate     float64   `json:"accuracy_rate"`
	AvgResponseMs    int       `json:"avg_response_ms"`
	StudyTimeSeconds int       `json:"study_time_seconds"`
}

// GetLearningProgress retrieves daily learning progress
func (as *AnalyticsService) GetLearningProgress(ctx context.Context, userID, wordlistID int64, days int) ([]LearningProgressStats, error) {
	query := `
		SELECT date, words_studied, words_mastered, total_quiz_attempts,
		       CASE 
		           WHEN total_quiz_attempts > 0 
		           THEN ROUND(correct_attempts::numeric / total_quiz_attempts::numeric * 100, 1)
		           ELSE 0
		       END as accuracy_rate,
		       average_response_time_ms, study_time_seconds
		FROM learning_progress
		WHERE user_id = $1 AND wordlist_id = $2 
		  AND date >= CURRENT_DATE - INTERVAL '%d days'
		ORDER BY date DESC
	`

	rows, err := as.db.Query(ctx, fmt.Sprintf(query, days), userID, wordlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var progress []LearningProgressStats
	for rows.Next() {
		var p LearningProgressStats
		err := rows.Scan(&p.Date, &p.WordsStudied, &p.WordsMastered,
			&p.TotalAttempts, &p.AccuracyRate, &p.AvgResponseMs, &p.StudyTimeSeconds)
		if err != nil {
			return nil, err
		}
		progress = append(progress, p)
	}

	return progress, nil
}

// UpdateBoxDistribution takes a snapshot of current box distribution
func (as *AnalyticsService) UpdateBoxDistribution(ctx context.Context, userID, wordlistID int64) error {
	query := `
		INSERT INTO box_distribution_snapshot (
			user_id, wordlist_id, snapshot_date,
			box_1_count, box_2_count, box_3_count, box_4_count,
			box_5_count, box_6_count, box_7_count
		)
		SELECT 
			$1, $2, CURRENT_DATE,
			COUNT(CASE WHEN box_id = 1 THEN 1 END),
			COUNT(CASE WHEN box_id = 2 THEN 1 END),
			COUNT(CASE WHEN box_id = 3 THEN 1 END),
			COUNT(CASE WHEN box_id = 4 THEN 1 END),
			COUNT(CASE WHEN box_id = 5 THEN 1 END),
			COUNT(CASE WHEN box_id = 6 THEN 1 END),
			COUNT(CASE WHEN box_id = 7 THEN 1 END)
		FROM leitner_system_tracking lst
		JOIN word_definitions wd ON lst.definition_id = wd.definition_id
		JOIN words w ON wd.word_id = w.id
		WHERE lst.user_id = $1 AND w.wordlist_id = $2
		ON CONFLICT (user_id, wordlist_id, snapshot_date) DO UPDATE SET
			box_1_count = EXCLUDED.box_1_count,
			box_2_count = EXCLUDED.box_2_count,
			box_3_count = EXCLUDED.box_3_count,
			box_4_count = EXCLUDED.box_4_count,
			box_5_count = EXCLUDED.box_5_count,
			box_6_count = EXCLUDED.box_6_count,
			box_7_count = EXCLUDED.box_7_count
	`

	_, err := as.db.Exec(ctx, query, userID, wordlistID)
	return err
}

// GetBoxDistributionHistory retrieves historical box distribution
func (as *AnalyticsService) GetBoxDistributionHistory(ctx context.Context, userID, wordlistID int64, days int) ([]map[string]interface{}, error) {
	query := `
		SELECT snapshot_date, box_1_count, box_2_count, box_3_count,
		       box_4_count, box_5_count, box_6_count, box_7_count
		FROM box_distribution_snapshot
		WHERE user_id = $1 AND wordlist_id = $2
		  AND snapshot_date >= CURRENT_DATE - INTERVAL '%d days'
		ORDER BY snapshot_date DESC
	`

	rows, err := as.db.Query(ctx, fmt.Sprintf(query, days), userID, wordlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var distributions []map[string]interface{}
	for rows.Next() {
		var date time.Time
		var b1, b2, b3, b4, b5, b6, b7 int

		err := rows.Scan(&date, &b1, &b2, &b3, &b4, &b5, &b6, &b7)
		if err != nil {
			return nil, err
		}

		dist := map[string]interface{}{
			"date": date.Format("2006-01-02"),
			"boxes": map[string]int{
				"box_1": b1,
				"box_2": b2,
				"box_3": b3,
				"box_4": b4,
				"box_5": b5,
				"box_6": b6,
				"box_7": b7,
			},
		}
		distributions = append(distributions, dist)
	}

	return distributions, nil
}

// DashboardStats holds all the pieces of data we need for the dashboard.
type DashboardStats struct {
	TotalWords        int      `json:"total_words"`
	WordsMastered     int      `json:"words_mastered"`
	AverageMastery    *float64 `json:"average_mastery"` // could be nil if no rows
	BestStreak        *int     `json:"best_streak"`     // could be nil if no data
	WordsStudiedToday int      `json:"words_studied_today"`
	QuizzesToday      int      `json:"quizzes_today"`
	AccuracyToday     float64  `json:"accuracy_today"`
	CurrentStreak     int      `json:"current_streak"`
}

// GetDashboardStats fetches and returns all pieces of data for a given user.
func (svc *AnalyticsService) GetDashboardStats(ctx context.Context, userID int64) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// 1) Total mastery summary
	if err := svc.fetchTotalMasteryStats(ctx, userID, stats); err != nil {
		return nil, fmt.Errorf("fetchTotalMasteryStats: %w", err)
	}

	// 2) Today’s activity summary
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
	const query = `
		SELECT 
			COUNT(DISTINCT wm.word_id) AS total_words,
			COUNT(DISTINCT CASE WHEN wm.mastery_level >= 0.8 THEN wm.word_id END) AS words_mastered,
			AVG(wm.mastery_level) AS avg_mastery,
			MAX(wm.streak_count) AS best_streak
		FROM word_mastery wm
		WHERE wm.user_id = $1;
	`

	var avgMastery *float64
	var bestStreak *int

	err := svc.db.QueryRow(ctx, query, userID).Scan(
		&stats.TotalWords,
		&stats.WordsMastered,
		&avgMastery,
		&bestStreak,
	)
	if err != nil {
		return err
	}

	stats.AverageMastery = avgMastery
	stats.BestStreak = bestStreak
	return nil
}

func (svc *AnalyticsService) fetchTodayStats(ctx context.Context, userID int64, stats *DashboardStats) error {
	const todayQuery = `
		SELECT 
			COALESCE(SUM(words_studied), 0)           AS words_studied_today,
			COALESCE(SUM(total_quiz_attempts), 0)      AS quizzes_today,
			COALESCE(
			  AVG(
			    CASE 
			      WHEN total_quiz_attempts > 0 
			      THEN correct_attempts::float / total_quiz_attempts * 100 
			      ELSE 0 
			    END
			  ), 
			  0
			) AS accuracy_today
		FROM learning_progress
		WHERE user_id = $1
		  AND date = CURRENT_DATE;
	`

	err := svc.db.QueryRow(ctx, todayQuery, userID).Scan(
		&stats.WordsStudiedToday,
		&stats.QuizzesToday,
		&stats.AccuracyToday,
	)
	return err
}

func (svc *AnalyticsService) fetchCurrentStreak(ctx context.Context, userID int64) (int, error) {
	const streakQuery = `
		WITH daily_activity AS (
			SELECT 
				date, 
				SUM(total_quiz_attempts) AS attempts
			FROM learning_progress
			WHERE user_id = $1
			GROUP BY date
			ORDER BY date DESC
		),
		streak_calc AS (
			SELECT 
				date,
				date - (ROW_NUMBER() OVER (ORDER BY date DESC))::int AS streak_group
			FROM daily_activity
			WHERE attempts > 0
		)
		SELECT COUNT(*) AS current_streak
		FROM streak_calc
		WHERE streak_group = (
			SELECT streak_group 
			FROM streak_calc 
			WHERE date = CURRENT_DATE
			LIMIT 1
		);
	`

	var currentStreak int
	err := svc.db.QueryRow(ctx, streakQuery, userID).Scan(&currentStreak)
	if err != nil {
		// If there’s no row for “date = CURRENT_DATE,” pgx returns ErrNoRows.
		// In that case, we treat the streak as zero.
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return currentStreak, nil
}
