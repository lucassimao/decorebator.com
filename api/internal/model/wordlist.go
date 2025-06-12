package model

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/pgtype"
)

type Wordlist struct {
	ID           int64              `json:"id"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	CreatedAt    pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt    pgtype.Timestamptz `json:"updatedAt"`
	UserID       int64              `json:"userId"`
	LanguageCode string             `json:"languageCode"`

	// Computed dinamically based on words table
	WordsCount        *int `json:"wordsCount"`
	WordsLearnedCount *int `json:"wordsLearnedCount"`
}

func (w Wordlist) MarshalJSON() ([]byte, error) {
	var (
		createdAtValue any
		updatedAtValue any
	)
	if w.CreatedAt.Status == pgtype.Present {
		createdAtValue = w.CreatedAt.Time.UTC().Format(time.RFC3339)
	}
	if w.UpdatedAt.Status == pgtype.Present {
		updatedAtValue = w.UpdatedAt.Time.UTC().Format(time.RFC3339)
	}

	m := map[string]any{
		"id":           w.ID,
		"name":         w.Name,
		"description":  w.Description,
		"languageCode": w.LanguageCode,
		"createdAt":    createdAtValue, // será string ou nil → "null"
		"updatedAt":    updatedAtValue, // será string ou nil → "null"
		"userId":       w.UserID,
	}

	if w.WordsCount != nil {
		m["wordsCount"] = *w.WordsCount
	}
	if w.WordsLearnedCount != nil {
		m["wordsLearnedCount"] = *w.WordsLearnedCount
	}

	return json.Marshal(m)
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
