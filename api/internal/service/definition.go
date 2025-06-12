package service

import (
	"fmt"
	"os"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/openai"
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
	// Set normalized part-of-speech for each definition before saving
	for _, def := range definitions {
		def.PartOfSpeechNormalized = NormalizePartOfSpeech(def.PartOfSpeech, def.Language)
	}

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

// NormalizePartOfSpeech converts a language-specific part-of-speech to normalized English
// This function uses the PartOfSpeechMappings from LANGUAGE_CONFIGS as the single source of truth
func NormalizePartOfSpeech(partOfSpeech, languageCode string) string {
	// Handle empty or nil cases
	if partOfSpeech == "" || languageCode == "" {
		return partOfSpeech
	}

	// Get language configuration
	config, exists := openai.LanguageConfigs[languageCode]
	if !exists {
		// Fallback to original value for unsupported languages
		return partOfSpeech
	}

	// Try exact match first
	if normalized, exists := config.PartOfSpeechMappings[partOfSpeech]; exists {
		return normalized
	}

	// Try case-insensitive match
	partOfSpeechLower := strings.ToLower(partOfSpeech)
	for original, normalized := range config.PartOfSpeechMappings {
		if strings.ToLower(original) == partOfSpeechLower {
			return normalized
		}
	}

	// Return original if no mapping found
	return partOfSpeech
}
