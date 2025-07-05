package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
)

type ErrorReportType string

const (
	UnrelatedImage   ErrorReportType = "_unrelated_image"
	MissingImage     ErrorReportType = "_missing_image"
	UnrelatedMeaning ErrorReportType = "_unrelated_meaning"
	UnrelatedExample ErrorReportType = "_unrelated_example"
	SoundNotPlaying  ErrorReportType = "_sound_not_playing"
	ProcessingFailed ErrorReportType = "_processing_failed" // For retrying failed word processing
)

func ReportError(ctx context.Context, errorType ErrorReportType, wordID int64, definitionID *int64, userID int64, flaggedContentIndex *int) error {
	logger := common.Logger.With("errorType", errorType, "wordID", wordID, "definitionID", definitionID, "userID", userID)

	// Log error report attempt for monitoring
	logger.Info("Error report attempt",
		"action", "error_report_attempt",
		"timestamp", time.Now().Unix(),
	)

	// Validate user owns the word
	if err := validateUserOwnsWord(wordID, userID, logger); err != nil {
		return err
	}

	// Check cooldown period
	err := checkErrorReportCooldown(ctx, userID, wordID, definitionID, string(errorType), logger)
	if err != nil {
		return err
	}

	// Execute error report in transaction
	return executeErrorReportTransaction(ctx, errorType, wordID, definitionID, userID, flaggedContentIndex, logger)
}

func validateUserOwnsWord(wordID int64, userID int64, logger *slog.Logger) error {
	isValid, err := didUserCreateWord(wordID, userID)
	if err != nil || !isValid {
		if err != nil {
			logger.Error("validation failed", "error", err)
		}
		return errors.New("validation failed")
	}
	return nil
}

func checkErrorReportCooldown(ctx context.Context, userID int64, wordID int64, definitionID *int64, errorType string, logger *slog.Logger) error {
	db, err := common.GetDBConnection()
	if err != nil {
		return err
	}

	repo := repository.NewErrorReportRepository(db)

	cooldownUntil, err := repo.CheckCooldown(ctx, userID, wordID, definitionID, errorType)
	if err != nil {
		logger.Error("failed to check cooldown", "error", err)
		return err
	}

	if cooldownUntil != nil {
		retryAfter := cooldownUntil.Sub(time.Now())
		logger.Warn("Error report blocked by cooldown",
			"action", "error_report_cooldown_blocked",
			"timestamp", time.Now().Unix(),
		)
		return CooldownError{
			Message:       fmt.Sprintf("Please wait before reporting this error again. You can retry in %d minutes.", int(retryAfter.Minutes())),
			CooldownUntil: *cooldownUntil,
			RetryAfter:    retryAfter,
		}
	}

	return nil
}

func executeErrorReportTransaction(ctx context.Context, errorType ErrorReportType, wordID int64, definitionID *int64, userID int64, flaggedContentIndex *int, logger *slog.Logger) error {
	db, err := common.GetDBConnection()
	if err != nil {
		return err
	}

	repo := repository.NewErrorReportRepository(db)

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				common.Logger.Error("failed to commit transaction in error reporting", "error", commitErr)
			}
		} else {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				common.Logger.Error("failed to rollback transaction in error reporting", "error", rollbackErr)
			}
		}
	}()

	// Process the error type and trigger appropriate workers
	report, err := processErrorType(errorType, wordID, definitionID, userID, &tx)
	if err != nil {
		common.Logger.Error("failed to trigger jobs", "error", err)
		return err
	}

	// Handle post-processing steps
	return completeErrorReport(ctx, repo, tx, errorType, wordID, definitionID, userID, flaggedContentIndex, report, logger)
}

