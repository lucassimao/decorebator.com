package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsRepository struct {
	db *pgxpool.Pool
}

func NewAnalyticsRepository(db *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// Quiz Performance Operations
func (r *AnalyticsRepository) RecordQuizPerformance(ctx context.Context, tx pgx.Tx, userID, wordlistID, wordID, definitionID, leitnerSystemTrackingID int64, quizType string, boxID int64, isCorrect bool, responseTimeMs int) error {
	query := `
		INSERT INTO quiz_performance (
			user_id, wordlist_id, word_id, definition_id, 
			leitner_system_tracking_id, quiz_type, box_id, 
			is_correct, response_time_ms
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := tx.Exec(ctx, query,
		userID, wordlistID, wordID, definitionID,
		leitnerSystemTrackingID, quizType, boxID,
		isCorrect, responseTimeMs,
	)

	return err
}

// Word Mastery Operations
func (r *AnalyticsRepository) UpsertWordMastery(ctx context.Context, tx pgx.Tx, userID, wordID int64, boxID int64, isCorrect bool) error {
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
					-- Handle NULL case by defaulting to 0 accuracy when no previous attempts
					(0.7 * ($4::decimal / 7.0) + 
					 0.3 * COALESCE(word_mastery.correct_attempts::decimal / NULLIF(word_mastery.total_attempts, 0), 0))
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

	_, err := tx.Exec(ctx, query, userID, wordID, isCorrect, boxID)
	return err
}

// Learning Progress Operations
func (r *AnalyticsRepository) UpsertLearningProgress(ctx context.Context, tx pgx.Tx, userID, wordlistID int64, date string, isCorrect bool, responseTimeMs int) error {
	query := `
		INSERT INTO learning_progress (
			user_id, wordlist_id, date, words_studied, 
			total_quiz_attempts, correct_attempts, average_response_time_ms
		) VALUES (
			$1, $2, $3::date, 1, 1, 
			CASE WHEN $4 THEN 1 ELSE 0 END, $5
		)
		ON CONFLICT (user_id, wordlist_id, date) DO UPDATE SET
			words_studied = (
				SELECT COUNT(DISTINCT qp.word_id) 
				FROM quiz_performance qp
				WHERE qp.user_id = $1 
				  AND qp.wordlist_id = $2
				  AND DATE(qp.created_at) = $3::date
			),
			total_quiz_attempts = learning_progress.total_quiz_attempts + 1,
			correct_attempts = learning_progress.correct_attempts + CASE WHEN $4 THEN 1 ELSE 0 END,
			average_response_time_ms = (
				(learning_progress.average_response_time_ms * learning_progress.total_quiz_attempts + $5) / 
				(learning_progress.total_quiz_attempts + 1)
			),
			updated_at = NOW()
	`

	_, err := tx.Exec(ctx, query, userID, wordlistID, date, isCorrect, responseTimeMs)
	return err
}

// Quiz Type Analytics Operations
func (r *AnalyticsRepository) UpsertQuizTypeAnalytics(ctx context.Context, tx pgx.Tx, userID int64, quizType string, isCorrect bool, responseTimeMs int) error {
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

	_, err := tx.Exec(ctx, query, userID, quizType, isCorrect, responseTimeMs)
	return err
}

// Box Distribution Operations
func (r *AnalyticsRepository) UpsertBoxDistribution(ctx context.Context, userID, wordlistID int64) error {
	query := `
		INSERT INTO box_distribution_snapshot (
			user_id, wordlist_id, snapshot_date,
			box_1_count, box_2_count, box_3_count, box_4_count,
			box_5_count, box_6_count, box_7_count
		)
		SELECT 
			$1::bigint, $2::bigint, CURRENT_DATE,
			COUNT(CASE WHEN lst.box_id = 1 THEN 1 END),
			COUNT(CASE WHEN lst.box_id = 2 THEN 1 END),
			COUNT(CASE WHEN lst.box_id = 3 THEN 1 END),
			COUNT(CASE WHEN lst.box_id = 4 THEN 1 END),
			COUNT(CASE WHEN lst.box_id = 5 THEN 1 END),
			COUNT(CASE WHEN lst.box_id = 6 THEN 1 END),
			COUNT(CASE WHEN lst.box_id = 7 THEN 1 END)
		FROM leitner_system_tracking lst
		JOIN words w ON lst.word_id = w.id
		WHERE lst.user_id = $1 AND w.wordlist_id = $2
		ON CONFLICT (user_id, wordlist_id, snapshot_date) DO UPDATE SET
			box_1_count = EXCLUDED.box_1_count,
			box_2_count = EXCLUDED.box_2_count,
			box_3_count = EXCLUDED.box_3_count,
			box_4_count = EXCLUDED.box_4_count,
			box_5_count = EXCLUDED.box_5_count,
			box_6_count = EXCLUDED.box_6_count,
			box_7_count = EXCLUDED.box_7_count,
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query, userID, wordlistID)
	if err != nil {
		return fmt.Errorf("failed to update box distribution: %w", err)
	}

	return nil
}

