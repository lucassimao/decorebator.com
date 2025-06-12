package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/pgtype"
)

// parseTimestamp converts a string timestamp to pgtype.Timestamptz
func parseTimestamp(timeStr *string) (pgtype.Timestamptz, error) {
	if timeStr != nil {
		parsedTime, err := time.Parse(time.RFC3339, *timeStr)
		if err != nil {
			return pgtype.Timestamptz{}, err
		}
		return pgtype.Timestamptz{Time: parsedTime, Status: pgtype.Present}, nil
	}
	return pgtype.Timestamptz{Status: pgtype.Null}, nil
}

type Word struct {
	ID            int64              `json:"id"`
	Name          string             `json:"name"`
	CreatedAt     pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt     pgtype.Timestamptz `json:"updatedAt"`
	WordlistID    int64              `json:"wordlistId"`
	UserID        int64              `json:"userId"`
	AudioURL      string             `json:"audioURL"`
	Notes         string             `json:"notes"`
	Pronunciation string             `json:"pronunciation"`
	Learned       bool               `json:"learned"`
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
        "userId": %d
    }`, w.ID, w.Name, createdAt, audioURL, w.Learned, updatedAt, w.WordlistID, w.Notes, w.Pronunciation, w.UserID)), nil
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
	var err error
	w.CreatedAt, err = parseTimestamp(aux.CreatedAt)
	if err != nil {
		return err
	}

	// Handle UpdatedAt
	w.UpdatedAt, err = parseTimestamp(aux.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}
