package service

import (
	"fmt"
	"os"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	repo "decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
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

func SaveDefinition(token string, tokenId int64, definitions []*model.Definition, tx pgx.Tx) ([]*model.Definition, error) {

	definitions, err := definitionRepository.Save(tokenId, definitions, tx)
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

func findDefinitionsByName(name string) ([]*model.Definition, error) {
	return definitionRepository.Find(repo.FindArgs{Name: &name})
}

func DeleteWordDefinitions(wordId int64, tx *pgx.Tx) error {
	return definitionRepository.DeleteWordDefinitions(wordId, tx)
}

func didUserCreateWord(wordId, userId int64) (bool, error) {
	res, err := definitionRepository.DidUserCreateWord(wordId, userId)
	if err != nil {
		return false, fmt.Errorf("validation failed for tuple wordId, userId. %w", err)
	}

	return res, nil
}

func GetDefinitionsByWordId(wordId, userId int64) ([]*model.Definition, error) {
	return definitionRepository.GetDefinitionsByWordId(wordId, userId)
}
