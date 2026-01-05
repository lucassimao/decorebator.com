package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/openai"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

type DefinitionFetcherArgs struct {
	WordId      int64        `json:"wordId"`
	UserID      *int64       `json:"userId"`
	ErrorReport *ErrorReport `json:"errorReport"`
}

func (DefinitionFetcherArgs) Kind() string { return "DefinitionFetcher" }

type DefinitionFetcherWorker struct {
	river.WorkerDefaults[DefinitionFetcherArgs]
	db                     *pgxpool.Pool
	wordService            *WordService
	definitionService      *DefinitionService
	jobService             JobService
	leitnerSystemStrategy  *LeitnerSystemStrategy
	leitnerTrackingService *LeitnerTrackingService
	userService            *UserService
}

func NewDefinitionFetcherWorker(db *pgxpool.Pool, wordService *WordService, definitionService *DefinitionService, jobService JobService, leitnerSystemStrategy *LeitnerSystemStrategy, leitnerTrackingService *LeitnerTrackingService, userService *UserService) *DefinitionFetcherWorker {
	return &DefinitionFetcherWorker{
		db:                     db,
		wordService:            wordService,
		definitionService:      definitionService,
		jobService:             jobService,
		leitnerSystemStrategy:  leitnerSystemStrategy,
		leitnerTrackingService: leitnerTrackingService,
		userService:            userService,
	}
}

