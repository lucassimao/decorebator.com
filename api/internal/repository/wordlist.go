package repository

import (
	"context"
	"errors"
	"fmt"
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
	Id                       *int64
	OwnerId                  *int64
	ComputeWordsCount        bool
	ComputeWordsLearnedCount bool
}

func (repository *WordlistRepository) Find(args FindWordlistArgs) ([]*Wordlist, error) {
	var (
		builder   strings.Builder
		queryArgs []interface{}
	)

	// Build SELECT clause
	builder.WriteString("SELECT wordlists.id, wordlists.name, wordlists.description, wordlists.user_id, wordlists.created_at, wordlists.updated_at, wordlists.language_code")
	if args.ComputeWordsCount || args.ComputeWordsLearnedCount {
		if args.ComputeWordsCount {
			builder.WriteString(", COUNT(words.id) AS word_count")
		}
		if args.ComputeWordsLearnedCount {
			builder.WriteString(", COUNT(words.id) FILTER (WHERE words.learned = true) AS word_learned_count")
		}
		builder.WriteString(" FROM wordlists")
		builder.WriteString(" LEFT JOIN words ON words.wordlist_id = wordlists.id")
	} else {
		builder.WriteString(" FROM wordlists")
	}

	// Build WHERE clause
	if args.Id != nil {
		builder.WriteString(" WHERE wordlists.id = $1")
		queryArgs = append(queryArgs, *args.Id)
	} else if args.OwnerId != nil {
		builder.WriteString(" WHERE wordlists.user_id = $1")
		queryArgs = append(queryArgs, *args.OwnerId)
	}

	// Add GROUP BY if aggregating
	if args.ComputeWordsCount || args.ComputeWordsLearnedCount {
		builder.WriteString(" GROUP BY wordlists.id")
	}

	// Add ORDER BY
	builder.WriteString(" ORDER BY wordlists.id DESC")

	query := builder.String()
	rows, err := repository.Db.Query(context.Background(), query, queryArgs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*Wordlist{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	var wordlists []*Wordlist
	for rows.Next() {
		w := Wordlist{}

		var dest []any
		dest = append(dest,
			&w.ID,
			&w.Name,
			&w.Description,
			&w.UserID,
			&w.CreatedAt,
			&w.UpdatedAt,
			&w.LanguageCode,
		)

		if args.ComputeWordsCount {
			w.WordsCount = new(int)
			dest = append(dest, w.WordsCount)
		}
		if args.ComputeWordsLearnedCount {
			w.WordsLearnedCount = new(int)
			dest = append(dest, w.WordsLearnedCount)
		}

		if err := rows.Scan(dest...); err != nil {
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

func (repository *WordlistRepository) DeleteAll(userId int64) (int64, error) {
	query := `DELETE FROM wordlists WHERE user_id=$1`
	result, err := repository.Db.Exec(context.Background(), query, userId)
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

func (repository *WordlistRepository) GetStats(userId int64) (*model.UserStats, error) {
	const statsQuery = `
    SELECT
      (SELECT COUNT(*) FROM words WHERE user_id = $1)             AS total_words,
      (SELECT COUNT(*) FROM wordlists WHERE user_id = $1)         AS wordlists,
      (SELECT COUNT(*) FROM words WHERE user_id = $1 AND learned) AS words_learned;
    `

	var s model.UserStats
	row := repository.Db.QueryRow(context.Background(), statsQuery, userId)
	if err := row.Scan(&s.TotalWords, &s.Wordlists, &s.WordsLearned); err != nil {
		return nil, fmt.Errorf("GetUserStats: scan failed: %w", err)
	}
	return &s, nil
}
