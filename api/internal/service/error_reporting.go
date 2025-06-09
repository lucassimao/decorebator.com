package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/repository"
)

type ErrorReportType string

const (
	UnrelatedImage   ErrorReportType = "_unrelated_image"
	MissingImage     ErrorReportType = "_missing_image"
	UnrelatedMeaning ErrorReportType = "_unrelated_meaning"
	UnrelatedExample ErrorReportType = "_unrelated_example"
	SoundNotPlaying  ErrorReportType = "_sound_not_playing"
)

func ReportError(errorType ErrorReportType, wordID int64, definitionID int64, userId int64, ctx context.Context) error {
	logger := common.Logger.With("errorType", errorType, "wordID", wordID, "definitionID", definitionID, "userId", userId)
	
	// Log error report attempt for monitoring
	logger.Info("Error report attempt",
		"action", "error_report_attempt",
		"timestamp", time.Now().Unix(),
	)

	isValid, err := didUserCreateWord(wordID, userId)

	if err != nil || !isValid {
		if err != nil {
			logger.Error("validation failed", "error", err)
		}
		return errors.New("validation failed")
	}

	db, err := common.GetDBConnection()
	if err != nil {
		return err
	}

	// Create repository
	repo := repository.NewErrorReportRepository(db)

	// Check cooldown period
	cooldownUntil, err := repo.CheckCooldown(ctx, userId, wordID, definitionID, string(errorType))
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

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			tx.Commit(ctx)
		} else {
			tx.Rollback(ctx)
		}
	}()

	var report ErrorReport

	switch errorType {
	case SoundNotPlaying:
		report = ErrorReport{WordId: &wordID, UserId: userId}
		_, err = TriggerTextToSpeechWorker(wordID, &report, &tx)

	case UnrelatedImage, MissingImage:
		report = ErrorReport{DefinitionId: &definitionID, UserId: userId}
		_, err = TriggerGenerateImageWorker(definitionID, &report, &tx)

	case UnrelatedExample, UnrelatedMeaning:
		err = DeleteWordDefinitions(wordID, &tx)
		if err == nil {
			report = ErrorReport{WordId: &wordID, UserId: userId}
			_, err = TriggerFetchDefinitionWorker(wordID, &report, &tx)
		}
	default:
		err = fmt.Errorf("invalid error type %s", errorType)
	}

	if err != nil {
		common.Logger.Error("failed to trigger jobs", "error", err)
		return err
	}

	// Update last regenerated timestamp
	switch errorType {
	case SoundNotPlaying:
		err = repo.UpdateWordAudioLastRegeneratedAt(ctx, tx, wordID)
	case UnrelatedImage, MissingImage, UnrelatedExample, UnrelatedMeaning:
		if definitionID > 0 {
			err = repo.UpdateDefinitionLastRegeneratedAt(ctx, tx, definitionID)
		}
	}
	if err != nil {
		logger.Error("failed to update last regenerated timestamp", "error", err)
		// Don't fail the whole operation for this
	}

	// Set cooldown for this error report
	cooldownDuration := time.Hour // 1 hour cooldown as agreed
	cooldownUntilTime := time.Now().Add(cooldownDuration)
	err = repo.SetCooldown(ctx, tx, userId, wordID, definitionID, string(errorType), cooldownUntilTime)
	if err != nil {
		logger.Error("failed to set cooldown", "error", err)
		// Don't fail the whole operation for this
	}

	// marking the definition temporarily as unavailable
	var strategy LeitnerSystemStrategy
	err = strategy.ReportError(userId, report, tx, ctx)

	if err != nil {
		common.Logger.Error("failed to report error", "error", err)
		return err
	}

	// Upsert error report
	err = repo.UpsertErrorReport(ctx, tx, userId, definitionID, wordID, string(errorType))
	if err != nil {
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

func DeleteUserErrorReports(userId int64) (int64, error) {
	db, err := common.GetDBConnection()
	if err != nil {
		return 0, err
	}

	repo := repository.NewErrorReportRepository(db)
	return repo.DeleteUserErrorReports(context.Background(), userId)
}

// CooldownError represents an error when a cooldown is active
type CooldownError struct {
	Message      string
	CooldownUntil time.Time
	RetryAfter   time.Duration
}

func (e CooldownError) Error() string {
	return e.Message
}