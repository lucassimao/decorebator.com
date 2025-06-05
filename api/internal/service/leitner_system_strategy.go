package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LeitnerSystemStrategy struct{}
type Quiz = model.Quiz
type QuizType = model.QuizType

type NextDefinition struct {
	Definition       *model.Definition
	LeitnerSystemID  int64
	BoxID            int64
	WordID           int64
	ImageUrl         string
	ImageDescription string
}

// Quiz difficulty progression
var boxToQuizTypes = map[int64][]model.QuizType{
	1: {model.GuessMeaning},                                    // Basic recognition
	2: {model.WordFromMeaning},                                 // Basic recall
	3: {model.WordFromImage, model.GuessMeaning},               // Visual association
	4: {model.CompleteSentence, model.WordFromMeaning},         // Contextual understanding
	5: {model.WriteWordFromDefinition, model.CompleteSentence}, // Active recall
	6: {model.WordFromAudio, model.WriteWordFromDefinition},    // Audio recognition
	7: {model.MeaningFromAudio, model.WordFromAudio},           // Advanced audio
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
			box_id ASC,           -- Prioritize lower boxes
			updated_at ASC        -- Then by oldest review
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
		ORDER BY lst.updated_at ASC NULLS FIRST
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

		// Select random example
		i := rand.Intn(len(def.Definition.Examples))
		value = def.Definition.Examples[i]

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

	// Proper Leitner system logic
	query := `UPDATE leitner_system_tracking 
	SET 
		updated_at = now(), 
		box_id = CASE 
			WHEN $1 AND box_id < 7 THEN box_id + 1  -- Move to next box on success
			WHEN $1 AND box_id = 7 THEN 7           -- Stay at max box
			ELSE 1                                   -- Reset to box 1 on failure
		END
	WHERE id = $2
	RETURNING box_id`

	var boxId int64
	row := tx.QueryRow(context.Background(), query, success, leitnerSystemTrackingId)
	err = row.Scan(&boxId)
	if err != nil {
		return err
	}

	// Record history
	_, err = tx.Exec(context.Background(), `INSERT INTO 
		leitner_system_history (at, box_id, leitner_system_tracking_id, success) 
		VALUES (now(), $1, $2, $3)`, boxId, leitnerSystemTrackingId, success)

	return err
}

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

func (LeitnerSystemStrategy) IncludeDefinitions(wordId, userId int64, definitionIds []int64, tx pgx.Tx) error {
	for _, definitionId := range definitionIds {
		query := `INSERT INTO leitner_system_tracking (user_id, definition_id, box_id, word_id)
		VALUES ($1, $2, $3, $4) RETURNING id`

		row := tx.QueryRow(context.Background(), query, userId, definitionId, 1, wordId)
		var leitnerSystemTrackingId int64
		err := row.Scan(&leitnerSystemTrackingId)
		if err != nil {
			return err
		}

		_, err = tx.Exec(context.Background(), `INSERT INTO 
		leitner_system_history (at, box_id, leitner_system_tracking_id, success) 
		VALUES (now(), $1, $2, NULL)`, 1, leitnerSystemTrackingId)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetSkippedDefinitions returns definitions that are temporarily skipped due to error reports
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

// MarkErrorResolved removes the temporary skip and marks the error as resolved
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

// Temporarily skip this definition by updating its next review time
// This prevents it from being selected again until the error is resolved
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