func (w *DefinitionFetcherWorker) Work(ctx context.Context, job *river.Job[DefinitionFetcherArgs]) error {
	// Validate user eligibility before processing (skip if userId is nil - admin context)
	if job.Args.UserID != nil {
		if err := w.userService.ValidateUserEligibilityForWorkers(ctx, *job.Args.UserID); err != nil {
			common.Logger.WarnContext(ctx, "User not eligible for definition fetching",
				"userId", *job.Args.UserID, "wordId", job.Args.WordId, "error", err)
			// Cancel job permanently - user needs to upgrade
			return river.JobCancel(err)
		}
	}
	logger := common.Logger.With("worker", "DefinitionFetcher")
	wordID := job.Args.WordId

	word, err := w.wordService.GetWordByID(ctx, wordID)

	if err != nil {
		if errors.Is(err, common.NotFoundError{}) {
			return river.JobCancel(errors.New("word not found"))
		}
		if updateErr := w.wordService.UpdateProcessingStatus(ctx, wordID, "failed", err.Error(), nil); updateErr != nil {
			common.Logger.ErrorContext(ctx, "failed to update processing status", "wordID", wordID, "error", updateErr)
		}
		return err
	}

	// Update word processing status to "processing"
	if updateErr := w.wordService.UpdateProcessingStatus(ctx, wordID, "processing", "", nil); updateErr != nil {
		logger.ErrorContext(ctx, "failed to update processing status to processing", "wordId", wordID, "error", updateErr)
		// Continue processing even if status update fails
	}

	// Get wordlist language and pronunciation system
	languageCode, pronunciationSystem, err := w.wordService.GetWordlistLanguageAndPronunciation(ctx, wordID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get wordlist language and pronunciation system", "error", err)
		if updateErr := w.wordService.UpdateProcessingStatus(ctx, wordID, "failed", "Failed to get wordlist language", nil); updateErr != nil {
			common.Logger.ErrorContext(ctx, "failed to update processing status", "wordID", wordID, "error", updateErr)
		}
		return err
	}

	logger.InfoContext(ctx, "fetching definitions", "word", word.Name, "language", languageCode, "pronunciationSystem", pronunciationSystem)

	definitionData, err := openai.GetDefinition(word.Name, languageCode, pronunciationSystem)
	if err != nil {
		logger.ErrorContext(ctx, "failed to fetch definitions using openai", "error", err)
		errorMsg := fmt.Sprintf("Failed to fetch definitions: %v", err)
		if updateErr := w.wordService.UpdateProcessingStatus(ctx, wordID, "failed", errorMsg, nil); updateErr != nil {
			logger.ErrorContext(ctx, "failed to update processing status", "wordID", wordID, "error", updateErr)
		}
		return err
	}

	if len(definitionData.Definitions) == 0 {
		if err = w.wordService.UpdateProcessingStatus(ctx, wordID, "failed", "No definitions found for this word", nil); err != nil {
			logger.ErrorContext(ctx, "failed to update processing status", "wordID", wordID, "error", err)
		}
		return river.JobCancel(errors.New("no definition found"))
	}

	// Validate the definitions received from ChatGPT
	validationErrors := validateDefinitions(word.Name, definitionData.Definitions)
	if len(validationErrors) > 0 {
		details := map[string]any{
			"definitionData":   definitionData,
			"validationErrors": validationErrors,
		}

		json, err := json.Marshal(details)
		if err != nil {
			if updateErr := w.wordService.UpdateProcessingStatus(ctx, wordID, "failed", "Failed to validate definitions", nil); updateErr != nil {
				logger.ErrorContext(ctx, "failed to update processing status", "wordID", wordID, "error", updateErr)
			}
			return err
		}

		errorMsg := fmt.Sprintf("Definition validation failed: %v", validationErrors)
		if err := w.wordService.UpdateProcessingStatus(ctx, wordID, "failed", errorMsg, nil); err != nil {
			logger.ErrorContext(ctx, "failed to update processing status", "wordID", wordID, "error", err)
		}
		return errors.New(string(json))
	}

	// Get database connection for transaction
	tx, err := w.db.Begin(ctx)
	if err != nil {
		if err = w.wordService.UpdateProcessingStatus(ctx, wordID, "failed", "Failed to start database transaction", nil); err != nil {
			logger.ErrorContext(ctx, "failed to update processing status", "wordId", wordID, "error", err)
		}
		return err
	}
	defer func() {
		if err == nil {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				logger.Error("failed to commit transaction in definition fetcher worker", "error", commitErr)
			}
		} else {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				logger.Error("failed to rollback transaction in definition fetcher worker", "error", rollbackErr)
			}
		}
	}()

	definitions, err := w.definitionService.SaveDefinition(ctx, word.ID, definitionData.Definitions, &tx)

	if err != nil {
		logger.Error("failed to save definitions", "error", err)
		if err = w.wordService.UpdateProcessingStatus(ctx, wordID, "failed", "Failed to save definitions", nil); err != nil {
			logger.Error("failed to update processing status", "wordId", wordID, "error", err)
		}
		return err
	}

	// Save pronunciation to word table
	if definitionData.Pronunciation != "" {
		err = w.wordService.UpdatePronunciation(ctx, word.ID, definitionData.Pronunciation, &tx)
		if err != nil {
			logger.Error("failed to save pronunciation", "error", err)
			if err = w.wordService.UpdateProcessingStatus(ctx, wordID, "failed", "Failed to save pronunciation", nil); err != nil {
				logger.Error("failed to update processing status", "wordId", wordID, "error", err)
			}
			return err
		}
		logger.Info("pronunciation saved", "pronunciation", definitionData.Pronunciation, "wordId", word.ID)
	}

	definitionIds := []int64{}
	for _, definition := range definitions {
		definitionIds = append(definitionIds, definition.ID)

		_, err = w.jobService.ScheduleImageJob(ctx, definition.ID, job.Args.UserID, nil, &tx)

		if err != nil {
			logger.Error("failed to trigger image generator", "definitionId", definition.ID, "error", err)
		}

		err = w.jobService.ScheduleExampleAudioJob(ctx, definition.ID, word.ID, job.Args.UserID, &tx)
		if err != nil {
			logger.Error("failed to queue example audio job", "definitionId", definition.ID, "wordId", word.ID, "error", err)
		}
	}
	if includeErr := w.leitnerTrackingService.IncludeDefinitions(ctx, word.ID, word.UserID, definitionIds, tx); includeErr != nil {
		logger.Error("failed to include definitions in quiz strategy", "wordId", word.ID, "error", includeErr)
	}

	// if this job was triggered by an error report, then mark the issue as solved
	if job.Args.ErrorReport != nil {
		if err := w.leitnerSystemStrategy.MarkErrorResolved(ctx, *job.Args.ErrorReport); err != nil {
			return err
		}
	}

	// Update word processing status to "completed"
	if err := w.wordService.UpdateProcessingStatus(ctx, wordID, "completed", "", &tx); err != nil {
		logger.Error("failed to update processing status to completed", "wordId", wordID, "error", err)
		// Don't fail the whole operation if status update fails
	}

	logger.Info("definitions fetched", "count", len(definitions))
	return nil
}

