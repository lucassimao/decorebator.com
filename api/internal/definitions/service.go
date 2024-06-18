package definitions

import (
	"fmt"
	"os"

	"decorebator.com/internal/common"
	"decorebator.com/internal/nlputils"
)

var repository *DefinitionRepository

func init() {
	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection: ", "error", err)
		os.Exit(1)
	}
	repository = &DefinitionRepository{db}
}

type TokenDefiner func(string) ([]Definition, error)

func FetchAndSave(token string, tokenId int64, definerFunc TokenDefiner) ([]Definition, error) {
	var definitions, err = definerFunc(token)

	if err != nil {
		return nil, fmt.Errorf("could not define %s(%d): %w", token, tokenId, err)
	}

	// wrapping token ocurrence within [ ] inside each example
	for _, definition := range definitions {
		for index, example := range definition.Examples {
			start, end, err := nlputils.FindTerm(definition.Token, example)
			if err != nil {
				common.Logger.Warn("failed to find term", "error", err, "token", definition.Token, "example", example)
				continue
			}

			definition.Examples[index] = fmt.Sprintf("%s[%s]%s", example[:start], example[start:end], example[end:])
		}
	}

	definitions, err = repository.save(tokenId, definitions)
	if err != nil {
		return nil, fmt.Errorf("failed to save definitions: %w", err)
	}

	return definitions, nil
}

func GetRandomMeanings(definitionIdsToIgnore []int, size int) ([]string, error) {
	return repository.getRandomMeanings(definitionIdsToIgnore, size)
}

func GetRandomExamples(filterOutIds []int, partOfSpeech string, size int) ([]string, error) {
	return repository.getRandomExamples(filterOutIds, partOfSpeech, size)
}

func GetRandomTokens(definitionIdsToIgnore []int, partOfSpeech string, size int) ([]string, error) {
	return repository.getRandomTokens(definitionIdsToIgnore, partOfSpeech, size)
}
