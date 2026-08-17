package analytics

import (
	"context"
	"fmt"
	"strconv"

	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BatchProgressRepository handles batch progress operations
type BatchProgressRepository struct {
	db *pgxpool.Pool
}

// NewBatchProgressRepository creates a new batch progress repository
func NewBatchProgressRepository(db *pgxpool.Pool) *BatchProgressRepository {
	return &BatchProgressRepository{db: db}
}

// GetAllWordlistsProgress retrieves comprehensive progress summary for all user's wordlists.
//
// Query: Optimized multi-CTE query combining wordlist statistics and streak calculations:
// 1. wordlist_stats CTE: Aggregates word mastery data per wordlist
//   - JOINs wordlists → words → word_mastery (more efficient join order)
//   - COUNT(DISTINCT w.id) as total_words in each wordlist
//   - COUNT(DISTINCT CASE WHEN mastery_level >= 0.8) as words_mastered (80% threshold)
//   - AVG(mastery_level) * 100 as progress_percent with NULL handling
//   - Filters: user_id match and learned = FALSE (active words only)
//
// 2. recent_activity CTE: Gets last practice date per wordlist
//   - MAX(date) where total_quiz_attempts > 0 for actual practice days
//
// 3. streaks CTE: Calculates current consecutive practice streaks per wordlist
//   - Uses recursive CTE approach for accurate streak counting
//   - Starts from most recent practice date (today or earlier)
//   - Counts backwards while consecutive practice days exist
//   - Handles edge cases: practiced today vs yesterday vs no recent activity
//
// 4. Main SELECT: Combines all CTEs using LEFT JOINs
//   - Returns: wordlist_id, name, language_code, totals, mastery stats, streak, last_activity
//
// Used for dashboard overview showing all wordlists with comprehensive progress metrics.
func (r *BatchProgressRepository) GetAllWordlistsProgress(ctx context.Context, userID int64, limit int, cursor *int64) ([]model.WordlistProgress, error) {
	queryArgs := []any{userID}
	cursorClause := ""
	if cursor != nil {
		cursorClause = " AND wl.id > $2"
		queryArgs = append(queryArgs, *cursor)
	}
	query := `
		WITH wordlist_stats AS (
			SELECT 
				wl.id as wordlist_id,
				wl.name as wordlist_name,
				wl.language_code,
				COUNT(DISTINCT w.id) as total_words,
				COUNT(DISTINCT CASE WHEN wm.mastery_level >= 0.8 THEN w.id END) as words_mastered,
				CASE 
					WHEN COUNT(DISTINCT w.id) > 0 
					THEN ROUND(AVG(COALESCE(wm.mastery_level, 0)) * 100, 2)
					ELSE 0 
				END as progress_percent
			FROM wordlists wl
			LEFT JOIN words w ON wl.id = w.wordlist_id AND w.user_id = $1 AND w.learned = FALSE
			LEFT JOIN word_mastery wm ON w.id = wm.word_id AND wm.user_id = $1
			WHERE wl.user_id = $1` + cursorClause + `
			GROUP BY wl.id, wl.name, wl.language_code
		),
		recent_activity AS (
			SELECT 
				wordlist_id,
				MAX(date) as last_activity_date,
				MAX(CASE WHEN total_quiz_attempts > 0 THEN date END) as last_practice_date
			FROM learning_progress
			WHERE user_id = $1
			GROUP BY wordlist_id
		),
		streaks AS (
			SELECT DISTINCT ON (wordlist_id)
				wordlist_id,
				current_streak
			FROM (
				-- For each wordlist, calculate its current streak
				SELECT 
					ra.wordlist_id,
					(
						WITH RECURSIVE streak_days(practice_date, day_count) AS (
							-- Base case: start from most recent practice date
							SELECT 
								ra.last_practice_date,
								1
							WHERE ra.last_practice_date IS NOT NULL
							  AND ra.last_practice_date >= CURRENT_DATE - INTERVAL '1 day'
							
							UNION ALL
							
							-- Recursive case: go back one day if practice exists
							SELECT 
								(sd.practice_date - INTERVAL '1 day')::date,
								sd.day_count + 1
							FROM streak_days sd
							WHERE EXISTS (
								SELECT 1 
								FROM learning_progress lp 
								WHERE lp.user_id = $1 
								  AND lp.wordlist_id = ra.wordlist_id
								  AND lp.date = (sd.practice_date - INTERVAL '1 day')::date
								  AND lp.total_quiz_attempts > 0
							)
							  AND sd.day_count < 365  -- Prevent infinite recursion
						)
						SELECT COALESCE(MAX(day_count), 0)
						FROM streak_days
					) as current_streak
				FROM recent_activity ra
			) streak_calc
		)
		SELECT 
			ws.wordlist_id,
			ws.wordlist_name,
			ws.language_code,
			ws.total_words,
			ws.words_mastered,
			ws.progress_percent,
			COALESCE(s.current_streak, 0) as current_streak,
			ra.last_activity_date
		FROM wordlist_stats ws
		LEFT JOIN recent_activity ra ON ws.wordlist_id = ra.wordlist_id
		LEFT JOIN streaks s ON ws.wordlist_id = s.wordlist_id
		ORDER BY ws.wordlist_id ASC
		LIMIT $` + strconv.Itoa(len(queryArgs)+1)
	queryArgs = append(queryArgs, limit)

	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get wordlists progress: %w", err)
	}
	defer rows.Close()

	var progress []model.WordlistProgress
	for rows.Next() {
		var wp model.WordlistProgress
		err := rows.Scan(
			&wp.WordlistID,
			&wp.WordlistName,
			&wp.LanguageCode,
			&wp.TotalWords,
			&wp.WordsMastered,
			&wp.ProgressPercent,
			&wp.CurrentStreak,
			&wp.LastActivityDate,
		)
		if err != nil {
			return nil, err
		}
		progress = append(progress, wp)
	}

	return progress, rows.Err()
}

// GetDueCounts returns a fresh per-wordlist count of definitions due now. It is
// intentionally separate from the cacheable progress aggregates because the
// answer changes as next_review_at crosses the current time.
func (r *BatchProgressRepository) GetDueCounts(ctx context.Context, userID int64, wordlistIDs []int64) (map[int64]int, error) {
	if len(wordlistIDs) == 0 {
		return map[int64]int{}, nil
	}
	query := `
		SELECT
			w.wordlist_id,
			COUNT(DISTINCT lst.id) AS due_count
		FROM leitner_system_tracking lst
		JOIN definitions def ON def.id = lst.definition_id
		JOIN words w ON w.id = lst.word_id
		JOIN word_definitions wd ON wd.definition_id = lst.definition_id AND wd.word_id = lst.word_id
		WHERE lst.user_id = $1
			AND w.wordlist_id = ANY($2)
			AND w.user_id = $1
			AND w.learned = FALSE
			AND def.meaning IS NOT NULL
			AND lst.next_review_at IS NOT NULL
			AND lst.next_review_at <= NOW()
			AND (lst.temporarily_skipped_until IS NULL OR lst.temporarily_skipped_until < NOW())
		GROUP BY w.wordlist_id
	`

	rows, err := r.db.Query(ctx, query, userID, wordlistIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get due counts: %w", err)
	}
	defer rows.Close()

	dueCounts := make(map[int64]int)
	for rows.Next() {
		var wordlistID int64
		var dueCount int
		if err := rows.Scan(&wordlistID, &dueCount); err != nil {
			return nil, fmt.Errorf("failed to scan due count: %w", err)
		}
		dueCounts[wordlistID] = dueCount
	}

	return dueCounts, rows.Err()
}
