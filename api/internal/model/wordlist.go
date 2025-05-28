package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/pgtype"
)

type UserStats struct {
	TotalWords   int `json:"totalWords"`
	Wordlists    int `json:"wordlists"`
	WordsLearned int `json:"wordsLearned"`
}

type Wordlist struct {
	ID           int64              `json:"id"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	CreatedAt    pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt    pgtype.Timestamptz `json:"updatedAt"`
	UserID       int64              `json:"userId"`
	LanguageCode string             `json:"languageCode"`
	WordsCount   int                `json:"wordsCount"`
}

func (w Wordlist) MarshalJSON() ([]byte, error) {
	createdAt := "null"
	updatedAt := "null"

	if w.CreatedAt.Status == pgtype.Present {
		createdAt = w.CreatedAt.Time.UTC().Format(time.RFC3339)
	}

	if w.UpdatedAt.Status == pgtype.Present {
		updatedAt = w.UpdatedAt.Time.UTC().Format(time.RFC3339)
	}

	return []byte(fmt.Sprintf(`{
        "id": %d,
        "wordsCount": %d,
        "name": %q,
        "description": %q,
        "languageCode": %q,
        "createdAt": %q,
        "updatedAt": %q,
        "userId": %d
    }`, w.ID, w.WordsCount, w.Name, w.Description, w.LanguageCode, createdAt, updatedAt, w.UserID)), nil
}

func (w *Wordlist) UnmarshalJSON(data []byte) error {
	type Alias Wordlist // Create an alias to avoid recursion

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
