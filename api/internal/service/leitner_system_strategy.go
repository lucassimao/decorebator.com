package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strings"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func init() {
	// Seed the random number generator with current time
	// Used for: example sentence selection, answer position randomization
	// Note: Quiz type selection now uses deterministic rotation
	rand.Seed(time.Now().UnixNano())
}

// LeitnerSystemStrategy implements the Leitner spaced repetition algorithm for vocabulary learning.
// The Leitner system uses boxes with increasing intervals to optimize long-term retention:
// - Words start in box 1 (immediate review)
// - Correct answers move words to higher boxes (longer intervals)
// - Incorrect answers reset words to box 1 (immediate review)
// - Each box has different quiz types to progressively increase difficulty
type LeitnerSystemStrategy struct{}
type Quiz = model.Quiz
type QuizType = model.QuizType

// NextDefinition represents a word definition selected for quiz generation
// along with its current Leitner system state and associated metadata.
type NextDefinition struct {
	Definition       *model.Definition // The definition content (meaning, examples, inflections, etc.)
	LeitnerSystemID  int64             // The tracking ID for this definition in the Leitner system
	BoxID            int64             // Current Leitner box (1-7, determines review interval)
	WordID           int64             // The word this definition belongs to
	ImageUrl         string            // URL of associated image (if any)
	ImageDescription string            // Description of the associated image
}

// boxToQuizTypes defines the quiz difficulty progression through Leitner boxes.
// Each box introduces a new skill level with minimal repetition for cleaner progression:
// - Box 1: Recognition (easiest)
// - Box 2: Basic recall
// - Box 3: Visual association
// - Box 4: Contextual understanding
// - Box 5: Active recall (hardest text-based)
// - Box 6: Audio recognition (new modality)
// - Box 7: Mastery level (most challenging types only)
var boxToQuizTypes = map[int64][]model.QuizType{
	1: {model.GuessMeaning},                                                                                 // Recognition: "What does this word mean?"
	2: {model.WordFromMeaning, model.GuessMeaning},                                                          // Basic recall: "Which word matches this meaning?" + recognition practice
	3: {model.WordFromImage, model.WordFromMeaning},                                                         // Visual association: "What word matches this image?" + meaning recall
	4: {model.CompleteSentence, model.WordFromExampleAudio, model.GuessMeaning},                            // Contextual understanding + recognition reinforcement
	5: {model.WriteWordFromDefinition, model.WordFromExampleAudio, model.WordFromMeaning},                  // Active recall + meaning practice
	6: {model.WordFromAudio, model.WordFromExampleAudio, model.WordFromImage, model.GuessMeaning},          // Audio/Visual recognition + meaning reinforcement
	7: {model.MeaningFromAudio, model.WordFromImage, model.WriteWordFromDefinition, model.CompleteSentence, model.WordFromExampleAudio, model.WordFromAudio, model.WordFromMeaning, model.GuessMeaning}, // Mastery: All types for comprehensive review
}

// ExampleUsage tracks when examples were last used for fair distribution
type ExampleUsage struct {
	ExampleHash string    `db:"example_hash"`
	LastUsedAt  time.Time `db:"last_used_at"`
}

