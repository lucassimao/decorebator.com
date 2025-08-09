package repository

import (
	"context"
	"fmt"
	"strings"

	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PublicQuizRepository struct {
	db *pgxpool.Pool
}

func NewPublicQuizRepository(db *pgxpool.Pool) *PublicQuizRepository {
	return &PublicQuizRepository{db: db}
}

// CreatePublicQuiz creates a new public quiz
func (r *PublicQuizRepository) CreatePublicQuiz(ctx context.Context, quiz *model.PublicQuiz) error {
	query := `
		INSERT INTO public_quizzes (
            slug, wordlist_id, creator_id, title, description, difficulty,
            time_limit_minutes, preview_image_url
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, published_at, created_at, updated_at
	`

	return r.db.QueryRow(ctx, query,
		quiz.Slug, quiz.WordlistID, quiz.CreatorID, quiz.Title, quiz.Description,
		quiz.Difficulty, quiz.TimeLimitMinutes, quiz.PreviewImageURL,
	).Scan(&quiz.ID, &quiz.PublishedAt, &quiz.CreatedAt, &quiz.UpdatedAt)
}

// ReactivateAndUpdatePublicQuiz reactivates an existing quiz and updates basic fields, keeping slug immutable
func (r *PublicQuizRepository) ReactivateAndUpdatePublicQuiz(ctx context.Context, id int64, title, description string, difficulty model.QuizDifficulty, timeLimit int) error {
	query := `
        UPDATE public_quizzes
        SET title = $1,
            description = $2,
            difficulty = $3,
            time_limit_minutes = $4,
            is_active = true,
            updated_at = NOW()
        WHERE id = $5`
	_, err := r.db.Exec(ctx, query, title, description, difficulty, timeLimit, id)
	return err
}

// Lookups should use Find with filters instead of bespoke methods

// GetUserPublicQuizzes retrieves all public quizzes created by a user
// GetUserPublicQuizzes omitted for MVP

// UpdatePublicQuiz updates a public quiz's settings
// UpdatePublicQuiz omitted for MVP

// DeactivatePublicQuiz deactivates a public quiz
func (r *PublicQuizRepository) DeactivatePublicQuiz(ctx context.Context, id int64) error {
	query := `
		UPDATE public_quizzes 
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// UpdatePreviewImageURL sets the preview image URL for a public quiz
func (r *PublicQuizRepository) UpdatePreviewImageURL(ctx context.Context, id int64, url string) error {
	query := `
        UPDATE public_quizzes
        SET preview_image_url = $1,
            updated_at = NOW()
        WHERE id = $2`
	_, err := r.db.Exec(ctx, query, url, id)
	return err
}

// Find pattern to retrieve public quizzes with flexible filters
type FindPublicQuizArgs struct {
	ID         *int64
	Slug       *string
	WordlistID *int64
	CreatorID  *int64
	OnlyActive *bool
	Limit      *int
}

func (r *PublicQuizRepository) Find(ctx context.Context, args FindPublicQuizArgs) ([]*model.PublicQuiz, error) {
	var builder strings.Builder
	var where []string
	var queryArgs []any
	idx := 0

	builder.WriteString(`SELECT pq.id, pq.slug, pq.wordlist_id, pq.creator_id, pq.title, pq.description,
        pq.difficulty, pq.time_limit_minutes, pq.preview_image_url, pq.play_count, pq.share_count,
        pq.average_score, pq.is_active, pq.published_at, pq.created_at, pq.updated_at,
        (u.first_name || ' ' || u.last_name) AS creator_name
        FROM public_quizzes pq
        LEFT JOIN users u ON pq.creator_id = u.id`)

	if args.ID != nil {
		idx++
		where = append(where, fmt.Sprintf("pq.id = $%d", idx))
		queryArgs = append(queryArgs, *args.ID)
	}
	if args.Slug != nil {
		idx++
		where = append(where, fmt.Sprintf("pq.slug = $%d", idx))
		queryArgs = append(queryArgs, *args.Slug)
	}
	if args.WordlistID != nil {
		idx++
		where = append(where, fmt.Sprintf("pq.wordlist_id = $%d", idx))
		queryArgs = append(queryArgs, *args.WordlistID)
	}
	if args.CreatorID != nil {
		idx++
		where = append(where, fmt.Sprintf("pq.creator_id = $%d", idx))
		queryArgs = append(queryArgs, *args.CreatorID)
	}
	if args.OnlyActive != nil {
		idx++
		where = append(where, fmt.Sprintf("pq.is_active = $%d", idx))
		queryArgs = append(queryArgs, *args.OnlyActive)
	}
	if len(where) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(where, " AND "))
	}
	builder.WriteString(" ORDER BY pq.id DESC")
	if args.Limit != nil && *args.Limit > 0 {
		builder.WriteString(fmt.Sprintf(" LIMIT %d", *args.Limit))
	}

	query := builder.String()
	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return []*model.PublicQuiz{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	var quizzes []*model.PublicQuiz
	for rows.Next() {
		q := &model.PublicQuiz{}
		if err := rows.Scan(
			&q.ID, &q.Slug, &q.WordlistID, &q.CreatorID, &q.Title, &q.Description,
			&q.Difficulty, &q.TimeLimitMinutes, &q.PreviewImageURL, &q.PlayCount, &q.ShareCount,
			&q.AverageScore, &q.IsActive, &q.PublishedAt, &q.CreatedAt, &q.UpdatedAt,
			&q.CreatorName,
		); err != nil {
			return nil, err
		}
		quizzes = append(quizzes, q)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return quizzes, nil
}

// Leaderboard entry DTO
type LeaderboardEntry struct {
	GuestName             string
	ScorePercentage       float64
	CompletionTimeSeconds int
}

// GetLeaderboardTopByQuizID returns top N attempts ordered by score desc, then time asc, then completed_at desc
func (r *PublicQuizRepository) GetLeaderboardTopByQuizID(ctx context.Context, quizID int64, limit int) ([]LeaderboardEntry, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	rows, err := r.db.Query(ctx, `
        SELECT COALESCE(guest_name, ''), score_percentage, completion_time_seconds
        FROM quiz_attempts
        WHERE public_quiz_id = $1
        ORDER BY score_percentage DESC, completion_time_seconds ASC, completed_at DESC
        LIMIT $2
    `, quizID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []LeaderboardEntry
	for rows.Next() {
		var e LeaderboardEntry
		if scanErr := rows.Scan(&e.GuestName, &e.ScorePercentage, &e.CompletionTimeSeconds); scanErr != nil {
			return nil, scanErr
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

// InsertQuizAttempt records a new attempt row for leaderboard (guest play)
func (r *PublicQuizRepository) InsertQuizAttempt(ctx context.Context, quizID int64, guestName *string, guestEmail *string, scorePercentage float64, accuracyPercentage float64, completionTimeSeconds int) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO quiz_attempts (public_quiz_id, player_id, guest_name, guest_email, score_percentage, accuracy_percentage, completion_time_seconds)
        VALUES ($1, NULL, $2, $3, $4, $5, $6)
        ON CONFLICT (public_quiz_id, guest_email)
        WHERE guest_email IS NOT NULL
        DO UPDATE SET
            guest_name = COALESCE(EXCLUDED.guest_name, quiz_attempts.guest_name),
            score_percentage = GREATEST(quiz_attempts.score_percentage, EXCLUDED.score_percentage),
            accuracy_percentage = GREATEST(quiz_attempts.accuracy_percentage, EXCLUDED.accuracy_percentage),
            completion_time_seconds = LEAST(quiz_attempts.completion_time_seconds, EXCLUDED.completion_time_seconds),
            completed_at = NOW()
    `, quizID, guestName, guestEmail, scorePercentage, accuracyPercentage, completionTimeSeconds)
	return err
}

