package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/openai"
	"github.com/riverqueue/river"
)

const (
	openAIErrorRateLimitExceeded = "rate_limit_exceeded"
	openAIErrorBillingLimit      = "billing_hard_limit_reached"
)

type ExampleAudioArgs struct {
	DefinitionID int64  `json:"definitionId"`
	WordID       int64  `json:"wordId"`
	UserID       *int64 `json:"userId"`
}

func (ExampleAudioArgs) Kind() string { return "ExampleAudio" }

func (w *ExampleAudioWorker) Timeout(*river.Job[ExampleAudioArgs]) time.Duration {
	return 10 * time.Minute
}

type ExampleAudioWorker struct {
	river.WorkerDefaults[ExampleAudioArgs]
	definitionService *DefinitionService
	wordService       *WordService
	userService       *UserService
	audioGenerator    AudioGenerator
}

type AudioGenerator func(ctx context.Context, text, languageCode string) (*openai.GenerateAudioResponse, error)

// NewExampleAudioWorker creates a new example audio worker with dependencies
func NewExampleAudioWorker(definitionService *DefinitionService, wordService *WordService, userService *UserService, generators ...AudioGenerator) *ExampleAudioWorker {
	generator := AudioGenerator(openai.GenerateAudio)
	if len(generators) > 0 && generators[0] != nil {
		generator = generators[0]
	}
	return &ExampleAudioWorker{
		definitionService: definitionService,
		wordService:       wordService,
		userService:       userService,
		audioGenerator:    generator,
	}
}

type ExampleAudioItem struct {
	ExampleText    string
	InflectionType string
}

func (w *ExampleAudioWorker) Work(ctx context.Context, job *river.Job[ExampleAudioArgs]) error {
	logger := common.Logger.With("worker", "exampleaudio", "DefinitionID", job.Args.DefinitionID, "WordID", job.Args.WordID, "UserID", job.Args.UserID)

	// Validate user eligibility before processing (skip for admin/system jobs)
	if job.Args.UserID != nil {
		if err := w.userService.ValidateUserEligibilityForWorkers(ctx, *job.Args.UserID); err != nil {
			logger.Warn("User not eligible for example audio generation",
				"userId", *job.Args.UserID, "wordId", job.Args.WordID, "error", err)
			// Cancel job permanently - user needs to upgrade
			return river.JobCancel(err)
		}
	}

	// 1. Fetch definition from database
	definition, err := w.definitionService.GetDefinitionByID(ctx, job.Args.DefinitionID)
	if err != nil {
		logger.Error("failed to get definition", "error", err)
		return exampleAudioDefinitionLookupError(err)
	}

	// 2. Get wordlist language for proper TTS voice selection
	wordlistLang, _, err := w.wordService.GetWordlistLanguageAndPronunciation(ctx, job.Args.WordID)
	if err != nil {
		logger.Error("failed to get wordlist language", "error", err)
		return err
	}

	// 3. Apply smart example selection based on part of speech
	selectedExamples := w.selectExamplesForAudio(definition)

	logger.Info("processing examples", "count", len(selectedExamples), "partOfSpeech", definition.PartOfSpeech)

	// 4. Generate audio for selected examples
	for _, exampleItem := range selectedExamples {
		// Generate audio using OpenAI TTS with language-specific voice
		response, err := w.audioGenerator(ctx, exampleItem.ExampleText, wordlistLang)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}
			// Log error but continue with other examples
			logger.Error("failed to generate audio for example", "example", exampleItem.ExampleText, "error", err)
			continue
		}

		if response.Error != nil {
			logger.Error("OpenAI error for example", "example", exampleItem.ExampleText, "error", response.Error)

			switch response.Error.Code {
			case openAIErrorRateLimitExceeded:
				// Return error to trigger retry for all examples
				//nolint:gosec // G404 - using weak random for rate limiting jitter, not cryptographic security
				return river.JobSnooze(time.Minute + (time.Duration(rand.Intn(60)) * time.Second))
			case openAIErrorBillingLimit:
				logger.Warn(response.Error.Message)
			}

			continue
		}

		// Upload audio to MinIO storage
		audioURL, err := w.uploadAudio(ctx, response.Data, job.Args.DefinitionID, exampleItem.ExampleText)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}
			logger.Error("failed to upload audio for example", "example", exampleItem.ExampleText, "error", err)
			continue
		}

		// Create new record in definition_example_audio table
		err = w.definitionService.CreateExampleAudio(
			ctx, job.Args.DefinitionID,
			exampleItem.ExampleText,
			audioURL,
			exampleItem.InflectionType,
		)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}
			logger.Error("failed to save audio record for example", "example", exampleItem.ExampleText, "error", err)
			continue
		}

		logger.Debug("example audio generated", "definitionId", job.Args.DefinitionID, "example", exampleItem.ExampleText, "url", audioURL)
	}

	return nil
}

func exampleAudioDefinitionLookupError(err error) error {
	var notFoundErr common.NotFoundError
	if errors.As(err, &notFoundErr) {
		return river.JobCancel(err)
	}
	return err
}

func (w *ExampleAudioWorker) selectExamplesForAudio(definition *model.Definition) []ExampleAudioItem {
	var selectedExamples []ExampleAudioItem

	// Smart selection based on part of speech
	if definition.IsVerbType {
		// For verbs: Skip main examples (should be empty per schema, or duplicates from fallback logic)
		// Only select the longest example from each inflection for cost optimization
		for _, inflection := range definition.Inflections {
			if longestExample := findLongestExample(inflection.Examples); longestExample != "" {
				selectedExamples = append(selectedExamples, ExampleAudioItem{
					ExampleText:    longestExample,
					InflectionType: inflection.Tense,
				})
			}
		}
	} else {
		for _, example := range definition.Examples {
			selectedExamples = append(selectedExamples, ExampleAudioItem{
				ExampleText:    example,
				InflectionType: "",
			})
		}
	}

	return selectedExamples
}

func (w *ExampleAudioWorker) uploadAudio(ctx context.Context, audioData []byte, definitionID int64, exampleText string) (string, error) {
	// Generate hash for consistent naming
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(exampleText)))[:8]
	filename := fmt.Sprintf("audio/example-%d-%s.mp3", definitionID, hash)

	audioURL, err := common.Upload(ctx, audioData, "decorebator", filename, "audio/mpeg")
	if err != nil {
		return "", err
	}

	return audioURL, nil
}

func findLongestExample(examples []string) string {
	longest := ""
	for _, example := range examples {
		if len(example) > len(longest) {
			longest = example
		}
	}
	return longest
}