func processErrorType(errorType ErrorReportType, wordID int64, definitionID *int64, userID int64, tx *pgx.Tx) (ErrorReport, error) {
	var report ErrorReport
	var err error

	switch errorType {
	case SoundNotPlaying:
		report = ErrorReport{WordId: &wordID, UserId: userID}
		_, err = TriggerTextToSpeechWorker(wordID, &userID, &report, tx)

	case UnrelatedImage, MissingImage:
		if definitionID == nil {
			return report, fmt.Errorf("definition ID required for image-related errors")
		}
		report = ErrorReport{DefinitionId: definitionID, UserId: userID}
		_, err = TriggerGenerateImageWorker(*definitionID, &userID, &report, tx)

	case UnrelatedExample, UnrelatedMeaning:
		err = DeleteWordDefinitions(wordID, tx)
		if err == nil {
			report = ErrorReport{WordId: &wordID, UserId: userID}
			_, err = TriggerFetchDefinitionWorker(wordID, &userID, &report, tx)
		}

	case ProcessingFailed:
		err = DeleteWordDefinitions(wordID, tx)
		if err == nil {
			report = ErrorReport{WordId: &wordID, UserId: userID}
			_, err = TriggerFetchDefinitionWorker(wordID, &userID, &report, tx)
		}

	default:
		err = fmt.Errorf("invalid error type %s", errorType)
	}

	return report, err
}

func completeErrorReport(ctx context.Context, repo *repository.ErrorReportRepository, tx pgx.Tx, errorType ErrorReportType, wordID int64, definitionID *int64, userID int64, flaggedContentIndex *int, report ErrorReport, logger *slog.Logger) error {
	// Update last regenerated timestamp
	if err := updateLastRegeneratedTimestamp(ctx, repo, tx, errorType, wordID, definitionID, logger); err != nil {
		logger.Error("failed to update last regenerated timestamp", "error", err)
		// Don't fail the whole operation for this
	}

	// Set cooldown for this error report
	if err := setCooldownPeriod(ctx, repo, tx, userID, wordID, definitionID, string(errorType), logger); err != nil {
		logger.Error("failed to set cooldown", "error", err)
		// Don't fail the whole operation for this
	}

	// Mark definition temporarily as unavailable
	strategy := DefaultLeitnerSystemStrategy()
	if err := strategy.ReportError(userID, report, tx, ctx); err != nil {
		common.Logger.Error("failed to report error", "error", err)
		return err
	}

	// Build content snapshot before potential deletion
	contentSnapshot, finalDefinitionID, err := buildContentSnapshot(ctx, tx, errorType, wordID, definitionID, flaggedContentIndex)
	if err != nil {
		logger.Error("failed to build content snapshot", "error", err)
		return err
	}

	// Upsert error report with content snapshot
	if err := repo.UpsertErrorReport(ctx, tx, userID, finalDefinitionID, wordID, string(errorType), contentSnapshot); err != nil {
		common.Logger.Error("failed to save error report", "error", err)
		return err
	}

	// Log successful error report for monitoring
	logger.Info("Error report processed successfully",
		"action", "error_report_success",
		"timestamp", time.Now().Unix(),
		"regeneration_type", errorType,
	)

	return nil
}

func updateLastRegeneratedTimestamp(ctx context.Context, repo *repository.ErrorReportRepository, tx pgx.Tx, errorType ErrorReportType, wordID int64, definitionID *int64, _ *slog.Logger) error {
	switch errorType {
	case SoundNotPlaying:
		return repo.UpdateWordAudioLastRegeneratedAt(ctx, tx, wordID)
	case UnrelatedImage, MissingImage, UnrelatedExample, UnrelatedMeaning:
		if definitionID != nil {
			return repo.UpdateDefinitionLastRegeneratedAt(ctx, tx, *definitionID)
		}
	}
	return nil
}

func setCooldownPeriod(ctx context.Context, repo *repository.ErrorReportRepository, tx pgx.Tx, userID int64, wordID int64, definitionID *int64, errorType string, _ *slog.Logger) error {
	cooldownDuration := time.Hour // 1 hour cooldown as agreed
	cooldownUntilTime := time.Now().Add(cooldownDuration)
	return repo.SetCooldown(ctx, tx, userID, wordID, definitionID, errorType, cooldownUntilTime)
}

