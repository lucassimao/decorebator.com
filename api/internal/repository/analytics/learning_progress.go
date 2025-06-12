package analytics

import (
	"context"

	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LearningProgressRepository handles learning progress operations
type LearningProgressRepository struct {
	db *pgxpool.Pool
}

// NewLearningProgressRepository creates a new learning progress repository
func NewLearningProgressRepository(db *pgxpool.Pool) *LearningProgressRepository {
	return &LearningProgressRepository{db: db}
}

// UpsertLearningProgress updates or inserts daily learning progress aggregates in learning_progress table.
//
// Query: INSERT ... ON CONFLICT DO UPDATE pattern for real-time daily progress tracking:
// - Inserts new record if user/wordlist/date combination doesn't exist with initial values
// - Updates existing record with new quiz attempt data:
//   * words_studied: Recalculates using COUNT(DISTINCT word_id) from quiz_performance for the date
//   * total_quiz_attempts: Increments by 1 for each quiz attempt
//   * correct_attempts: Increments by 1 if answer was correct
//   * average_response_time_ms: Updates using weighted average formula:
//     (old_avg * old_count + new_time) / (old_count + 1)
//   * updated_at: Sets to NOW() timestamp
//
// This maintains accurate daily learning statistics with real-time updates on each quiz attempt.
func (r *LearningProgressRepository) UpsertLearningProgress(ctx context.Context, tx pgx.Tx, userID, wordlistID int64, date string, isCorrect bool, responseTimeMs int) error {
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

// GetLearningProgress retrieves daily learning progress statistics for a specific time period.
//
// Query: SELECT from learning_progress table with date range filtering:
// - Selects daily aggregated learning statistics for a specific wordlist and user
// - Calculates accuracy_rate as percentage: (correct_attempts / total_quiz_attempts) * 100
// - Handles division by zero with CASE WHEN total_quiz_attempts > 0 condition
// - Filters by date range: date >= CURRENT_DATE - INTERVAL days
// - Orders by date DESC to show most recent days first
// - Returns: date, words_studied, words_mastered, total_quiz_attempts, accuracy_rate, 
//   average_response_time_ms, study_time_seconds
//
// Used for analytics charts showing learning progress over time.
func (r *LearningProgressRepository) GetLearningProgress(ctx context.Context, userID, wordlistID int64, days int) ([]model.LearningProgressStats, error) {
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

	var progress []model.LearningProgressStats
	for rows.Next() {
		var p model.LearningProgressStats
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

// GetWordlistTodayStats retrieves today's activity statistics for a specific wordlist.
//
// Query: Aggregates today's learning progress data from learning_progress table:
// - SUM(words_studied) as total words studied today
// - SUM(total_quiz_attempts) as total quiz attempts today
// - AVG(accuracy calculation) as average accuracy across today's sessions
// - Accuracy calculation: (correct_attempts / total_quiz_attempts) * 100 per session
// - Uses COALESCE to handle NULL values and return 0 for no activity
// - Filters by date = CURRENT_DATE for today's data only
// - Returns: words_studied_today, quizzes_today, accuracy_today
//
// Used for dashboard "Today's Stats" widgets showing current day performance.
func (r *LearningProgressRepository) GetWordlistTodayStats(ctx context.Context, userID, wordlistID int64) (int, int, float64, error) {
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
		  AND wordlist_id = $2
		  AND date = CURRENT_DATE;
	`

	var wordsStudiedToday, quizzesToday int
	var accuracyToday float64

	err := r.db.QueryRow(ctx, query, userID, wordlistID).Scan(
		&wordsStudiedToday, &quizzesToday, &accuracyToday,
	)

	return wordsStudiedToday, quizzesToday, accuracyToday, err
}

// GetWordlistCurrentStreak calculates the current consecutive daily practice streak for a wordlist.
//
// Query: Complex CTE-based streak calculation using gap-and-island technique:
// 1. daily_activity CTE: Groups learning_progress by date and sums total_quiz_attempts
// 2. streak_calc CTE: Uses ROW_NUMBER() window function to identify consecutive date groups:
//    - Subtracts row number from date to create "streak_group" identifier
//    - Consecutive dates with activity will have the same streak_group value
// 3. Main query: Counts days in the streak group that includes CURRENT_DATE
//    - Only includes dates where attempts > 0 (active practice days)
//    - Returns 0 if no activity found for current date
//
// This algorithm efficiently calculates streaks by identifying gaps in daily activity.
// Example: dates [2024-01-01, 2024-01-02, 2024-01-03] become streak_group values
// that group consecutive practice days together.
func (r *LearningProgressRepository) GetWordlistCurrentStreak(ctx context.Context, userID, wordlistID int64) (int, error) {
	query := `
		WITH daily_activity AS (
			SELECT 
				date, 
				SUM(total_quiz_attempts) AS attempts
			FROM learning_progress
			WHERE user_id = $1 AND wordlist_id = $2
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
	err := r.db.QueryRow(ctx, query, userID, wordlistID).Scan(&currentStreak)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return currentStreak, nil
}