// Query Operations

// WordMasteryStats represents word mastery statistics
type WordMasteryStats struct {
	WordID       int64      `json:"wordId"`
	Word         string     `json:"word"`
	MasteryLevel float64    `json:"masteryLevel"`
	Accuracy     float64    `json:"accuracy"`
	StreakCount  int        `json:"streakCount"`
	LastSeenAt   *time.Time `json:"lastSeenAt"`
	HighestBox   int        `json:"highestBox"`
}

func (r *AnalyticsRepository) GetWordMastery(ctx context.Context, userID, wordlistID int64) ([]WordMasteryStats, error) {
	query := `
		SELECT word_id, word, mastery_level, accuracy_rate, 
		       streak_count, last_seen_at, highest_box_reached
		FROM mv_word_mastery_current
		WHERE user_id = $1 AND wordlist_id = $2
		ORDER BY mastery_level DESC, last_seen_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID, wordlistID)
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}

// QuizTypePerformance represents quiz type performance statistics
type QuizTypePerformance struct {
	QuizType      string    `json:"quizType"`
	TotalAttempts int       `json:"totalAttempts"`
	SuccessRate   float64   `json:"successRate"`
	AvgResponseMs int       `json:"avgResponseMs"`
	LastUpdated   time.Time `json:"lastUpdated"`
}

func (r *AnalyticsRepository) GetQuizTypePerformance(ctx context.Context, userID int64) ([]QuizTypePerformance, error) {
	query := `
		SELECT quiz_type, total_attempts, success_rate, 
		       average_response_time_ms, last_updated
		FROM mv_quiz_type_performance
		WHERE user_id = $1
		ORDER BY success_rate DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return performances, nil
}

// LearningProgressStats represents daily learning progress statistics
type LearningProgressStats struct {
	Date             time.Time `json:"date"`
	WordsStudied     int       `json:"wordsStudied"`
	WordsMastered    int       `json:"wordsMastered"`
	TotalAttempts    int       `json:"totalAttempts"`
	AccuracyRate     float64   `json:"accuracyRate"`
	AvgResponseMs    int       `json:"avgResponseMs"`
	StudyTimeSeconds int       `json:"studyTimeSeconds"`
}

