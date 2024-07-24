package api

import (
	"errors"
	"strings"

	"decorebator.com/internal/common"
)

var wordRepository *WordRepository

func init() {
	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection", "error", err)
	}
	wordRepository = &WordRepository{db}
}

func GetWordByWordlist(wordlistId, userId int64) ([]Word, error) {
	return wordRepository.getAllFromWordlist(wordlistId, userId)
}

func GetWordById(id int64) (*Word, error) {
	return wordRepository.getById(id)
}

func SaveWord(dto *Word) (*Word, error) {
	var lowerCasedName = strings.ToLower(dto.Name)
	var trimmedName = strings.TrimSpace(lowerCasedName)
	word, err := wordRepository.save(trimmedName, dto.UserID, dto.WordlistID)
	logger := common.Logger.With("token", dto.Name, "userId", dto.UserID, "wordId", word.ID, "token", dto.Name, "func", "SaveWord")

	if err != nil {
		logger.Error("failed to save word", "error", err)
		return nil, errors.New("could not save word")
	}

	TriggerDefinitionFetcher(word.ID)

	return word, nil
}

func DeleteWord(id, userId int64) (int64, error) {
	count, err := wordRepository.delete(userId, id)
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
	count, err := wordRepository.update(word)
	if err != nil {
		common.Logger.Error("failed to update word", "error", err)
		return errors.New("failed to update word")
	}

	if count == 0 {
		return common.NotFoundError{ID: word.ID, Entity: "Word"}
	}
	return nil
}
