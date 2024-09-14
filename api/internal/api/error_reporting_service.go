package api

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

func SaveErrorReport(errorType ErrorType, quiz Quiz, userId int64) error {
	logger := common.Logger.With("errorType", errorType, "quiz", quiz, "userId", userId)

	isValid, err := IsValidWordDefinition(quiz.WordID, quiz.DefinitionID, userId)

	if err != nil || !isValid {
		if err != nil {
			logger.Error("validation failed", "error", err)
		}
		return errors.New("validation failed")
	}

	success := false

	switch errorType {
	case SoundNotPlaying:
		_, err := TriggerTextToSpeech(quiz.WordID)
		success = (err == nil)
	case UnrelatedImage, MissingImage:
		// default to the longest example
		_, err = TriggerImageGenerator(quiz.DefinitionID, "")
		success = (err == nil)
	case UnrelatedExample, UnrelatedMeaning:
		err := DeleteWordDefinitions(quiz.WordID)
		fmt.Println(err)
		if err == nil {
			_, err = TriggerDefinitionFetcher(quiz.WordID)
			fmt.Println(err)
			success = (err == nil)
		}
	default:
		return fmt.Errorf("invalid error type %s", errorType)
	}

	// marking the quiz as failed
	var strategy LeitnerSystemStrategy
	strategy.SaveQuizResult(quiz.ID, false)

	// persist report
	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection", "error", err)
		return err
	}

	_, err = db.Exec(context.Background(), `INSERT into 
				error_reports (is_resolved,error_type,quiz,user_id) 
				VALUES($1,$2,$3,$4)`, success, errorType, quiz, userId)

	if err != nil {
		common.Logger.Error("failed to save error report", "error", err)
		return err
	}
	return nil
}
