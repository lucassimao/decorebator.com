package analytics

import (
	"context"
	"strconv"

	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// QuizPerformanceRepository handles quiz performance operations
type QuizPerformanceRepository struct {
	db *pgxpool.Pool
}

const maxQuizPerformanceTypes = 20

// NewQuizPerformanceRepository creates a new quiz performance repository
func NewQuizPerformanceRepository(db *pgxpool.Pool) *QuizPerformanceRepository {
	return &QuizPerformanceRepository{db: db}
}

// RecordQuizPerformance records a quiz attempt in the quiz_performance table.
//
// Query: INSERT INTO quiz_performance with all quiz attempt details including:
// - user_id, wordlist_id, word_id, definition_id, leitner_system_tracking_id
// - quiz_type (e.g., "guess_meaning", "word_from_meaning", etc.)
// - box_id (current Leitner box level 1-7)
// - is_correct (boolean result of the quiz attempt)
// - response_time_ms (time taken to answer in milliseconds)
//
// This data is used for analytics aggregation and performance tracking.
func (r *QuizPerformanceRepository) RecordQuizPerformance(ctx context.Context, tx pgx.Tx, userID, wordlistID, wordID, definitionID, leitnerSystemTrackingID int64, quizType string, boxID int64, isCorrect bool, responseTimeMs int) error {
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

// GetQuizTypePerformance retrieves quiz type performance statistics for a wordlist.
//
// Query: Aggregates quiz_performance data by quiz_type for a specific user and wordlist:
// - Groups by quiz_type to get performance metrics per quiz mode
// - Calculates total_attempts (COUNT(*))
// - Calculates success_rate as percentage: (correct_attempts / total_attempts) * 100
// - Calculates average_response_time_ms with NULL handling and outlier filtering:
//   - Only includes response times between 200ms and 30000ms (realistic range)
//   - Uses COALESCE to handle NULL values in average calculation
//
// - Gets last_updated timestamp (MAX of created_at)
// - Orders by success_rate DESC to show best performing quiz types first
//
// Returns array of QuizTypePerformance with performance data for each quiz type.
func (r *QuizPerformanceRepository) GetQuizTypePerformance(ctx context.Context, userID int64, wordlistID int64) ([]model.QuizTypePerformance, error) {
	query := `
		SELECT 
			qp.quiz_type,
			COUNT(*) as total_attempts,
			CASE 
				WHEN COUNT(*) > 0 
				THEN ROUND(SUM(CASE WHEN qp.is_correct THEN 1 ELSE 0 END)::numeric / COUNT(*)::numeric * 100, 1)
				ELSE 0
			END AS success_rate,
			COALESCE(
				AVG(
					CASE 
						WHEN qp.response_time_ms BETWEEN 200 AND 30000 
						THEN qp.response_time_ms 
						ELSE NULL 
					END
				)::INT, 
				0
			) as average_response_time_ms,
			MAX(qp.created_at) as last_updated
		FROM quiz_performance qp
		WHERE qp.user_id = $1 AND qp.wordlist_id = $2
		GROUP BY qp.quiz_type
		ORDER BY success_rate DESC, qp.quiz_type ASC
		LIMIT ` + strconv.Itoa(maxQuizPerformanceTypes) + `
	`

	rows, err := r.db.Query(ctx, query, userID, wordlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var performances []model.QuizTypePerformance
	for rows.Next() {
		var p model.QuizTypePerformance
		scanErr := rows.Scan(&p.QuizType, &p.TotalAttempts, &p.SuccessRate,
			&p.AvgResponseMs, &p.LastUpdated)
		if scanErr != nil {
			return nil, scanErr
		}
		performances = append(performances, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return performances, nil
}

// UpsertQuizTypeAnalytics updates quiz type analytics aggregates in quiz_type_analytics table.
//
// Query: INSERT ... ON CONFLICT DO UPDATE pattern for real-time analytics:
// - Inserts new record if user/quiz_type combination doesn't exist
// - Updates existing record with new attempt data:
//   - Increments total_attempts by 1
//   - Increments correct_attempts by 1 if answer was correct
//   - Recalculates average_response_time_ms using weighted average formula with NULL handling:
//     COALESCE(old_avg, 0) * old_count + new_time) / (old_count + 1)
//   - Only includes response times in realistic range (200ms-30000ms) for average calculation
//   - Updates last_updated timestamp to NOW()
//
// This maintains real-time aggregated statistics per user per quiz type.
func (r *QuizPerformanceRepository) UpsertQuizTypeAnalytics(ctx context.Context, tx pgx.Tx, userID int64, quizType string, isCorrect bool, responseTimeMs int) error {
	// Validate response time is in realistic range for analytics
	filteredResponseTime := 0
	if responseTimeMs >= 200 && responseTimeMs <= 30000 {
		filteredResponseTime = responseTimeMs
	}

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
			average_response_time_ms = 
				CASE 
					WHEN $4 > 0 THEN  -- Only update average if we have a valid filtered response time
						(COALESCE(quiz_type_analytics.average_response_time_ms, 0) * quiz_type_analytics.total_attempts + $4) / 
						(quiz_type_analytics.total_attempts + 1)
					ELSE 
						quiz_type_analytics.average_response_time_ms  -- Keep existing average
				END,
			last_updated = NOW()
	`

	_, err := tx.Exec(ctx, query, userID, quizType, isCorrect, filteredResponseTime)
	return err
}
