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
	"github.com/jackc/pgx/v5"
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
}

// getWordlistLanguageAndPronunciation retrieves the language code and pronunciation system for a wordlist from a word ID
func getWordlistLanguageAndPronunciation(wordID int64) (string, model.PronunciationSystem, error) {
	db, err := common.GetDBConnection()
	if err != nil {
		return "", "", err
	}

	query := `
		SELECT w.language_code, w.pronunciation_system
		FROM words wd 
		JOIN wordlists w ON wd.wordlist_id = w.id 
		WHERE wd.id = $1`

	var languageCode string
	var pronunciationSystem model.PronunciationSystem
	err = db.QueryRow(context.Background(), query, wordID).Scan(&languageCode, &pronunciationSystem)
	if err != nil {
		return "", "", err
	}

	return languageCode, pronunciationSystem, nil
}

// getWordlistLanguage is a backward compatibility wrapper for other workers that only need language code
func getWordlistLanguage(wordID int64) (string, error) {
	languageCode, _, err := getWordlistLanguageAndPronunciation(wordID)
	return languageCode, err
}

func (w *DefinitionFetcherWorker) Work(ctx context.Context, job *river.Job[DefinitionFetcherArgs]) error {
	// Validate user eligibility before processing (skip if userId is nil - admin context)
	if job.Args.UserID != nil {
		if err := ValidateUserEligibilityForWorkers(*job.Args.UserID); err != nil {
			common.Logger.Warn("User not eligible for definition fetching",
				"userId", *job.Args.UserID, "wordId", job.Args.WordId, "error", err)
			// Cancel job permanently - user needs to upgrade
			return river.JobCancel(err)
		}
	}
	logger := common.Logger.With("worker", "DefinitionFetcher")
	wordID := job.Args.WordId
	word, err := GetWordById(wordID)

	if err != nil {
		if errors.Is(err, common.NotFoundError{}) {
			return river.JobCancel(errors.New("word not found"))
		}
		return err
	}

	// Get wordlist language and pronunciation system
	languageCode, pronunciationSystem, err := getWordlistLanguageAndPronunciation(wordID)
	if err != nil {
		logger.Error("failed to get wordlist language and pronunciation system", "error", err)
		return river.JobCancel(err)
	}

	logger.Info("fetching definitions", "word", word.Name, "language", languageCode, "pronunciationSystem", pronunciationSystem)

	definitionData, err := openai.GetDefinition(word.Name, languageCode, pronunciationSystem)
	if err != nil {
		logger.Error("failed to fetch definitions using openai", "error", err)
		return err
	}

	if len(definitionData.Definitions) == 0 {
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
			return err
		}

		return errors.New(string(json))
	}

	db, err := common.GetDBConnection()
	if err != nil {
		return err
	}

	tx, err := db.Begin(ctx)
	if err != nil {
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

	definitions, err := SaveDefinition(word.Name, word.ID, definitionData.Definitions, tx)

	if err != nil {
		logger.Error("failed to save definitions", "error", err)
		return err
	}

	// Save pronunciation to word table
	if definitionData.Pronunciation != "" {
		_, err = tx.Exec(ctx, "UPDATE words SET pronunciation = $1 WHERE id = $2", definitionData.Pronunciation, word.ID)
		if err != nil {
			logger.Error("failed to save pronunciation", "error", err)
			return err
		}
		logger.Info("pronunciation saved", "pronunciation", definitionData.Pronunciation, "wordId", word.ID)
	}

	definitionIds := []int64{}
	for _, definition := range definitions {
		definitionIds = append(definitionIds, definition.ID)

		_, err = TriggerGenerateImageWorker(definition.ID, job.Args.UserID, nil, &tx)

		if err != nil {
			logger.Error("failed to trigger image generator", "definitionId", definition.ID, "error", err)
		}

		// Queue example audio generation job
		err = QueueExampleAudioJob(definition.ID, word.ID, job.Args.UserID, &tx)
		if err != nil {
			logger.Error("failed to queue example audio job", "definitionId", definition.ID, "wordId", word.ID, "error", err)
		}
	}
	strategy := LeitnerSystemStrategy{}
	if includeErr := strategy.IncludeDefinitions(word.ID, word.UserID, definitionIds, tx); includeErr != nil {
		logger.Error("failed to include definitions in quiz strategy", "wordId", word.ID, "error", includeErr)
	}

	// if this job was triggered by an error report, then mark the issue as solved
	if job.Args.ErrorReport != nil {
		return strategy.MarkErrorResolved(*job.Args.ErrorReport)
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

// QueueExampleAudioJob queues a job to generate example audio for a definition
func QueueExampleAudioJob(definitionID int64, wordID int64, userID *int64, tx *pgx.Tx) error {
	client, err := GetRiverClient()
	if err != nil {
		return err
	}

	args := ExampleAudioArgs{
		DefinitionID: definitionID,
		WordID:       wordID,
		UserID:       userID,
	}

	opts := &river.InsertOpts{
		Queue: EXAMPLE_AUDIO_QUEUE,
	}

	if tx != nil {
		_, err = client.InsertTx(context.Background(), *tx, args, opts)
	} else {
		_, err = client.Insert(context.Background(), args, opts)
	}

	return err
}
