package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/openai"
	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

type DefinitionFetcherArgs struct {
	WordID      int64        `json:"wordId"`
	UserID      *int64       `json:"userId"`
	ErrorReport *ErrorReport `json:"errorReport"`
}

func (DefinitionFetcherArgs) Kind() string { return "DefinitionFetcher" }

func (w *DefinitionFetcherWorker) Timeout(*river.Job[DefinitionFetcherArgs]) time.Duration {
	return 6 * time.Minute
}

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

func (w *DefinitionFetcherWorker) Work(ctx context.Context, job *river.Job[DefinitionFetcherArgs]) error { //nolint:gocyclo // Staged retry-sensitive worker pipeline.
	txn := sentry.StartTransaction(ctx, "worker.DefinitionFetcher", sentry.WithOpName("queue.process"), sentry.WithTransactionSource(sentry.SourceTask))
	defer txn.Finish()
	ctx = txn.Context()

	// Validate user eligibility before processing (skip if userID is nil - admin context)
	if job.Args.UserID != nil {
		if err := w.userService.ValidateUserEligibilityForWorkers(ctx, *job.Args.UserID); err != nil {
			common.Logger.WarnContext(ctx, "User not eligible for definition fetching",
				"userId", *job.Args.UserID, "wordId", job.Args.WordID, "error", err)
			// Cancel job permanently - user needs to upgrade
			return river.JobCancel(err)
		}
	}
	logger := common.Logger.With("worker", "DefinitionFetcher")
	wordID := job.Args.WordID

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

	span := sentry.StartSpan(ctx, "openai.definition.generate", sentry.WithDescription("openai.GetDefinition"))
	definitionData, err := openai.GetDefinition(span.Context(), word.Name, languageCode, pronunciationSystem)
	span.Finish()
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

		detailsJSON, marshalErr := json.Marshal(details)
		if marshalErr != nil {
			if updateErr := w.wordService.UpdateProcessingStatus(ctx, wordID, "failed", "Failed to validate definitions", nil); updateErr != nil {
				logger.ErrorContext(ctx, "failed to update processing status", "wordID", wordID, "error", updateErr)
			}
			return marshalErr
		}

		errorMsg := fmt.Sprintf("Definition validation failed: %v", validationErrors)
		if updateErr := w.wordService.UpdateProcessingStatus(ctx, wordID, "failed", errorMsg, nil); updateErr != nil {
			logger.ErrorContext(ctx, "failed to update processing status", "wordID", wordID, "error", updateErr)
		}
		return errors.New(string(detailsJSON))
	}

	// Get database connection for transaction
	tx, err := w.db.Begin(ctx)
	if err != nil {
		if err = w.wordService.UpdateProcessingStatus(ctx, wordID, "failed", "Failed to start database transaction", nil); err != nil {
			logger.ErrorContext(ctx, "failed to update processing status", "wordId", wordID, "error", err)
		}
		return err
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	span = sentry.StartSpan(ctx, "db.definition.save", sentry.WithDescription("definitionService.SaveDefinition"))
	definitions, err := w.definitionService.SaveDefinition(span.Context(), word.ID, definitionData.Definitions, &tx)
	span.Finish()

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

	definitionIDs := []int64{}
	for _, definition := range definitions {
		definitionIDs = append(definitionIDs, definition.ID)

		_, err = w.jobService.ScheduleImageJob(ctx, definition.ID, job.Args.UserID, nil, &tx)

		if err != nil {
			logger.Error("failed to trigger image generator", "definitionId", definition.ID, "error", err)
		}

		err = w.jobService.ScheduleExampleAudioJob(ctx, definition.ID, word.ID, job.Args.UserID, &tx)
		if err != nil {
			logger.Error("failed to queue example audio job", "definitionId", definition.ID, "wordId", word.ID, "error", err)
		}

		if definition.Meaning != "" {
			// Meaning audio is quiz content, not best-effort decoration. Persist the
			// definition and its durable outbox job as one unit so a queue insert
			// failure retries the definition job instead of committing a definition
			// that can never gain meaning audio.
			if err := scheduleRequiredMeaningAudio(ctx, w.jobService, definition.ID, word.ID, job.Args.UserID, &tx); err != nil {
				return err
			}
		}
	}
	if includeErr := w.leitnerTrackingService.IncludeDefinitions(ctx, word.ID, word.UserID, definitionIDs, tx); includeErr != nil {
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
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit definition transaction: %w", err)
	}
	return nil
}