// buildContentSnapshot creates a content snapshot and determines final foreign key strategy
func buildContentSnapshot(ctx context.Context, tx pgx.Tx, errorType ErrorReportType, wordID int64, definitionID *int64, flaggedContentIndex *int) (map[string]interface{}, *int64, error) {
	var snapshot map[string]any
	var finalDefinitionID *int64

	switch errorType {
	case UnrelatedMeaning, UnrelatedExample:
		// These will delete definitions, so capture full snapshot and nullify foreign key
		if definitionID == nil {
			return nil, nil, fmt.Errorf("definition ID required for %s", errorType)
		}

		defSnapshot, err := fetchCompleteDefinitionSnapshot(ctx, tx, *definitionID, wordID)
		if err != nil {
			return nil, nil, err
		}

		// Add flagged content index for examples
		if errorType == UnrelatedExample && flaggedContentIndex != nil {
			defSnapshot["flagged_content_index"] = *flaggedContentIndex
		}

		snapshot = defSnapshot
		finalDefinitionID = nil // Set to NULL since definition will be deleted

	case UnrelatedImage, MissingImage:
		// These regenerate images but keep definitions, so preserve foreign key
		if definitionID == nil {
			return nil, nil, fmt.Errorf("definition ID required for %s", errorType)
		}

		defSnapshot, err := fetchCompleteDefinitionSnapshot(ctx, tx, *definitionID, wordID)
		if err != nil {
			return nil, nil, err
		}

		snapshot = defSnapshot
		finalDefinitionID = definitionID // Keep foreign key

	case SoundNotPlaying:
		// This regenerates audio but keeps word, so preserve foreign key
		wordSnapshot, err := fetchWordSnapshot(ctx, tx, wordID)
		if err != nil {
			return nil, nil, err
		}

		snapshot = wordSnapshot
		finalDefinitionID = definitionID // Keep foreign key (might be nil for word-level errors)

	case ProcessingFailed:
		// This is a processing retry, no content quality issue
		snapshot = nil
		finalDefinitionID = nil // Set to NULL since definition will be deleted

	default:
		return nil, nil, fmt.Errorf("unsupported error type: %s", errorType)
	}

	return snapshot, finalDefinitionID, nil
}

// fetchCompleteDefinitionSnapshot fetches complete definition data for historical preservation
func fetchCompleteDefinitionSnapshot(ctx context.Context, tx pgx.Tx, definitionID int64, wordID int64) (map[string]interface{}, error) {
	// Use definition service to fetch the definition
	definition, err := GetDefinitionById(definitionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch definition: %w", err)
	}
	if definition == nil {
		return nil, fmt.Errorf("definition not found with ID: %d", definitionID)
	}

	// Build comprehensive snapshot
	snapshot := map[string]interface{}{
		"definition_id":   definitionID,
		"meaning":         definition.Meaning,
		"part_of_speech":  definition.PartOfSpeech,
		"source":          definition.Source,
		"language":        definition.Language,
		"token":           definition.Token,
		"examples":        definition.Examples,
		"captured_at":     time.Now().Format(time.RFC3339),
	}

	return snapshot, nil
}

// fetchWordSnapshot fetches word data for audio-related errors
func fetchWordSnapshot(ctx context.Context, tx pgx.Tx, wordID int64) (map[string]interface{}, error) {
	// Get database connection and create word service
	db, err := common.GetDBConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}
	
	wordService := NewWordService(db, nil) // Pass db connection, nil for moderation service since we're just reading
	word, err := wordService.GetWordByID(wordID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch word: %w", err)
	}
	if word == nil {
		return nil, fmt.Errorf("word not found with ID: %d", wordID)
	}

	// Get language from wordlist - we still need this query since word doesn't contain language directly
	var language string
	err = tx.QueryRow(ctx, `
		SELECT wl.language_code
		FROM wordlists wl
		WHERE wl.id = $1
	`, word.WordlistID).Scan(&language)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch wordlist language: %w", err)
	}

	// Build word snapshot
	snapshot := map[string]interface{}{
		"word_id":     wordID,
		"token":       word.Name,
		"language":    language,
		"audio_url":   word.AudioURL,
		"captured_at": time.Now().Format(time.RFC3339),
	}

	return snapshot, nil
}

func DeleteUserErrorReports(userID int64) (int64, error) {
	db, err := common.GetDBConnection()
	if err != nil {
		return 0, err
	}

	repo := repository.NewErrorReportRepository(db)
	return repo.DeleteUserErrorReports(context.Background(), userID)
}

// CooldownError represents an error when a cooldown is active
type CooldownError struct {
	Message       string
	CooldownUntil time.Time
	RetryAfter    time.Duration
}

func (e CooldownError) Error() string {
	return e.Message
}
