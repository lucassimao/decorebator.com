package repository

import (
	"context"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Word = model.Word

type WordRepository struct {
	Db *pgxpool.Pool
}

func (repository *WordRepository) Save(name string, userId, wordlistId int64) (*Word, error) {
	query := `
		INSERT INTO words (name, wordlist_id, user_id, created_at)
		VALUES ($1, $2,$3, now())
		RETURNING id, created_at, updated_at`

	var createdAt pgtype.Timestamp
	var updatedAt pgtype.Timestamp
	var wordID int64

	err := repository.Db.QueryRow(context.Background(), query, name, wordlistId, userId).Scan(&wordID, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	return &Word{wordID, name, createdAt, updatedAt, wordlistId, userId, ""}, nil
}

func (repository *WordRepository) GetAllFromWordlist(wordlistId, userId int64) ([]Word, error) {
	query := `SELECT id , name, created_at, updated_At, COALESCE(audio_url,'') FROM words WHERE wordlist_id=$1 AND user_id=$2 order by id desc`
	rows, err := repository.Db.Query(context.Background(), query, wordlistId, userId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	words := []Word{}
	for rows.Next() {
		w := Word{WordlistID: wordlistId, UserID: userId}
		err := rows.Scan(&w.ID, &w.Name, &w.CreatedAt, &w.UpdatedAt, &w.AudioURL)
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
	query := `SELECT id , name, created_at, updated_At, wordlist_id, user_id, COALESCE(audio_url,'') FROM words WHERE id=$1`
	row := repository.Db.QueryRow(context.Background(), query, wordId)
	var w Word

	err := row.Scan(&w.ID, &w.Name, &w.CreatedAt, &w.UpdatedAt, &w.WordlistID, &w.UserID, &w.AudioURL)
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

func (repository *WordRepository) Update(word *Word) (int64, error) {
	query := `UPDATE words SET name=$1, updated_at=NOW(), audio_url=$4 WHERE user_id=$2 AND ID=$3`
	result, err := repository.Db.Exec(context.Background(), query, word.Name, word.UserID, word.ID, word.AudioURL)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}