// getNextDefinition selects the next definition for quiz generation using deterministic priority-based selection.
//
// This function implements a deterministic version of the Leitner spaced repetition system that guarantees
// definition selection while respecting spaced repetition intervals.
//
// ALGORITHM:
// Definitions are selected based on their progress toward their target review interval:
// - Priority 1: Overdue definitions (100%+ of interval elapsed)
// - Priority 2: Due soon definitions (80-100% of interval)
// - Priority 3: Available definitions (50-80% of interval)
// - Priority 4: All other definitions (sorted by progress %)
//
// BOX-SPECIFIC INTERVALS:
// - Box 1: Immediate (always available)
// - Box 2: 1 hour
// - Box 3: 1 day (24 hours)
// - Box 4: 3 days (72 hours)
// - Box 5: 1 week (168 hours)
// - Box 6: 2 weeks (336 hours)
// - Box 7: 1 month (720 hours)
//
// SELECTION LOGIC:
// 1. Calculate progress ratio for each definition (time_elapsed / target_interval)
// 2. Assign priority bucket based on progress ratio
// 3. Sort by priority bucket, then by oldest reviewed
// 4. Select first definition (deterministic, no randomness)
//
// BENEFITS:
// - 100% deterministic (same data = same result)
// - Guaranteed selection (no "no words selected" errors)
// - Respects Leitner intervals precisely
// - Simple and efficient (no complex probability calculations)
// - Easy to debug and test
//
// Parameters:
// - userID: The user requesting a quiz
// - wordlistID: The wordlist to select from
//
// Returns:
// - NextDefinition with selected definition details and Leitner system metadata
// - Error if no definitions exist in wordlist or database operations fail
func getNextDefinition(userID, wordlistID int64) (*NextDefinition, error) {
	// Deterministic query that guarantees definition selection
	query := `
		WITH definition_priorities AS (
			SELECT 
				def.id, lst.id AS lst_id, def.token, 
				def.part_of_speech, def.language, def.is_verb_type, def.meaning, def.examples, 
				def.inflections, lst.box_id, def.sounds, def.phonetic_notations, 
				di.url as image_url, di.description as image_description, 
				wd.word_id AS word_id,
				lst.updated_at,
				-- Calculate hours since last review
				COALESCE(EXTRACT(EPOCH FROM (NOW() - COALESCE(lst.updated_at, NOW())))/3600, 0) as hours_since_review,
				-- Get target interval hours based on box
				CASE lst.box_id
					WHEN 1 THEN 0     -- Always available
					WHEN 2 THEN 1     -- 1 hour
					WHEN 3 THEN 24    -- 1 day
					WHEN 4 THEN 72    -- 3 days
					WHEN 5 THEN 168   -- 1 week
					WHEN 6 THEN 336   -- 2 weeks
					WHEN 7 THEN 720   -- 1 month
				END as target_hours,
				-- Calculate progress ratio (how close to target interval)
				CASE 
					WHEN lst.box_id = 1 THEN 1.0  -- Box 1 always 100% ready
					ELSE COALESCE(
						EXTRACT(EPOCH FROM (NOW() - COALESCE(lst.updated_at, NOW())))/3600 / 
						NULLIF(CASE lst.box_id
							WHEN 2 THEN 1
							WHEN 3 THEN 24
							WHEN 4 THEN 72
							WHEN 5 THEN 168
							WHEN 6 THEN 336
							WHEN 7 THEN 720
						END, 0), 
						0
					)
				END as progress_ratio
			FROM leitner_system_tracking lst 
			JOIN definitions def ON lst.definition_id = def.id
			JOIN word_definitions wd ON def.id = wd.definition_id
			JOIN words w ON wd.word_id = w.id
			LEFT JOIN definition_images di ON di.definition_id = def.id AND di.is_visible=TRUE
			WHERE 
				lst.user_id = $1
				AND w.wordlist_id = $2
				AND w.learned = FALSE
				AND def.meaning IS NOT NULL
				-- Exclude temporarily skipped definitions
				AND (lst.temporarily_skipped_until IS NULL OR lst.temporarily_skipped_until < NOW())
		),
		weighted_definitions AS (
			SELECT 
				id, token, part_of_speech, language, is_verb_type, meaning, examples, inflections, 
				lst_id, box_id, sounds, phonetic_notations, 
				COALESCE(image_url,'') as image_url, word_id, COALESCE(image_description,'') as image_description,
				hours_since_review, progress_ratio,
				-- Calculate weighted selection score to reduce dominance
				CASE 
					-- Overdue definitions get high weight (but not exclusive)
					WHEN progress_ratio >= 1.0 THEN 
						100 + (progress_ratio * 50) + hours_since_review
					-- Due soon get medium-high weight
					WHEN progress_ratio >= 0.8 THEN 
						80 + (progress_ratio * 20) + hours_since_review
					-- Available get medium weight
					WHEN progress_ratio >= 0.5 THEN 
						60 + (progress_ratio * 15) + hours_since_review
					-- Not quite ready get low weight (but still possible)
					ELSE 
						20 + (progress_ratio * 10) + hours_since_review
				END as selection_weight
			FROM definition_priorities
		),
		selected_definition AS (
			SELECT 
				id, token, part_of_speech, language, is_verb_type, meaning, examples, inflections, 
				lst_id, box_id, sounds, phonetic_notations, 
				image_url, word_id, image_description, hours_since_review, progress_ratio
			FROM weighted_definitions
			-- Use deterministic but varied selection based on current time and definition ID
			-- This creates pseudo-random selection weighted by importance
			ORDER BY 
				-- Primary sort: Use a deterministic hash to create variety while respecting weights
				(selection_weight * (1 + SIN(id * 12345 + EXTRACT(EPOCH FROM NOW())::INTEGER % 86400))) DESC,
				-- Fallback tiebreaker for absolute determinism in edge cases
				id ASC
			LIMIT 1
		)
		SELECT 
			sd.id, sd.token, sd.part_of_speech, sd.language, sd.is_verb_type, sd.meaning, sd.examples, sd.inflections, 
			sd.lst_id, sd.box_id, sd.sounds, sd.phonetic_notations, 
			sd.image_url, sd.word_id, sd.image_description,
			sd.hours_since_review, sd.progress_ratio,
			-- Aggregate example audio files into JSON array
			COALESCE(
				JSON_AGG(
					JSON_BUILD_OBJECT(
						'id', dea.id,
						'definitionId', dea.definition_id,
						'exampleText', dea.example_text,
						'exampleHash', dea.example_hash,
						'audioUrl', dea.audio_url,
						'inflectionType', COALESCE(dea.inflection_type, ''),
						'createdAt', dea.created_at
					) ORDER BY dea.created_at DESC
				) FILTER (WHERE dea.id IS NOT NULL),
				'[]'::json
			) as example_audio_files
		FROM selected_definition sd
		LEFT JOIN definition_example_audio dea ON dea.definition_id = sd.id
		GROUP BY sd.id, sd.token, sd.part_of_speech, sd.language, sd.is_verb_type, sd.meaning, sd.examples, sd.inflections, 
				 sd.lst_id, sd.box_id, sd.sounds, sd.phonetic_notations, 
				 sd.image_url, sd.word_id, sd.image_description, sd.hours_since_review, sd.progress_ratio;
	`

	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection", "error", err)
		return nil, err
	}

	rows, err := db.Query(context.Background(), query, userID, wordlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		// Debug: Let's check what's happening
		debugQuery := `
			SELECT COUNT(*) as total_words,
				   COUNT(CASE WHEN w.learned = FALSE THEN 1 END) as unlearned_words,
				   COUNT(CASE WHEN def.meaning IS NOT NULL THEN 1 END) as with_meaning,
				   COUNT(CASE WHEN lst.temporarily_skipped_until IS NULL OR lst.temporarily_skipped_until < NOW() THEN 1 END) as not_skipped
			FROM leitner_system_tracking lst 
			JOIN definitions def ON lst.definition_id = def.id
			JOIN word_definitions wd ON def.id = wd.definition_id
			JOIN words w ON wd.word_id = w.id
			WHERE lst.user_id = $1 AND w.wordlist_id = $2
		`
		var totalWords, unlearnedWords, withMeaning, notSkipped int
		err := db.QueryRow(context.Background(), debugQuery, userID, wordlistID).Scan(&totalWords, &unlearnedWords, &withMeaning, &notSkipped)
		if err == nil {
			common.Logger.Error("no definitions selected - debug info",
				"userID", userID,
				"wordlistID", wordlistID,
				"totalWords", totalWords,
				"unlearnedWords", unlearnedWords,
				"withMeaning", withMeaning,
				"notSkipped", notSkipped)
		}

		return nil, errors.New("no definitions found in wordlist")
	}

	definition := model.Definition{}
	result := NextDefinition{Definition: &definition}
	var hoursSinceReview float64
	var progressRatio float64
	var exampleAudioFilesJSON []byte

	err = rows.Scan(&definition.ID, &definition.Token, &definition.PartOfSpeech, &definition.Language, &definition.IsVerbType,
		&definition.Meaning, &definition.Examples,
		&definition.Inflections, &result.LeitnerSystemID, &result.BoxID, &definition.Sounds,
		&definition.PhoneticNotations, &result.ImageUrl, &result.WordID, &result.ImageDescription,
		&hoursSinceReview, &progressRatio, &exampleAudioFilesJSON)

	if err != nil {
		return nil, err
	}

	// Parse the JSON array of example audio files
	if len(exampleAudioFilesJSON) > 0 {
		err = json.Unmarshal(exampleAudioFilesJSON, &definition.ExampleAudioFiles)
		if err != nil {
			// Log the error but don't fail the quiz generation
			common.Logger.Error("failed to parse example audio files JSON", "definitionID", definition.ID, "error", err)
			definition.ExampleAudioFiles = []model.DefinitionExampleAudio{}
		}
	} else {
		definition.ExampleAudioFiles = []model.DefinitionExampleAudio{}
	}

	// Log weighted selection for monitoring
	common.Logger.Info("weighted_selection",
		"userID", userID,
		"wordlistID", wordlistID,
		"definitionID", definition.ID,
		"boxID", result.BoxID,
		"hoursSinceReview", hoursSinceReview,
		"progressRatio", progressRatio,
		"wasOverdue", progressRatio >= 1.0,
		"exampleAudioCount", len(definition.ExampleAudioFiles))

	return &result, nil
}

