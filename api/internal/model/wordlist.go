package model

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/pgtype"
)

type Wordlist struct {
	ID          int64            `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	CreatedAt   pgtype.Timestamp `json:"createdAt"`
	UpdatedAt   pgtype.Timestamp `json:"updatedAt"`
	UserID      int64            `json:"userId"`
}

func (w Wordlist) MarshalJSON() ([]byte, error) {
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
        "description": "%s",
        "createdAt": %s,
        "updatedAt": %s,
        "userId": %d
    }`, w.ID, w.Name, w.Description, createdAt, updatedAt, w.UserID)), nil
}
