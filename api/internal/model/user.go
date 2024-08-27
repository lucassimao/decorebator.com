package model

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/pgtype"
)

type User struct {
	ID           int64            `json:"id"`
	FirstName    string           `json:"firstName"`
	LastName     string           `json:"lastName"`
	PasswordHash string           `json:"passwordHash"`
	Email        string           `json:"email"`
	CreatedAt    pgtype.Timestamp `json:"createdAt"`
	UpdatedAt    pgtype.Timestamp `json:"updatedAt"`
}

func (u User) MarshalJSON() ([]byte, error) {
	userMap := map[string]interface{}{
		"id":           u.ID,
		"firstName":    u.FirstName,
		"lastName":     u.LastName,
		"passwordHash": u.PasswordHash,
		"email":        u.Email,
	}

	if u.CreatedAt.Status == pgtype.Present {
		userMap["createdAt"] = u.CreatedAt.Time.UTC().Format(time.RFC3339)
	}

	if u.UpdatedAt.Status == pgtype.Present {
		userMap["updatedAt"] = u.UpdatedAt.Time.UTC().Format(time.RFC3339)
	}

	return json.Marshal(userMap)
}
