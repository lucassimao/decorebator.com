package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
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
// Each box introduces more challenging quiz types as the user demonstrates mastery:
// - Lower boxes focus on recognition and basic recall
// - Higher boxes require active recall, contextual understanding, and audio recognition
var boxToQuizTypes = map[int64][]model.QuizType{
	1: {model.GuessMeaning},                                    // Basic recognition
	2: {model.WordFromMeaning},                                 // Basic recall
	3: {model.WordFromImage, model.GuessMeaning},               // Visual association
	4: {model.CompleteSentence, model.WordFromMeaning},         // Contextual understanding
	5: {model.WriteWordFromDefinition, model.CompleteSentence}, // Active recall
	6: {model.WordFromAudio, model.WriteWordFromDefinition},    // Audio recognition
	7: {model.MeaningFromAudio, model.WordFromAudio},           // Advanced audio
}

// ExampleUsage tracks when examples were last used for fair distribution
type ExampleUsage struct {
	ExampleHash string    `db:"example_hash"`
	LastUsedAt  time.Time `db:"last_used_at"`
}

func getNextDefinition(userID, wordlistID int64) (*NextDefinition, error) {
	// Improved query that respects Leitner intervals
	query := `
		WITH due_definitions AS (
			SELECT 
				def.id, lst.id AS lst_id, def.token, 
				def.part_of_speech, def.meaning, def.examples, 
				def.inflections, lst.box_id, def.sounds, def.phonetic_notations, 
				di.url as image_url, di.description as image_description, 
				wd.word_id AS word_id,
				lst.updated_at,
				-- Calculate if definition is due for review
				CASE 
					WHEN lst.box_id = 1 THEN TRUE														-- Immediate review
					WHEN lst.box_id = 2 THEN EXTRACT(EPOCH FROM (NOW() - lst.updated_at))/3600 >= 1 	-- 1 hours  
					WHEN lst.box_id = 3 THEN EXTRACT(EPOCH FROM (NOW() - lst.updated_at))/3600 >= 24 	-- 1 day
					WHEN lst.box_id = 4 THEN EXTRACT(EPOCH FROM (NOW() - lst.updated_at))/3600 >= 72 	-- 3 days
					WHEN lst.box_id = 5 THEN EXTRACT(EPOCH FROM (NOW() - lst.updated_at))/3600 >= 168 	-- 1 week
					WHEN lst.box_id = 6 THEN EXTRACT(EPOCH FROM (NOW() - lst.updated_at))/3600 >= 336 	-- 2 weeks
					WHEN lst.box_id = 7 THEN EXTRACT(EPOCH FROM (NOW() - lst.updated_at))/3600 >= 720	-- 1 month
				END AS is_due
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
		)
		SELECT 
			id, token, part_of_speech, meaning, examples, inflections, 
			lst_id, box_id, sounds, phonetic_notations, 
			COALESCE(image_url,''), word_id, COALESCE(image_description,'')
		FROM due_definitions
		WHERE is_due = TRUE
		ORDER BY 
			box_id ASC,                                    -- Prioritize lower boxes
			EXTRACT(EPOCH FROM (NOW() - updated_at)) DESC, -- Prioritize older reviews
			RANDOM()                                       -- Add randomization for ties
		LIMIT 1;
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
		// No due definitions, get the oldest reviewed one
		return getOldestDefinition(userID, wordlistID)
	}

	definition := model.Definition{}
	result := NextDefinition{Definition: &definition}

	err = rows.Scan(&definition.ID, &definition.Token, &definition.PartOfSpeech,
		&definition.Meaning, &definition.Examples,
		&definition.Inflections, &result.LeitnerSystemID, &result.BoxID, &definition.Sounds,
		&definition.PhoneticNotations, &result.ImageUrl, &result.WordID, &result.ImageDescription)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func getOldestDefinition(userID, wordlistID int64) (*NextDefinition, error) {
	// Enhanced fallback query that also excludes temporarily skipped definitions
	query := `
		SELECT 
			def.id,def.token, 
			def.part_of_speech, def.meaning, def.examples, 
			def.inflections,  lst.id AS lst_id, lst.box_id, def.sounds, def.phonetic_notations, 
			COALESCE(di.url,'') as image_url, COALESCE(di.description,'') as image_description, 
			wd.word_id AS word_id
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
		ORDER BY 
			lst.box_id ASC,                                    -- Prioritize lower boxes
			lst.updated_at ASC NULLS FIRST,                    -- Then by oldest review
			RANDOM()                                           -- Add randomization for ties
		LIMIT 1;
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

	if !rows.Next() {
		return nil, errors.New("no definitions found")
	}

	definition := model.Definition{}
	result := NextDefinition{Definition: &definition}

	err = rows.Scan(&definition.ID, &definition.Token, &definition.PartOfSpeech,
		&definition.Meaning, &definition.Examples,
		&definition.Inflections, &result.LeitnerSystemID, &result.BoxID, &definition.Sounds,
		&definition.PhoneticNotations, &result.ImageUrl, &result.ImageDescription, &result.WordID)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateQuiz generates a new quiz question for the user based on the Leitner spaced repetition system.
// It selects the most appropriate word/definition that is due for review, considering:
// - Leitner box intervals (box 1: immediate, box 2: 1 hour, box 3: 1 day, etc.)
// - Word difficulty progression (different quiz types for different boxes)
// - Temporarily skipped words (to avoid immediate repetition)
// - Available content (images, audio, examples, inflections)
//
// Returns a Quiz object with the question, options, correct answer, and metadata.
// Returns an error if no words are available for review or if database operations fail.
func (LeitnerSystemStrategy) CreateQuiz(wordlistID, userID int64) (*Quiz, error) {
	nextDefinition, err := getNextDefinition(userID, wordlistID)
	if err != nil {
		return nil, err
	}

	word, err := GetWordById(nextDefinition.WordID)
	if err != nil {
		return nil, err
	}

	// Select appropriate quiz type based on box and available content
	quizType, err := selectQuizType(nextDefinition, word)
	if err != nil {
		return nil, err
	}

	// Create quiz based on selected type
	return createQuizForType(quizType, nextDefinition, word)
}

func selectQuizType(def *NextDefinition, word *model.Word) (model.QuizType, error) {
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

	// Randomly select from available types
	return availableTypes[rand.Intn(len(availableTypes))], nil
}

func isQuizTypeAvailable(qt model.QuizType, def *NextDefinition, word *model.Word) bool {
	switch qt {
	case model.WordFromImage:
		return def.ImageUrl != ""
	case model.CompleteSentence:
		// For verbs and phrasal verbs, check inflections; for others, check examples
		isVerb := def.Definition.PartOfSpeech == "verb" || def.Definition.PartOfSpeech == "phrasal verb"
		if isVerb {
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
	var err error

	switch quizType {
	case model.MeaningFromAudio:
		options, err = GetRandomMeanings([]int{int(def.Definition.ID)}, 3)
		if err != nil {
			return nil, err
		}
		quizAnswer = def.Definition.Meaning
		value = ""

	case model.WordFromAudio:
		options, err = GetRandomTokens([]int{int(def.Definition.ID)}, def.Definition.PartOfSpeech, 3)
		if err != nil {
			return nil, err
		}
		quizAnswer = word.Name
		value = ""

	case model.WriteWordFromDefinition:
		value = def.Definition.Meaning
		quizAnswer = word.Name
		// No options for write-in quiz

	case model.CompleteSentence:
		options, err = GetRandomTokens([]int{int(def.Definition.ID)}, def.Definition.PartOfSpeech, 3)
		if err != nil {
			return nil, err
		}

		// For verbs and phrasal verbs, use inflection examples; for others, use regular examples
		isVerb := def.Definition.PartOfSpeech == "verb" || def.Definition.PartOfSpeech == "phrasal verb"
		var availableExamples []string

		if isVerb {
			// Collect all examples from inflections
			for _, inflection := range def.Definition.Inflections {
				availableExamples = append(availableExamples, inflection.Examples...)
			}
		} else {
			availableExamples = def.Definition.Examples
		}

		// Select example using fair distribution to avoid repetition
		selectedExample, err := selectFairExample(def.Definition.ID, def.WordID, availableExamples)
		if err != nil {
			// Fallback to random selection if fair selection fails
			common.Logger.Warn("fair example selection failed, using random", "definitionId", def.Definition.ID, "error", err)
			i := rand.Intn(len(availableExamples))
			selectedExample = availableExamples[i]
		}
		value = selectedExample

		// Extract answer from brackets
		quizAnswer = extractAnswerFromExample(value, def.Definition.Token)

	case model.WordFromImage:
		options, err = GetRandomTokens([]int{int(def.Definition.ID)}, def.Definition.PartOfSpeech, 3)
		if err != nil {
			return nil, err
		}
		value = def.ImageUrl
		quizAnswer = extractAnswerFromImageDescription(def.ImageDescription, def.Definition.Token)

	case model.WordFromMeaning:
		options, err = GetRandomTokens([]int{int(def.Definition.ID)}, def.Definition.PartOfSpeech, 3)
		if err != nil {
			return nil, err
		}
		value = def.Definition.Meaning
		quizAnswer = def.Definition.Token

	case model.GuessMeaning:
		options, err = GetRandomMeanings([]int{int(def.Definition.ID)}, 3)
		if err != nil {
			return nil, err
		}
		value = def.Definition.Token
		quizAnswer = def.Definition.Meaning

	default:
		return nil, fmt.Errorf("unexpected quiz type: %v", quizType)
	}

	// Add correct answer to options and fshuffle
	options = append(options, quizAnswer)
	answerIndex := rand.Intn(len(options))
	copy(options[answerIndex+1:], options[answerIndex:])
	options[answerIndex] = quizAnswer

	return &Quiz{
		Value:            value,
		Options:          options,
		AnswerIndex:      answerIndex,
		ID:               def.LeitnerSystemID,
		Type:             quizType,
		PartOfSpeech:     def.Definition.PartOfSpeech,
		Pronunciation:    word.Pronunciation,
		ImageDescription: def.ImageDescription,
		AudioURL:         word.AudioURL,
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
func selectFairExample(definitionID, wordID int64, availableExamples []string) (string, error) {
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
	analyticsService, err := NewAnalyticsService()
	if err != nil {
		return err
	}

	err = analyticsService.TrackQuizPerformance(ctx, quizResult, tx)
	if err != nil {
		// Log error but don't fail the transaction
		common.Logger.Error("failed to track quiz performance", "error", err)
		fmt.Print(err)
	}

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
		query := `INSERT INTO leitner_system_tracking (user_id, definition_id, box_id, word_id)
		VALUES ($1, $2, $3, $4)`

		_, err := tx.Exec(context.Background(), query, userId, definitionId, 1, wordId)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetSkippedDefinitions retrieves all definitions that are currently being skipped due to error reports.
// These are definitions that users have reported as having issues (wrong images, incorrect meanings, etc.)
// and are temporarily excluded from quiz generation until the issues are resolved.
//
// This function is useful for:
// - Admin interfaces to review reported issues
// - Analytics to understand content quality problems
// - Debugging quiz generation issues
//
// Parameters:
// - userID: The ID of the user whose skipped definitions to retrieve
// - wordlistID: The ID of the wordlist to check for skipped definitions
//
// Returns a slice of NextDefinition objects with error reporting details.
func (s LeitnerSystemStrategy) GetSkippedDefinitions(userID, wordlistID int64) ([]NextDefinition, error) {
	query := `
		SELECT 
			def.id, def.token, def.part_of_speech, def.meaning, def.examples, 
			def.inflections, lst.id AS lst_id, lst.box_id, def.sounds, def.phonetic_notations, 
			COALESCE(di.url,'') as image_url, COALESCE(di.description,'') as image_description, 
			wd.word_id AS word_id,
			der.error_type, der.description as error_description, der.reported_at
		FROM leitner_system_tracking lst 
		JOIN definitions def ON lst.definition_id = def.id
		JOIN word_definitions wd ON def.id = wd.definition_id
		JOIN words w ON wd.word_id = w.id
		LEFT JOIN definition_images di ON di.definition_id = def.id AND di.is_visible=TRUE
		LEFT JOIN error_reports der ON der.definition_id = def.id AND der.status = 'pending'
		WHERE 
			lst.user_id = $1
			AND w.wordlist_id = $2
			AND w.learned = FALSE
			AND lst.temporarily_skipped_until > NOW()
		ORDER BY der.reported_at DESC;
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

	var results []NextDefinition
	for rows.Next() {
		definition := model.Definition{}
		result := NextDefinition{Definition: &definition}
		var errorType, errorDescription string
		var reportedAt time.Time

		err = rows.Scan(&definition.ID, &definition.Token, &definition.PartOfSpeech,
			&definition.Meaning, &definition.Examples,
			&definition.Inflections, &result.LeitnerSystemID, &result.BoxID, &definition.Sounds,
			&definition.PhoneticNotations, &result.ImageUrl, &result.ImageDescription, &result.WordID,
			&errorType, &errorDescription, &reportedAt)

		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
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
