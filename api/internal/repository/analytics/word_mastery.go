package analytics

import (
	"context"

	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WordMasteryRepository handles word mastery operations
type WordMasteryRepository struct {
	db *pgxpool.Pool
}

// NewWordMasteryRepository creates a new word mastery repository
func NewWordMasteryRepository(db *pgxpool.Pool) *WordMasteryRepository {
	return &WordMasteryRepository{db: db}
}

// UpsertWordMastery updates or inserts word mastery data in word_mastery table.
//
// Query: INSERT ... ON CONFLICT DO UPDATE pattern with sophisticated mastery calculation:
// - Inserts new record if user/word combination doesn't exist with initial values
// - Updates existing record using weighted mastery algorithm:
//   * mastery_level: 70% current box progress + 30% historical accuracy
//   * Box progress: boxID / 7.0 (scales Leitner box 1-7 to 0.0-1.0)
//   * On correct answer: increases mastery, increments streak
//   * On incorrect answer: reduces mastery by 0.1 (minimum 0), resets streak
//   * Tracks total_attempts, correct_attempts, max_streak
//   * Updates last_seen_at and updated_at timestamps
//
// This provides a nuanced mastery score combining current Leitner progress with historical accuracy.
func (r *WordMasteryRepository) UpsertWordMastery(ctx context.Context, tx pgx.Tx, userID, wordID int64, boxID int64, isCorrect bool) error {
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
			max_streak = CASE 
				WHEN $3 THEN 
					-- On correct answer, compare current max with new streak
					GREATEST(word_mastery.max_streak, word_mastery.streak_count + 1)
				ELSE 
					-- On incorrect answer, streak resets to 0, so max_streak stays unchanged
					word_mastery.max_streak
			END,
			last_seen_at = NOW(),
			updated_at = NOW()
	`

	_, err := tx.Exec(ctx, query, userID, wordID, isCorrect, boxID)
	return err
}

// GetWordMastery retrieves word mastery statistics for a wordlist.
//
// Query: Complex JOIN query combining word_mastery, words, and leitner_system_tracking:
// - JOINs word_mastery with words table to get word names and filter by wordlist
// - LEFT JOINs leitner_system_tracking to get highest box reached per word
// - Calculates accuracy_rate as correct_attempts / total_attempts
// - Groups by word to handle multiple definitions per word
// - Uses MAX(box_id) to get the highest Leitner box achieved for each word
// - Filters out learned words (learned = FALSE) to focus on active learning
// - Orders by mastery_level DESC, then last_seen_at DESC for best-first sorting
//
// Returns array of WordMasteryStats with comprehensive learning metrics per word.
func (r *WordMasteryRepository) GetWordMastery(ctx context.Context, userID, wordlistID int64) ([]model.WordMasteryStats, error) {
	query := `
		SELECT 
			wm.word_id,
			w.name AS word,
			wm.mastery_level,
			CASE 
				WHEN wm.total_attempts > 0 
				THEN ROUND(wm.correct_attempts::numeric / wm.total_attempts::numeric, 2)
				ELSE 0
			END AS accuracy_rate,
			wm.streak_count,
			wm.last_seen_at,
			MAX(lst.box_id) AS highest_box_reached
		FROM word_mastery wm
		JOIN words w ON wm.word_id = w.id
		LEFT JOIN leitner_system_tracking lst 
			ON lst.word_id = wm.word_id AND lst.user_id = wm.user_id
		WHERE wm.user_id = $1 
		  AND w.wordlist_id = $2 
		  AND w.learned = FALSE  -- Only show active learning words
		GROUP BY wm.word_id, w.name, wm.mastery_level, 
				 wm.total_attempts, wm.correct_attempts, wm.streak_count, wm.last_seen_at
		ORDER BY wm.mastery_level DESC, wm.last_seen_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID, wordlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []model.WordMasteryStats
	for rows.Next() {
		var s model.WordMasteryStats
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

// GetWordlistMasteryStats retrieves aggregate mastery statistics for a specific wordlist.
//
// Query: Aggregates mastery data for all words in wordlist (not just those with mastery entries):
// - COUNT(DISTINCT w.id) as total_words in the wordlist (all active learning words)
// - COUNT(DISTINCT ... WHERE mastery_level >= 0.8) as words_mastered (80% threshold)
// - AVG(mastery_level) as average mastery across words with mastery data
// - MAX(streak_count) as best_streak achieved in this wordlist
// - Uses LEFT JOIN to include all words, even those without mastery data
// - Filters out learned words (learned = FALSE) to focus on active learning
// - Filters by user_id and wordlist_id for user-specific wordlist stats
//
// Returns: totalWords, wordsMastered, avgMastery, bestStreak for dashboard display.
func (r *WordMasteryRepository) GetWordlistMasteryStats(ctx context.Context, userID, wordlistID int64) (int, int, *float64, *int, error) {
	query := `
		SELECT 
			COUNT(DISTINCT w.id) AS total_words,  -- Count all words in wordlist, not just those with mastery data
			COUNT(DISTINCT CASE WHEN wm.mastery_level >= 0.8 THEN w.id END) AS words_mastered,
			AVG(wm.mastery_level) AS avg_mastery,  -- Average only among words with mastery data
			MAX(wm.streak_count) AS best_streak
		FROM words w
		LEFT JOIN word_mastery wm ON w.id = wm.word_id AND wm.user_id = $1
		WHERE w.user_id = $1 
		  AND w.wordlist_id = $2 
		  AND w.learned = FALSE;  -- Only count active learning words
	`

	var totalWords, wordsMastered int
	var avgMastery *float64
	var bestStreak *int

	err := r.db.QueryRow(ctx, query, userID, wordlistID).Scan(
		&totalWords, &wordsMastered, &avgMastery, &bestStreak,
	)

	return totalWords, wordsMastered, avgMastery, bestStreak, err
}