package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/pgtype"
)

type Word struct {
	ID                    int64              `json:"id"`
	Name                  string             `json:"name"`
	CreatedAt             pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt             pgtype.Timestamptz `json:"updatedAt"`
	WordlistID            int64              `json:"wordlistId"`
	UserID                int64              `json:"userId"`
	AudioURL              string             `json:"audioURL"`
	Notes                 string             `json:"notes"`
	Pronunciation         string             `json:"pronunciation"`
	Learned               bool               `json:"learned"`
	ProcessingStatus      string             `json:"processingStatus"`
	ProcessingError       string             `json:"processingError"`
	ProcessingStartedAt   pgtype.Timestamptz `json:"processingStartedAt"`
	ProcessingCompletedAt pgtype.Timestamptz `json:"processingCompletedAt"`
}

func (w Word) MarshalJSON() ([]byte, error) {
	createdAt := "null"
	updatedAt := "null"
	processingStartedAt := "null"
	processingCompletedAt := "null"

	if w.CreatedAt.Status == pgtype.Present {
		createdAt = w.CreatedAt.Time.UTC().Format(time.RFC3339)
	}

	if w.UpdatedAt.Status == pgtype.Present {
		updatedAt = w.UpdatedAt.Time.UTC().Format(time.RFC3339)
	}

	if w.ProcessingStartedAt.Status == pgtype.Present {
		processingStartedAt = w.ProcessingStartedAt.Time.UTC().Format(time.RFC3339)
	}

	if w.ProcessingCompletedAt.Status == pgtype.Present {
		processingCompletedAt = w.ProcessingCompletedAt.Time.UTC().Format(time.RFC3339)
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
        "userId": %d,
        "processingStatus": %q,
        "processingError": %q,
        "processingStartedAt": %q,
        "processingCompletedAt": %q
    }`, w.ID, w.Name, createdAt, audioURL, w.Learned, updatedAt, w.WordlistID, w.Notes, w.Pronunciation, w.UserID,
		w.ProcessingStatus, w.ProcessingError, processingStartedAt, processingCompletedAt)), nil
}

func (w *Word) UnmarshalJSON(data []byte) error {
	type Alias Word // Create an alias to avoid recursion

	aux := &struct {
		CreatedAt             *string `json:"createdAt"`
		UpdatedAt             *string `json:"updatedAt"`
		ProcessingStartedAt   *string `json:"processingStartedAt"`
		ProcessingCompletedAt *string `json:"processingCompletedAt"`
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

	// Handle ProcessingStartedAt
	if aux.ProcessingStartedAt != nil {
		processingStartedAtTime, err := time.Parse(time.RFC3339, *aux.ProcessingStartedAt)
		if err != nil {
			return err
		}
		w.ProcessingStartedAt = pgtype.Timestamptz{Time: processingStartedAtTime, Status: pgtype.Present}
	} else {
		w.ProcessingStartedAt = pgtype.Timestamptz{Status: pgtype.Null}
	}

	// Handle ProcessingCompletedAt
	if aux.ProcessingCompletedAt != nil {
		processingCompletedAtTime, err := time.Parse(time.RFC3339, *aux.ProcessingCompletedAt)
		if err != nil {
			return err
		}
		w.ProcessingCompletedAt = pgtype.Timestamptz{Time: processingCompletedAtTime, Status: pgtype.Present}
	} else {
		w.ProcessingCompletedAt = pgtype.Timestamptz{Status: pgtype.Null}
	}

	return nil
}
