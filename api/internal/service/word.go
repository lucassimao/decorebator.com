package service

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	repo "decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
)

var wordRepository *repo.WordRepository

type Word = model.Word

func init() {
	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection", "error", err)
	}
	wordRepository = &repo.WordRepository{Db: db}
}

// GetWordByWordlist returns words from wordlist with optional filtering
// onlyWithDefinitions: if true, returns only words that have definitions with meanings
func GetWordByWordlist(wordlistId, userId int64, onlyWithDefinitions bool) ([]Word, error) {
	return wordRepository.GetWordsByWordlist(wordlistId, userId, onlyWithDefinitions)
}

func GetWordById(id int64) (*Word, error) {
	return wordRepository.GetById(id)
}

func SaveWord(dto *Word, ctx context.Context) (*Word, error) {
	var lowerCasedName = strings.ToLower(dto.Name)
	var trimmedName = strings.TrimSpace(lowerCasedName)

	// count runes (Unicode characters), not bytes
	if utf8.RuneCountInString(trimmedName) > 15 {
		return nil, common.BusinessError{Message: "words must be limited to 15 chars"}
	}

	tx, err := wordRepository.Db.Begin(ctx)

	if err != nil {
		return nil, err
	}

	// db transaction mgmt
	defer func() {
		if err == nil {
			tx.Commit(ctx)
		} else {
			tx.Rollback(ctx)
		}
	}()

	word, err := wordRepository.Save(trimmedName, dto.Notes, dto.UserID, dto.WordlistID, &tx)
	if err != nil {
		return nil, err
	}

	// check if there are definitions for this word already
	definitions, _ := findDefinitionsByName(word.Name)

	if len(definitions) > 0 {
		definitionIds := []int64{}

		for _, def := range definitions {
			definitionIds = append(definitionIds, def.ID)
		}
		wordRepository.ReuseDefinitions(word.ID, definitionIds, tx)

		quizStrategy := LeitnerSystemStrategy{}
		quizStrategy.IncludeDefinitions(word.ID, word.UserID, definitionIds, tx)

		var latestAudioURL string
		latestAudioURL, err = wordRepository.GetLatestAudioUrl(trimmedName)

		if err != nil {
			TriggerTextToSpeechWorker(word.ID, nil, &tx)
			err = nil // fine if triggering the worker fails somehow
		} else {
			word.AudioURL = latestAudioURL
			err = UpdateWord(word, &tx)
		}
	} else {
		TriggerFetchDefinitionWorker(word.ID, nil, &tx)
		TriggerTextToSpeechWorker(word.ID, nil, &tx)
	}

	if err != nil {
		return nil, err
	}

	return word, nil
}

func DeleteWord(id, userId int64) (int64, error) {
	word, err := GetWordById(id)
	if err != nil {
		return 0, err
	}

	if word == nil {
		return 0, common.NotFoundError{ID: id, Entity: "Word"}
	}

	count, err := wordRepository.Delete(userId, id)
	if err != nil {
		common.Logger.Error("failed to delete word", "error", err)
		return 0, errors.New("failed to delete word")
	}

	return count, nil
}

func UpdateWord(word *Word, tx *pgx.Tx) error {
	count, err := wordRepository.Update(word, tx)
	if err != nil {
		return err
	}

	if count == 0 {
		return common.NotFoundError{ID: word.ID, Entity: "Word"}
	}
	return nil
}
