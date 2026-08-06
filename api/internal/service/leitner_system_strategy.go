package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	mathrand "math/rand"
	"regexp"
	"sort"
	"strings"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// cryptoRandInt generates a cryptographically secure random int in range [0, max)
func cryptoRandInt(limit int) int {
	if limit <= 0 {
		return 0
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(limit)))
	if err != nil {
		// Fallback to math/rand if crypto/rand fails
		//nolint:gosec // G404 - fallback random when crypto/rand is unavailable
		return mathrand.Intn(limit)
	}
	return int(n.Int64())
}

// LeitnerSystemStrategy implements the Leitner spaced repetition algorithm for vocabulary learning.
// The Leitner system uses boxes with increasing intervals to optimize long-term retention:
// - Words start in box 1 (immediate review)
// - Correct answers move words to higher boxes (longer intervals)
// - Incorrect answers reset words to box 1 (immediate review)
// - Each box has different quiz types to progressively increase difficulty
type LeitnerSystemStrategy struct {
	db                     *pgxpool.Pool
	wordService            *WordService
	definitionService      *DefinitionService
	analyticsWriter        *LeitnerAnalyticsWriter
	leitnerTrackingService *LeitnerTrackingService
}

func NewLeitnerSystemStrategy(db *pgxpool.Pool, wordService *WordService, definitionService *DefinitionService, analyticsWriter *LeitnerAnalyticsWriter, leitnerTrackingService *LeitnerTrackingService) *LeitnerSystemStrategy {
	return &LeitnerSystemStrategy{
		db:                     db,
		wordService:            wordService,
		definitionService:      definitionService,
		analyticsWriter:        analyticsWriter,
		leitnerTrackingService: leitnerTrackingService,
	}
}

type Quiz = model.Quiz
type QuizType = model.QuizType

// NextDefinition represents a word definition selected for quiz generation
// along with its current Leitner system state and associated metadata.
type NextDefinition struct {
	Definition                   *model.Definition // The definition content (meaning, examples, inflections, etc.)
	LeitnerSystemID              int64             // The tracking ID for this definition in the Leitner system
	BoxID                        int64             // Current Leitner box (1-7, determines review interval)
	WordID                       int64             // The word this definition belongs to
	WordAudioURL                 string            // Audio URL for the word (if any)
	ImageURL                     string            // URL of associated image (if any)
	ImageDescription             string            // Description of the associated image
	ScopedDistractorsChecked     bool
	ScopedTokenDistractorCount   int
	ScopedMeaningDistractorCount int
}

type WordlistAvailability struct {
	Total      int
	Unlearned  int
	Completed  int
	InProgress int
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
	1: {model.GuessMeaning},                                                                                                                                         // Recognition: "What does this word mean?"
	2: {model.WordFromMeaning, model.GuessMeaning},                                                                                                                  // Basic recall: "Which word matches this meaning?" + recognition practice
	3: {model.WordFromImage, model.WordFromMeaning},                                                                                                                 // Visual association: "What word matches this image?" + meaning recall
	4: {model.CompleteSentence, model.WordFromExampleAudio, model.GuessMeaning},                                                                                     // Contextual understanding + recognition reinforcement
	5: {model.WriteWordFromDefinition, model.WordFromExampleAudio, model.WordFromMeaning},                                                                           // Active recall + meaning practice
	6: {model.WordFromAudio, model.WordFromMeaningAudio, model.WordFromExampleAudio, model.WordFromImage, model.GuessMeaning},                                       // Audio/Visual recognition + meaning reinforcement
	7: {model.MeaningFromAudio, model.WordFromMeaningAudio, model.WordFromImage, model.WriteWordFromDefinition, model.CompleteSentence, model.WordFromExampleAudio}, // Mastery
}

const multipleChoiceDistractors = 2
const quizTypeFilteredCandidateLimit = 50

func newDistractorScope(userID, wordlistID int64, definition *model.Definition) DistractorScope {
	return DistractorScope{
		UserID:                 userID,
		WordlistID:             wordlistID,
		Language:               definition.Language,
		PartOfSpeechNormalized: NormalizePartOfSpeech(definition.PartOfSpeech, definition.Language),
	}
}

func requiresTokenDistractors(quizType model.QuizType) bool {
	switch quizType {
	case model.WordFromAudio,
		model.WordFromMeaningAudio,
		model.CompleteSentence,
		model.WordFromImage,
		model.WordFromMeaning,
		model.WordFromExampleAudio:
		return true
	default:
		return false
	}
}

func requiresMeaningDistractors(quizType model.QuizType) bool {
	return quizType == model.MeaningFromAudio || quizType == model.GuessMeaning
}

func hasRequiredScopedDistractors(quizType model.QuizType, definition *NextDefinition) bool {
	if !definition.ScopedDistractorsChecked {
		return true
	}
	if requiresTokenDistractors(quizType) {
		return definition.ScopedTokenDistractorCount >= multipleChoiceDistractors
	}
	if requiresMeaningDistractors(quizType) {
		return definition.ScopedMeaningDistractorCount >= multipleChoiceDistractors
	}
	return true
}

// ExampleUsage tracks when examples were last used for fair distribution
type ExampleUsage struct {
	ExampleHash string    `db:"example_hash"`
	LastUsedAt  time.Time `db:"last_used_at"`
}

