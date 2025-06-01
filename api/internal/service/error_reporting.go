package service

import (
	"context"
	"errors"
	"fmt"

	"decorebator.com/internal/common"
)

type ErrorType string

const (
	UnrelatedImage   ErrorType = "_unrelated_image"
	MissingImage     ErrorType = "_missing_image"
	UnrelatedMeaning ErrorType = "_unrelated_meaning"
	UnrelatedExample ErrorType = "_unrelated_example"
	SoundNotPlaying  ErrorType = "_sound_not_playing"
)

func SaveErrorReport(errorType ErrorType, quiz Quiz, userId int64, ctx context.Context) error {
	logger := common.Logger.With("errorType", errorType, "quiz", quiz, "userId", userId)

	isValid, err := IsValidWordDefinition(quiz.WordID, quiz.DefinitionID, userId)

	if err != nil || !isValid {
		if err != nil {
			logger.Error("validation failed", "error", err)
		}
		return errors.New("validation failed")
	}

	success := false

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

	switch errorType {
	case SoundNotPlaying:
		_, err = TriggerTextToSpeechWorker(quiz.WordID, &tx)

	case UnrelatedImage, MissingImage:
		// default to the longest example
		_, err = TriggerGenerateImageWorker(quiz.DefinitionID, "", &tx)
	case UnrelatedExample, UnrelatedMeaning:
		err := DeleteWordDefinitions(quiz.WordID, tx)
		if err == nil {
			_, err = TriggerFetchDefinitionWorker(quiz.WordID, tx)
		}
	default:
		return fmt.Errorf("invalid error type %s", errorType)
	}

	success = (err == nil)

	// marking the quiz as failed
	var strategy LeitnerSystemStrategy
	err = strategy.SaveQuizResult(quiz.ID, false, &tx)

	if err != nil {
		common.Logger.Error("failed to flag quiz as failed", "error", err)
		return err
	}

	_, err = tx.Exec(context.Background(), `INSERT into 
				error_reports (is_resolved,error_type,quiz,user_id) 
				VALUES($1,$2,$3,$4)`, success, errorType, quiz, userId)

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
