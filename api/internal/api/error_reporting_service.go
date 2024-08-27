package api

import (
	"context"
	"errors"
	"fmt"
	"os"

	"decorebator.com/internal/common"
)

type ErrorType string

const (
	UnrelatedImage   ErrorType = "_unrelated_image"
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
		wordID, err := getWordID(userId, quiz.ID)
		if err != nil {
			fmt.Println(err)
			return err
		}
		TriggerTextToSpeech(wordID)
	}
	// persist report
	return nil
}

func getWordID(userID, leitnerSystemId int64) (int64, error) {

	query := `
			SELECT wd.word_id FROM leitner_system_tracking lst 
			JOIN word_definitions wd ON wd.definition_id = lst.definition_id
			JOIN words w ON wd.word_id = w.id
			WHERE w.user_id=$1 AND lst.id = $2
	`

	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection", "error", err)
		os.Exit(1)
	}

	rows, err := db.Query(context.Background(), query, userID, leitnerSystemId)
	if err != nil {
		return -1, err
	}

	defer rows.Close()

	hasRow := rows.Next()
	if !hasRow {
		return -1, errors.New("no word found")
	}

	var wordId int64

	err = rows.Scan(&wordId)

	if err != nil {
		return -1, err
	}

	return wordId, nil
}
