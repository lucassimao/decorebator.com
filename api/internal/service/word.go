package service

import (
	"context"
	"errors"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	repo "decorebator.com/internal/repository"
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

func GetWordByWordlist(wordlistId, userId int64) ([]Word, error) {
	return wordRepository.GetAllFromWordlist(wordlistId, userId)
}

func GetWordById(id int64) (*Word, error) {
	return wordRepository.GetById(id)
}

func SaveWord(dto *Word, ctx context.Context) (*Word, error) {
	var lowerCasedName = strings.ToLower(dto.Name)
	var trimmedName = strings.TrimSpace(lowerCasedName)

	if len(trimmedName) > 15 {
		return nil, common.BusinessError{Message: "words must be limited to 15 chars"}
	}

	word, err := wordRepository.Save(trimmedName, dto.UserID, dto.WordlistID)

	if err != nil {
		return nil, err
	}

	// check if there are definitions for this word already
	definitions, _ := findDefinitionsByName(word.Name)
	container, ok := common.GetDigContainerFromContext(ctx)
	if !ok {
		return nil, errors.New("dig container not found")
	}

	if len(definitions) > 0 {
		definitionIds := []int64{}
		for _, def := range definitions {
			definitionIds = append(definitionIds, def.ID)
		}
		reuseDefinitions(word.ID, definitionIds)

		quizStrategy := LeitnerSystemStrategy{}
		quizStrategy.IncludeDefinitions(word.ID, word.UserID, definitionIds)

		container.Invoke(func(trigger common.ContentGenerationService) {
			trigger.TextToSpeech(word.ID)
		})
	} else {
		container.Invoke(func(trigger common.ContentGenerationService) {
			trigger.FetchDefinition(word.ID)
			trigger.TextToSpeech(word.ID)
		})
	}

	return word, nil
}

func DeleteWord(id, userId int64) (int64, error) {
	count, err := wordRepository.Delete(userId, id)
	if err != nil {
		common.Logger.Error("failed to delete word", "error", err)
		return 0, errors.New("failed to delete word")
	}

	if count == 0 {
		return 0, common.NotFoundError{ID: id, Entity: "Word"}
	}

	return count, nil
}

func UpdateWord(word *Word) error {
	count, err := wordRepository.Update(word)
	if err != nil {
		common.Logger.Error("failed to update word", "error", err)
		return errors.New("failed to update word")
	}

	if count == 0 {
		return common.NotFoundError{ID: word.ID, Entity: "Word"}
	}
	return nil
}
