package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/pgtype"
)

type Word struct {
	ID                 int64              `json:"id"`
	Name               string             `json:"name"`
	CreatedAt          pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt          pgtype.Timestamptz `json:"updatedAt"`
	WordlistID         int64              `json:"wordlistId"`
	UserID             int64              `json:"userId"`
	AudioURL           string             `json:"audioURL"`
	Notes              string             `json:"notes"`
	Pronunciation      string             `json:"pronunciation"`
	Learned            bool               `json:"learned"`
	ContentStatus      ContentStatus      `json:"contentStatus"`
	FlaggedReason      *string            `json:"flaggedReason,omitempty"`
	ContentReviewedAt  *pgtype.Timestamptz `json:"contentReviewedAt,omitempty"`
}

func (w Word) MarshalJSON() ([]byte, error) {

	createdAt := "null"
	updatedAt := "null"

	if w.CreatedAt.Status == pgtype.Present {
		createdAt = w.CreatedAt.Time.UTC().Format(time.RFC3339)
	}

	if w.UpdatedAt.Status == pgtype.Present {
		updatedAt = w.UpdatedAt.Time.UTC().Format(time.RFC3339)
	}

	audioURL := ""
	if w.AudioURL != "" {
		audioURL = w.AudioURL
	}

	contentReviewedAt := "null"
	if w.ContentReviewedAt != nil && w.ContentReviewedAt.Status == pgtype.Present {
		contentReviewedAt = w.ContentReviewedAt.Time.UTC().Format(time.RFC3339)
	}

	flaggedReason := "null"
	if w.FlaggedReason != nil {
		flaggedReason = fmt.Sprintf("%q", *w.FlaggedReason)
	}

	return []byte(fmt.Sprintf(`{
        "id": %d,
        "name": %q,
        "createdAt": %q,
        "audioURL": %q,
        "learned": %v,
        "updatedAt": %q,
        "wordlistId": %d,
        "notes": %q,
		"pronunciation": %q,
        "userId": %d,
		"contentStatus": %q,
		"flaggedReason": %s,
		"contentReviewedAt": %q
    }`, w.ID, w.Name, createdAt, audioURL, w.Learned, updatedAt, w.WordlistID, w.Notes, w.Pronunciation, w.UserID, w.ContentStatus, flaggedReason, contentReviewedAt)), nil
}

func (w *Word) UnmarshalJSON(data []byte) error {
	type Alias Word // Create an alias to avoid recursion

	aux := &struct {
		CreatedAt *string `json:"createdAt"`
		UpdatedAt *string `json:"updatedAt"`
		*Alias
	}{
		Alias: (*Alias)(w),
	}

	// Unmarshal into aux struct
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Handle CreatedAt
	if aux.CreatedAt != nil {
		createdAtTime, err := time.Parse(time.RFC3339, *aux.CreatedAt)
		if err != nil {
			return err
		}
		w.CreatedAt = pgtype.Timestamptz{Time: createdAtTime, Status: pgtype.Present}
	} else {
		w.CreatedAt = pgtype.Timestamptz{Status: pgtype.Null}
	}

	// Handle UpdatedAt
	if aux.UpdatedAt != nil {
		updatedAtTime, err := time.Parse(time.RFC3339, *aux.UpdatedAt)
		if err != nil {
			return err
		}
		w.UpdatedAt = pgtype.Timestamptz{Time: updatedAtTime, Status: pgtype.Present}
	} else {
		w.UpdatedAt = pgtype.Timestamptz{Status: pgtype.Null}
	}

	return nil
}
