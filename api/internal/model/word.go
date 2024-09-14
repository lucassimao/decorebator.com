package model

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/pgtype"
)

type Word struct {
	ID         int64            `json:"id"`
	Name       string           `json:"name"`
	CreatedAt  pgtype.Timestamp `json:"createdAt"`
	UpdatedAt  pgtype.Timestamp `json:"updatedAt"`
	WordlistID int64            `json:"wordlistId"`
	UserID     int64            `json:"userId"`
	AudioURL   string           `json:"audioUrl"`
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
