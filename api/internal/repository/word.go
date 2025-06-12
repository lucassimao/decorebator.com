package repository

import (
	"context"
	"fmt"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Word = model.Word

type WordRepository struct {
	Db *pgxpool.Pool
}

func (repository *WordRepository) Save(name, notes string, userId, wordlistId int64, tx *pgx.Tx) (*Word, error) {
	query := `
		INSERT INTO words (name, wordlist_id, user_id, created_at, notes)
		VALUES ($1, $2,$3, now(),$4)
		RETURNING id, created_at, updated_at`

	var createdAt pgtype.Timestamptz
	var updatedAt pgtype.Timestamptz
	var wordID int64

	args := []any{name, wordlistId, userId, notes}

	var row pgx.Row
	if tx != nil {
		row = (*tx).QueryRow(context.Background(), query, args...)
	} else {
		row = repository.Db.QueryRow(context.Background(), query, args...)
	}

	err := row.Scan(&wordID, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	return &Word{ID: wordID, Name: name, CreatedAt: createdAt, Notes: notes,
		UpdatedAt: updatedAt, WordlistID: wordlistId, UserID: userId, AudioURL: ""}, nil
}

func (repository *WordRepository) ReuseDefinitions(wordId int64, definitionIds []int64, tx pgx.Tx) error {

	var strBuilder strings.Builder

	strBuilder.WriteString("INSERT INTO word_definitions (word_id, definition_id) VALUES ")
	parameters := []string{}
	values := []any{}
	index := 0

	for _, definitionId := range definitionIds {
		parameters = append(parameters, fmt.Sprintf("($%d, $%d)", index+1, index+2))
		values = append(values, wordId, definitionId)
		index = index + 2
	}

	strBuilder.WriteString(strings.Join(parameters, ","))
	sql := strBuilder.String()

	_, err := tx.Exec(context.Background(), sql, values...)

	if err != nil {
		return err
	}

	return nil
}

// GetWordsByWordlist returns words from wordlist with optional filtering
// onlyWithDefinitions: if true, returns only words that have definitions with meanings
func (repository *WordRepository) GetWordsByWordlist(wordlistId, userId int64, onlyWithDefinitions bool) ([]Word, error) {
	var query string

	if onlyWithDefinitions {
		query = `SELECT DISTINCT w.id, w.name, w.created_at, w.updated_at, 
					COALESCE(w.audio_url,''), COALESCE(w.notes,''), 
					COALESCE(w.pronunciation,''), w.learned
				FROM words w
				INNER JOIN word_definitions wd ON w.id = wd.word_id
				INNER JOIN definitions d ON wd.definition_id = d.id
				WHERE w.wordlist_id=$1 AND w.user_id=$2 AND d.meaning IS NOT NULL AND d.meaning != ''
				ORDER BY w.id DESC`
	} else {
		query = `SELECT id, name, created_at, updated_at, COALESCE(audio_url,''), 
					COALESCE(notes,''), COALESCE(pronunciation,''), learned
				FROM words WHERE wordlist_id=$1 AND user_id=$2 ORDER BY id DESC`
	}

	rows, err := repository.Db.Query(context.Background(), query, wordlistId, userId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	words := []Word{}
	for rows.Next() {
		w := Word{WordlistID: wordlistId, UserID: userId}
		err := rows.Scan(&w.ID, &w.Name, &w.CreatedAt, &w.UpdatedAt, &w.AudioURL, &w.Notes, &w.Pronunciation, &w.Learned)
		if err != nil {
			return nil, err
		}
		words = append(words, w)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return words, nil
}

func (repository *WordRepository) GetById(wordId int64) (*Word, error) {
	query := `SELECT id, name, created_at, updated_at, wordlist_id, user_id, 
				COALESCE(audio_url,''), COALESCE(notes,''), COALESCE(pronunciation,''), learned
			FROM words WHERE id=$1`
	row := repository.Db.QueryRow(context.Background(), query, wordId)
	var w Word

	err := row.Scan(&w.ID, &w.Name, &w.CreatedAt, &w.UpdatedAt, &w.WordlistID, &w.UserID, &w.AudioURL, &w.Notes, &w.Pronunciation, &w.Learned)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.NotFoundError{ID: wordId, Entity: "word"}
		}

		return nil, err
	}

	return &w, nil
}

func (repository *WordRepository) Delete(userId, wordID int64) (int64, error) {
	query := `DELETE FROM words WHERE user_id=$1 AND id=$2`
	result, err := repository.Db.Exec(context.Background(), query, userId, wordID)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

func (repository *WordRepository) Update(word *Word, tx *pgx.Tx) (int64, error) {
	query := `UPDATE words SET name=$1, updated_at=NOW(), audio_url=$4, wordlist_id=$5, notes=$6, learned=$7 WHERE user_id=$2 AND ID=$3`

	var result pgconn.CommandTag
	var err error

	exec := repository.Db.Exec
	if tx != nil {
		exec = (*tx).Exec
	}
	result, err = exec(context.Background(), query, word.Name, word.UserID,
		word.ID, word.AudioURL, word.WordlistID, word.Notes, word.Learned)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

func (repository *WordRepository) GetLatestAudioUrl(word string) (string, error) {
	query := `SELECT audio_url
			  FROM words 
			  WHERE name=$1 AND audio_url is not null AND LENGTH(audio_url) > 0 order by id desc`

	var audioURL string
	err := repository.Db.QueryRow(context.Background(), query, word).Scan(&audioURL)

	if err != nil {
		return "", err
	}
	return audioURL, nil
}
