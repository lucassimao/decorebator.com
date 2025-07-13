package service

import (
	"fmt"
	"strings"

	"decorebator.com/internal/model"
	"decorebator.com/internal/openai"
	repo "decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DefinitionService handles definition-related operations with dependency injection
type DefinitionService struct {
	definitionRepository *repo.DefinitionRepository
}

// NewDefinitionService creates a new DefinitionService with injected dependencies
func NewDefinitionService(db *pgxpool.Pool) *DefinitionService {
	return &DefinitionService{
		definitionRepository: &repo.DefinitionRepository{Db: db},
	}
}

func (s *DefinitionService) SaveDefinition(tokenID int64, definitions []*model.Definition, tx *pgx.Tx) ([]*model.Definition, error) {
	// Set normalized part-of-speech for each definition before saving
	for _, def := range definitions {
		def.PartOfSpeechNormalized = NormalizePartOfSpeech(def.PartOfSpeech, def.Language)
	}

	definitions, err := s.definitionRepository.Save(tokenID, definitions, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to save definitions: %w", err)
	}

	return definitions, nil
}

func (s *DefinitionService) GetRandomMeanings(definitionIDsToIgnore []int, size int) ([]string, error) {
	return s.definitionRepository.GetRandomMeanings(definitionIDsToIgnore, size)
}

func (s *DefinitionService) GetRandomTokens(definitionIDsToIgnore []int, partOfSpeech string, size int) ([]string, error) {
	return s.definitionRepository.GetRandomTokens(definitionIDsToIgnore, partOfSpeech, size)
}

func (s *DefinitionService) GetDefinitionByID(id int64) (*model.Definition, error) {
	results, err := s.definitionRepository.Find(repo.FindArgs{ID: &id})
	if err != nil || len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

func (s *DefinitionService) findDefinitionsByName(name string) ([]*model.Definition, error) {
	return s.definitionRepository.Find(repo.FindArgs{Name: &name})
}

func (s *DefinitionService) DeleteWordDefinitions(wordID int64, tx *pgx.Tx) error {
	return s.definitionRepository.DeleteWordDefinitions(wordID, tx)
}

func (s *DefinitionService) didUserCreateWord(wordID, userID int64) (bool, error) {
	res, err := s.definitionRepository.DidUserCreateWord(wordID, userID)
	if err != nil {
		return false, fmt.Errorf("validation failed for tuple wordID, userID. %w", err)
	}

	return res, nil
}

func (s *DefinitionService) GetDefinitionsByWordID(wordID, userID int64) ([]*model.Definition, error) {
	return s.definitionRepository.GetDefinitionsByWordID(wordID, userID)
}

// NormalizePartOfSpeech converts a language-specific part-of-speech to normalized English
// This function uses the PartOfSpeechMappings from LANGUAGE_CONFIGS as the single source of truth
func NormalizePartOfSpeech(partOfSpeech, languageCode string) string {
	// Handle empty or nil cases
	if partOfSpeech == "" || languageCode == "" {
		return partOfSpeech
	}

	// Get language configuration
	config, exists := openai.LANGUAGE_CONFIGS[languageCode]
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
