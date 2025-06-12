package analytics

import (
	"context"
	"fmt"

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
// Query: Complex multi-CTE query combining wordlist statistics and streak calculations:
// 1. wordlist_stats CTE: Aggregates word mastery and activity data per wordlist
//   - JOINs words → wordlists → word_mastery → learning_progress
//   - COUNT(DISTINCT w.id) as total_words in each wordlist
//   - COUNT(DISTINCT CASE WHEN mastery_level >= 0.8) as words_mastered (80% threshold)
//   - AVG(mastery_level) * 100 as progress_percent with NULL handling
//   - MAX(date) as last_activity_date from learning_progress
//   - Filters: learned = FALSE (active words only)
//
// 2. streaks CTE: Calculates current consecutive practice streaks using gap-and-island technique
//   - Uses ROW_NUMBER() window function to identify consecutive date groups
//   - Filters by 30-day window and active practice days (total_quiz_attempts > 0)
//   - Groups consecutive dates with same streak_group identifier
//   - Only includes current streak (ending on CURRENT_DATE)
//
// 3. Main SELECT: Combines wordlist stats with streak data using LEFT JOIN
//   - Returns: wordlist_id, name, language_code, totals, mastery stats, streak, last_activity
//
// Used for dashboard overview showing all wordlists with comprehensive progress metrics.
func (r *BatchProgressRepository) GetAllWordlistsProgress(ctx context.Context, userID int64) ([]model.WordlistProgress, error) {
	query := `
		WITH wordlist_stats AS (
			SELECT 
				w.wordlist_id,
				wl.name as wordlist_name,
				wl.language_code,
				COUNT(DISTINCT w.id) as total_words,
				COUNT(DISTINCT CASE WHEN wm.mastery_level >= 0.8 THEN w.id END) as words_mastered,
				CASE 
					WHEN COUNT(DISTINCT w.id) > 0 
					THEN ROUND(AVG(COALESCE(wm.mastery_level, 0)) * 100, 2)
					ELSE 0 
				END as progress_percent,
				MAX(lp.date) as last_activity_date
			FROM words w
			JOIN wordlists wl ON w.wordlist_id = wl.id
			LEFT JOIN word_mastery wm ON w.id = wm.word_id AND w.user_id = wm.user_id
			LEFT JOIN learning_progress lp ON w.wordlist_id = lp.wordlist_id AND w.user_id = lp.user_id
			WHERE w.user_id = $1 
				AND w.learned = FALSE
			GROUP BY w.wordlist_id, wl.name, wl.language_code
		),
		streaks AS (
			SELECT 
				wordlist_id,
				COUNT(*) as current_streak
			FROM (
				SELECT 
					wordlist_id,
					date,
					date - (ROW_NUMBER() OVER (PARTITION BY wordlist_id ORDER BY date DESC))::int AS streak_group
				FROM learning_progress
				WHERE user_id = $1 
					AND total_quiz_attempts > 0
					AND date >= CURRENT_DATE - INTERVAL '30 days'
			) streak_calc
			WHERE streak_group = (
				SELECT streak_group 
				FROM (
					SELECT 
						date - (ROW_NUMBER() OVER (ORDER BY date DESC))::int AS streak_group
					FROM learning_progress
					WHERE user_id = $1 
						AND wordlist_id = streak_calc.wordlist_id
						AND total_quiz_attempts > 0
				) sg
				WHERE date = CURRENT_DATE
				LIMIT 1
			)
			GROUP BY wordlist_id
		)
		SELECT 
			ws.wordlist_id,
			ws.wordlist_name,
			ws.language_code,
			ws.total_words,
			ws.words_mastered,
			ws.progress_percent,
			COALESCE(s.current_streak, 0) as current_streak,
			ws.last_activity_date
		FROM wordlist_stats ws
		LEFT JOIN streaks s ON ws.wordlist_id = s.wordlist_id
		ORDER BY ws.wordlist_id;
	`

	rows, err := r.db.Query(ctx, query, userID)
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
