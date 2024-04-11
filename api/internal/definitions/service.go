package definitions

import (
	"fmt"
	"log"

	"decorebator.com/internal/common"
)

var repository *DefinitionRepository

func init() {
	db, err := common.GetDBConnection()
	if err != nil {
		log.Fatal("failed to open db connection: ", err)
	}
	repository = &DefinitionRepository{db}
}

type TokenDefiner func(string) ([]Definition, error)

func FetchAndSave(token string, tokenId int64, definerFunc TokenDefiner) ([]Definition, error) {
	var definitions, err = definerFunc(token)

	if err != nil {
		return nil, fmt.Errorf("could not define %s(%d): %w", token, tokenId, err)
	}

	definitions, err = repository.save(tokenId, definitions)
	if err != nil {
		return nil, fmt.Errorf("faild to save definitions: %w", err)
	}

	return definitions, nil
}

func GetRandomMeanings(filterOutIds []int, size int) ([]string, error) {
	return repository.getRandomMeanings(filterOutIds, size)
}

func GetRandomExamples(filterOutIds []int, partOfSpeech string, size int) ([]string, error) {
	return repository.getRandomExamples(filterOutIds, partOfSpeech, size)
}

func GetRandomTokens(filterOutIds []int, partOfSpeech string, size int) ([]string, error) {
	return repository.getRandomTokens(filterOutIds, partOfSpeech, size)
}
