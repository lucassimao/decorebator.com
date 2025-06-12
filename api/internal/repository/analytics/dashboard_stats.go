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

// GetPracticeTime retrieves daily practice time statistics calculated from quiz response times.
//
// Query: Aggregates quiz_performance data by date to calculate daily practice time:
// - Groups by DATE(created_at) to aggregate quiz attempts by day
// - SUM(response_time_ms) as total_practice_time_ms for the day
// - Converts milliseconds to minutes: SUM(response_time_ms) / 60000.0 with ROUND to 1 decimal
// - COUNT(*) as quiz_count to show number of quiz attempts per day
// - Uses COALESCE to handle NULL response times (defaults to 0)
// - Filters by date range: created_at >= CURRENT_DATE - INTERVAL days
// - Orders by practice_date DESC to show most recent days first
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
			SUM(COALESCE(qp.response_time_ms, 0)) as total_practice_time_ms,
			ROUND(SUM(COALESCE(qp.response_time_ms, 0))::numeric / 60000.0, 1) as practice_time_minutes,
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