func (s *LeitnerSystemStrategy) loadScopedDistractorAvailability(
	ctx context.Context,
	definition *NextDefinition,
	userID, wordlistID int64,
	allowedTypes []model.QuizType,
) error {
	scope := newDistractorScope(userID, wordlistID, definition.Definition)
	ignored := []int{int(definition.Definition.ID)}
	needsTokens := false
	needsMeanings := false
	for _, quizType := range allowedTypes {
		needsTokens = needsTokens || requiresTokenDistractors(quizType)
		needsMeanings = needsMeanings || requiresMeaningDistractors(quizType)
	}

	if needsTokens {
		tokens, err := s.definitionService.GetRandomTokens(ctx, scope, ignored, multipleChoiceDistractors)
		if err != nil {
			return fmt.Errorf("count scoped token distractors: %w", err)
		}
		definition.ScopedTokenDistractorCount = len(tokens)
	}
	if needsMeanings {
		meanings, err := s.definitionService.GetRandomMeanings(ctx, scope, ignored, multipleChoiceDistractors)
		if err != nil {
			return fmt.Errorf("count scoped meaning distractors: %w", err)
		}
		definition.ScopedMeaningDistractorCount = len(meanings)
	}
	definition.ScopedDistractorsChecked = true
	return nil
}

// getNextDefinition selects the next due definition for quiz generation.
//
// Scheduling uses next_review_at as the source of truth and orders by:
// 1. Earliest next_review_at
// 2. Oldest updated_at
// 3. Lowest definition ID (deterministic tiebreaker)
//
// Returns:
// - NextDefinition with selected definition details and Leitner system metadata
// - QuizUnavailableError if no definitions are due
// - Error if database operations fail
func (s *LeitnerSystemStrategy) getNextDefinition(ctx context.Context, userID, wordlistID int64, allowedTypes []model.QuizType) (*NextDefinition, error) { //nolint:gocyclo // Deterministic selection algorithm with dedicated integration coverage.
	queryTemplate := `
		WITH due_definitions AS (
			SELECT 
				def.id, lst.id AS lst_id, def.token, 
				def.part_of_speech, def.language, def.is_verb_type, def.meaning, COALESCE(def.meaning_audio_url, '') as meaning_audio_url, def.examples, 
				def.inflections, lst.box_id, def.sounds, def.phonetic_notations, 
				COALESCE(di.url, '') as image_url, COALESCE(di.description, '') as image_description, 
				w.id AS word_id,
				COALESCE(w.audio_url, '') as word_audio_url,
				lst.updated_at,
				lst.next_review_at,
				EXTRACT(EPOCH FROM (NOW() - lst.updated_at))/3600 as hours_since_review
			FROM leitner_system_tracking lst 
			JOIN definitions def ON lst.definition_id = def.id
			JOIN words w ON w.id = lst.word_id
			JOIN word_definitions wd ON wd.definition_id = def.id AND wd.word_id = w.id
			LEFT JOIN definition_images di ON di.definition_id = def.id AND di.is_visible=TRUE
			WHERE 
				lst.user_id = $1
				AND w.wordlist_id = $2
				AND w.learned = FALSE
				AND def.meaning IS NOT NULL
				AND lst.next_review_at IS NOT NULL
				%s
				-- Exclude temporarily skipped definitions
				AND (lst.temporarily_skipped_until IS NULL OR lst.temporarily_skipped_until < NOW())
		)
		SELECT 
			dd.id, dd.token, dd.part_of_speech, dd.language, dd.is_verb_type, dd.meaning, dd.meaning_audio_url, dd.examples, dd.inflections, 
			dd.lst_id, dd.box_id, dd.sounds, dd.phonetic_notations, 
			dd.image_url, dd.word_id, dd.image_description, dd.word_audio_url,
			dd.hours_since_review, dd.next_review_at,
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
		FROM due_definitions dd
		LEFT JOIN definition_example_audio dea ON dea.definition_id = dd.id
		GROUP BY dd.id, dd.token, dd.part_of_speech, dd.language, dd.is_verb_type, dd.meaning, dd.meaning_audio_url, dd.examples, dd.inflections, 
				 dd.lst_id, dd.box_id, dd.sounds, dd.phonetic_notations, 
				 dd.image_url, dd.word_id, dd.image_description, dd.word_audio_url, dd.hours_since_review, dd.next_review_at, dd.updated_at
		ORDER BY 
			dd.next_review_at ASC,
			dd.updated_at ASC,
			dd.id ASC
		LIMIT $3;
	`

	candidateLimit := 1
	if len(allowedTypes) > 0 {
		// When a quiz-type filter is active, we widen the candidate pool so we can
		// pick the first top-priority definition that supports any selected type.
		candidateLimit = quizTypeFilteredCandidateLimit
	}

	type candidateMeta struct {
		definition        *NextDefinition
		hoursSinceReview  float64
		nextReviewAt      time.Time
		exampleAudioCount int
	}

	scanCandidates := func(rows pgx.Rows) ([]candidateMeta, error) {
		var candidates []candidateMeta
		for rows.Next() {
			definition := model.Definition{}
			result := NextDefinition{Definition: &definition}
			var hoursSinceReview float64
			var nextReviewAt time.Time
			var exampleAudioFilesJSON []byte

			err := rows.Scan(&definition.ID, &definition.Token, &definition.PartOfSpeech, &definition.Language, &definition.IsVerbType,
				&definition.Meaning, &definition.MeaningAudioURL, &definition.Examples,
				&definition.Inflections, &result.LeitnerSystemID, &result.BoxID, &definition.Sounds,
				&definition.PhoneticNotations, &result.ImageURL, &result.WordID, &result.ImageDescription, &result.WordAudioURL,
				&hoursSinceReview, &nextReviewAt, &exampleAudioFilesJSON)

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

			candidates = append(candidates, candidateMeta{
				definition:        &result,
				hoursSinceReview:  hoursSinceReview,
				nextReviewAt:      nextReviewAt,
				exampleAudioCount: len(definition.ExampleAudioFiles),
			})
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}

		return candidates, nil
	}

	query := fmt.Sprintf(queryTemplate, "AND lst.next_review_at <= NOW()")
	rows, err := s.db.Query(ctx, query, userID, wordlistID, candidateLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	candidates, err := scanCandidates(rows)
	if err != nil {
		return nil, err
	}

	practiceAhead := false
	if len(candidates) == 0 {
		futureQuery := fmt.Sprintf(queryTemplate, "")
		futureRows, futureErr := s.db.Query(ctx, futureQuery, userID, wordlistID, candidateLimit)
		if futureErr != nil {
			return nil, futureErr
		}
		defer futureRows.Close()

		candidates, err = scanCandidates(futureRows)
		if err != nil {
			return nil, err
		}
		if len(candidates) > 0 {
			practiceAhead = true
			common.Logger.Info("no_due_items_practice_ahead",
				"userID", userID,
				"wordlistID", wordlistID,
				"nextReviewAt", candidates[0].nextReviewAt)
		}
	}

	if len(candidates) == 0 {
		// Debug: Let's check what's happening
		debugQuery := `
			SELECT COUNT(*) as total_words,
				   COUNT(CASE WHEN w.learned = FALSE THEN 1 END) as unlearned_words,
				   COUNT(CASE WHEN def.meaning IS NOT NULL THEN 1 END) as with_meaning,
				   COUNT(CASE WHEN lst.temporarily_skipped_until IS NULL OR lst.temporarily_skipped_until < NOW() THEN 1 END) as not_skipped,
				   COUNT(CASE WHEN lst.next_review_at IS NOT NULL AND lst.next_review_at <= NOW() THEN 1 END) as due_now
			FROM leitner_system_tracking lst 
			JOIN definitions def ON lst.definition_id = def.id
			JOIN word_definitions wd ON def.id = wd.definition_id
			JOIN words w ON wd.word_id = w.id
			WHERE lst.user_id = $1 AND w.wordlist_id = $2
		`
		var totalWords, unlearnedWords, withMeaning, notSkipped, dueNow int
		debugErr := s.db.QueryRow(ctx, debugQuery, userID, wordlistID).Scan(&totalWords, &unlearnedWords, &withMeaning, &notSkipped, &dueNow)
		if debugErr == nil {
			common.Logger.Info("no definitions selected - debug info",
				"userID", userID,
				"wordlistID", wordlistID,
				"totalWords", totalWords,
				"unlearnedWords", unlearnedWords,
				"withMeaning", withMeaning,
				"notSkipped", notSkipped,
				"dueNow", dueNow)
		}

		return nil, common.QuizUnavailableError{
			Reason:  common.QuizUnavailableNoDueItems,
			Message: "no due items available",
		}
	}

	if practiceAhead {
		// Practice-ahead mode: shuffle the earliest upcoming items so we don't
		// repeatedly quiz the exact same definition when nothing is due.
		poolSize := 10
		if len(candidates) < poolSize {
			poolSize = len(candidates)
		}
		for i := poolSize - 1; i > 0; i-- {
			j := cryptoRandInt(i + 1)
			candidates[i], candidates[j] = candidates[j], candidates[i]
		}
	}

	if len(allowedTypes) == 0 {
		selected := candidates[0]
		logSelectedDefinition(userID, wordlistID, selected.definition, selected.hoursSinceReview, selected.nextReviewAt, selected.exampleAudioCount)
		return selected.definition, nil
	}

	for _, candidate := range candidates {
		if err := s.loadScopedDistractorAvailability(ctx, candidate.definition, userID, wordlistID, allowedTypes); err != nil {
			return nil, err
		}
		availableTypes := availableQuizTypesForDefinition(candidate.definition)
		if len(filterQuizTypes(availableTypes, allowedTypes)) > 0 {
			logSelectedDefinition(userID, wordlistID, candidate.definition, candidate.hoursSinceReview, candidate.nextReviewAt, candidate.exampleAudioCount)
			return candidate.definition, nil
		}
	}

	return nil, common.QuizUnavailableError{
		Reason:  common.QuizUnavailableNoMatchingQuizTypes,
		Message: "no quiz types available for selection",
	}
}

// getWordlistBoxDistribution provides analytics on word distribution across boxes
func (s *LeitnerSystemStrategy) getWordlistBoxDistribution(ctx context.Context, userID, wordlistID int64) (map[int64]int, error) {
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

	rows, err := s.db.Query(ctx, query, userID, wordlistID)
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

// getWordlistAvailability checks whether a wordlist has enough completed, unlearned words to generate quizzes.
func (s *LeitnerSystemStrategy) getWordlistAvailability(ctx context.Context, userID, wordlistID int64) (*WordlistAvailability, error) {
	query := `
		SELECT
			COUNT(*) as total_words,
			COUNT(*) FILTER (WHERE learned = FALSE) as unlearned_words,
			COUNT(*) FILTER (WHERE processing_status = 'completed') as completed_words,
			COUNT(*) FILTER (WHERE processing_status IN ('pending', 'processing')) as in_progress_words
		FROM words
		WHERE user_id = $1 AND wordlist_id = $2
	`

	var availability WordlistAvailability
	err := s.db.QueryRow(ctx, query, userID, wordlistID).Scan(
		&availability.Total,
		&availability.Unlearned,
		&availability.Completed,
		&availability.InProgress,
	)
	if err != nil {
		return nil, err
	}

	return &availability, nil
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
func (s LeitnerSystemStrategy) CreateQuiz(ctx context.Context, wordlistID, userID int64, allowedTypes []model.QuizType) (*Quiz, error) {
	availability, err := s.getWordlistAvailability(ctx, userID, wordlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to check wordlist status: %w", err)
	}

	if availability.Total == 0 {
		return nil, common.QuizUnavailableError{
			Reason:  common.QuizUnavailableWordlistEmpty,
			Message: "wordlist has no words",
		}
	}

	if availability.Completed < 1 {
		return nil, common.QuizUnavailableError{
			Reason:  common.QuizUnavailableWordlistProcessing,
			Message: "wordlist is still processing",
		}
	}

	if availability.Unlearned == 0 {
		return nil, common.QuizUnavailableError{
			Reason:  common.QuizUnavailableNoUnlearnedWords,
			Message: "no unlearned words in wordlist",
		}
	}

	// Log box distribution for monitoring (but don't fail if it errors)
	if distribution, distErr := s.getWordlistBoxDistribution(ctx, userID, wordlistID); distErr == nil {
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

	nextDefinition, err := s.getNextDefinition(ctx, userID, wordlistID, allowedTypes)
	if err != nil {
		return nil, err
	}

	word, err := s.wordService.GetWordByID(ctx, nextDefinition.WordID)
	if err != nil {
		return nil, err
	}

	// Select appropriate quiz type based on box and available content
	quizType := s.selectQuizType(ctx, nextDefinition, word, userID, wordlistID, allowedTypes)

	// Additional logging for debugging WordFromImage availability
	if nextDefinition.BoxID == 3 || nextDefinition.BoxID == 7 {
		common.Logger.Info("image_quiz_availability",
			"boxID", nextDefinition.BoxID,
			"hasImage", nextDefinition.ImageURL != "",
			"imageUrl", nextDefinition.ImageURL,
			"imageDescription", nextDefinition.ImageDescription,
			"selectedQuizType", quizType)
	}

	// Create quiz based on selected type
	distractorScope := newDistractorScope(userID, wordlistID, nextDefinition.Definition)

	allowWriteFallback := len(allowedTypes) == 0
	for _, allowedType := range allowedTypes {
		if allowedType == model.WriteWordFromDefinition {
			allowWriteFallback = true
			break
		}
	}

	return s.createQuizForType(ctx, quizType, nextDefinition, word, s.definitionService, distractorScope, allowWriteFallback)
}

func (s *LeitnerSystemStrategy) selectQuizType(ctx context.Context, def *NextDefinition, word *model.Word, userID, wordlistID int64, allowedTypes []model.QuizType) model.QuizType {
	possibleTypes := boxToQuizTypes[def.BoxID]

	// Filter out quiz types that require unavailable content
	var availableTypes []model.QuizType
	for _, qt := range possibleTypes {
		if isQuizTypeAvailable(qt, def, word.AudioURL) {
			availableTypes = append(availableTypes, qt)
		}
	}

	if len(allowedTypes) > 0 {
		allowedSet := map[model.QuizType]struct{}{}
		for _, qt := range allowedTypes {
			allowedSet[qt] = struct{}{}
		}
		var filteredTypes []model.QuizType
		for _, qt := range availableTypes {
			if _, ok := allowedSet[qt]; ok {
				filteredTypes = append(filteredTypes, qt)
			}
		}
		availableTypes = filteredTypes
	}

	if len(availableTypes) == 0 {
		// Fallback to basic quiz if no appropriate quiz available
		return model.GuessMeaning
	}

	// Global quiz type balancing: favor least recently used types
	// This ensures better distribution across all quiz types
	selectedType, err := s.selectBalancedQuizType(ctx, userID, wordlistID, availableTypes)
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
		"imageUrl", def.ImageURL,
		"hasImage", def.ImageURL != "")

	return selectedType
}

func availableQuizTypesForDefinition(def *NextDefinition) []model.QuizType {
	possibleTypes := boxToQuizTypes[def.BoxID]
	var availableTypes []model.QuizType
	for _, qt := range possibleTypes {
		if isQuizTypeAvailable(qt, def, def.WordAudioURL) {
			availableTypes = append(availableTypes, qt)
		}
	}
	return availableTypes
}

func filterQuizTypes(availableTypes, allowedTypes []model.QuizType) []model.QuizType {
	if len(allowedTypes) == 0 {
		return availableTypes
	}
	allowedSet := map[model.QuizType]struct{}{}
	for _, qt := range allowedTypes {
		allowedSet[qt] = struct{}{}
	}
	var filteredTypes []model.QuizType
	for _, qt := range availableTypes {
		if _, ok := allowedSet[qt]; ok {
			filteredTypes = append(filteredTypes, qt)
		}
	}
	return filteredTypes
}

func logSelectedDefinition(userID, wordlistID int64, def *NextDefinition, hoursSinceReview float64, nextReviewAt time.Time, exampleAudioCount int) {
	common.Logger.Info("due_item_selection",
		"userID", userID,
		"wordlistID", wordlistID,
		"definitionID", def.Definition.ID,
		"boxID", def.BoxID,
		"hoursSinceReview", hoursSinceReview,
		"nextReviewAt", nextReviewAt,
		"isOverdue", nextReviewAt.Before(time.Now()),
		"exampleAudioCount", exampleAudioCount)
}

func isQuizTypeAvailable(qt model.QuizType, def *NextDefinition, wordAudioURL string) bool {
	if !hasRequiredScopedDistractors(qt, def) {
		return false
	}

	switch qt {
	case model.WordFromImage:
		return def.ImageURL != ""
	case model.CompleteSentence:
		// Use centralized method to check if definition has examples
		return def.Definition.HasExamples()
	case model.WordFromAudio, model.MeaningFromAudio:
		return wordAudioURL != ""
	case model.WordFromMeaningAudio:
		return def.Definition.MeaningAudioURL != ""
	case model.WordFromExampleAudio:
		return len(def.Definition.ExampleAudioFiles) > 0
	case model.WriteWordFromDefinition:
		return def.Definition.Meaning != ""
	default:
		return true
	}
}

// createCompleteSentenceQuiz handles the complex logic for CompleteSentence quiz type
func (s *LeitnerSystemStrategy) createCompleteSentenceQuiz(ctx context.Context, def *NextDefinition, definitionService *DefinitionService, distractorScope DistractorScope) (string, string, []string, error) {
	// Use centralized method to get all appropriate examples based on part of speech
	availableExamples := def.Definition.GetAllExamples()

	// Select example using fair distribution to avoid repetition
	selectedExample, err := s.selectFairExample(ctx, def.Definition.ID, availableExamples)
	if err != nil {
		// Fallback to random selection if fair selection fails
		common.Logger.Warn("fair example selection failed, using random", "definitionId", def.Definition.ID, "error", err)
		//nolint:gosec // G404 - fallback random selection, not security-critical
		i := mathrand.Intn(len(availableExamples))
		selectedExample = availableExamples[i]
	}

	// Extract answer from brackets
	quizAnswer := extractAnswerFromExample(selectedExample, def.Definition.Token)

	// Get random options (database automatically excludes tokens from ignored definition IDs)
	options, err := definitionService.GetRandomTokens(ctx, distractorScope, []int{int(def.Definition.ID)}, multipleChoiceDistractors)
	if err != nil {
		return "", "", nil, err
	}

	return selectedExample, quizAnswer, options, nil
}

// createWordFromExampleAudioQuiz handles the complex logic for WordFromExampleAudio quiz type
// Returns (quizAnswer, options, audioURL, error) - no visual value needed for audio quizzes
func (s *LeitnerSystemStrategy) createWordFromExampleAudioQuiz(ctx context.Context, def *NextDefinition, definitionService *DefinitionService, distractorScope DistractorScope) (string, []string, string, error) {
	options, err := definitionService.GetRandomTokens(ctx, distractorScope, []int{int(def.Definition.ID)}, multipleChoiceDistractors)
	if err != nil {
		return "", nil, "", err
	}

	// Use the already-loaded example audio files for better performance and consistency
	if len(def.Definition.ExampleAudioFiles) == 0 {
		return "", nil, "", errors.New("no example audio files available for WordFromExampleAudio quiz")
	}

	// Select example audio using fair distribution to avoid repetition
	selectedExampleAudio, err := s.selectFairExampleAudio(ctx, def.Definition.ID, def.Definition.ExampleAudioFiles)
	if err != nil {
		// Fallback to first available if fair selection fails
		common.Logger.Warn("fair example audio selection failed, using first available",
			"definitionId", def.Definition.ID, "error", err)
		selectedExampleAudio = &def.Definition.ExampleAudioFiles[0]
	}

	// Extract answer from brackets in the example text
	quizAnswer := extractAnswerFromExample(selectedExampleAudio.ExampleText, def.Definition.Token)
	audioURL := selectedExampleAudio.AudioURL // Example audio URL

	return quizAnswer, options, audioURL, nil
}

func (s *LeitnerSystemStrategy) handleInsufficientDistractors(
	ctx context.Context,
	quizType model.QuizType,
	options []string,
	def *NextDefinition,
	word *model.Word,
	definitionService *DefinitionService,
	distractorScope DistractorScope,
	allowWriteFallback bool,
) (*Quiz, error) {
	if quizType == model.WriteWordFromDefinition || len(options) >= multipleChoiceDistractors {
		return nil, nil
	}

	common.Logger.WarnContext(ctx, "insufficient scoped quiz distractors",
		"userID", distractorScope.UserID,
		"wordlistID", distractorScope.WordlistID,
		"definitionID", def.Definition.ID,
		"quizType", quizType,
		"distractorCount", len(options))

	if allowWriteFallback && def.Definition.Meaning != "" {
		return s.createQuizForType(
			ctx,
			model.WriteWordFromDefinition,
			def,
			word,
			definitionService,
			distractorScope,
			false,
		)
	}

	return nil, common.QuizUnavailableError{
		Reason:  common.QuizUnavailableNoMatchingQuizTypes,
		Message: "not enough scoped distractors for the selected quiz types",
	}
}

func (s *LeitnerSystemStrategy) finalizeQuiz(
	ctx context.Context,
	quizType model.QuizType,
	value, quizAnswer, audioURL string,
	options []string,
	def *NextDefinition,
	word *model.Word,
	definitionService *DefinitionService,
	distractorScope DistractorScope,
	allowWriteFallback bool,
) (*Quiz, error) {
	if fallbackQuiz, fallbackErr := s.handleInsufficientDistractors(
		ctx,
		quizType,
		options,
		def,
		word,
		definitionService,
		distractorScope,
		allowWriteFallback,
	); fallbackQuiz != nil || fallbackErr != nil {
		return fallbackQuiz, fallbackErr
	}

	answerIndex := 0
	if quizType != model.WriteWordFromDefinition {
		answerIndex = cryptoRandInt(len(options) + 1)
		options = append(options, "")
		copy(options[answerIndex+1:], options[answerIndex:])
		options[answerIndex] = quizAnswer
	} else {
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

func (s *LeitnerSystemStrategy) createQuizForType(ctx context.Context, quizType model.QuizType, def *NextDefinition, word *model.Word, definitionService *DefinitionService, distractorScope DistractorScope, allowWriteFallback bool) (*Quiz, error) {
	var options []string
	var value string
	var quizAnswer string
	var audioURL string
	var err error

	// Default audioURL to word's audio (only overridden for specific quiz types)
	audioURL = word.AudioURL

	switch quizType {
	case model.MeaningFromAudio:
		quizAnswer = def.Definition.Meaning
		options, err = definitionService.GetRandomMeanings(ctx, distractorScope, []int{int(def.Definition.ID)}, multipleChoiceDistractors)
		if err != nil {
			return nil, err
		}
		value = ""

	case model.WordFromAudio:
		quizAnswer = word.Name
		options, err = definitionService.GetRandomTokens(ctx, distractorScope, []int{int(def.Definition.ID)}, multipleChoiceDistractors)
		if err != nil {
			return nil, err
		}
		value = ""

	case model.WordFromMeaningAudio:
		quizAnswer = word.Name
		options, err = definitionService.GetRandomTokens(ctx, distractorScope, []int{int(def.Definition.ID)}, multipleChoiceDistractors)
		if err != nil {
			return nil, err
		}
		audioURL = def.Definition.MeaningAudioURL
		value = ""

	case model.WriteWordFromDefinition:
		value = def.Definition.Meaning
		quizAnswer = word.Name
		// No options for write-in quiz

	case model.CompleteSentence:
		value, quizAnswer, options, err = s.createCompleteSentenceQuiz(ctx, def, definitionService, distractorScope)
		if err != nil {
			return nil, err
		}

	case model.WordFromImage:
		quizAnswer = extractAnswerFromImageDescription(def.ImageDescription, def.Definition.Token)
		options, err = definitionService.GetRandomTokens(ctx, distractorScope, []int{int(def.Definition.ID)}, multipleChoiceDistractors)
		if err != nil {
			return nil, err
		}
		value = def.ImageURL

	case model.WordFromMeaning:
		quizAnswer = def.Definition.Token
		options, err = definitionService.GetRandomTokens(ctx, distractorScope, []int{int(def.Definition.ID)}, multipleChoiceDistractors)
		if err != nil {
			return nil, err
		}
		value = def.Definition.Meaning

	case model.GuessMeaning:
		quizAnswer = def.Definition.Meaning
		options, err = definitionService.GetRandomMeanings(ctx, distractorScope, []int{int(def.Definition.ID)}, multipleChoiceDistractors)
		if err != nil {
			return nil, err
		}
		value = def.Definition.Token

	case model.WordFromExampleAudio:
		quizAnswer, options, audioURL, err = s.createWordFromExampleAudioQuiz(ctx, def, definitionService, distractorScope)
		if err != nil {
			return nil, err
		}
		value = "" // No visual value needed for audio quiz

	default:
		return nil, fmt.Errorf("unexpected quiz type: %v", quizType)
	}

	return s.finalizeQuiz(
		ctx,
		quizType,
		value,
		quizAnswer,
		audioURL,
		options,
		def,
		word,
		definitionService,
		distractorScope,
		allowWriteFallback,
	)
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
// 2. Check database for recently used examples (within last 7 days)
// 3. If there are unused examples, pick one at random
// 4. Otherwise, pick the least recently used example
// 4. Record the selected example usage in the database
//
// Parameters:
// - definitionID: The definition these examples belong to
// - wordID: The word being quizzed (for additional context)
// - availableExamples: Slice of example sentences to choose from
//
// Returns the selected example string and any database error
func (s *LeitnerSystemStrategy) selectFairExample(ctx context.Context, definitionID int64, availableExamples []string) (string, error) {
	if len(availableExamples) == 0 {
		return "", errors.New("no examples available")
	}

	if len(availableExamples) == 1 {
		// Only one example, record its usage and return it
		err := s.recordExampleUsage(ctx, definitionID, availableExamples[0])
		return availableExamples[0], err
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
		AND last_used_at > NOW() - INTERVAL '7 days'`

	rows, err := s.db.Query(ctx, query, definitionID, hashes)
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

	// Select from unused examples at random to avoid deterministic repetition
	var unused []exampleInfo
	for _, info := range exampleInfos {
		if info.lastUsedAt == nil {
			unused = append(unused, info)
		}
	}

	var selectedExample string
	if len(unused) > 0 {
		selectedExample = unused[cryptoRandInt(len(unused))].example
	} else {
		// All examples used recently: pick least recently used
		sort.Slice(exampleInfos, func(i, j int) bool {
			return exampleInfos[i].lastUsedAt.Before(*exampleInfos[j].lastUsedAt)
		})
		selectedExample = exampleInfos[0].example
	}

	// Record the usage
	err = s.recordExampleUsage(ctx, definitionID, selectedExample)
	if err != nil {
		common.Logger.Error("failed to record example usage", "definitionId", definitionID, "error", err)
		// Don't fail the quiz generation if usage recording fails
	}

	return selectedExample, nil
}

// hashExample creates a consistent hash for an example to track its usage
func hashExample(example string) string {
	hash := sha256.Sum256([]byte(strings.TrimSpace(strings.ToLower(example))))
	return hex.EncodeToString(hash[:])
}

// recordExampleUsage records when an example was used to prevent immediate repetition
func (s *LeitnerSystemStrategy) recordExampleUsage(ctx context.Context, definitionID int64, example string) error {
	hash := hashExample(example)
	query := `
		INSERT INTO example_usage (definition_id, example_hash, last_used_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (definition_id, example_hash)
		DO UPDATE SET last_used_at = NOW()`

	_, err := s.db.Exec(ctx, query, definitionID, hash)
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
func (s *LeitnerSystemStrategy) selectFairExampleAudio(ctx context.Context, definitionID int64, audioFiles []model.DefinitionExampleAudio) (*model.DefinitionExampleAudio, error) {
	if len(audioFiles) == 0 {
		return nil, errors.New("no audio files available")
	}

	if len(audioFiles) == 1 {
		// Only one audio file, record its usage and return it
		err := s.recordExampleAudioUsage(ctx, definitionID, audioFiles[0].ID)
		return &audioFiles[0], err
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

	rows, err := s.db.Query(ctx, query, definitionID, audioFileIDs)
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
	err = s.recordExampleAudioUsage(ctx, definitionID, selectedAudioFile.ID)
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
func (s *LeitnerSystemStrategy) recordExampleAudioUsage(ctx context.Context, definitionID, audioFileID int64) error {
	query := `
		INSERT INTO example_audio_usage (definition_id, example_audio_id, last_used_at, usage_count)
		VALUES ($1, $2, NOW(), 1)
		ON CONFLICT (definition_id, example_audio_id)
		DO UPDATE SET last_used_at = NOW(), usage_count = example_audio_usage.usage_count + 1`

	_, err := s.db.Exec(ctx, query, definitionID, audioFileID)
	return err
}

// clearEarliestSkipsIfNeeded prevents all definitions from being blocked by clearing
// the 3 definitions that are closest to becoming available again.
// This ensures that quiz generation never fails due to all definitions being temporarily skipped.
func (s *LeitnerSystemStrategy) clearEarliestSkipsIfNeeded(ctx context.Context, userID, wordlistID, currentLeitnerTrackingID int64, tx pgx.Tx) (int, error) {
	// First, check if applying the skip to current definition would block all definitions
	availableCountQuery := `
		SELECT COUNT(*) as available_count 
		FROM leitner_system_tracking lst 
		JOIN word_definitions wd ON lst.definition_id = wd.definition_id
		JOIN words w ON wd.word_id = w.id 
		WHERE lst.user_id = $1 AND w.wordlist_id = $2 
		AND w.learned = FALSE 
		AND (lst.temporarily_skipped_until IS NULL OR lst.temporarily_skipped_until < NOW())
		AND lst.id != $3  -- Exclude current one being updated
	`

	var availableCount int
	err := tx.QueryRow(ctx, availableCountQuery, userID, wordlistID, currentLeitnerTrackingID).Scan(&availableCount)
	if err != nil {
		return 0, fmt.Errorf("failed to count available definitions: %w", err)
	}

	// If we still have available definitions, no need to clear any skips
	if availableCount > 0 {
		return 0, nil
	}

	// All definitions would be blocked - clear the 3 earliest skipped definitions
	clearSkipsQuery := `
		UPDATE leitner_system_tracking 
		SET temporarily_skipped_until = NULL 
		WHERE id IN (
			SELECT lst.id 
			FROM leitner_system_tracking lst 
			JOIN word_definitions wd ON lst.definition_id = wd.definition_id
			JOIN words w ON wd.word_id = w.id 
			WHERE lst.user_id = $1 AND w.wordlist_id = $2 
			AND w.learned = FALSE 
			AND lst.temporarily_skipped_until IS NOT NULL
			AND lst.id != $3  -- Don't clear the current one being updated
			ORDER BY lst.temporarily_skipped_until ASC  -- Clear earliest to expire first
			LIMIT 3
		)
	`

	result, err := tx.Exec(ctx, clearSkipsQuery, userID, wordlistID, currentLeitnerTrackingID)
	if err != nil {
		return 0, fmt.Errorf("failed to clear earliest skips: %w", err)
	}

	clearedCount := int(result.RowsAffected())

	// Log the automatic skip clearing for monitoring
	if clearedCount > 0 {
		common.Logger.Info("cleared_earliest_skips_to_prevent_blocking",
			"userID", userID,
			"wordlistID", wordlistID,
			"currentLeitnerTrackingID", currentLeitnerTrackingID,
			"clearedCount", clearedCount,
			"reason", "all_definitions_would_be_blocked")
	}

	return clearedCount, nil
}

func (s *LeitnerSystemStrategy) updateLeitnerSystemTracking(ctx context.Context, quizResult QuizResult, transactionPtr *pgx.Tx) error {
	var tx pgx.Tx
	var err error

	if transactionPtr == nil {
		tx, err = s.db.Begin(ctx)
		if err != nil {
			return err
		}
		defer func() {
			if err == nil {
				if commitErr := tx.Commit(ctx); commitErr != nil {
					common.Logger.Error("failed to commit transaction in leitner system", "error", commitErr)
				}
			} else {
				if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
					common.Logger.Error("failed to rollback transaction in leitner system", "error", rollbackErr)
				}
			}
		}()
	} else {
		tx = *transactionPtr
	}

	// If answer is incorrect, check if we need to clear some skips to prevent blocking all definitions
	if !quizResult.IsCorrect {
		clearedCount, clearErr := s.clearEarliestSkipsIfNeeded(ctx, quizResult.UserID, quizResult.WordlistID, quizResult.LeitnerSystemTrackingID, tx)
		if clearErr != nil {
			return fmt.Errorf("failed to clear earliest skips: %w", clearErr)
		}
		if clearedCount > 0 {
			common.Logger.Info("preventive_skip_clearing_completed",
				"userID", quizResult.UserID,
				"wordlistID", quizResult.WordlistID,
				"leitnerSystemTrackingID", quizResult.LeitnerSystemTrackingID,
				"clearedDefinitionsCount", clearedCount)
		}
	}

	// Proper Leitner system logic with temporary skip on incorrect answers
	query := `
		WITH updated AS (
			SELECT id,
				CASE 
					WHEN $1 AND box_id < 7 THEN box_id + 1  -- Move to next box on success
					WHEN $1 AND box_id = 7 THEN 7           -- Stay at max box
					ELSE 1                                   -- Reset to box 1 on failure
				END AS new_box
			FROM leitner_system_tracking
			WHERE id = $2
		)
		UPDATE leitner_system_tracking lst
		SET 
			updated_at = NOW(),
			box_id = updated.new_box,
			next_review_at = CASE updated.new_box
				WHEN 1 THEN NOW()
				WHEN 2 THEN NOW() + INTERVAL '6 hours'
				WHEN 3 THEN NOW() + INTERVAL '24 hours'
				WHEN 4 THEN NOW() + INTERVAL '72 hours'
				WHEN 5 THEN NOW() + INTERVAL '168 hours'
				WHEN 6 THEN NOW() + INTERVAL '336 hours'
				WHEN 7 THEN NOW() + INTERVAL '720 hours'
				ELSE NOW()
			END,
			temporarily_skipped_until = CASE 
				WHEN NOT $1 THEN NOW() + INTERVAL '10 minutes'  -- Skip for 10 minutes on incorrect answer
				ELSE NULL                                        -- Clear skip on correct answer
			END
		FROM updated
		WHERE lst.id = updated.id
		RETURNING updated.new_box`

	var boxID int64
	row := tx.QueryRow(ctx, query, quizResult.IsCorrect, quizResult.LeitnerSystemTrackingID)
	err = row.Scan(&boxID)
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
	ctx context.Context,
	quizResult QuizResult,
	isPremium bool,
	transactionPtr *pgx.Tx) (returnErr error) {
	var tx pgx.Tx
	var err error
	ownsTransaction := transactionPtr == nil

	if ownsTransaction {
		tx, err = s.db.Begin(ctx)
		if err != nil {
			return err
		}
		defer func() {
			if returnErr != nil {
				if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
					common.Logger.Error("failed to rollback transaction in leitner system", "error", rollbackErr)
				}
			}
		}()
	} else {
		tx = *transactionPtr
	}

	var trackingExists bool
	verifyErr := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM leitner_system_tracking lst
			JOIN words w ON w.id = lst.word_id
			JOIN word_definitions wd ON wd.word_id = lst.word_id AND wd.definition_id = lst.definition_id
			WHERE lst.id = $1
				AND lst.user_id = $2
				AND lst.definition_id = $3
				AND lst.word_id = $4
				AND w.wordlist_id = $5
		)
	`, quizResult.LeitnerSystemTrackingID, quizResult.UserID, quizResult.DefinitionID, quizResult.WordID, quizResult.WordlistID).Scan(&trackingExists)
	if verifyErr != nil {
		return verifyErr
	}
	if !trackingExists {
		return common.NotFoundError{
			ID:     quizResult.LeitnerSystemTrackingID,
			Entity: "leitner_system_tracking",
		}
	}

	// First, get the current box_id before update
	var currentBoxID int64
	err = tx.QueryRow(ctx, "SELECT box_id FROM leitner_system_tracking WHERE id = $1", quizResult.LeitnerSystemTrackingID).Scan(&currentBoxID)
	if err != nil {
		return err
	}

	// Update the Leitner system tracking
	err = s.updateLeitnerSystemTracking(ctx, quizResult, &tx)
	if err != nil {
		return err
	}

	// Set the quiz result data
	quizResult.BoxID = currentBoxID

	// Track analytics using injected writer
	if err = (*s.analyticsWriter).TrackQuiz(ctx, quizResult, tx); err != nil {
		return fmt.Errorf("failed to track quiz performance: %w", err)
	}

	if ownsTransaction {
		if err = tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit quiz result: %w", err)
		}
	}

	// Update box distribution snapshot and invalidate cache (outside transaction)
	go func() {
		// Intentionally use context.Background() for detached background operations
		// This ensures cache invalidation completes even if the HTTP request times out,
		// maintaining data consistency for analytics
		ctx := context.Background()

		// First, update the box distribution snapshot
		common.Logger.Info("updating box distribution snapshot",
			"userId", quizResult.UserID,
			"wordlistId", quizResult.WordlistID)
		err := (*s.analyticsWriter).UpdateBoxDistribution(ctx, quizResult.UserID, quizResult.WordlistID)
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
// - report: ErrorReport containing either DefinitionID or WordID to resolve
//
// Returns an error if the database operations fail.
func (s LeitnerSystemStrategy) MarkErrorResolved(ctx context.Context, report ErrorReport) error {
	if report.DefinitionID == nil && report.WordID == nil {
		return errors.New("definition or word missing")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				common.Logger.Error("failed to commit transaction in mark error resolved", "error", commitErr)
			}
		} else {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				common.Logger.Error("failed to rollback transaction in mark error resolved", "error", rollbackErr)
			}
		}
	}()

	// Clear temporary skip using tracking service
	err = s.leitnerTrackingService.ClearTemporarySkip(ctx, report, tx)
	if err != nil {
		return err
	}

	// Mark error reports as resolved (this is not part of tracking, so we handle it here)
	var errorReportsUpdate string
	var queryArgs []interface{}

	if report.DefinitionID != nil {
		errorReportsUpdate = `UPDATE error_reports SET status = 'resolved', resolved_at = NOW() WHERE user_id = $1 AND definition_id = $2`
		queryArgs = []interface{}{report.UserID, *report.DefinitionID}
	} else if report.WordID != nil {
		errorReportsUpdate = `UPDATE error_reports SET status = 'resolved', resolved_at = NOW() WHERE user_id = $1 AND word_id = $2`
		queryArgs = []interface{}{report.UserID, *report.WordID}
	} else {
		return errors.New("definition or word missing")
	}

	// Mark error reports as resolved
	_, err = tx.Exec(ctx, errorReportsUpdate, queryArgs...)

	return err
}

type ErrorReport struct {
	DefinitionID *int64 `json:"definitionId"`
	WordID       *int64 `json:"wordId"`
	UserID       int64  `json:"userId"`
}

// selectBalancedQuizType selects a quiz type from available types, favoring those that have been used less recently.
// This helps balance quiz type distribution across the session to provide better variety for users.
func (s *LeitnerSystemStrategy) selectBalancedQuizType(ctx context.Context, userID, wordlistID int64, availableTypes []model.QuizType) (model.QuizType, error) {
	if len(availableTypes) == 1 {
		return availableTypes[0], nil
	}

	// Get recent quiz type usage for this user/wordlist (last 2 hours)
	query := `
		SELECT quiz_type, COUNT(*) as usage_count, MAX(created_at) as last_used_at
		FROM quiz_performance 
		WHERE user_id = $1 AND wordlist_id = $2 
		AND created_at > NOW() - INTERVAL '2 hours'
		GROUP BY quiz_type`

	rows, err := s.db.Query(ctx, query, userID, wordlistID)
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
