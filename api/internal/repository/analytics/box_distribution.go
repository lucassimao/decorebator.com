package analytics

import (
	"context"
	"fmt"
	"time"

	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoxDistributionRepository handles box distribution operations
type BoxDistributionRepository struct {
	db *pgxpool.Pool
}

// NewBoxDistributionRepository creates a new box distribution repository
func NewBoxDistributionRepository(db *pgxpool.Pool) *BoxDistributionRepository {
	return &BoxDistributionRepository{db: db}
}

// UpsertBoxDistribution updates or inserts daily box distribution snapshot in box_distribution_snapshot table.
//
// Query: INSERT ... ON CONFLICT DO UPDATE with complex CTE for box distribution calculation:
// 1. word_min_boxes CTE: Finds minimum box level for each word across all its definitions
//   - JOINs leitner_system_tracking with words table to filter by wordlist
//   - Uses MIN(box_id) to get the lowest Leitner box for each word (most recent learning level)
//   - Groups by word_id since words can have multiple definitions at different box levels
//
// 2. Main INSERT: Counts words at each box level (1-7) using conditional aggregation
//   - Uses COUNT(CASE WHEN min_box_id = X THEN 1 END) pattern for each box
//   - Sets snapshot_date to CURRENT_DATE for daily tracking
//
// 3. ON CONFLICT: Updates existing daily snapshot with fresh counts
//
// This creates daily snapshots of word distribution across Leitner boxes for historical analysis.
func (r *BoxDistributionRepository) UpsertBoxDistribution(ctx context.Context, userID, wordlistID int64) error {
	// Use the same logic as GetCurrentBoxDistribution - count words by their minimum box level
	query := `
		INSERT INTO box_distribution_snapshot (
			user_id, wordlist_id, snapshot_date,
			box_1_count, box_2_count, box_3_count, box_4_count,
			box_5_count, box_6_count, box_7_count
		)
		WITH word_min_boxes AS (
			SELECT 
				lst.word_id,
				MIN(lst.box_id) as min_box_id
			FROM leitner_system_tracking lst
			JOIN words w ON lst.word_id = w.id
			WHERE lst.user_id = $1 AND w.wordlist_id = $2
			GROUP BY lst.word_id
		)
		SELECT 
			$1::bigint, $2::bigint, CURRENT_DATE,
			COUNT(CASE WHEN min_box_id = 1 THEN 1 END),
			COUNT(CASE WHEN min_box_id = 2 THEN 1 END),
			COUNT(CASE WHEN min_box_id = 3 THEN 1 END),
			COUNT(CASE WHEN min_box_id = 4 THEN 1 END),
			COUNT(CASE WHEN min_box_id = 5 THEN 1 END),
			COUNT(CASE WHEN min_box_id = 6 THEN 1 END),
			COUNT(CASE WHEN min_box_id = 7 THEN 1 END)
		FROM word_min_boxes
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

	result, err := r.db.Exec(ctx, query, userID, wordlistID)
	if err != nil {
		return fmt.Errorf("failed to update box distribution: %w", err)
	}

	rowsAffected := result.RowsAffected()
	// Note: RowsAffected() for INSERT ... ON CONFLICT can return 1 for INSERT or 2 for UPDATE
	// We just want to log for debugging
	_ = rowsAffected

	return nil
}

// GetCurrentBoxDistribution retrieves the current real-time distribution of words across Leitner boxes.
//
// Query: CTE-based word distribution calculation from leitner_system_tracking table:
// 1. word_min_boxes CTE: Calculates minimum box level for each word
//   - JOINs leitner_system_tracking with words to filter by wordlist_id
//   - Uses MIN(box_id) since words can have multiple definitions at different box levels
//   - Groups by word_id to get one entry per word (using the lowest/most recent box level)
//
// 2. Main SELECT: Counts words at each box level using conditional aggregation
//   - COUNT(CASE WHEN min_box_id = X THEN 1 END) pattern for boxes 1-7
//   - COUNT(*) for total_words across all boxes
//
// Returns current BoxDistribution with counts for each box level and total words.
// Used for real-time box distribution charts and progress visualization.
func (r *BoxDistributionRepository) GetCurrentBoxDistribution(ctx context.Context, userID, wordlistID int64) (*model.BoxDistribution, error) {
	// First, get the minimum box for each word (since a word can have multiple definitions in different boxes)
	// Then count how many words are at each minimum box level
	query := `
		WITH word_min_boxes AS (
			SELECT 
				lst.word_id,
				MIN(lst.box_id) as min_box_id
			FROM leitner_system_tracking lst
			JOIN words w ON lst.word_id = w.id
			WHERE lst.user_id = $1 AND w.wordlist_id = $2
			GROUP BY lst.word_id
		)
		SELECT 
			COUNT(CASE WHEN min_box_id = 1 THEN 1 END) as box_1,
			COUNT(CASE WHEN min_box_id = 2 THEN 1 END) as box_2,
			COUNT(CASE WHEN min_box_id = 3 THEN 1 END) as box_3,
			COUNT(CASE WHEN min_box_id = 4 THEN 1 END) as box_4,
			COUNT(CASE WHEN min_box_id = 5 THEN 1 END) as box_5,
			COUNT(CASE WHEN min_box_id = 6 THEN 1 END) as box_6,
			COUNT(CASE WHEN min_box_id = 7 THEN 1 END) as box_7,
			COUNT(*) as total_words
		FROM word_min_boxes
	`

	var dist model.BoxDistribution
	err := r.db.QueryRow(ctx, query, userID, wordlistID).Scan(
		&dist.Box1Count, &dist.Box2Count, &dist.Box3Count, &dist.Box4Count,
		&dist.Box5Count, &dist.Box6Count, &dist.Box7Count, &dist.TotalWords,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get box distribution: %w", err)
	}

	return &dist, nil
}

// GetBoxDistributionHistory retrieves historical box distribution snapshots for chart visualization.
//
// Query: Simple SELECT from box_distribution_snapshot table with date range filtering:
// - Selects snapshot_date and all box counts (box_1_count through box_7_count)
// - Filters by user_id, wordlist_id and date range (CURRENT_DATE - days interval)
// - Orders by snapshot_date DESC to show most recent snapshots first
// - Returns raw historical data for time-series analytics
//
// Transforms results into map[string]interface{} format with:
// - "date": formatted date string (YYYY-MM-DD)
// - "boxes": BoxDistribution struct with individual box counts and calculated total
//
// Used for historical box distribution charts showing learning progress over time.
func (r *BoxDistributionRepository) GetBoxDistributionHistory(ctx context.Context, userID, wordlistID int64, days int) ([]map[string]interface{}, error) {
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

		if scanErr := rows.Scan(&date, &b1, &b2, &b3, &b4, &b5, &b6, &b7); scanErr != nil {
			return nil, scanErr
		}

		dist := map[string]interface{}{
			"date": date.Format("2006-01-02"),
			"boxes": model.BoxDistribution{
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
