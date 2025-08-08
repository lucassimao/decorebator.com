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

// GetPublicQuizBySlug retrieves a public quiz by its slug
func (r *PublicQuizRepository) GetPublicQuizBySlug(ctx context.Context, slug string) (*model.PublicQuiz, error) {
	query := `
		SELECT 
            pq.id, pq.slug, pq.wordlist_id, pq.creator_id, pq.title, pq.description,
            pq.difficulty, pq.time_limit_minutes, pq.preview_image_url,
			pq.play_count, pq.share_count, pq.average_score, pq.is_active,
			pq.published_at, pq.created_at, pq.updated_at,
			u.first_name || ' ' || u.last_name as creator_name
		FROM public_quizzes pq
		JOIN users u ON pq.creator_id = u.id
		WHERE pq.slug = $1 AND pq.is_active = true
	`

	quiz := &model.PublicQuiz{}
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&quiz.ID, &quiz.Slug, &quiz.WordlistID, &quiz.CreatorID, &quiz.Title, &quiz.Description,
		&quiz.Difficulty, &quiz.TimeLimitMinutes, &quiz.PreviewImageURL,
		&quiz.PlayCount, &quiz.ShareCount, &quiz.AverageScore, &quiz.IsActive,
		&quiz.PublishedAt, &quiz.CreatedAt, &quiz.UpdatedAt, &quiz.CreatorName,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return quiz, nil
}

// GetPublicQuizByID retrieves a public quiz by its ID
func (r *PublicQuizRepository) GetPublicQuizByID(ctx context.Context, id int64) (*model.PublicQuiz, error) {
	query := `
		SELECT 
            pq.id, pq.slug, pq.wordlist_id, pq.creator_id, pq.title, pq.description,
            pq.difficulty, pq.time_limit_minutes, pq.preview_image_url,
			pq.play_count, pq.share_count, pq.average_score, pq.is_active,
			pq.published_at, pq.created_at, pq.updated_at,
			u.first_name || ' ' || u.last_name as creator_name
		FROM public_quizzes pq
		JOIN users u ON pq.creator_id = u.id
		WHERE pq.id = $1
	`

	quiz := &model.PublicQuiz{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&quiz.ID, &quiz.Slug, &quiz.WordlistID, &quiz.CreatorID, &quiz.Title, &quiz.Description,
		&quiz.Difficulty, &quiz.TimeLimitMinutes, &quiz.PreviewImageURL,
		&quiz.PlayCount, &quiz.ShareCount, &quiz.AverageScore, &quiz.IsActive,
		&quiz.PublishedAt, &quiz.CreatedAt, &quiz.UpdatedAt, &quiz.CreatorName,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return quiz, nil
}

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

// CreateQuizAttempt creates a new quiz attempt
// CreateQuizAttempt omitted for MVP

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
