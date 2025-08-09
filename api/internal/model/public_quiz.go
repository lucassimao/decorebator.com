package model

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/pgtype"
)

// QuizDifficulty represents the difficulty level of a public quiz
type QuizDifficulty string

const (
	QuizDifficultyEasy   QuizDifficulty = "easy"
	QuizDifficultyMedium QuizDifficulty = "medium"
	QuizDifficultyHard   QuizDifficulty = "hard"
)

// PublicQuiz represents a published quiz that can be played by anyone
type PublicQuiz struct {
	ID               int64              `json:"id"`
	Slug             string             `json:"slug"`
	WordlistID       int64              `json:"wordlistId"`
	CreatorID        int64              `json:"creatorId"`
	Title            string             `json:"title"`
	Description      *string            `json:"description,omitempty"`
	Difficulty       QuizDifficulty     `json:"difficulty"`
	TimeLimitMinutes int                `json:"timeLimitMinutes"`
	PreviewImageURL  *string            `json:"previewImageUrl,omitempty"`
	PlayCount        int                `json:"playCount"`
	ShareCount       int                `json:"shareCount"`
	AverageScore     *float64           `json:"averageScore,omitempty"`
	IsActive         bool               `json:"isActive"`
	PublishedAt      pgtype.Timestamptz `json:"publishedAt"`
	CreatedAt        pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt        pgtype.Timestamptz `json:"updatedAt"`

	// Related data (populated when needed)
	CreatorName *string   `json:"creatorName,omitempty"`
	Wordlist    *Wordlist `json:"wordlist,omitempty"`
}

// QuizAttempt represents a player's attempt at a public quiz
type QuizAttempt struct {
	ID                    int64              `json:"id"`
	PublicQuizID          int64              `json:"publicQuizId"`
	PlayerID              *int64             `json:"playerId,omitempty"` // NULL for anonymous
	GuestName             *string            `json:"guestName,omitempty"`
	ScorePercentage       float64            `json:"scorePercentage"`
	AccuracyPercentage    float64            `json:"accuracyPercentage"`
	CompletionTimeSeconds int                `json:"completionTimeSeconds"`
	StartedAt             pgtype.Timestamptz `json:"startedAt"`
	CompletedAt           pgtype.Timestamptz `json:"completedAt"`

	// Related data (populated when needed)
	PublicQuiz *PublicQuiz `json:"publicQuiz,omitempty"`
	Player     *User       `json:"player,omitempty"`
}

// LeaderboardEntry represents a player's ranking on a quiz
type LeaderboardEntry struct {
	QuizSlug              string  `json:"quizSlug"`
	QuizTitle             string  `json:"quizTitle"`
	PlayerID              *int64  `json:"playerId,omitempty"`
	GuestName             *string `json:"guestName,omitempty"`
	ScorePercentage       float64 `json:"scorePercentage"`
	CompletionTimeSeconds int     `json:"completionTimeSeconds"`
	Rank                  int     `json:"rank"`
}

// PublishQuizDTO represents the configuration for publishing a quiz
type PublishQuizDTO struct {
	Title            string         `json:"title" binding:"required"`
	Description      *string        `json:"description,omitempty"`
	Difficulty       QuizDifficulty `json:"difficulty" binding:"required"`
	TimeLimitMinutes int            `json:"timeLimitMinutes" binding:"required,min=1,max=15"`
}

// QuizResults represents the results of a completed quiz attempt
type QuizResults struct {
	ScorePercentage       float64 `json:"scorePercentage"`
	AccuracyPercentage    float64 `json:"accuracyPercentage"`
	CompletionTimeSeconds int     `json:"completionTimeSeconds"`
	Rank                  *int    `json:"rank,omitempty"`
	TotalPlayers          int     `json:"totalPlayers"`
}

// MarshalJSON customizes JSON output for PublicQuiz
func (pq PublicQuiz) MarshalJSON() ([]byte, error) {
	var (
		publishedAtValue any
		createdAtValue   any
		updatedAtValue   any
	)
	if pq.PublishedAt.Status == pgtype.Present {
		publishedAtValue = pq.PublishedAt.Time.UTC().Format(time.RFC3339)
	}
	if pq.CreatedAt.Status == pgtype.Present {
		createdAtValue = pq.CreatedAt.Time.UTC().Format(time.RFC3339)
	}
	if pq.UpdatedAt.Status == pgtype.Present {
		updatedAtValue = pq.UpdatedAt.Time.UTC().Format(time.RFC3339)
	}

	m := map[string]any{
		"id":               pq.ID,
		"slug":             pq.Slug,
		"wordlistId":       pq.WordlistID,
		"creatorId":        pq.CreatorID,
		"title":            pq.Title,
		"difficulty":       pq.Difficulty,
		"timeLimitMinutes": pq.TimeLimitMinutes,
		"playCount":        pq.PlayCount,
		"shareCount":       pq.ShareCount,
		"isActive":         pq.IsActive,
		"publishedAt":      publishedAtValue,
		"createdAt":        createdAtValue,
		"updatedAt":        updatedAtValue,
	}

	if pq.Description != nil {
		m["description"] = *pq.Description
	}
	if pq.PreviewImageURL != nil {
		m["previewImageUrl"] = *pq.PreviewImageURL
	}
	if pq.AverageScore != nil {
		m["averageScore"] = *pq.AverageScore
	}
	if pq.CreatorName != nil {
		m["creatorName"] = *pq.CreatorName
	}
	if pq.Wordlist != nil {
		m["wordlist"] = pq.Wordlist
	}

	return json.Marshal(m)
}

// MarshalJSON customizes JSON output for QuizAttempt
func (qa QuizAttempt) MarshalJSON() ([]byte, error) {
	var (
		startedAtValue   any
		completedAtValue any
	)
	if qa.StartedAt.Status == pgtype.Present {
		startedAtValue = qa.StartedAt.Time.UTC().Format(time.RFC3339)
	}
	if qa.CompletedAt.Status == pgtype.Present {
		completedAtValue = qa.CompletedAt.Time.UTC().Format(time.RFC3339)
	}

	m := map[string]any{
		"id":                    qa.ID,
		"publicQuizId":          qa.PublicQuizID,
		"scorePercentage":       qa.ScorePercentage,
		"accuracyPercentage":    qa.AccuracyPercentage,
		"completionTimeSeconds": qa.CompletionTimeSeconds,
		"startedAt":             startedAtValue,
		"completedAt":           completedAtValue,
	}

	if qa.PlayerID != nil {
		m["playerId"] = *qa.PlayerID
	}
	if qa.GuestName != nil {
		m["guestName"] = *qa.GuestName
	}
	if qa.PublicQuiz != nil {
		m["publicQuiz"] = qa.PublicQuiz
	}
	if qa.Player != nil {
		m["player"] = qa.Player
	}

	return json.Marshal(m)
}

// Validate checks if the public quiz settings are valid
func (pqs PublishQuizDTO) Validate() error {
	if pqs.Title == "" {
		return errors.New("title is required")
	}
	if pqs.TimeLimitMinutes < 1 || pqs.TimeLimitMinutes > 15 {
		return errors.New("time limit must be between 1 and 15 minutes")
	}
	if pqs.Difficulty != QuizDifficultyEasy && pqs.Difficulty != QuizDifficultyMedium && pqs.Difficulty != QuizDifficultyHard {
		return errors.New("invalid difficulty level")
	}
	return nil
}
