package api

import (
	"fmt"
	"os"
	"regexp"

	"decorebator.com/internal/common"
)

var definitionRepository *DefinitionRepository

func init() {
	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection: ", "error", err)
		os.Exit(1)
	}
	definitionRepository = &DefinitionRepository{db}
}

func SaveDefinition(token string, tokenId int64, definitions []Definition) ([]Definition, error) {

	// wrapping token ocurrence within [ ] inside each example
	for _, definition := range definitions {
		for _, example := range definition.Examples {
			if !containsBracketedWord(example) {
				return nil, fmt.Errorf("example not wrapped into brackets: %s", example)
			}
		}
	}

	definitions, err := definitionRepository.save(tokenId, definitions)
	if err != nil {
		return nil, fmt.Errorf("failed to save definitions: %w", err)
	}

	return definitions, nil
}

func GetRandomMeanings(definitionIdsToIgnore []int, size int) ([]string, error) {
	return definitionRepository.getRandomMeanings(definitionIdsToIgnore, size)
}

func GetRandomTokens(definitionIdsToIgnore []int, partOfSpeech string, size int) ([]string, error) {
	return definitionRepository.getRandomTokens(definitionIdsToIgnore, partOfSpeech, size)
}

func GetDefinitionById(id int64) (*Definition, error) {
	return definitionRepository.getById(id)
}

func containsBracketedWord(s string) bool {
	// Compile the regular expression
	re := regexp.MustCompile(`\[[^\]]+\]`)

	// Check if the string matches the pattern
	return re.MatchString(s)
}
