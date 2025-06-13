package analytics

import (
	"context"

	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DashboardStatsRepository handles dashboard-specific analytics operations
type DashboardStatsRepository struct {
	db *pgxpool.Pool
}

// NewDashboardStatsRepository creates a new dashboard stats repository
func NewDashboardStatsRepository(db *pgxpool.Pool) *DashboardStatsRepository {
	return &DashboardStatsRepository{db: db}
}

// GetPracticeTime retrieves daily practice time statistics calculated from filtered quiz response times.
//
// Query: Aggregates quiz_performance data by date with outlier filtering for realistic practice time:
// - Groups by DATE(created_at) to aggregate quiz attempts by day
// - Filters response times to exclude unrealistic values:
//   * >= 200ms (minimum realistic response time)
//   * <= 30000ms (30 seconds maximum to exclude "stepped away" cases)
// - SUM(response_time_ms) as total_practice_time_ms for the day
// - Converts to minutes: SUM(response_time_ms) / 60000.0 with ROUND to 1 decimal
// - COUNT(*) as quiz_count to show total quiz attempts (including filtered ones)
// - Uses COALESCE to handle NULL response times (defaults to 0)
// - Filters by date range: created_at >= CURRENT_DATE - INTERVAL days
// - Orders by practice_date DESC to show most recent days first
//
// Note: This is an approximation of actual practice time based on quiz response times.
// The filtering helps remove obvious outliers (accidental clicks, stepping away) for more realistic estimates.
//
// Returns array of PracticeTimeStats with daily practice metrics for time-based analytics.
// Used for practice time charts and daily activity visualization.
func (r *DashboardStatsRepository) GetPracticeTime(ctx context.Context, userID, wordlistID int64, days int) ([]model.PracticeTimeStats, error) {
	// Validate days parameter
	if days < 0 {
		days = 0
	}

	query := `
		SELECT 
			DATE(qp.created_at) as practice_date,
			SUM(
				CASE 
					WHEN qp.response_time_ms BETWEEN 200 AND 30000 
					THEN qp.response_time_ms 
					ELSE 0 
				END
			) as total_practice_time_ms,
			ROUND(
				SUM(
					CASE 
						WHEN qp.response_time_ms BETWEEN 200 AND 30000 
						THEN qp.response_time_ms 
						ELSE 0 
					END
				)::numeric / 60000.0, 
				1
			) as practice_time_minutes,
			COUNT(*) as quiz_count
		FROM quiz_performance qp
		WHERE qp.user_id = $1 
		  AND qp.wordlist_id = $2
		  AND qp.created_at >= CURRENT_DATE - INTERVAL '1 day' * $3
		GROUP BY DATE(qp.created_at)
		ORDER BY practice_date DESC
	`

	rows, err := r.db.Query(ctx, query, userID, wordlistID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var practiceTime []model.PracticeTimeStats
	for rows.Next() {
		var p model.PracticeTimeStats
		err := rows.Scan(&p.Date, &p.PracticeTimeMs, &p.PracticeTimeMinutes, &p.QuizCount)
		if err != nil {
			return nil, err
		}
		practiceTime = append(practiceTime, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return practiceTime, nil
}
