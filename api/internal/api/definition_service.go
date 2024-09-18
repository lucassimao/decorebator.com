package api

import (
	"fmt"
	"os"
	"regexp"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	repo "decorebator.com/internal/repository"
)

var definitionRepository *repo.DefinitionRepository

func init() {
	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection: ", "error", err)
		os.Exit(1)
	}
	definitionRepository = &repo.DefinitionRepository{Db: db}
}

func SaveDefinition(token string, tokenId int64, definitions []*model.Definition) ([]*model.Definition, error) {

	// wrapping token ocurrence within [ ] inside each example
	for _, definition := range definitions {
		for _, example := range definition.Examples {
			if !containsBracketedWord(example) {
				return nil, fmt.Errorf("example not wrapped into brackets: %s", example)
			}
		}
	}

	definitions, err := definitionRepository.Save(tokenId, definitions)
	if err != nil {
		return nil, fmt.Errorf("failed to save definitions: %w", err)
	}

	return definitions, nil
}

func GetRandomMeanings(definitionIdsToIgnore []int, size int) ([]string, error) {
	return definitionRepository.GetRandomMeanings(definitionIdsToIgnore, size)
}

func GetRandomTokens(definitionIdsToIgnore []int, partOfSpeech string, size int) ([]string, error) {
	return definitionRepository.GetRandomTokens(definitionIdsToIgnore, partOfSpeech, size)
}

func GetDefinitionById(id int64) (*model.Definition, error) {
	results, err := definitionRepository.Find(repo.FindArgs{Id: &id})
	if err != nil || len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

func FindDefinitionsByName(name string) ([]*model.Definition, error) {
	return definitionRepository.Find(repo.FindArgs{Name: &name})
}

func DeleteWordDefinitions(wordId int64) error {
	return definitionRepository.DeleteWordDefinitions(wordId)
}

func IsValidWordDefinition(wordId, definitionId, userId int64) (bool, error) {
	isValid, err := definitionRepository.IsValidWordDefinition(wordId, definitionId, userId)
	if err != nil {
		return false, fmt.Errorf("validation failed for tuple wordId, definitionId, userId. %w", err)
	}

	return isValid, nil
}

func containsBracketedWord(s string) bool {
	// Compile the regular expression
	re := regexp.MustCompile(`\[[^\]]+\]`)

	// Check if the string matches the pattern
	return re.MatchString(s)
}
