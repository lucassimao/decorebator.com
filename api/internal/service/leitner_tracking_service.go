package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LeitnerTrackingService manages the lifecycle and state of learning progress tracking records
// in the leitner_system_tracking table. This service is responsible for:
// - Creating tracking records for new definitions (IncludeDefinitions)
// - Updating box progression and skip status after quizzes (UpdateQuizProgress)
// - Managing temporary skips for error handling (ClearTemporarySkip, SetTemporarySkip)
type LeitnerTrackingService struct {
	db *pgxpool.Pool
}

// NewLeitnerTrackingService creates a new tracking service with database dependency
func NewLeitnerTrackingService(db *pgxpool.Pool) *LeitnerTrackingService {
	return &LeitnerTrackingService{
		db: db,
	}
}

// IncludeDefinitions initializes tracking records for new definitions in the Leitner system.
// Each definition starts in box 1 (immediate review) and will progress through boxes based on quiz performance.
//
// Parameters:
// - wordID: The ID of the word these definitions belong to
// - userID: The ID of the user who will be quizzed on these definitions
// - definitionIDs: Array of definition IDs to include in the Leitner system
// - tx: Database transaction to ensure atomicity
//
// Returns an error if any database operations fail.
func (s *LeitnerTrackingService) IncludeDefinitions(ctx context.Context, wordID, userID int64, definitionIDs []int64, tx pgx.Tx) error {
	for _, definitionID := range definitionIDs {
		query := `INSERT INTO leitner_system_tracking (user_id, definition_id, box_id, word_id, updated_at, next_review_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())`

		_, err := tx.Exec(ctx, query, userID, definitionID, 1, wordID)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateQuizProgress updates the box progression and temporary skip status for a tracking record
// based on quiz performance. This implements the core Leitner algorithm:
// - Correct answers: Move to next box (max box 7) and clear temporary skip
// - Incorrect answers: Reset to box 1 and set 10-minute temporary skip
//
// Parameters:
// - trackingID: The ID of the leitner_system_tracking record to update
// - isCorrect: Whether the quiz answer was correct
// - tx: Database transaction (optional)
//
// Returns an error if database operations fail.
func (s *LeitnerTrackingService) UpdateQuizProgress(ctx context.Context, trackingID int64, isCorrect bool, tx *pgx.Tx) error {
	var execTx pgx.Tx
	if tx != nil {
		execTx = *tx
	} else {
		// Create transaction if none provided
		newTx, err := s.db.Begin(ctx)
		if err != nil {
			return err
		}
		defer func() {
			_ = newTx.Rollback(ctx) // Log rollback error but don't override the original error
		}()
		execTx = newTx
	}

	// Proper Leitner system logic with temporary skip on incorrect answers
	query := `
		WITH updated AS (
			SELECT id,
				CASE 
					WHEN $1 AND box_id < 7 THEN box_id + 1  -- Move to next box on success
					WHEN $1 AND box_id = 7 THEN 7           -- Stay at max box
					ELSE 1                                   -- Reset to box 1 on failure
				END AS new_box
			FROM leitner_system_tracking
			WHERE id = $2
		)
		UPDATE leitner_system_tracking lst
		SET 
			updated_at = NOW(),
			box_id = updated.new_box,
			next_review_at = CASE updated.new_box
				WHEN 1 THEN NOW()
				WHEN 2 THEN NOW() + INTERVAL '6 hours'
				WHEN 3 THEN NOW() + INTERVAL '24 hours'
				WHEN 4 THEN NOW() + INTERVAL '72 hours'
				WHEN 5 THEN NOW() + INTERVAL '168 hours'
				WHEN 6 THEN NOW() + INTERVAL '336 hours'
				WHEN 7 THEN NOW() + INTERVAL '720 hours'
				ELSE NOW()
			END,
			temporarily_skipped_until = CASE 
				WHEN NOT $1 THEN NOW() + INTERVAL '10 minutes'  -- Skip for 10 minutes on incorrect answer
				ELSE NULL                                        -- Clear skip on correct answer
			END
		FROM updated
		WHERE lst.id = updated.id`

	_, err := execTx.Exec(ctx, query, isCorrect, trackingID)
	if err != nil {
		return err
	}

	// If we created the transaction, commit it
	if tx == nil {
		err = execTx.Commit(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// ClearTemporarySkip removes the temporary skip status from tracking records.
// This is called when errors are resolved (e.g., new definitions fetched, images generated).
//
// Parameters:
// - report: ErrorReport containing either DefinitionID or WordID to identify records
// - tx: Database transaction to ensure atomicity
//
// Returns an error if database operations fail.
func (s *LeitnerTrackingService) ClearTemporarySkip(ctx context.Context, report ErrorReport, tx pgx.Tx) error {
	baseQuery := `UPDATE leitner_system_tracking SET temporarily_skipped_until = NULL `
	selection, queryArgs, err := buildQuerySelectionFromErrorReport(report)
	if err != nil {
		return err
	}

	query := baseQuery + selection
	_, err = tx.Exec(ctx, query, queryArgs...)
	return err
}

// SetTemporarySkip adds a temporary skip status to tracking records for error handling.
// This prevents problematic definitions from appearing in quizzes for 1 hour.
//
// Parameters:
// - report: ErrorReport containing either DefinitionID or WordID to identify records
// - tx: Database transaction to ensure atomicity
//
// Returns an error if database operations fail.
func (s *LeitnerTrackingService) SetTemporarySkip(ctx context.Context, report ErrorReport, tx pgx.Tx) error {
	if report.DefinitionID == nil && report.WordID == nil {
		return errors.New("definition or word missing")
	}

	baseQuery := `UPDATE leitner_system_tracking SET temporarily_skipped_until = NOW() + INTERVAL '1 hour' `
	selection, queryArgs, err := buildQuerySelectionFromErrorReport(report)
	if err != nil {
		return err
	}

	query := baseQuery + selection
	_, err = tx.Exec(ctx, query, queryArgs...)
	return err
}

// buildQuerySelectionFromErrorReport is a helper function that builds WHERE clause
// and query arguments for leitner_system_tracking operations based on ErrorReport.
// This is extracted from the original LeitnerSystemStrategy implementation.
func buildQuerySelectionFromErrorReport(report ErrorReport) (string, []interface{}, error) {
	var selection string
	var queryArgs []interface{}

	if report.DefinitionID != nil {
		selection = `WHERE definition_id = $1 AND user_id = $2`
		queryArgs = []interface{}{*report.DefinitionID, report.UserID}
	} else if report.WordID != nil {
		selection = `WHERE word_id = $1 AND user_id = $2`
		queryArgs = []interface{}{*report.WordID, report.UserID}
	} else {
		return "", nil, errors.New("either definition_id or word_id must be provided")
	}

	return selection, queryArgs, nil
}
