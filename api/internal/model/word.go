package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/pgtype"
)

type Word struct {
	ID         int64              `json:"id"`
	Name       string             `json:"name"`
	CreatedAt  pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt  pgtype.Timestamptz `json:"updatedAt"`
	WordlistID int64              `json:"wordlistId"`
	UserID     int64              `json:"userId"`
	AudioURL   string             `json:"audioUrl"`
}

func (w Word) MarshalJSON() ([]byte, error) {
	createdAt := "null"
	updatedAt := "null"

	if w.CreatedAt.Status == pgtype.Present {
		createdAt = `"` + w.CreatedAt.Time.UTC().Format(time.RFC3339) + `"`
	}

	if w.UpdatedAt.Status == pgtype.Present {
		updatedAt = `"` + w.UpdatedAt.Time.UTC().Format(time.RFC3339) + `"`
	}

	return []byte(fmt.Sprintf(`{
        "id": %d,
        "name": "%s",
        "createdAt": %s,
        "updatedAt": %s,
        "wordlistId": %d,
        "userId": %d
    }`, w.ID, w.Name, createdAt, updatedAt, w.WordlistID, w.UserID)), nil
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
