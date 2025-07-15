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

func (repository *WordRepository) Save(ctx context.Context, name, notes string, userID, wordlistID int64, tx *pgx.Tx) (*Word, error) {
	query := `
		INSERT INTO words (name, wordlist_id, user_id, created_at, notes)
		VALUES ($1, $2,$3, now(),$4)
		RETURNING id, created_at, updated_at, processing_status, processing_error, processing_started_at, processing_completed_at`

	var createdAt pgtype.Timestamptz
	var updatedAt pgtype.Timestamptz
	var processingStartedAt pgtype.Timestamptz
	var processingCompletedAt pgtype.Timestamptz
	var wordID int64
	var processingStatus string
	var processingError *string

	args := []any{name, wordlistID, userID, notes}

	var row pgx.Row
	if tx != nil {
		row = (*tx).QueryRow(ctx, query, args...)
	} else {
		row = repository.Db.QueryRow(ctx, query, args...)
	}

	err := row.Scan(&wordID, &createdAt, &updatedAt, &processingStatus, &processingError, &processingStartedAt, &processingCompletedAt)

	if err != nil {
		return nil, err
	}

	processingErrorStr := ""
	if processingError != nil {
		processingErrorStr = *processingError
	}

	return &Word{
		ID: wordID, Name: name, CreatedAt: createdAt, Notes: notes,
		UpdatedAt: updatedAt, WordlistID: wordlistID, UserID: userID, AudioURL: "",
		ProcessingStatus: processingStatus, ProcessingError: processingErrorStr,
		ProcessingStartedAt: processingStartedAt, ProcessingCompletedAt: processingCompletedAt,
	}, nil
}