func scheduleRequiredMeaningAudio(ctx context.Context, jobService JobService, definitionID, wordID int64, userID *int64, tx *pgx.Tx) error {
	if err := jobService.ScheduleMeaningAudioJob(ctx, definitionID, wordID, userID, tx); err != nil {
		return fmt.Errorf("failed to queue required meaning audio for definition %d: %w", definitionID, err)
	}
	return nil
}

// validateDefinitions checks the quality and completeness of definitions received from ChatGPT
func validateDefinitions(word string, definitions []*model.Definition) []string {
	var validationErrors []string

	for i, def := range definitions {
		validationErrors = append(validationErrors, validateDefinition(word, i+1, def)...)
	}

	return validationErrors
}

func validateDefinition(word string, index int, def *model.Definition) []string {
	validationErrors := validateRequiredDefinitionFields(index, def)

	normalizedPOS := NormalizePartOfSpeech(def.PartOfSpeech, def.Language)
	if normalizedPOS == "verb" || normalizedPOS == "phrasal verb" {
		validationErrors = append(validationErrors, validateVerbDefinition(index, def)...)
	} else {
		validationErrors = append(validationErrors, validateNonVerbDefinition(index, def)...)
	}

	if def.Token != "" && !strings.EqualFold(def.Token, word) {
		validationErrors = append(validationErrors, fmt.Sprintf("definition %d: token mismatch - expected '%s', got '%s'", index, word, def.Token))
	}
	return validationErrors
}

func validateRequiredDefinitionFields(index int, def *model.Definition) []string {
	var validationErrors []string
	if def.PartOfSpeech == "" {
		validationErrors = append(validationErrors, fmt.Sprintf("definition %d: missing part_of_speech", index))
	}
	if def.Meaning == "" {
		validationErrors = append(validationErrors, fmt.Sprintf("definition %d: missing meaning", index))
	}
	if def.Language == "" {
		validationErrors = append(validationErrors, fmt.Sprintf("definition %d: missing language", index))
	}
	if def.Token == "" {
		validationErrors = append(validationErrors, fmt.Sprintf("definition %d: missing token", index))
	}
	return validationErrors
}

func validateVerbDefinition(index int, def *model.Definition) []string {
	var validationErrors []string
	if len(def.Examples) > 0 {
		validationErrors = append(validationErrors, fmt.Sprintf("definition %d: verb/phrasal verb must have empty examples array but has %d examples", index, len(def.Examples)))
	}
	if len(def.Inflections) == 0 {
		return append(validationErrors, fmt.Sprintf("definition %d: verb missing inflections", index))
	}

	hasPresent, hasPast, hasPastParticiple := false, false, false
	allExamples := make(map[string][]string)
	for _, inflection := range def.Inflections {
		hasPresent = hasPresent || inflection.Tense == "present"
		hasPast = hasPast || inflection.Tense == "past"
		hasPastParticiple = hasPastParticiple || inflection.Tense == "past participle"
		if inflection.Inflection == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("definition %d: inflection missing form", index))
		}
		if len(inflection.Examples) == 0 {
			validationErrors = append(validationErrors, fmt.Sprintf("definition %d: inflection %s missing examples", index, inflection.Tense))
		}
		for _, example := range inflection.Examples {
			if !strings.Contains(example, "[") || !strings.Contains(example, "]") {
				validationErrors = append(validationErrors, fmt.Sprintf("definition %d: inflection example missing brackets: %s", index, example))
			}
			allExamples[example] = append(allExamples[example], inflection.Tense)
		}
	}

	for example, tenses := range allExamples {
		if len(tenses) > 1 {
			validationErrors = append(validationErrors, fmt.Sprintf("definition %d: duplicate example across inflections '%s' found in tenses: %v", index, example, tenses))
		}
	}
	if !hasPresent {
		validationErrors = append(validationErrors, fmt.Sprintf("definition %d: verb missing present tense inflection", index))
	}
	if !hasPast {
		validationErrors = append(validationErrors, fmt.Sprintf("definition %d: verb missing past tense inflection", index))
	}
	if !hasPastParticiple {
		validationErrors = append(validationErrors, fmt.Sprintf("definition %d: verb missing past participle inflection", index))
	}
	return validationErrors
}

func validateNonVerbDefinition(index int, def *model.Definition) []string {
	if len(def.Examples) == 0 {
		return []string{fmt.Sprintf("definition %d: non-verb missing examples", index)}
	}

	var validationErrors []string
	for _, example := range def.Examples {
		if !strings.Contains(example, "[") || !strings.Contains(example, "]") {
			validationErrors = append(validationErrors, fmt.Sprintf("definition %d: example missing brackets: %s", index, example))
		}
	}
	return validationErrors
}
