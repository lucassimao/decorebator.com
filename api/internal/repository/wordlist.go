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

func (repository *WordlistRepository) Save(name, description, languageCode string, userID int64) (*Wordlist, error) {
	query := `
		INSERT INTO wordlists (name, description, user_id,language_code)
		VALUES ($1, $2, $3,$4)
		RETURNING id, created_at, updated_at`

	var wordlistID int64
	var createdAt pgtype.Timestamptz
	var updatedAt pgtype.Timestamptz

	err := repository.Db.
		QueryRow(context.Background(), query, name, description, userID, languageCode).
		Scan(&wordlistID, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	return &Wordlist{ID: wordlistID, Name: name, Description: description,
		CreatedAt: createdAt, UpdatedAt: updatedAt, UserID: userID}, nil
}

type FindWordlistArgs struct {
	Id      *int64
	OwnerId *int64
}

func (repository *WordlistRepository) Find(args FindWordlistArgs) ([]*Wordlist, error) {
	var builder strings.Builder
	builder.WriteString(`SELECT id, name, description, user_id, created_at, updated_at, language_code, words_count FROM wordlists`)
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
		err := rows.Scan(&w.ID, &w.Name, &w.Description, &w.UserID, &w.CreatedAt, &w.UpdatedAt, &w.LanguageCode, &w.WordsCount)
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
	query := `UPDATE wordlists SET name=$1, description=$2,language_code=$3 updated_at=NOW() WHERE user_id=$4 AND ID=$5`
	result, err := repository.Db.Exec(context.Background(), query, wordlist.Name, wordlist.Description, wordlist.LanguageCode, wordlist.UserID, wordlist.ID)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

func (repository *WordlistRepository) updateWordCount(wordlistId int64, count int, tx *pgx.Tx) error {
	query := `UPDATE wordlists SET words_count = words_count + $1, updated_at=NOW() WHERE ID=$2`

	var err error

	if tx != nil {
		_, err = (*tx).Exec(context.Background(), query, count, wordlistId)
	} else {
		_, err = repository.Db.Exec(context.Background(), query, count, wordlistId)
	}

	return err
}

func (repository *WordlistRepository) IncWordCount(wordlistId int64, tx *pgx.Tx) error {
	return repository.updateWordCount(wordlistId, 1, tx)
}

func (repository *WordlistRepository) DecWordCount(wordlistId int64, tx *pgx.Tx) error {
	return repository.updateWordCount(wordlistId, -1, tx)
}
