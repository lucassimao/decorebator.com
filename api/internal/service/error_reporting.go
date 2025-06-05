package service

import (
	"context"
	"errors"
	"fmt"

	"decorebator.com/internal/common"
)

type ErrorReportType string

const (
	UnrelatedImage   ErrorReportType = "_unrelated_image"
	MissingImage     ErrorReportType = "_missing_image"
	UnrelatedMeaning ErrorReportType = "_unrelated_meaning"
	UnrelatedExample ErrorReportType = "_unrelated_example"
	SoundNotPlaying  ErrorReportType = "_sound_not_playing"
)

func ReportError(errorType ErrorReportType, quiz Quiz, userId int64, ctx context.Context) error {
	logger := common.Logger.With("errorType", errorType, "quiz", quiz, "userId", userId)

	isValid, err := IsValidWordDefinition(quiz.WordID, quiz.DefinitionID, userId)

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
		report = ErrorReport{WordId: &quiz.WordID, UserId: userId}
		_, err = TriggerTextToSpeechWorker(quiz.WordID, &report, &tx)

	case UnrelatedImage, MissingImage:
		report = ErrorReport{DefinitionId: &quiz.DefinitionID, UserId: userId}
		_, err = TriggerGenerateImageWorker(quiz.DefinitionID, "", &report, &tx)

	case UnrelatedExample, UnrelatedMeaning:
		err = DeleteWordDefinitions(quiz.WordID, &tx)
		if err == nil {
			report = ErrorReport{WordId: &quiz.WordID, UserId: userId}
			_, err = TriggerFetchDefinitionWorker(quiz.WordID, &report, &tx)
		}
	default:
		err = fmt.Errorf("invalid error type %s", errorType)
	}

	if err != nil {
		common.Logger.Error("failed to trigger jobs", "error", err)
		return err
	}

	// marking the definition temporarily as unavailable
	var strategy LeitnerSystemStrategy
	err = strategy.ReportError(userId, report, tx, ctx)

	if err != nil {
		common.Logger.Error("failed to report error", "error", err)
		return err
	}

	// Insert error report
	_, err = tx.Exec(ctx, `
		INSERT INTO error_reports 
		(user_id, definition_id,word_id, error_type, reported_at, status) 
		VALUES ($1, $2, $3, $4, NOW(), 'pending')
		ON CONFLICT (user_id, definition_id, word_id) 
		DO UPDATE SET 
			error_type = EXCLUDED.error_type,
			reported_at = NOW(),
			status = 'pending'`,
		userId, quiz.DefinitionID, quiz.WordID, string(errorType))

	if err != nil {
		common.Logger.Error("failed to save error report", "error", err)
		return err
	}
	return nil
}

func DeleteUserErrorReports(userId int64) (int64, error) {

	db, err := common.GetDBConnection()
	if err != nil {
		return 0, err
	}

	query := `DELETE FROM error_reports WHERE user_id=$1`
	result, err := db.Exec(context.Background(), query, userId)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}
