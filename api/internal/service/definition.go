package service

import (
	"context"
	"fmt"
	"strings"

	"decorebator.com/internal/common"
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

type DistractorScope struct {
	UserID                 int64
	WordlistID             int64
	Language               string
	PartOfSpeechNormalized string
}

// NewDefinitionService creates a new DefinitionService with injected dependencies
func NewDefinitionService(db *pgxpool.Pool) *DefinitionService {
	return &DefinitionService{
		definitionRepository: &repo.DefinitionRepository{Db: db},
	}
}

func (s *DefinitionService) SaveDefinition(ctx context.Context, tokenID int64, definitions []*model.Definition, tx *pgx.Tx) ([]*model.Definition, error) {
	normalizeDefinitionsForPersistence(definitions)

	definitions, err := s.definitionRepository.Save(ctx, tokenID, definitions, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to save definitions: %w", err)
	}

	return definitions, nil
}

func (s *DefinitionService) RegenerateDefinitions(
	ctx context.Context,
	wordID, userID int64,
	reportedDefinitionID *int64,
	definitions []*model.Definition,
	tx pgx.Tx,
) ([]*model.Definition, error) {
	normalizeDefinitionsForPersistence(definitions)
	if reportedDefinitionID == nil {
		return s.definitionRepository.Save(ctx, wordID, definitions, &tx)
	}
	regenerated, err := s.definitionRepository.ReplaceReportedDefinition(
		ctx, wordID, userID, *reportedDefinitionID, definitions, tx,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to replace reported definition: %w", err)
	}
	return regenerated, nil
}

func normalizeDefinitionsForPersistence(definitions []*model.Definition) {
	for _, def := range definitions {
		def.PartOfSpeechNormalized = NormalizePartOfSpeech(def.PartOfSpeech, def.Language)
		for index, example := range def.Examples {
			def.Examples[index] = strings.ToValidUTF8(example, "")
		}
	}
}

func (s *DefinitionService) GetRandomMeanings(ctx context.Context, scope DistractorScope, definitionIDsToIgnore []int, size int) ([]string, error) {
	return s.definitionRepository.GetRandomMeanings(
		ctx,
		scope.UserID,
		scope.WordlistID,
		scope.Language,
		scope.PartOfSpeechNormalized,
		definitionIDsToIgnore,
		size,
	)
}

func (s *DefinitionService) GetRandomTokens(ctx context.Context, scope DistractorScope, definitionIDsToIgnore []int, size int) ([]string, error) {
	return s.definitionRepository.GetRandomTokens(
		ctx,
		scope.UserID,
		scope.WordlistID,
		scope.Language,
		scope.PartOfSpeechNormalized,
		definitionIDsToIgnore,
		size,
	)
}

func (s *DefinitionService) GetDefinitionByID(ctx context.Context, id int64) (*model.Definition, error) {
	results, err := s.definitionRepository.Find(ctx, repo.FindArgs{ID: &id})
	if err != nil {
		return nil, fmt.Errorf("failed to find definition %d: %w", id, err)
	}
	if len(results) == 0 {
		return nil, common.NotFoundError{ID: id, Entity: "definition"}
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

func (s *DefinitionService) GetDefinitionsByWordID(ctx context.Context, wordlistID, wordID, userID int64) ([]*model.Definition, error) {
	return s.definitionRepository.GetDefinitionsByWordID(ctx, wordlistID, wordID, userID)
}

// GetDefinitionsByWordIDs returns a denormalized response including wordID, token(name) and its definitions
func (s *DefinitionService) GetDefinitionsByWordIDs(ctx context.Context, wordlistID, userID int64, wordIDs []int64) ([]repo.WordDefinitionsResponse, error) {
	return s.definitionRepository.GetDefinitionsByWordIDs(ctx, wordlistID, userID, wordIDs)
}

func (s *DefinitionService) CreateExampleAudio(ctx context.Context, definitionID int64, exampleText, audioURL, inflectionType string) error {
	return s.definitionRepository.CreateExampleAudio(ctx, definitionID, exampleText, audioURL, inflectionType)
}

func (s *DefinitionService) SetMeaningAudioURLIfEmpty(ctx context.Context, definitionID int64, audioURL string) (bool, error) {
	return s.definitionRepository.SetMeaningAudioURLIfEmpty(ctx, definitionID, audioURL)
}

// NormalizePartOfSpeech converts a language-specific part-of-speech to normalized English
// This function uses the PartOfSpeechMappings from LanguageConfigs as the single source of truth
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
