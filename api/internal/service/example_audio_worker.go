package service

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/openai"
	"decorebator.com/internal/repository"
	"github.com/riverqueue/river"
)

type ExampleAudioArgs struct {
	DefinitionID int64 `json:"definitionId"`
	WordID       int64 `json:"wordId"`
}

func (ExampleAudioArgs) Kind() string { return "ExampleAudio" }

type ExampleAudioWorker struct {
	river.WorkerDefaults[ExampleAudioArgs]
}

type ExampleAudioItem struct {
	ExampleText    string
	InflectionType string
}

func (w *ExampleAudioWorker) Work(ctx context.Context, job *river.Job[ExampleAudioArgs]) error {
	logger := common.Logger.With("worker", "exampleaudio", "DefinitionID", job.Args.DefinitionID, "WordID", job.Args.WordID)

	// Validate user eligibility before processing
	if err := ValidateWordEligibilityForWorkers(job.Args.WordID); err != nil {
		logger.Warn("User not eligible for example audio generation",
			"wordId", job.Args.WordID, "error", err)
		// Cancel job permanently - user needs to upgrade
		return river.JobCancel(err)
	}

	// 1. Fetch definition from database
	db, err := common.GetDBConnection()
	if err != nil {
		logger.Error("failed to get database connection", "error", err)
		return err
	}

	definitionRepo := repository.NewDefinitionRepository(db)
	definition, err := definitionRepo.GetDefinitionByID(job.Args.DefinitionID)
	if err != nil && errors.Is(err, common.NotFoundError{}) {
		return river.JobCancel(errors.New("definition not found"))
	}
	if err != nil {
		logger.Error("failed to get definition", "error", err)
		return err
	}

	// 2. Get wordlist language for proper TTS voice selection
	wordlistLang, err := getWordlistLanguage(job.Args.WordID)
	if err != nil {
		logger.Error("failed to get wordlist language", "error", err)
		return err
	}

	// 3. Select appropriate voice based on language
	voice := getVoiceForLanguage(wordlistLang)

	// 4. Apply smart example selection based on part of speech
	selectedExamples := w.selectExamplesForAudio(definition)

	logger.Info("processing examples", "count", len(selectedExamples), "partOfSpeech", definition.PartOfSpeech)

	// 5. Generate audio for selected examples
	for _, exampleItem := range selectedExamples {
		// Generate audio using OpenAI TTS with language-specific voice
		response, err := openai.GenerateAudio(exampleItem.ExampleText, voice)
		if err != nil {
			// Log error but continue with other examples
			logger.Error("failed to generate audio for example", "example", exampleItem.ExampleText, "error", err)
			continue
		}

		if response.Error != nil {
			logger.Error("OpenAI error for example", "example", exampleItem.ExampleText, "error", response.Error)

			switch response.Error.Code {
			case "rate_limit_exceeded":
				// Return error to trigger retry for all examples
				return river.JobSnooze(time.Minute + (time.Duration(rand.Intn(60)) * time.Second))
			case "billing_hard_limit_reached":
				logger.Warn(response.Error.Message)
			}

			continue
		}

		// Upload audio to MinIO storage
		audioURL, err := w.uploadAudio(response.Data, job.Args.DefinitionID, exampleItem.ExampleText)
		if err != nil {
			logger.Error("failed to upload audio for example", "example", exampleItem.ExampleText, "error", err)
			continue
		}

		// Create new record in definition_example_audio table
		err = definitionRepo.CreateExampleAudio(
			job.Args.DefinitionID,
			exampleItem.ExampleText,
			audioURL,
			exampleItem.InflectionType,
		)
		if err != nil {
			logger.Error("failed to save audio record for example", "example", exampleItem.ExampleText, "error", err)
			continue
		}

		logger.Debug("example audio generated", "definitionId", job.Args.DefinitionID, "example", exampleItem.ExampleText, "url", audioURL)
	}

	return nil
}

func (w *ExampleAudioWorker) selectExamplesForAudio(definition *model.Definition) []ExampleAudioItem {
	var selectedExamples []ExampleAudioItem

	// Smart selection based on part of speech
	if definition.IsVerbType {
		// For verbs: Only select the longest example from main examples
		if longestExample := findLongestExample(definition.Examples); longestExample != "" {
			selectedExamples = append(selectedExamples, ExampleAudioItem{
				ExampleText:    longestExample,
				InflectionType: "", // Main example
			})
		}

		// For verbs: Only select the longest example from each inflection
		for _, inflection := range definition.Inflections {
			if longestExample := findLongestExample(inflection.Examples); longestExample != "" {
				selectedExamples = append(selectedExamples, ExampleAudioItem{
					ExampleText:    longestExample,
					InflectionType: inflection.Tense,
				})
			}
		}
	} else {
		// For non-verbs: Select all examples
		for _, example := range definition.Examples {
			selectedExamples = append(selectedExamples, ExampleAudioItem{
				ExampleText:    example,
				InflectionType: "",
			})
		}

		for _, inflection := range definition.Inflections {
			for _, example := range inflection.Examples {
				selectedExamples = append(selectedExamples, ExampleAudioItem{
					ExampleText:    example,
					InflectionType: inflection.Tense,
				})
			}
		}
	}

	return selectedExamples
}

func (w *ExampleAudioWorker) uploadAudio(audioData []byte, definitionID int64, exampleText string) (string, error) {
	// Generate hash for consistent naming
	hash := fmt.Sprintf("%x", md5.Sum([]byte(exampleText)))[:8]
	filename := fmt.Sprintf("audio/example-%d-%s.mp3", definitionID, hash)

	audioURL, err := common.Upload(audioData, "decorebator", filename, "audio/mpeg")
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

// Language to voice mapping for OpenAI TTS
func getVoiceForLanguage(language string) string {
	voiceMap := map[string]string{
		"en": "alloy",   // English - Clear, versatile voice
		"es": "nova",    // Spanish - Natural pronunciation with proper accent
		"fr": "shimmer", // French - Elegant pronunciation with proper liaison
		"de": "echo",    // German - Clear consonants for complex grammar
		"it": "fable",   // Italian - Expressive speech with proper intonation
		"pt": "onyx",    // Portuguese - Deep, clear pronunciation
		"ja": "alloy",   // Japanese - Optimized for Japanese phonetics
	}

	if voice, exists := voiceMap[language]; exists {
		return voice
	}
	return "alloy" // Default fallback
}
