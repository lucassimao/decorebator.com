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
	"github.com/riverqueue/river"
)

type DefinitionFetcherArgs struct {
	WordId      int64        `json:"wordId"`
	ErrorReport *ErrorReport `json:"errorReport"`
}

func (DefinitionFetcherArgs) Kind() string { return "DefinitionFetcher" }

type DefinitionFetcherWorker struct {
	river.WorkerDefaults[DefinitionFetcherArgs]
}

func (w *DefinitionFetcherWorker) Work(ctx context.Context, job *river.Job[DefinitionFetcherArgs]) error {
	logger := common.Logger.With("worker", "DefinitionFetcher")
	wordID := job.Args.WordId
	word, err := GetWordById(wordID)

	if err != nil {
		if errors.Is(err, common.NotFoundError{}) {
			return river.JobCancel(errors.New("word not found"))
		}
		return err
	}

	definitionData, err := openai.GetDefinition(word.Name)
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
		json, err := json.Marshal(definitionData)
		if err != nil {
			logger.Debug(string(json))
		} else {
			logger.Error("fail", "err", err)
		}

		for _, validationErr := range validationErrors {
			logger.Warn("definition validation warning", "word", word.Name, "issue", validationErr)
		}
	}

	// Check if we have at least one valid definition
	hasValidDefinition := false
	for _, def := range definitionData.Definitions {
		if def.PartOfSpeech != "" && def.Meaning != "" {
			hasValidDefinition = true
			break
		}
	}

	if !hasValidDefinition {
		logger.Error("no valid definitions received from ChatGPT", "word", word.Name)
		return errors.New("all definitions received are invalid")
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
			tx.Commit(ctx)
		} else {
			tx.Rollback(ctx)
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

		_, err = TriggerGenerateImageWorker(definition.ID, "", nil, &tx)

		if err != nil {
			logger.Error("failed to trigger image generator", "definitionId", definition.ID, "error", err)
		}
	}
	strategy := LeitnerSystemStrategy{}
	strategy.IncludeDefinitions(word.ID, word.UserID, definitionIds, tx)

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

		if def.Token == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("definition %d: missing token", i+1))
		}

		// Check part of speech specific requirements
		if def.PartOfSpeech == "verb" || def.PartOfSpeech == "phrasal verb" {
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
		if def.Token != "" && strings.ToLower(def.Token) != strings.ToLower(word) {
			validationErrors = append(validationErrors, fmt.Sprintf("definition %d: token mismatch - expected '%s', got '%s'", i+1, word, def.Token))
		}

		// Check that meaning is not too short
		if len(def.Meaning) < 10 {
			validationErrors = append(validationErrors, fmt.Sprintf("definition %d: meaning too short (%d chars)", i+1, len(def.Meaning)))
		}
	}

	return validationErrors
}
