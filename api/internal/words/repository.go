package words

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Word struct {
	ID         int64            `json:"id"`
	Name       string           `json:"name"`
	CreatedAt  pgtype.Timestamp `json:"createdAt"`
	UpdatedAt  pgtype.Timestamp `json:"updatedAt"`
	WordlistID int64            `json:"wordlistId"`
	UserID     int64            `json:"userId"`
}

type WordRepository struct {
	db *pgxpool.Pool
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

func (repository *WordRepository) save(name string, userId, wordlistId int64) (*Word, error) {
	query := `
		INSERT INTO words (name, wordlist_id, user_id, created_at)
		VALUES ($1, $2,$3, now())
		RETURNING id, created_at, updated_at`

	var createdAt pgtype.Timestamp
	var updatedAt pgtype.Timestamp
	var wordID int64

	err := repository.db.QueryRow(context.Background(), query, name, wordlistId, userId).Scan(&wordID, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	return &Word{wordID, name, createdAt, updatedAt, wordlistId, userId}, nil
}

func (repository *WordRepository) getAllFromWordlist(wordlistId, userId int64) ([]*Word, error) {
	query := `SELECT id , name, created_at, updated_At FROM words WHERE wordlist_id=$1 AND user_id=$2`
	rows, err := repository.db.Query(context.Background(), query, wordlistId, userId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	words := []*Word{}
	for rows.Next() {
		w := Word{WordlistID: wordlistId, UserID: userId}
		err := rows.Scan(&w.ID, &w.Name, &w.CreatedAt, &w.UpdatedAt)
		if err != nil {
			return nil, err
		}
		words = append(words, &w)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return words, nil
}

func (repository *WordRepository) delete(userId, wordID int64) (int64, error) {
	query := `DELETE FROM words WHERE user_id=$1 AND id=$2`
	result, err := repository.db.Exec(context.Background(), query, userId, wordID)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

func (repository *WordRepository) update(word *Word) (int64, error) {
	query := `UPDATE words SET name=$1, updated_at=NOW() WHERE user_id=$2 AND ID=$3`
	result, err := repository.db.Exec(context.Background(), query, word.Name, word.UserID, word.ID)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}
