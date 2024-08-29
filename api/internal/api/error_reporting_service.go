package api

import (
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

func init() {
	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection", "error", err)
	}
	wordRepository = &WordRepository{db}
}

func SaveErrorReport(errorType ErrorType, quiz Quiz, userId int64) error {

	fmt.Println(quiz, errorType)

	if errorType == SoundNotPlaying {
		word, err := GetWordById(quiz.WordID)
		if err != nil {
			fmt.Println(err)
			return err
		}

		if word.UserID != userId {
			return errors.New("no word found for user")
		}
		TriggerTextToSpeech(quiz.WordID)
	}

	if errorType == UnrelatedImage {
		TriggerImageGenerator(quiz.DefinitionID, "") // default to the longest example
	}

	var strategy LeitnerSystemStrategy
	strategy.SaveQuizResult(quiz.ID, false)

	// persist report
	return nil
}
