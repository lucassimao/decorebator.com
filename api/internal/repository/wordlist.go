package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WordlistRepository struct {
	Db *pgxpool.Pool
}

type Wordlist = model.Wordlist

func (repository *WordlistRepository) Save(ctx context.Context, name, description, languageCode string, userID int64, pronunciationSystem model.PronunciationSystem) (*Wordlist, error) {
	query := `
		INSERT INTO wordlists (name, description, user_id, language_code, pronunciation_system)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	var wordlistID int64
	var createdAt pgtype.Timestamptz
	var updatedAt pgtype.Timestamptz

	err := repository.Db.
		QueryRow(ctx, query, name, description, userID, languageCode, pronunciationSystem).
		Scan(&wordlistID, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	return &Wordlist{
		ID:                  wordlistID,
		Name:                name,
		Description:         description,
		LanguageCode:        languageCode,
		PronunciationSystem: pronunciationSystem,
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
		UserID:              userID,
	}, nil
}

type FindWordlistArgs struct {
	ID                       *int64
	OwnerID                  *int64
	ComputeWordsCount        bool
	ComputeWordsLearnedCount bool
}

func (repository *WordlistRepository) Find(ctx context.Context, args FindWordlistArgs) ([]*Wordlist, error) {
	var (
		builder         strings.Builder
		queryArgs       []any
		whereConditions []string
		index           int
	)

	// Build SELECT clause
	builder.WriteString("SELECT wordlists.id, wordlists.name, wordlists.description, wordlists.user_id, wordlists.created_at, wordlists.updated_at, wordlists.language_code, wordlists.pronunciation_system")
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
	if args.ID != nil {
		index = index + 1
		whereConditions = append(whereConditions, fmt.Sprintf("wordlists.id = $%d", index))
		queryArgs = append(queryArgs, *args.ID)
	}

	if args.OwnerID != nil {
		index = index + 1
		whereConditions = append(whereConditions, fmt.Sprintf("wordlists.user_id = $%d", index))
		queryArgs = append(queryArgs, *args.OwnerID)
	}

	if len(whereConditions) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(whereConditions, " AND "))
	}

	// Add GROUP BY if aggregating
	if args.ComputeWordsCount || args.ComputeWordsLearnedCount {
		builder.WriteString(" GROUP BY wordlists.id")
	}

	// Add ORDER BY
	builder.WriteString(" ORDER BY wordlists.id DESC")

	query := builder.String()
	rows, err := repository.Db.Query(ctx, query, queryArgs...)
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
			&w.PronunciationSystem,
		)
		if args.ComputeWordsCount {
			w.WordsCount = new(int)
			dest = append(dest, w.WordsCount)
		}
		if args.ComputeWordsLearnedCount {
			w.WordsLearnedCount = new(int)
			dest = append(dest, w.WordsLearnedCount)
		}

		if scanErr := rows.Scan(dest...); scanErr != nil {
			return nil, scanErr
		}
		wordlists = append(wordlists, &w)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return wordlists, nil
}

func (repository *WordlistRepository) Delete(ctx context.Context, wordlistID, userID int64) (int64, error) {
	query := `DELETE FROM wordlists WHERE user_id=$1 AND ID=$2`
	result, err := repository.Db.Exec(ctx, query, userID, wordlistID)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

func (repository *WordlistRepository) DeleteAll(ctx context.Context, userID int64) (int64, error) {
	query := `DELETE FROM wordlists WHERE user_id=$1`
	result, err := repository.Db.Exec(ctx, query, userID)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

type UpdateWordlistArgs struct {
	ID                  int64
	OwnerID             int64
	Name                string
	Description         string
	LanguageCode        string
	PronunciationSystem *model.PronunciationSystem
}

func (repository *WordlistRepository) Update(ctx context.Context, args UpdateWordlistArgs) (int64, error) {
	var pronunciationSystem any
	if args.PronunciationSystem != nil {
		pronunciationSystem = string(*args.PronunciationSystem)
	}

	query := `
		UPDATE wordlists
		SET name=$1,
			description=$2,
			language_code=$3::varchar,
			pronunciation_system=CASE
				WHEN $4::varchar IS NOT NULL THEN $4::varchar
				WHEN language_code=$3::varchar THEN pronunciation_system
				WHEN $3::varchar='ja' THEN 'romaji'
				WHEN $3::varchar IN ('zh', 'zh-cn', 'zh-tw') THEN 'pinyin'
				WHEN $3::varchar IN ('ko', 'kr') THEN 'hangul'
				ELSE 'ipa'
			END,
			updated_at=NOW()
		WHERE user_id=$5 AND id=$6`
	result, err := repository.Db.Exec(
		ctx,
		query,
		args.Name,
		args.Description,
		args.LanguageCode,
		pronunciationSystem,
		args.OwnerID,
		args.ID,
	)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}