func (repository *WordRepository) ReuseDefinitions(wordID int64, definitionIDs []int64, tx pgx.Tx) error {
	var strBuilder strings.Builder

	strBuilder.WriteString("INSERT INTO word_definitions (word_id, definition_id) VALUES ")
	parameters := []string{}
	values := []any{}
	index := 0

	for _, definitionID := range definitionIDs {
		parameters = append(parameters, fmt.Sprintf("($%d, $%d)", index+1, index+2))
		values = append(values, wordID, definitionID)
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
func (repository *WordRepository) GetWordsByWordlist(ctx context.Context, wordlistID, userID int64, onlyWithDefinitions bool) ([]Word, error) {
	var query string

	if onlyWithDefinitions {
		query = `SELECT DISTINCT w.id, w.name, w.created_at, w.updated_at, 
					COALESCE(w.audio_url,''), COALESCE(w.notes,''), 
					COALESCE(w.pronunciation,''), w.learned,
					w.processing_status, COALESCE(w.processing_error,''), 
					w.processing_started_at, w.processing_completed_at
				FROM words w
				INNER JOIN word_definitions wd ON w.id = wd.word_id
				INNER JOIN definitions d ON wd.definition_id = d.id
				WHERE w.wordlist_id=$1 AND w.user_id=$2 AND d.meaning IS NOT NULL AND d.meaning != ''
				ORDER BY w.id DESC`
	} else {
		query = `SELECT id, name, created_at, updated_at, COALESCE(audio_url,''), 
					COALESCE(notes,''), COALESCE(pronunciation,''), learned,
					processing_status, COALESCE(processing_error,''), 
					processing_started_at, processing_completed_at
				FROM words WHERE wordlist_id=$1 AND user_id=$2 ORDER BY id DESC`
	}

	rows, err := repository.Db.Query(ctx, query, wordlistID, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	words := []Word{}
	for rows.Next() {
		var scanErr error
		w := Word{WordlistID: wordlistID, UserID: userID}
		scanErr = rows.Scan(&w.ID, &w.Name, &w.CreatedAt, &w.UpdatedAt, &w.AudioURL, &w.Notes,
			&w.Pronunciation, &w.Learned, &w.ProcessingStatus, &w.ProcessingError,
			&w.ProcessingStartedAt, &w.ProcessingCompletedAt)
		if scanErr != nil {
			return nil, scanErr
		}
		words = append(words, w)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return words, nil
}

func (repository *WordRepository) GetByID(ctx context.Context, wordID int64) (*Word, error) {
	query := `SELECT id, name, created_at, updated_at, wordlist_id, user_id, 
				COALESCE(audio_url,''), COALESCE(notes,''), COALESCE(pronunciation,''), learned,
				processing_status, COALESCE(processing_error,''), 
				processing_started_at, processing_completed_at
			FROM words WHERE id=$1`
	row := repository.Db.QueryRow(ctx, query, wordID)
	var w Word

	err := row.Scan(&w.ID, &w.Name, &w.CreatedAt, &w.UpdatedAt, &w.WordlistID, &w.UserID,
		&w.AudioURL, &w.Notes, &w.Pronunciation, &w.Learned,
		&w.ProcessingStatus, &w.ProcessingError, &w.ProcessingStartedAt, &w.ProcessingCompletedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.NotFoundError{ID: wordID, Entity: "word"}
		}

		return nil, err
	}

	return &w, nil
}

func (repository *WordRepository) Delete(ctx context.Context, userID, wordID int64) (int64, error) {
	query := `DELETE FROM words WHERE user_id=$1 AND id=$2`
	result, err := repository.Db.Exec(ctx, query, userID, wordID)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

func (repository *WordRepository) Update(ctx context.Context, word *Word, tx *pgx.Tx) (int64, error) {
	updates := map[string]interface{}{
		"name":        word.Name,
		"audio_url":   word.AudioURL,
		"wordlist_id": word.WordlistID,
		"notes":       word.Notes,
		"learned":     word.Learned,
	}

	// Build query with user_id check for security
	var setParts []string
	var args []interface{}
	argIndex := 1

	// Always update the updated_at timestamp
	setParts = append(setParts, "updated_at = NOW()")

	// Handle all the update fields
	for field, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	// Build the complete query with user_id check for security
	query := fmt.Sprintf("UPDATE words SET %s WHERE user_id = $%d AND id = $%d",
		strings.Join(setParts, ", "), argIndex, argIndex+1)
	args = append(args, word.UserID, word.ID)

	// Execute the query
	var result pgconn.CommandTag
	var err error

	if tx != nil {
		result, err = (*tx).Exec(ctx, query, args...)
	} else {
		result, err = repository.Db.Exec(ctx, query, args...)
	}

	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

func (repository *WordRepository) GetLatestAudioURL(word string) (string, error) {
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

// UpdateFields updates specified fields for a word using a flexible map-based approach
func (repository *WordRepository) UpdateFields(wordID int64, updates map[string]interface{}, tx *pgx.Tx) error {
	if len(updates) == 0 {
		return fmt.Errorf("no fields specified for update")
	}

	var setParts []string
	var args []interface{}
	argIndex := 1

	// Always update the updated_at timestamp
	setParts = append(setParts, "updated_at = NOW()")

	// Handle special processing status logic
	if status, hasStatus := updates["processing_status"]; hasStatus {
		statusStr, ok := status.(string)
		if !ok {
			return fmt.Errorf("processing_status must be a string")
		}

		setParts = append(setParts, fmt.Sprintf("processing_status = $%d", argIndex))
		args = append(args, statusStr)
		argIndex++

		switch statusStr {
		case "processing":
			setParts = append(setParts, "processing_started_at = NOW()")
		case "completed":
			setParts = append(setParts, "processing_completed_at = NOW()")
			setParts = append(setParts, "processing_error = NULL")
		case "failed":
			setParts = append(setParts, "processing_completed_at = NOW()")
			if errorMsg, hasError := updates["processing_error"]; hasError {
				setParts = append(setParts, fmt.Sprintf("processing_error = $%d", argIndex))
				args = append(args, errorMsg)
				argIndex++
			}
		default:
			return fmt.Errorf("invalid processing status: %s", statusStr)
		}
		// Remove processing_status and processing_error from further processing
		delete(updates, "processing_status")
		delete(updates, "processing_error")
	}

	// Handle other fields
	validFields := map[string]bool{
		"name":          true,
		"audio_url":     true,
		"notes":         true,
		"learned":       true,
		"wordlist_id":   true,
		"pronunciation": true,
	}

	for field, value := range updates {
		if !validFields[field] {
			return fmt.Errorf("invalid field for update: %s", field)
		}

		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	// Build the complete query
	query := fmt.Sprintf("UPDATE words SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)
	args = append(args, wordID)

	// Execute the query
	var err error
	if tx != nil {
		_, err = (*tx).Exec(context.Background(), query, args...)
	} else {
		_, err = repository.Db.Exec(context.Background(), query, args...)
	}

	return err
}