func (r *AnalyticsRepository) GetLearningProgress(ctx context.Context, userID, wordlistID int64, days int) ([]LearningProgressStats, error) {
	// Validate days parameter
	if days < 0 {
		days = 0
	}
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
		  AND date >= CURRENT_DATE - INTERVAL '1 day' * $3
		ORDER BY date DESC
	`

	rows, err := r.db.Query(ctx, query, userID, wordlistID, days)
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return progress, nil
}

func (r *AnalyticsRepository) GetBoxDistributionHistory(ctx context.Context, userID, wordlistID int64, days int) ([]map[string]interface{}, error) {
	// Validate days parameter
	if days < 0 {
		days = 0
	}
	query := `
		SELECT snapshot_date, box_1_count, box_2_count, box_3_count,
		       box_4_count, box_5_count, box_6_count, box_7_count
		FROM box_distribution_snapshot
		WHERE user_id = $1 AND wordlist_id = $2
		  AND snapshot_date >= CURRENT_DATE - INTERVAL '1 day' * $3
		ORDER BY snapshot_date DESC
	`

	rows, err := r.db.Query(ctx, query, userID, wordlistID, days)
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
			"boxes": BoxDistribution{
				Box1Count:  b1,
				Box2Count:  b2,
				Box3Count:  b3,
				Box4Count:  b4,
				Box5Count:  b5,
				Box6Count:  b6,
				Box7Count:  b7,
				TotalWords: b1 + b2 + b3 + b4 + b5 + b6 + b7,
			},
		}
		distributions = append(distributions, dist)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return distributions, nil
}

// Dashboard Stats Operations

// DashboardStats holds all the pieces of data we need for the dashboard
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

func (r *AnalyticsRepository) GetTotalMasteryStats(ctx context.Context, userID int64) (int, int, *float64, *int, error) {
	query := `
		SELECT 
			COUNT(DISTINCT wm.word_id) AS total_words,
			COUNT(DISTINCT CASE WHEN wm.mastery_level >= 0.8 THEN wm.word_id END) AS words_mastered,
			AVG(wm.mastery_level) AS avg_mastery,
			MAX(wm.streak_count) AS best_streak
		FROM word_mastery wm
		WHERE wm.user_id = $1;
	`

	var totalWords, wordsMastered int
	var avgMastery *float64
	var bestStreak *int

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&totalWords, &wordsMastered, &avgMastery, &bestStreak,
	)

	return totalWords, wordsMastered, avgMastery, bestStreak, err
}

func (r *AnalyticsRepository) GetTodayStats(ctx context.Context, userID int64) (int, int, float64, error) {
	query := `
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

	var wordsStudiedToday, quizzesToday int
	var accuracyToday float64

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&wordsStudiedToday, &quizzesToday, &accuracyToday,
	)

	return wordsStudiedToday, quizzesToday, accuracyToday, err
}

func (r *AnalyticsRepository) GetCurrentStreak(ctx context.Context, userID int64) (int, error) {
	query := `
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
	err := r.db.QueryRow(ctx, query, userID).Scan(&currentStreak)
	if err != nil {
		// If there's no row for "date = CURRENT_DATE," pgx returns ErrNoRows.
		// In that case, we treat the streak as zero.
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return currentStreak, nil
}

// BoxDistribution represents the current distribution of words across Leitner boxes
type BoxDistribution struct {
	Box1Count  int `json:"box1"`
	Box2Count  int `json:"box2"`
	Box3Count  int `json:"box3"`
	Box4Count  int `json:"box4"`
	Box5Count  int `json:"box5"`
	Box6Count  int `json:"box6"`
	Box7Count  int `json:"box7"`
	TotalWords int `json:"totalWords"`
}

// GetCurrentBoxDistribution retrieves the current distribution of words across Leitner boxes
func (r *AnalyticsRepository) GetCurrentBoxDistribution(ctx context.Context, userID, wordlistID int64) (*BoxDistribution, error) {
	query := `
		SELECT 
			COUNT(CASE WHEN lst.box_id = 1 THEN 1 END) as box_1,
			COUNT(CASE WHEN lst.box_id = 2 THEN 1 END) as box_2,
			COUNT(CASE WHEN lst.box_id = 3 THEN 1 END) as box_3,
			COUNT(CASE WHEN lst.box_id = 4 THEN 1 END) as box_4,
			COUNT(CASE WHEN lst.box_id = 5 THEN 1 END) as box_5,
			COUNT(CASE WHEN lst.box_id = 6 THEN 1 END) as box_6,
			COUNT(CASE WHEN lst.box_id = 7 THEN 1 END) as box_7,
			COUNT(*) as total_words
		FROM leitner_system_tracking lst
		JOIN words w ON lst.word_id = w.id
		WHERE lst.user_id = $1 AND w.wordlist_id = $2
	`

	var dist BoxDistribution
	err := r.db.QueryRow(ctx, query, userID, wordlistID).Scan(
		&dist.Box1Count, &dist.Box2Count, &dist.Box3Count, &dist.Box4Count,
		&dist.Box5Count, &dist.Box6Count, &dist.Box7Count, &dist.TotalWords,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get box distribution: %w", err)
	}

	return &dist, nil
}