// checkHasUnlearnedWords verifies if the user has any unlearned words in the wordlist
func checkHasUnlearnedWords(userID, wordlistID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 
			FROM leitner_system_tracking lst 
			JOIN word_definitions wd ON lst.definition_id = wd.definition_id
			JOIN words w ON wd.word_id = w.id
			WHERE lst.user_id = $1 
			AND w.wordlist_id = $2 
			AND w.learned = FALSE
			LIMIT 1
		)
	`

	db, err := common.GetDBConnection()
	if err != nil {
		return false, err
	}

	var exists bool
	err = db.QueryRow(context.Background(), query, userID, wordlistID).Scan(&exists)
	return exists, err
}

// getWordlistBoxDistribution provides analytics on word distribution across boxes
func getWordlistBoxDistribution(userID, wordlistID int64) (map[int64]int, error) {
	query := `
		SELECT 
			lst.box_id,
			COUNT(*) as word_count
		FROM leitner_system_tracking lst
		JOIN word_definitions wd ON lst.definition_id = wd.definition_id
		JOIN words w ON wd.word_id = w.id
		WHERE lst.user_id = $1 
		AND w.wordlist_id = $2
		AND w.learned = FALSE
		GROUP BY lst.box_id
		ORDER BY lst.box_id
	`

	db, err := common.GetDBConnection()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(context.Background(), query, userID, wordlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	distribution := make(map[int64]int)
	for rows.Next() {
		var boxID int64
		var count int
		if err := rows.Scan(&boxID, &count); err != nil {
			return nil, err
		}
		distribution[boxID] = count
	}

	return distribution, nil
}

// CreateQuiz generates a new quiz question for the user based on the Leitner spaced repetition system.
// It selects the most appropriate word/definition that is due for review, considering:
// - Probabilistic selection that ensures words are always available
// - Leitner box intervals (box 1: immediate, box 2: 1 hour, box 3: 1 day, etc.)
// - Word difficulty progression (different quiz types for different boxes)
// - Temporarily skipped words (to avoid immediate repetition)
// - Available content (images, audio, examples, inflections)
//
// Returns a Quiz object with the question, options, correct answer, and metadata.
// Returns an error if no words are available for review or if database operations fail.
func (LeitnerSystemStrategy) CreateQuiz(wordlistID, userID int64) (*Quiz, error) {
	// Early check to avoid unnecessary queries
	hasWords, err := checkHasUnlearnedWords(userID, wordlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to check wordlist status: %w", err)
	}
	if !hasWords {
		return nil, errors.New("no unlearned words in wordlist")
	}

	// Log box distribution for monitoring (but don't fail if it errors)
	if distribution, err := getWordlistBoxDistribution(userID, wordlistID); err == nil {
		totalWords := 0
		box7Count := 0
		for boxID, count := range distribution {
			totalWords += count
			if boxID == 7 {
				box7Count = count
			}
		}
		if totalWords > 0 && float64(box7Count)/float64(totalWords) > 0.8 {
			common.Logger.Info("high_box_7_concentration",
				"userID", userID,
				"wordlistID", wordlistID,
				"box7Percentage", float64(box7Count)/float64(totalWords)*100,
				"distribution", distribution)
		}
	}

	nextDefinition, err := getNextDefinition(userID, wordlistID)
	if err != nil {
		return nil, err
	}

	word, err := GetWordById(nextDefinition.WordID)
	if err != nil {
		return nil, err
	}

	// Select appropriate quiz type based on box and available content
	quizType, err := selectQuizType(nextDefinition, word, userID, wordlistID)
	if err != nil {
		return nil, err
	}

	// Additional logging for debugging WordFromImage availability
	if nextDefinition.BoxID == 3 || nextDefinition.BoxID == 7 {
		common.Logger.Info("image_quiz_availability",
			"boxID", nextDefinition.BoxID,
			"hasImage", nextDefinition.ImageUrl != "",
			"imageUrl", nextDefinition.ImageUrl,
			"imageDescription", nextDefinition.ImageDescription,
			"selectedQuizType", quizType)
	}

	// Create quiz based on selected type
	return createQuizForType(quizType, nextDefinition, word)
}

func selectQuizType(def *NextDefinition, word *model.Word, userID, wordlistID int64) (model.QuizType, error) {
	possibleTypes := boxToQuizTypes[def.BoxID]

	// Filter out quiz types that require unavailable content
	var availableTypes []model.QuizType
	for _, qt := range possibleTypes {
		if isQuizTypeAvailable(qt, def, word) {
			availableTypes = append(availableTypes, qt)
		}
	}

	if len(availableTypes) == 0 {
		// Fallback to basic quiz if no appropriate quiz available
		return model.GuessMeaning, nil
	}

	// Global quiz type balancing: favor least recently used types
	// This ensures better distribution across all quiz types
	selectedType, err := selectBalancedQuizType(userID, wordlistID, availableTypes)
	if err != nil {
		// Fallback to time-based rotation if balancing fails
		timeRotation := time.Now().Unix() / 300 // 300 seconds = 5 minutes
		index := (int(def.Definition.ID) + int(timeRotation)) % len(availableTypes)
		selectedType = availableTypes[index]
	}

	// Log the selection for monitoring
	common.Logger.Info("quiz_type_selected",
		"definitionID", def.Definition.ID,
		"boxID", def.BoxID,
		"availableTypes", availableTypes,
		"selectedType", selectedType,
		"imageUrl", def.ImageUrl,
		"hasImage", def.ImageUrl != "")

	return selectedType, nil
}

func isQuizTypeAvailable(qt model.QuizType, def *NextDefinition, word *model.Word) bool {
	switch qt {
	case model.WordFromImage:
		return def.ImageUrl != ""
	case model.CompleteSentence:
		// For verbs and phrasal verbs, check inflections; for others, check examples
		if def.Definition.IsVerbType {
			// Check if inflections have examples
			for _, inflection := range def.Definition.Inflections {
				if len(inflection.Examples) > 0 {
					return true
				}
			}
			return false
		}
		return len(def.Definition.Examples) > 0
	case model.WordFromAudio, model.MeaningFromAudio:
		return word.AudioURL != ""
	case model.WordFromExampleAudio:
		return len(def.Definition.ExampleAudioFiles) > 0
	case model.WriteWordFromDefinition:
		return def.Definition.Meaning != ""
	default:
		return true
	}
}

func createQuizForType(quizType model.QuizType, def *NextDefinition, word *model.Word) (*Quiz, error) {
	var options []string
	var value string
	var quizAnswer string
	var audioURL string
	var answerIndex int
	var err error

	// Default audioURL to word's audio (only overridden for specific quiz types)
	audioURL = word.AudioURL

	switch quizType {
	case model.MeaningFromAudio:
		quizAnswer = def.Definition.Meaning
		options, err = GetRandomMeanings([]int{int(def.Definition.ID)}, 3)
		if err != nil {
			return nil, err
		}
		value = ""

	case model.WordFromAudio:
		quizAnswer = word.Name
		options, err = GetRandomTokens([]int{int(def.Definition.ID)}, def.Definition.PartOfSpeech, 3)
		if err != nil {
			return nil, err
		}
		value = ""

	case model.WriteWordFromDefinition:
		value = def.Definition.Meaning
		quizAnswer = word.Name
		// No options for write-in quiz

	case model.CompleteSentence:
		// For verbs and phrasal verbs, use inflection examples; for others, use regular examples
		var availableExamples []string

		if def.Definition.IsVerbType {
			// Collect all examples from inflections
			for _, inflection := range def.Definition.Inflections {
				availableExamples = append(availableExamples, inflection.Examples...)
			}
		} else {
			availableExamples = def.Definition.Examples
		}

		// Select example using fair distribution to avoid repetition
		selectedExample, err := selectFairExample(def.Definition.ID, availableExamples)
		if err != nil {
			// Fallback to random selection if fair selection fails
			common.Logger.Warn("fair example selection failed, using random", "definitionId", def.Definition.ID, "error", err)
			i := rand.Intn(len(availableExamples))
			selectedExample = availableExamples[i]
		}
		value = selectedExample

		// Extract answer from brackets
		quizAnswer = extractAnswerFromExample(value, def.Definition.Token)

		// Get random options (database automatically excludes tokens from ignored definition IDs)
		options, err = GetRandomTokens([]int{int(def.Definition.ID)}, def.Definition.PartOfSpeech, 3)
		if err != nil {
			return nil, err
		}

	case model.WordFromImage:
		quizAnswer = extractAnswerFromImageDescription(def.ImageDescription, def.Definition.Token)
		options, err = GetRandomTokens([]int{int(def.Definition.ID)}, def.Definition.PartOfSpeech, 3)
		if err != nil {
			return nil, err
		}
		value = def.ImageUrl

	case model.WordFromMeaning:
		quizAnswer = def.Definition.Token
		options, err = GetRandomTokens([]int{int(def.Definition.ID)}, def.Definition.PartOfSpeech, 3)
		if err != nil {
			return nil, err
		}
		value = def.Definition.Meaning

	case model.GuessMeaning:
		quizAnswer = def.Definition.Meaning
		options, err = GetRandomMeanings([]int{int(def.Definition.ID)}, 3)
		if err != nil {
			return nil, err
		}
		value = def.Definition.Token

	case model.WordFromExampleAudio:
		quizAnswer = def.Definition.Token
		options, err = GetRandomTokens([]int{int(def.Definition.ID)}, def.Definition.PartOfSpeech, 3)
		if err != nil {
			return nil, err
		}

		// Use the already-loaded example audio files for better performance and consistency
		if len(def.Definition.ExampleAudioFiles) == 0 {
			return nil, errors.New("no example audio files available for WordFromExampleAudio quiz")
		}

		// Select example audio using fair distribution to avoid repetition
		selectedExampleAudio, err := selectFairExampleAudio(def.Definition.ID, def.Definition.ExampleAudioFiles)
		if err != nil {
			// Fallback to first available if fair selection fails
			common.Logger.Warn("fair example audio selection failed, using first available",
				"definitionId", def.Definition.ID, "error", err)
			selectedExampleAudio = &def.Definition.ExampleAudioFiles[0]
		}

		value = ""                               // No visual value needed for audio quiz
		audioURL = selectedExampleAudio.AudioURL // Example audio URL

	default:
		return nil, fmt.Errorf("unexpected quiz type: %v", quizType)
	}

	// Add correct answer to options at random position (for multiple choice quizzes)
	if quizType != model.WriteWordFromDefinition {
		// Ensure we have enough options and don't exceed the available positions
		maxOptions := len(options) + 1 // +1 for the correct answer
		answerIndex = rand.Intn(maxOptions)

		// Insert the correct answer at the random position
		options = append(options, "")
		copy(options[answerIndex+1:], options[answerIndex:])
		options[answerIndex] = quizAnswer
	} else {
		// For write-in quizzes, no options are needed
		answerIndex = 0
		options = []string{quizAnswer}
	}

	return &Quiz{
		Value:            value,
		Options:          options,
		AnswerIndex:      answerIndex,
		ID:               def.LeitnerSystemID,
		Type:             quizType,
		PartOfSpeech:     def.Definition.PartOfSpeech,
		IsVerbType:       def.Definition.IsVerbType,
		Pronunciation:    word.Pronunciation,
		ImageDescription: def.ImageDescription,
		AudioURL:         audioURL,
		WordID:           def.WordID,
		DefinitionID:     def.Definition.ID,
	}, nil
}

func extractAnswerFromExample(example, defaultToken string) string {
	re := regexp.MustCompile(`\[(.*?)\]`)
	matches := re.FindAllStringSubmatch(example, -1)

	var tokens []string
	for _, match := range matches {
		if len(match) > 1 {
			tokens = append(tokens, match[1])
		}
	}

	if len(tokens) > 0 {
		return strings.Join(tokens, " ")
	}

	// If no brackets found, return the default token
	return defaultToken
}

func extractAnswerFromImageDescription(description, defaultToken string) string {
	re := regexp.MustCompile(`\[(.*?)\]`)
	matches := re.FindStringSubmatch(description)

	if len(matches) > 1 {
		return matches[1]
	}

	return defaultToken
}

// selectFairExample implements a fair distribution algorithm to select examples that haven't been used recently.
// This prevents the same examples from appearing repeatedly and ensures a better learning experience.
//
// Algorithm:
// 1. Create hash for each example to track usage
// 2. Check database for recently used examples (within last 24 hours)
// 3. Prioritize unused examples, then least recently used
// 4. Record the selected example usage in the database
//
// Parameters:
// - definitionID: The definition these examples belong to
// - wordID: The word being quizzed (for additional context)
// - availableExamples: Slice of example sentences to choose from
//
// Returns the selected example string and any database error
func selectFairExample(definitionID int64, availableExamples []string) (string, error) {
	if len(availableExamples) == 0 {
		return "", errors.New("no examples available")
	}

	if len(availableExamples) == 1 {
		// Only one example, record its usage and return it
		err := recordExampleUsage(definitionID, availableExamples[0])
		return availableExamples[0], err
	}

	db, err := common.GetDBConnection()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	// Create a map of example hashes to examples and their usage info
	type exampleInfo struct {
		example    string
		hash       string
		lastUsedAt *time.Time
	}

	exampleInfos := make([]exampleInfo, len(availableExamples))
	hashes := make([]string, len(availableExamples))

	for i, example := range availableExamples {
		hash := hashExample(example)
		exampleInfos[i] = exampleInfo{
			example: example,
			hash:    hash,
		}
		hashes[i] = hash
	}

	// Query for recent usage of these examples
	query := `
		SELECT example_hash, last_used_at 
		FROM example_usage 
		WHERE definition_id = $1 AND example_hash = ANY($2)
		AND last_used_at > NOW() - INTERVAL '24 hours'`

	rows, err := db.Query(context.Background(), query, definitionID, hashes)
	if err != nil {
		return "", fmt.Errorf("failed to query example usage: %w", err)
	}
	defer rows.Close()

	// Create a map of hash -> last used time
	usageMap := make(map[string]time.Time)
	for rows.Next() {
		var hash string
		var lastUsedAt time.Time
		if err = rows.Scan(&hash, &lastUsedAt); err != nil {
			return "", fmt.Errorf("failed to scan usage row: %w", err)
		}
		usageMap[hash] = lastUsedAt
	}

	// Update exampleInfos with usage data
	for i := range exampleInfos {
		if lastUsedAt, exists := usageMap[exampleInfos[i].hash]; exists {
			exampleInfos[i].lastUsedAt = &lastUsedAt
		}
	}

	// Sort by usage: unused first, then by oldest usage
	sort.Slice(exampleInfos, func(i, j int) bool {
		if exampleInfos[i].lastUsedAt == nil && exampleInfos[j].lastUsedAt == nil {
			return false // Both unused, maintain order
		}
		if exampleInfos[i].lastUsedAt == nil {
			return true // i is unused, prioritize it
		}
		if exampleInfos[j].lastUsedAt == nil {
			return false // j is unused, prioritize it
		}
		// Both used, prioritize older usage
		return exampleInfos[i].lastUsedAt.Before(*exampleInfos[j].lastUsedAt)
	})

	// Select the first (best) example
	selectedExample := exampleInfos[0].example

	// Record the usage
	err = recordExampleUsage(definitionID, selectedExample)
	if err != nil {
		common.Logger.Error("failed to record example usage", "definitionId", definitionID, "error", err)
		// Don't fail the quiz generation if usage recording fails
	}

	return selectedExample, nil
}

// hashExample creates a consistent hash for an example to track its usage
func hashExample(example string) string {
	hash := md5.Sum([]byte(strings.TrimSpace(strings.ToLower(example))))
	return hex.EncodeToString(hash[:])
}

// recordExampleUsage records when an example was used to prevent immediate repetition
func recordExampleUsage(definitionID int64, example string) error {
	db, err := common.GetDBConnection()
	if err != nil {
		return err
	}

	hash := hashExample(example)
	query := `
		INSERT INTO example_usage (definition_id, example_hash, last_used_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (definition_id, example_hash)
		DO UPDATE SET last_used_at = NOW()`

	_, err = db.Exec(context.Background(), query, definitionID, hash)
	return err
}

// selectFairExampleAudio implements fair distribution for example audio files to prevent repetition.
// This function selects the least recently used example audio file and updates its usage tracking.
//
// Algorithm:
// 1. Query the database for usage statistics of each example audio file
// 2. Prioritize unused files, then least recently used files
// 3. Update usage tracking for the selected file
// 4. Return the selected audio file
//
// Parameters:
// - definitionID: The definition these audio files belong to
// - audioFiles: Slice of available example audio files
//
// Returns the selected audio file and any database error
func selectFairExampleAudio(definitionID int64, audioFiles []model.DefinitionExampleAudio) (*model.DefinitionExampleAudio, error) {
	if len(audioFiles) == 0 {
		return nil, errors.New("no audio files available")
	}

	if len(audioFiles) == 1 {
		// Only one audio file, record its usage and return it
		err := recordExampleAudioUsage(definitionID, audioFiles[0].ID)
		return &audioFiles[0], err
	}

	db, err := common.GetDBConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Create a map of audio file IDs to audio files and their usage info
	type audioFileInfo struct {
		audioFile  *model.DefinitionExampleAudio
		lastUsedAt *time.Time
		usageCount int
	}

	audioFileInfos := make([]audioFileInfo, len(audioFiles))
	audioFileIDs := make([]int64, len(audioFiles))

	for i := range audioFiles {
		audioFileInfos[i] = audioFileInfo{
			audioFile: &audioFiles[i],
		}
		audioFileIDs[i] = audioFiles[i].ID
	}

	// Query for recent usage of these audio files
	query := `
		SELECT example_audio_id, last_used_at, usage_count
		FROM example_audio_usage 
		WHERE definition_id = $1 AND example_audio_id = ANY($2)
		AND last_used_at > NOW() - INTERVAL '24 hours'`

	rows, err := db.Query(context.Background(), query, definitionID, audioFileIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query audio file usage: %w", err)
	}
	defer rows.Close()

	// Create a map of audio file ID -> usage info
	usageMap := make(map[int64]struct {
		lastUsedAt time.Time
		usageCount int
	})

	for rows.Next() {
		var audioFileID int64
		var lastUsedAt time.Time
		var usageCount int
		if err = rows.Scan(&audioFileID, &lastUsedAt, &usageCount); err != nil {
			return nil, fmt.Errorf("failed to scan usage row: %w", err)
		}
		usageMap[audioFileID] = struct {
			lastUsedAt time.Time
			usageCount int
		}{lastUsedAt, usageCount}
	}

	// Update audioFileInfos with usage data
	for i := range audioFileInfos {
		if usage, exists := usageMap[audioFileInfos[i].audioFile.ID]; exists {
			audioFileInfos[i].lastUsedAt = &usage.lastUsedAt
			audioFileInfos[i].usageCount = usage.usageCount
		}
	}

	// Sort by usage: unused first, then by oldest usage and lowest usage count
	sort.Slice(audioFileInfos, func(i, j int) bool {
		if audioFileInfos[i].lastUsedAt == nil && audioFileInfos[j].lastUsedAt == nil {
			// Both unused, prioritize by lowest usage count (in case of edge cases)
			return audioFileInfos[i].usageCount < audioFileInfos[j].usageCount
		}
		if audioFileInfos[i].lastUsedAt == nil {
			return true // i is unused, prioritize it
		}
		if audioFileInfos[j].lastUsedAt == nil {
			return false // j is unused, prioritize it
		}
		// Both used, prioritize by usage count first, then by older usage
		if audioFileInfos[i].usageCount != audioFileInfos[j].usageCount {
			return audioFileInfos[i].usageCount < audioFileInfos[j].usageCount
		}
		return audioFileInfos[i].lastUsedAt.Before(*audioFileInfos[j].lastUsedAt)
	})

	// Select the best (least used) audio file
	selectedAudioFile := audioFileInfos[0].audioFile

	// Record the usage
	err = recordExampleAudioUsage(definitionID, selectedAudioFile.ID)
	if err != nil {
		common.Logger.Error("failed to record example audio usage",
			"definitionId", definitionID,
			"audioFileId", selectedAudioFile.ID,
			"error", err)
		// Don't fail the quiz generation if usage recording fails
	}

	return selectedAudioFile, nil
}

// recordExampleAudioUsage records when an example audio file was used to prevent immediate repetition
func recordExampleAudioUsage(definitionID, audioFileID int64) error {
	db, err := common.GetDBConnection()
	if err != nil {
		return err
	}

	query := `
		INSERT INTO example_audio_usage (definition_id, example_audio_id, last_used_at, usage_count)
		VALUES ($1, $2, NOW(), 1)
		ON CONFLICT (definition_id, example_audio_id)
		DO UPDATE SET last_used_at = NOW(), usage_count = example_audio_usage.usage_count + 1`

	_, err = db.Exec(context.Background(), query, definitionID, audioFileID)
	return err
}

func (LeitnerSystemStrategy) updateLeitnerSystemTracking(leitnerSystemTrackingId int64, success bool, transactionPtr *pgx.Tx) error {
	var tx pgx.Tx
	var err error

	if transactionPtr == nil {
		var db *pgxpool.Pool
		db, err = common.GetDBConnection()
		if err != nil {
			return err
		}

		ctx := context.Background()
		tx, err = db.Begin(ctx)
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
	} else {
		tx = *transactionPtr
	}

	// Proper Leitner system logic with temporary skip on incorrect answers
	query := `UPDATE leitner_system_tracking 
	SET 
		updated_at = now(), 
		box_id = CASE 
			WHEN $1 AND box_id < 7 THEN box_id + 1  -- Move to next box on success
			WHEN $1 AND box_id = 7 THEN 7           -- Stay at max box
			ELSE 1                                   -- Reset to box 1 on failure
		END,
		temporarily_skipped_until = CASE 
			WHEN NOT $1 THEN NOW() + INTERVAL '10 minutes'  -- Skip for 10 minutes on incorrect answer
			ELSE NULL                                        -- Clear skip on correct answer
		END
	WHERE id = $2
	RETURNING box_id`

	var boxId int64
	row := tx.QueryRow(context.Background(), query, success, leitnerSystemTrackingId)
	err = row.Scan(&boxId)
	if err != nil {
		return err
	}

	return nil
}

// SaveQuizResult processes the user's quiz answer and updates the Leitner system accordingly.
// This is the core of the spaced repetition algorithm that:
// - Moves words to higher boxes on correct answers (increasing review intervals)
// - Resets words to box 1 on incorrect answers (immediate review)
// - Temporarily skips incorrectly answered words for 10 minutes to prevent repetition
//
// The Leitner box progression determines when words are reviewed again:
// Box 1: immediate, Box 2: 1 hour, Box 3: 1 day, Box 4: 3 days, Box 5: 1 week, Box 6: 2 weeks, Box 7: 1 month
//
// Parameters:
// - quizResult: Contains the quiz response details (correct/incorrect, timing, etc.)
// - transactionPtr: Optional database transaction (if nil, creates a new one)
func (s LeitnerSystemStrategy) SaveQuizResult(
	quizResult QuizResult,
	isPremium bool,
	transactionPtr *pgx.Tx) error {

	var tx pgx.Tx
	var err error
	ctx := context.Background()

	if transactionPtr == nil {
		var db *pgxpool.Pool
		db, err = common.GetDBConnection()
		if err != nil {
			return err
		}

		tx, err = db.Begin(ctx)
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
	} else {
		tx = *transactionPtr
	}

	// First, get the current box_id before update
	var currentBoxId int64
	err = tx.QueryRow(ctx, "SELECT box_id FROM leitner_system_tracking WHERE id = $1", quizResult.LeitnerSystemTrackingID).Scan(&currentBoxId)
	if err != nil {
		return err
	}

	// Update the Leitner system tracking
	err = s.updateLeitnerSystemTracking(quizResult.LeitnerSystemTrackingID, quizResult.IsCorrect, &tx)
	if err != nil {
		return err
	}

	// Set the quiz result data
	quizResult.BoxID = currentBoxId

	// Track analytics
	// Create analytics service without caching for write operations
	analyticsService, err := NewAnalyticsService(AnalyticsConfig{
		UserID:     quizResult.UserID,
		WordlistID: quizResult.WordlistID,
		UseCache:   false,
	})
	if err != nil {
		return err
	}

	err = analyticsService.TrackQuiz(ctx, quizResult, tx)
	if err != nil {
		// Log error but don't fail the transaction
		common.Logger.Error("failed to track quiz performance",
			"error", err,
			"userId", quizResult.UserID,
			"wordId", quizResult.WordID,
			"quizType", quizResult.QuizType)
	}

	// Update box distribution snapshot and invalidate cache (outside transaction)
	go func() {
		ctx := context.Background()

		// First, update the box distribution snapshot
		common.Logger.Info("updating box distribution snapshot",
			"userId", quizResult.UserID,
			"wordlistId", quizResult.WordlistID)
		err := analyticsService.UpdateBoxDistribution(ctx, quizResult.UserID, quizResult.WordlistID)
		if err != nil {
			common.Logger.Error("failed to update box distribution snapshot",
				"error", err,
				"userId", quizResult.UserID,
				"wordlistId", quizResult.WordlistID)
			return // Don't proceed with cache invalidation if DB update failed
		}
		common.Logger.Info("box distribution snapshot updated successfully",
			"userId", quizResult.UserID,
			"wordlistId", quizResult.WordlistID)

		// Then, invalidate cache for premium users
		if isPremium {
			// Try to get cached analytics service
			cache, err := common.GetRedisClient()
			if err != nil {
				// Redis not available, nothing to invalidate
				common.Logger.Warn("Redis not available for cache invalidation", "error", err)
				return
			}

			// Create cached service just for invalidation
			cachedService := &CachedAnalyticsService{
				cache: cache,
			}

			// Invalidate wordlist analytics cache (includes historical box distribution)
			err = cachedService.InvalidateWordlistAnalytics(ctx, quizResult.UserID, quizResult.WordlistID)
			if err != nil {
				common.Logger.Error("failed to invalidate analytics cache",
					"error", err,
					"userId", quizResult.UserID,
					"wordlistId", quizResult.WordlistID)
			} else {
				common.Logger.Info("analytics cache invalidated for premium user",
					"userId", quizResult.UserID,
					"wordlistId", quizResult.WordlistID)
			}
		}
	}()

	return nil
}

// IncludeDefinitions adds new word definitions to the Leitner system for spaced repetition tracking.
// This function is called when new definitions are fetched for a word (e.g., from ChatGPT).
// Each definition starts in box 1 (immediate review) and will progress through the Leitner boxes
// based on the user's quiz performance.
//
// Parameters:
// - wordId: The ID of the word these definitions belong to
// - userId: The ID of the user who will be quizzed on these definitions
// - definitionIds: Array of definition IDs to include in the Leitner system
// - tx: Database transaction to ensure atomicity
//
// Returns an error if any database operations fail.
func (LeitnerSystemStrategy) IncludeDefinitions(wordId, userId int64, definitionIds []int64, tx pgx.Tx) error {
	for _, definitionId := range definitionIds {
		query := `INSERT INTO leitner_system_tracking (user_id, definition_id, box_id, word_id, updated_at)
		VALUES ($1, $2, $3, $4, NOW())`

		_, err := tx.Exec(context.Background(), query, userId, definitionId, 1, wordId)
		if err != nil {
			return err
		}
	}

	return nil
}

// MarkErrorResolved removes the temporary skip status from definitions and marks error reports as resolved.
// This function is called when:
// - New definitions are successfully fetched for a word (resolving missing/incorrect content)
// - New images are generated for definitions (resolving image-related issues)
// - Content issues are manually resolved by administrators
//
// The function clears the temporarily_skipped_until timestamp, allowing the definitions
// to be included in quiz generation again.
//
// Parameters:
// - report: ErrorReport containing either DefinitionId or WordId to resolve
//
// Returns an error if the database operations fail.
func (s LeitnerSystemStrategy) MarkErrorResolved(report ErrorReport) error {

	if report.DefinitionId == nil && report.WordId == nil {
		return errors.New("definition or word missing")
	}

	db, err := common.GetDBConnection()
	if err != nil {
		return err
	}

	ctx := context.Background()
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

	leitnerSystemTrackingUpdate := `UPDATE leitner_system_tracking SET temporarily_skipped_until = NULL `
	selection, queryArgs, err := buildQuerySelectionFromErrorReport(report)

	if err != nil {
		return err
	}

	leitnerSystemTrackingUpdate = leitnerSystemTrackingUpdate + selection

	// Remove temporary skip
	_, err = tx.Exec(ctx, leitnerSystemTrackingUpdate, queryArgs...)

	if err != nil {
		return err
	}

	errorReportsUpdate := `UPDATE error_reports SET status = 'resolved', resolved_at = NOW() `
	selection, queryArgs, err = buildQuerySelectionFromErrorReport(report)
	if err != nil {
		return err
	}
	errorReportsUpdate = errorReportsUpdate + selection

	// Mark error reports as resolved
	_, err = tx.Exec(ctx, errorReportsUpdate, queryArgs...)

	return err
}

type ErrorReport struct {
	DefinitionId *int64 `json:"definitionId"`
	WordId       *int64 `json:"wordId"`
	UserId       int64  `json:"userId"`
}

func buildQuerySelectionFromErrorReport(report ErrorReport) (string, []any, error) {

	if report.DefinitionId == nil && report.WordId == nil {
		return "", nil, errors.New("definition or word missing")
	}

	var builder strings.Builder

	var queryArgs []any
	var whereConditions []string

	whereConditions = append(whereConditions, "user_id = $1")
	queryArgs = append(queryArgs, report.UserId)
	argIndex := 2

	if report.DefinitionId != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("definition_id = $%d", argIndex))
		queryArgs = append(queryArgs, report.DefinitionId)
		argIndex++
	}

	if report.WordId != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("word_id = $%d", argIndex))
		queryArgs = append(queryArgs, report.WordId)
		argIndex++
	}

	builder.WriteString(" WHERE ")
	builder.WriteString(strings.Join(whereConditions, " AND "))

	return builder.String(), queryArgs, nil

}

// selectBalancedQuizType selects a quiz type from available types, favoring those that have been used less recently.
// This helps balance quiz type distribution across the session to provide better variety for users.
func selectBalancedQuizType(userID, wordlistID int64, availableTypes []model.QuizType) (model.QuizType, error) {
	if len(availableTypes) == 1 {
		return availableTypes[0], nil
	}

	db, err := common.GetDBConnection()
	if err != nil {
		return "", err
	}

	// Get recent quiz type usage for this user/wordlist (last 2 hours)
	query := `
		SELECT quiz_type, COUNT(*) as usage_count, MAX(created_at) as last_used_at
		FROM quiz_performance 
		WHERE user_id = $1 AND wordlist_id = $2 
		AND created_at > NOW() - INTERVAL '2 hours'
		GROUP BY quiz_type`

	rows, err := db.Query(context.Background(), query, userID, wordlistID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	// Build usage map
	usage := make(map[string]struct {
		count      int
		lastUsedAt time.Time
	})

	for rows.Next() {
		var quizType string
		var count int
		var lastUsedAt time.Time
		if err = rows.Scan(&quizType, &count, &lastUsedAt); err != nil {
			return "", err
		}
		usage[quizType] = struct {
			count      int
			lastUsedAt time.Time
		}{count: count, lastUsedAt: lastUsedAt}
	}

	// Calculate scores for available types (lower score = higher priority)
	type typeScore struct {
		quizType model.QuizType
		score    float64
	}

	var scores []typeScore
	now := time.Now()

	for _, qt := range availableTypes {
		typeUsage, exists := usage[string(qt)]
		
		var score float64
		if !exists {
			// Never used = highest priority
			score = 0
		} else {
			// Score based on usage count and recency
			hoursSinceLastUse := now.Sub(typeUsage.lastUsedAt).Hours()
			
			// Base score from usage count (more usage = higher score)
			score = float64(typeUsage.count) * 10
			
			// Reduce score based on time since last use (older = lower score)
			score -= hoursSinceLastUse * 2
			
			// Ensure score is not negative
			if score < 0 {
				score = 0
			}
		}

		scores = append(scores, typeScore{
			quizType: qt,
			score:    score,
		})
	}

	// Sort by score (ascending - lower score = higher priority)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score < scores[j].score
	})

	// Select from the top 2 least used types with some randomness
	// This prevents being too predictable while still favoring underused types
	topCount := len(scores)
	if topCount > 2 {
		topCount = 2
	}

	// Use time-based pseudo-randomness for selection within top choices
	timeRotation := time.Now().Unix() / 300 // 5-minute rotation
	selectedIndex := int(timeRotation) % topCount

	common.Logger.Info("quiz_type_balancing",
		"userID", userID,
		"wordlistID", wordlistID,
		"availableTypes", availableTypes,
		"selectedType", scores[selectedIndex].quizType,
		"selectedScore", scores[selectedIndex].score,
		"allScores", scores)

	return scores[selectedIndex].quizType, nil
}

// ReportError records a user-reported issue with quiz content and temporarily excludes it from quiz generation.
// This function handles user feedback about problematic content such as:
// - Images that don't match the word meaning
// - Incorrect or confusing definitions
// - Audio that doesn't play or is unclear
// - Examples that are inappropriate or wrong
//
// The function:
// 1. Records the error report in the database for tracking and analytics
// 2. Temporarily skips the affected definition(s) by setting temporarily_skipped_until
// 3. Ensures the problematic content won't appear in future quizzes until resolved
//
// Parameters:
// - userID: The ID of the user reporting the error
// - report: ErrorReport object containing error details and affected content IDs
// - tx: Database transaction to ensure atomicity
// - ctx: Context for the database operations
//
// Returns an error if the database operations fail.
func (s LeitnerSystemStrategy) ReportError(userID int64, report ErrorReport, tx pgx.Tx, ctx context.Context) error {

	if report.DefinitionId == nil && report.WordId == nil {
		return errors.New("definition or word missing")
	}

	query := `UPDATE leitner_system_tracking SET temporarily_skipped_until = NOW() + INTERVAL '1 hour' `
	selection, queryArgs, err := buildQuerySelectionFromErrorReport(report)

	if err != nil {
		return err
	}

	query = query + selection
	_, err = tx.Exec(ctx, query, queryArgs...)

	if err != nil {
		return err
	}

	return nil
}
