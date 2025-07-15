package service

import (
	"context"
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

func (s *DefinitionService) SaveDefinition(ctx context.Context, tokenID int64, definitions []*model.Definition, tx *pgx.Tx) ([]*model.Definition, error) {
	// Set normalized part-of-speech for each definition before saving
	for _, def := range definitions {
		def.PartOfSpeechNormalized = NormalizePartOfSpeech(def.PartOfSpeech, def.Language)
	}

	definitions, err := s.definitionRepository.Save(ctx, tokenID, definitions, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to save definitions: %w", err)
	}

	return definitions, nil
}

func (s *DefinitionService) GetRandomMeanings(ctx context.Context, definitionIDsToIgnore []int, size int) ([]string, error) {
	return s.definitionRepository.GetRandomMeanings(ctx, definitionIDsToIgnore, size)
}

func (s *DefinitionService) GetRandomTokens(ctx context.Context, definitionIDsToIgnore []int, partOfSpeech string, size int) ([]string, error) {
	return s.definitionRepository.GetRandomTokens(ctx, definitionIDsToIgnore, partOfSpeech, size)
}

func (s *DefinitionService) GetDefinitionByID(ctx context.Context, id int64) (*model.Definition, error) {
	results, err := s.definitionRepository.Find(ctx, repo.FindArgs{ID: &id})
	if err != nil || len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

func (s *DefinitionService) findDefinitionsByName(ctx context.Context, name string) ([]*model.Definition, error) {
	return s.definitionRepository.Find(ctx, repo.FindArgs{Name: &name})
}

func (s *DefinitionService) DeleteWordDefinitions(ctx context.Context, wordID int64, tx *pgx.Tx) error {
	return s.definitionRepository.DeleteWordDefinitions(ctx, wordID, tx)
}

func (s *DefinitionService) didUserCreateWord(ctx context.Context, wordID, userID int64) (bool, error) {
	res, err := s.definitionRepository.DidUserCreateWord(ctx, wordID, userID)
	if err != nil {
		return false, fmt.Errorf("validation failed for tuple wordID, userID. %w", err)
	}

	return res, nil
}

func (s *DefinitionService) GetDefinitionsByWordID(ctx context.Context, wordID, userID int64) ([]*model.Definition, error) {
	return s.definitionRepository.GetDefinitionsByWordID(ctx, wordID, userID)
}

func (s *DefinitionService) CreateExampleAudio(ctx context.Context, definitionID int64, exampleText, audioURL, inflectionType string) error {
	return s.definitionRepository.CreateExampleAudio(ctx, definitionID, exampleText, audioURL, inflectionType)
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
