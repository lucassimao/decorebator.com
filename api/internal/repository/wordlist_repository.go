package repository

import (
	"context"
	"errors"
	"strings"

	"decorebator.com/internal/model"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WordlistRepository struct {
	Db *pgxpool.Pool
}

type Wordlist = model.Wordlist

func (repository *WordlistRepository) Save(name, description string, userID int64) (*Wordlist, error) {
	query := `
		INSERT INTO wordlists (name, description, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	var wordlistID int64
	var createdAt pgtype.Timestamp
	var updatedAt pgtype.Timestamp

	err := repository.Db.QueryRow(context.Background(), query, name, description, userID).Scan(&wordlistID, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	return &Wordlist{wordlistID, name, description, createdAt, updatedAt, userID}, nil
}

type FindWordlistArgs struct {
	Id      *int64
	OwnerId *int64
}

func (repository *WordlistRepository) Find(args FindWordlistArgs) ([]*Wordlist, error) {
	var builder strings.Builder
	builder.WriteString(`SELECT id,name, description, user_id, created_at, updated_at FROM wordlists`)
	var queryArgs []interface{}

	if args.Id != nil {
		builder.WriteString(` WHERE id = $1`)
		queryArgs = append(queryArgs, args.Id)
	} else if args.OwnerId != nil {
		builder.WriteString(` WHERE user_id = $1`)
		queryArgs = append(queryArgs, args.OwnerId)
	}

	wordlists := []*Wordlist{}
	query := builder.String()
	rows, err := repository.Db.Query(context.Background(), query, queryArgs...)
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

func (repository *WordlistRepository) Delete(wordlistID, userId int64) (int64, error) {
	query := `DELETE FROM wordlists WHERE user_id=$1 AND ID=$2`
	result, err := repository.Db.Exec(context.Background(), query, userId, wordlistID)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

func (repository *WordlistRepository) Update(wordlist *Wordlist) (int64, error) {
	query := `UPDATE wordlists SET name=$1, description=$2, updated_at=NOW() WHERE user_id=$3 AND ID=$4`
	result, err := repository.Db.Exec(context.Background(), query, wordlist.Name, wordlist.Description, wordlist.UserID, wordlist.ID)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}