// UpdateAggregatesForAttempt updates play_count and average_score on public_quizzes
func (r *PublicQuizRepository) UpdateAggregatesForAttempt(ctx context.Context, quizID int64, scorePct float64) error {
	_, err := r.db.Exec(ctx, `
        UPDATE public_quizzes
        SET play_count = play_count + 1,
            average_score = CASE 
                WHEN play_count = 0 THEN ($1::float)
                WHEN average_score IS NULL THEN ($1::float)
                ELSE ((average_score * play_count) + $1::float) / (play_count + 1)
            END,
            updated_at = NOW()
        WHERE id = $2
    `, scorePct, quizID)
	return err
}

// GetCreatorEntitlement retrieves creator entitlements for a user
// Creator entitlement methods omitted for MVP

// UpdateCreatorEntitlement updates creator entitlements
// Creator entitlement methods omitted for MVP

// GetWeeklyLeaderboard retrieves the weekly leaderboard for a quiz
// Leaderboard methods omitted for MVP

// GetAllTimeLeaderboard retrieves the all-time leaderboard for a quiz
// Leaderboard methods omitted for MVP

// IncrementShareCount increments the share count for a quiz
// IncrementShareCount omitted for MVP

// GenerateUniqueSlug generates a unique slug for a quiz title
func (r *PublicQuizRepository) GenerateUniqueSlug(ctx context.Context, title string) (string, error) {
	baseSlug := generateSlugFromTitle(title)
	slug := baseSlug
	counter := 1

	for {
		exists, err := r.slugExists(ctx, slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}
}

// slugExists checks if a slug already exists
func (r *PublicQuizRepository) slugExists(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM public_quizzes WHERE slug = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, slug).Scan(&exists)
	return exists, err
}

// generateSlugFromTitle creates a URL-friendly slug from a title
func generateSlugFromTitle(title string) string {
	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters except hyphens
	var result strings.Builder
	for _, char := range slug {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}

	// Remove multiple consecutive hyphens
	slug = result.String()
	slug = strings.ReplaceAll(slug, "--", "-")
	slug = strings.Trim(slug, "-")

	// Limit length
	if len(slug) > 50 {
		slug = slug[:50]
		slug = strings.TrimRight(slug, "-")
	}

	return slug
}
