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

func ReportError(ctx context.Context, errorType ErrorReportType, wordID int64, definitionID *int64, userID int64) error {
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
	repo, err := checkErrorReportCooldown(ctx, userID, wordID, definitionID, string(errorType), logger)
	if err != nil {
		return err
	}

	// Execute error report in transaction
	return executeErrorReportTransaction(ctx, repo, errorType, wordID, definitionID, userID, logger)
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

func checkErrorReportCooldown(ctx context.Context, userID int64, wordID int64, definitionID *int64, errorType string, logger *slog.Logger) (*repository.ErrorReportRepository, error) {
	db, err := common.GetDBConnection()
	if err != nil {
		return nil, err
	}

	repo := repository.NewErrorReportRepository(db)

	cooldownUntil, err := repo.CheckCooldown(ctx, userID, wordID, definitionID, errorType)
	if err != nil {
		logger.Error("failed to check cooldown", "error", err)
		return nil, err
	}

	if cooldownUntil != nil {
		retryAfter := cooldownUntil.Sub(time.Now())
		logger.Warn("Error report blocked by cooldown",
			"action", "error_report_cooldown_blocked",
			"timestamp", time.Now().Unix(),
		)
		return nil, CooldownError{
			Message:       fmt.Sprintf("Please wait before reporting this error again. You can retry in %d minutes.", int(retryAfter.Minutes())),
			CooldownUntil: *cooldownUntil,
			RetryAfter:    retryAfter,
		}
	}

	return repo, nil
}

func executeErrorReportTransaction(ctx context.Context, repo *repository.ErrorReportRepository, errorType ErrorReportType, wordID int64, definitionID *int64, userID int64, logger *slog.Logger) error {
	db, err := common.GetDBConnection()
	if err != nil {
		return err
	}

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
	return completeErrorReport(ctx, repo, tx, errorType, wordID, definitionID, userID, report, logger)
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

func completeErrorReport(ctx context.Context, repo *repository.ErrorReportRepository, tx pgx.Tx, errorType ErrorReportType, wordID int64, definitionID *int64, userID int64, report ErrorReport, logger *slog.Logger) error {
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

	// Upsert error report
	if err := repo.UpsertErrorReport(ctx, tx, userID, definitionID, wordID, string(errorType)); err != nil {
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
