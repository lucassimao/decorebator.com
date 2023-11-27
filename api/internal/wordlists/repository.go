package wordlists

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Wordlist struct {
	ID          int64            `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	CreatedAt   pgtype.Timestamp `json:"createdAt"`
	UpdatedAt   pgtype.Timestamp `json:"updatedAt"`
	UserID      int64            `json:"userId"`
}

type WordlistRepository struct {
	db *pgxpool.Pool
}

func (w Wordlist) MarshalJSON() ([]byte, error) {
	type Alias Wordlist
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

func (repository *WordlistRepository) save(name, description string, userID int64) (*Wordlist, error) {
	query := `
		INSERT INTO wordlists (name, description, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	var wordlistID int64
	var createdAt pgtype.Timestamp
	var updatedAt pgtype.Timestamp

	err := repository.db.QueryRow(context.Background(), query, name, description, userID).Scan(&wordlistID, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	return &Wordlist{wordlistID, name, description, createdAt, updatedAt, userID}, nil
}

type FindArgs struct {
	id      *int64
	ownerId *int64
}

func (repository *WordlistRepository) find(args FindArgs) ([]*Wordlist, error) {
	var builder strings.Builder
	builder.WriteString(`SELECT id,name, description, user_id, created_at, updated_at FROM wordlists`)
	var queryArgs []interface{}

	if args.id != nil {
		builder.WriteString(` WHERE id = $1`)
		queryArgs = append(queryArgs, args.id)
	} else if args.ownerId != nil {
		builder.WriteString(` WHERE ownerId = $1`)
		queryArgs = append(queryArgs, args.ownerId)
	}

	wordlists := []*Wordlist{}
	query := builder.String()
	rows, err := repository.db.Query(context.Background(), query, queryArgs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return wordlists, nil
		}
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		w := Wordlist{}
		err := rows.Scan(&w.ID, &w.Name, &w.Description, &w.UserID, &w.CreatedAt, &w.UpdatedAt)
		if err != nil {
			return nil, err
		}
		wordlists = append(wordlists, &w)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return wordlists, nil
}

func (repository *WordlistRepository) delete(userId, wordlistID int64) (int64, error) {
	query := `DELETE FROM wordlists WHERE user_id=$1 AND ID=$2`
	result, err := repository.db.Exec(context.Background(), query, userId, wordlistID)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

func (repository *WordlistRepository) update(wordlist *Wordlist) (int64, error) {
	query := `UPDATE wordlists SET name=$1, description=$2, updated_at=NOW() WHERE user_id=$3 AND ID=$4`
	result, err := repository.db.Exec(context.Background(), query, wordlist.Name, wordlist.Description, wordlist.UserID, wordlist.ID)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}