// validateDefinitions checks the quality and completeness of definitions received from ChatGPT
func validateDefinitions(word string, definitions []*model.Definition) []string {
	var validationErrors []string

	for i, def := range definitions {
		// Check required fields
		if def.PartOfSpeech == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("definition %d: missing part_of_speech", i+1))
		}

		if def.Meaning == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("definition %d: missing meaning", i+1))
		}

		if def.Language == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("definition %d: missing language", i+1))
		}

		if def.Token == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("definition %d: missing token", i+1))
		}

		// Check part of speech specific requirements
		normalizedPos := NormalizePartOfSpeech(def.PartOfSpeech, def.Language)
		if normalizedPos == "verb" || normalizedPos == "phrasal verb" {
			// CRITICAL: Verbs and phrasal verbs MUST have empty main examples array
			if len(def.Examples) > 0 {
				validationErrors = append(validationErrors, fmt.Sprintf("definition %d: verb/phrasal verb must have empty examples array but has %d examples", i+1, len(def.Examples)))
			}
			if len(def.Inflections) == 0 {
				validationErrors = append(validationErrors, fmt.Sprintf("definition %d: verb missing inflections", i+1))
			} else {
				// Validate inflections
				hasPresent := false
				hasPast := false
				hasPastParticiple := false

				for _, inflection := range def.Inflections {
					if inflection.Tense == "present" {
						hasPresent = true
					}
					if inflection.Tense == "past" {
						hasPast = true
					}
					if inflection.Tense == "past participle" {
						hasPastParticiple = true
					}

					if inflection.Inflection == "" {
						validationErrors = append(validationErrors, fmt.Sprintf("definition %d: inflection missing form", i+1))
					}

					if len(inflection.Examples) == 0 {
						validationErrors = append(validationErrors, fmt.Sprintf("definition %d: inflection %s missing examples", i+1, inflection.Tense))
					}

					// Check that examples contain the word in brackets
					for _, example := range inflection.Examples {
						if !strings.Contains(example, "[") || !strings.Contains(example, "]") {
							validationErrors = append(validationErrors, fmt.Sprintf("definition %d: inflection example missing brackets: %s", i+1, example))
						}
					}
				}

				// Check for duplicate examples across inflections
				allExamples := make(map[string][]string) // example -> list of tenses that have it
				for _, inflection := range def.Inflections {
					for _, example := range inflection.Examples {
						allExamples[example] = append(allExamples[example], inflection.Tense)
					}
				}

				for example, tenses := range allExamples {
					if len(tenses) > 1 {
						validationErrors = append(validationErrors, fmt.Sprintf("definition %d: duplicate example across inflections '%s' found in tenses: %v", i+1, example, tenses))
					}
				}

				if !hasPresent {
					validationErrors = append(validationErrors, fmt.Sprintf("definition %d: verb missing present tense inflection", i+1))
				}
				if !hasPast {
					validationErrors = append(validationErrors, fmt.Sprintf("definition %d: verb missing past tense inflection", i+1))
				}
				if !hasPastParticiple {
					validationErrors = append(validationErrors, fmt.Sprintf("definition %d: verb missing past participle inflection", i+1))
				}
			}
		} else {
			// For non-verbs, check that examples are provided
			if len(def.Examples) == 0 {
				validationErrors = append(validationErrors, fmt.Sprintf("definition %d: non-verb missing examples", i+1))
			} else {
				// Check that examples contain the word in brackets
				for _, example := range def.Examples {
					if !strings.Contains(example, "[") || !strings.Contains(example, "]") {
						validationErrors = append(validationErrors, fmt.Sprintf("definition %d: example missing brackets: %s", i+1, example))
					}
				}
			}
		}

		// Check that the token matches the word being defined (case insensitive)
		if def.Token != "" && !strings.EqualFold(def.Token, word) {
			validationErrors = append(validationErrors, fmt.Sprintf("definition %d: token mismatch - expected '%s', got '%s'", i+1, word, def.Token))
		}
	}

	return validationErrors
}
