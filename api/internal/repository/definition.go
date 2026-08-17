package repository

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Definition = model.Definition

type DefinitionRepository struct {
	Db *pgxpool.Pool
}

func (repository *DefinitionRepository) Save(ctx context.Context, tokenID int64, definitions []*Definition, tx *pgx.Tx) ([]*Definition, error) {
	if tx != nil {
		return repository.saveTx(ctx, tokenID, definitions, *tx)
	}

	var saved []*Definition
	err := pgx.BeginFunc(ctx, repository.Db, func(managedTx pgx.Tx) error {
		var saveErr error
		saved, saveErr = repository.saveTx(ctx, tokenID, definitions, managedTx)
		return saveErr
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save definitions transaction: %w", err)
	}
	return saved, nil
}

func (repository *DefinitionRepository) saveTx(
	ctx context.Context,
	tokenID int64,
	definitions []*Definition,
	tx pgx.Tx,
) ([]*Definition, error) {
	// Prepare the definitions insert
	definitionsInsert := `
        INSERT INTO 
			definitions (token, language, part_of_speech, part_of_speech_normalized, meaning, examples, inflections, source, 
						source_id,sounds,phonetic_notations)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        RETURNING id, created_at, updated_at`

	wordDefinitionsInsert := `INSERT INTO word_definitions (word_id, definition_id) VALUES ($1, $2)`

	for i, def := range definitions {
		var createdAt pgtype.Timestamp
		var updatedAt pgtype.Timestamp

		// remove invalid utf8 chars
		for index, example := range def.Examples {
			def.Examples[index] = strings.ToValidUTF8(example, "")
		}

		meaning := strings.ToValidUTF8(def.Meaning, "")

		// Execute the query within the transaction
		err := tx.QueryRow(ctx, definitionsInsert, def.Token,
			def.Language, def.PartOfSpeech, def.PartOfSpeechNormalized, meaning, def.Examples, def.Inflections,
			def.Source, def.SourceID, def.Sounds, def.PhoneticNotations).Scan(&def.ID, &createdAt, &updatedAt)

		if err != nil {
			jsonString, _ := json.Marshal(def)
			common.Logger.Error("failed to insert definition", "definition", jsonString, "tokenID", tokenID)
			return nil, err
		}

		// Update the definition object with the returned values
		def.CreatedAt = createdAt
		def.UpdatedAt = updatedAt
		definitions[i] = def

		_, err = tx.Exec(ctx, wordDefinitionsInsert, tokenID, def.ID)

		if err != nil {
			common.Logger.Error("failed to insert word_definition", "def.ID", def.ID, "tokenID", tokenID)
			return nil, err
		}
	}

	return definitions, nil
}

// ReplaceReportedDefinition installs regenerated content without deleting the
// reported word's learning state. Exclusive definitions are updated in place;
// shared definitions are copied and only the reported word's relationship and
// history are moved to the replacement.
func (repository *DefinitionRepository) ReplaceReportedDefinition(
	ctx context.Context,
	wordID, userID, reportedDefinitionID int64,
	definitions []*Definition,
	tx pgx.Tx,
) ([]*Definition, error) {
	if len(definitions) == 0 {
		return nil, errors.New("replacement definition missing")
	}

	var lockedWordID int64
	if err := tx.QueryRow(ctx, `
		SELECT id FROM words WHERE id=$1 AND user_id=$2 FOR UPDATE
	`, wordID, userID).Scan(&lockedWordID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NotFoundError{ID: wordID, Entity: "word"}
		}
		return nil, err
	}

	var existingNormalizedPOS string
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE(NULLIF(d.part_of_speech_normalized, ''), d.part_of_speech)
		FROM definitions d
		JOIN word_definitions wd ON wd.definition_id=d.id
		WHERE d.id=$1 AND wd.word_id=$2
		FOR UPDATE OF d, wd
	`, reportedDefinitionID, wordID).Scan(&existingNormalizedPOS); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NotFoundError{ID: reportedDefinitionID, Entity: "word definition"}
		}
		return nil, err
	}

	replacement, err := selectReportedDefinitionReplacement(
		ctx, tx, wordID, reportedDefinitionID, existingNormalizedPOS, definitions,
	)
	if err != nil {
		return nil, err
	}

	var relationshipCount int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(DISTINCT word_id)
		FROM word_definitions
		WHERE definition_id=$1
	`, reportedDefinitionID).Scan(&relationshipCount); err != nil {
		return nil, err
	}

	if relationshipCount == 1 {
		if _, err := tx.Exec(ctx, `DELETE FROM definition_images WHERE definition_id=$1`, reportedDefinitionID); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM definition_example_audio WHERE definition_id=$1`, reportedDefinitionID); err != nil {
			return nil, err
		}
		if err := tx.QueryRow(ctx, `
			UPDATE definitions SET
				token=$2, language=$3, part_of_speech=$4,
				part_of_speech_normalized=$5, meaning=$6, examples=$7,
				inflections=$8, source=$9, source_id=$10, sounds=$11,
				phonetic_notations=$12, meaning_audio_url=NULL, updated_at=NOW()
			WHERE id=$1
			RETURNING created_at, updated_at
		`, reportedDefinitionID, replacement.Token, replacement.Language,
			replacement.PartOfSpeech, replacement.PartOfSpeechNormalized,
			strings.ToValidUTF8(replacement.Meaning, ""), replacement.Examples,
			replacement.Inflections, replacement.Source, replacement.SourceID,
			replacement.Sounds, replacement.PhoneticNotations,
		).Scan(&replacement.CreatedAt, &replacement.UpdatedAt); err != nil {
			return nil, err
		}
		replacement.ID = reportedDefinitionID
	} else {
		inserted, err := repository.Save(ctx, wordID, []*Definition{replacement}, &tx)
		if err != nil {
			return nil, err
		}
		replacement = inserted[0]
		if _, err := tx.Exec(ctx, `
			UPDATE leitner_system_tracking
			SET definition_id=$1
			WHERE definition_id=$2 AND word_id=$3 AND user_id=$4
		`, replacement.ID, reportedDefinitionID, wordID, userID); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE quiz_performance
			SET definition_id=$1
			WHERE definition_id=$2 AND word_id=$3 AND user_id=$4
		`, replacement.ID, reportedDefinitionID, wordID, userID); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `
			DELETE FROM word_definitions
			WHERE definition_id=$1 AND word_id=$2
		`, reportedDefinitionID, wordID); err != nil {
			return nil, err
		}
	}

	// A targeted report replaces only the reported meaning/example. Other
	// provider results describe relationships that are already present and must
	// not be appended as duplicate definitions on every regeneration.
	return []*Definition{replacement}, nil
}

func canonicalDefinitionMeaning(meaning string) string {
	return strings.ToLower(strings.TrimSpace(meaning))
}

func selectReportedDefinitionReplacement(
	ctx context.Context,
	tx pgx.Tx,
	wordID, reportedDefinitionID int64,
	existingNormalizedPOS string,
	definitions []*Definition,
) (*Definition, error) {
	otherMeanings := make(map[string]struct{})
	rows, err := tx.Query(ctx, `
		SELECT meaning
		FROM definitions d
		JOIN word_definitions wd ON wd.definition_id=d.id
		WHERE wd.word_id=$1 AND d.id<>$2
	`, wordID, reportedDefinitionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var meaning string
		if err := rows.Scan(&meaning); err != nil {
			return nil, err
		}
		otherMeanings[canonicalDefinitionMeaning(meaning)] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, definition := range definitions {
		if !strings.EqualFold(definition.PartOfSpeechNormalized, existingNormalizedPOS) {
			continue
		}
		if _, duplicate := otherMeanings[canonicalDefinitionMeaning(definition.Meaning)]; !duplicate {
			return definition, nil
		}
	}
	for _, definition := range definitions {
		if strings.EqualFold(definition.PartOfSpeechNormalized, existingNormalizedPOS) {
			return definition, nil
		}
	}
	return definitions[0], nil
}

// all other defitions for the same word defined by the the records which ids are in definitionIdsToIgnore will be ignored too
func (repository *DefinitionRepository) GetRandomMeanings(
	ctx context.Context,
	userID, wordlistID int64,
	language, partOfSpeechNormalized string,
	definitionIDsToIgnore []int,
	limit int,
) ([]string, error) {
	query := `
        WITH excluded_words AS (
			SELECT DISTINCT wd.word_id
			FROM word_definitions wd
			WHERE wd.definition_id = ANY($5)
			),
        candidates AS (
			SELECT DISTINCT def.meaning
			FROM definitions def
			JOIN word_definitions wd ON wd.definition_id = def.id
			JOIN words w ON w.id = wd.word_id
			WHERE w.user_id = $1
			AND w.wordlist_id = $2
			AND def.language = $3
			AND LOWER(COALESCE(NULLIF(def.part_of_speech_normalized, ''), def.part_of_speech)) = LOWER($4)
			AND wd.word_id NOT IN (SELECT word_id FROM excluded_words)
			AND def.id <> ALL($5)
			AND def.meaning NOT IN (SELECT meaning FROM definitions WHERE id = ANY($5))
		)
        SELECT meaning
        FROM candidates
        ORDER BY RANDOM(), length(meaning) ASC
		LIMIT $6;
	`

	rows, err := repository.Db.Query(
		ctx,
		query,
		userID,
		wordlistID,
		language,
		partOfSpeechNormalized,
		definitionIDsToIgnore,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query random meanings for definitions %v: %w", definitionIDsToIgnore, err)
	}
	defer rows.Close()

	var meanings []string
	for rows.Next() {
		var meaning string
		err = rows.Scan(&meaning)
		if err != nil {
			return nil, err
		}
		meanings = append(meanings, meaning)
	}

	return meanings, nil
}

func (repository *DefinitionRepository) GetRandomTokens(
	ctx context.Context,
	userID, wordlistID int64,
	language, partOfSpeechNormalized string,
	definitionIDsToIgnore []int,
	limit int,
) ([]string, error) {
	query := `
		WITH candidate_tokens AS (
			SELECT DISTINCT def.token
			FROM 
				definitions def
			JOIN 
				word_definitions wd ON wd.definition_id = def.id
			JOIN
				words w ON w.id = wd.word_id
			WHERE 
				w.user_id = $1
				AND w.wordlist_id = $2
				AND def.language = $3
				AND LOWER(COALESCE(NULLIF(def.part_of_speech_normalized, ''), def.part_of_speech)) = LOWER($4)
				AND wd.word_id NOT IN (select word_id FROM word_definitions WHERE definition_id = ANY($5))
				AND token NOT IN (SELECT token FROM definitions WHERE id = ANY($5))
		)
		SELECT token
		FROM candidate_tokens
		ORDER BY random()
		LIMIT $6;
	`
	rows, err := repository.Db.Query(
		ctx,
		query,
		userID,
		wordlistID,
		language,
		partOfSpeechNormalized,
		definitionIDsToIgnore,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		err = rows.Scan(&token)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}

func (repository *DefinitionRepository) Find(ctx context.Context, args FindArgs) ([]*Definition, error) {
	var builder strings.Builder

	builder.WriteString(`SELECT id, token, language, part_of_speech, part_of_speech_normalized, is_verb_type, meaning, examples, inflections, source, 
						source_id,sounds,phonetic_notations, COALESCE(meaning_audio_url,''), created_at, updated_at
		FROM definitions`)

	queryArgs := []any{}
	filters := []string{}
	index := 1

	if args.ID != nil {
		filters = append(filters, fmt.Sprintf("id = $%d", index))
		index++
		queryArgs = append(queryArgs, strconv.FormatInt(*args.ID, 10))
	}

	if args.Name != nil {
		filters = append(filters, fmt.Sprintf("token = $%d", index))
		queryArgs = append(queryArgs, *args.Name)
	}

	if len(filters) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(filters, " AND "))
	}

	query := builder.String()
	rows, err := repository.Db.Query(ctx, query, queryArgs...)

	if err != nil {
		if err == pgx.ErrNoRows {
			return []*Definition{}, nil
		}
		return nil, err
	}

	results := []*Definition{}

	for rows.Next() {
		var def Definition

		err = rows.Scan(&def.ID,
			&def.Token, &def.Language, &def.PartOfSpeech, &def.PartOfSpeechNormalized, &def.IsVerbType, &def.Meaning, &def.Examples, &def.Inflections,
			&def.Source, &def.SourceID, &def.Sounds, &def.PhoneticNotations, &def.MeaningAudioURL,
			&def.CreatedAt, &def.UpdatedAt)

		if err != nil {
			return nil, err
		}

		results = append(results, &def)
	}

	return results, nil
}

// DeleteWordDefinitions removes every definition relationship for a word.
// Definitions used only by this word are deleted; shared definitions remain for
// their other words. Tracking/history for removed shared relationships is also
// deleted so no orphaned learning state survives without a word-definition link.
func (repository *DefinitionRepository) DeleteWordDefinitions(ctx context.Context, wordID int64, tx *pgx.Tx) error {
	if tx != nil {
		return repository.deleteWordDefinitionsTx(ctx, wordID, *tx)
	}

	err := pgx.BeginFunc(ctx, repository.Db, func(managedTx pgx.Tx) error {
		return repository.deleteWordDefinitionsTx(ctx, wordID, managedTx)
	})
	if err != nil {
		return fmt.Errorf("failed to delete word definitions transaction: %w", err)
	}
	return nil
}

func (repository *DefinitionRepository) deleteWordDefinitionsTx(ctx context.Context, wordID int64, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, `
		DELETE FROM leitner_system_tracking lst
		USING word_definitions wd
		WHERE wd.word_id=$1
		  AND lst.word_id=wd.word_id
		  AND lst.definition_id=wd.definition_id
	`, wordID); err != nil {
		return fmt.Errorf("failed to delete word definition tracking: %w", err)
	}

	// Delete each exclusive definition independently. The old aggregate count
	// preserved every exclusive definition whenever any definition in the target
	// set was also linked to another word.
	if _, err := tx.Exec(ctx, `
		DELETE FROM definitions d
		WHERE EXISTS (
			SELECT 1 FROM word_definitions target
			WHERE target.word_id=$1 AND target.definition_id=d.id
		)
		AND NOT EXISTS (
			SELECT 1 FROM word_definitions other
			WHERE other.definition_id=d.id AND other.word_id<>$1
		)
	`, wordID); err != nil {
		return fmt.Errorf("failed to delete exclusive definitions: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM word_definitions WHERE word_id=$1`, wordID); err != nil {
		return fmt.Errorf("failed to delete shared word definitions: %w", err)
	}
	return nil
}

func (repository *DefinitionRepository) DidUserCreateWord(ctx context.Context, wordID, userID int64) (bool, error) {
	query := `SELECT count(*) FROM words w WHERE w.id=$1 and user_id=$2`
	row := repository.Db.QueryRow(ctx, query, wordID, userID)
	var count int
	err := row.Scan(&count)

	if err != nil {
		return false, err
	}

	return count == 1, nil
}

func (repository *DefinitionRepository) GetDefinitionsByWordID(ctx context.Context, wordlistID, wordID, userID int64, limit int, cursor *int64) ([]*Definition, error) {
	queryArgs := []any{wordID, wordlistID, userID}
	cursorClause := ""
	if cursor != nil {
		cursorClause = " AND d.id > $4"
		queryArgs = append(queryArgs, *cursor)
	}
	query := `
		SELECT d.id, d.token, d.language, d.part_of_speech, d.part_of_speech_normalized, d.is_verb_type, d.meaning, d.examples, d.inflections, 
			   d.source, d.source_id, d.sounds, d.phonetic_notations, COALESCE(d.meaning_audio_url,''), d.created_at, d.updated_at
		FROM definitions d
		JOIN word_definitions wd ON wd.definition_id = d.id
		JOIN words w ON w.id = wd.word_id
		WHERE w.id = $1 AND w.wordlist_id = $2 AND w.user_id = $3` + cursorClause + `
		ORDER BY d.id ASC LIMIT $` + strconv.Itoa(len(queryArgs)+1)
	queryArgs = append(queryArgs, limit)

	rows, err := repository.Db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	definitions := make([]*Definition, 0)
	for rows.Next() {
		var def Definition

		err = rows.Scan(&def.ID,
			&def.Token, &def.Language, &def.PartOfSpeech, &def.PartOfSpeechNormalized, &def.IsVerbType, &def.Meaning, &def.Examples, &def.Inflections,
			&def.Source, &def.SourceID, &def.Sounds, &def.PhoneticNotations, &def.MeaningAudioURL,
			&def.CreatedAt, &def.UpdatedAt)

		if err != nil {
			return nil, err
		}

		definitions = append(definitions, &def)
	}

	return definitions, nil
}

type FindArgs struct {
	ID   *int64
	Name *string
}

const (
	MaxDefinitionBatchWordIDs    = 50
	MaxDefinitionsPerBatchWordID = 50
)

func NewDefinitionRepository(db *pgxpool.Pool) *DefinitionRepository {
	return &DefinitionRepository{Db: db}
}

// WordDefinitionsResponse is used by the batch endpoint to return grouped data
type WordDefinitionsResponse struct {
	WordID      int64        `json:"wordId"`
	Name        string       `json:"name"` // token/name of the word
	Definitions []Definition `json:"definitions"`
}

// DefinitionBatchPage keeps the existing response array while exposing a
// per-word keyset for callers that need every definition.
type DefinitionBatchPage struct {
	Results     []WordDefinitionsResponse
	NextCursors map[int64]int64
}

// GetDefinitionsByWordIDs fetches definitions for multiple word IDs, scoped to user and wordlist
func (repository *DefinitionRepository) GetDefinitionsByWordIDs(ctx context.Context, wordlistID, userID int64, wordIDs []int64, definitionCursors map[int64]int64) (DefinitionBatchPage, error) {
	if len(wordIDs) == 0 {
		return DefinitionBatchPage{Results: []WordDefinitionsResponse{}}, nil
	}
	if len(wordIDs) > MaxDefinitionBatchWordIDs {
		return DefinitionBatchPage{}, fmt.Errorf("definition batch may contain at most %d word IDs", MaxDefinitionBatchWordIDs)
	}
	requestedWordIDs := make([]int64, len(wordIDs))
	afterDefinitionIDs := make([]int64, len(wordIDs))
	copy(requestedWordIDs, wordIDs)
	for index, wordID := range wordIDs {
		afterDefinitionIDs[index] = definitionCursors[wordID]
	}

	query := `
		WITH requested_words AS (
			SELECT word_id, after_definition_id
			FROM unnest($3::bigint[], $4::bigint[]) AS requested(word_id, after_definition_id)
		), matched_definitions AS (
		SELECT
            w.id as word_id,
            d.token as name,
            d.id, d.token, d.language, d.part_of_speech, d.part_of_speech_normalized, d.is_verb_type, d.meaning, d.examples, d.inflections,
			d.source, d.source_id, d.sounds, d.phonetic_notations, COALESCE(d.meaning_audio_url,'') AS meaning_audio_url, d.created_at, d.updated_at,
			ROW_NUMBER() OVER (PARTITION BY w.id ORDER BY d.id ASC) AS definition_position
        FROM words w
        JOIN word_definitions wd ON wd.word_id = w.id
        JOIN definitions d ON d.id = wd.definition_id
		JOIN requested_words requested ON requested.word_id = w.id
        WHERE w.wordlist_id = $1 AND w.user_id = $2 AND d.id > requested.after_definition_id
		)
		SELECT word_id, name, id, token, language, part_of_speech, part_of_speech_normalized, is_verb_type, meaning, examples, inflections,
		source, source_id, sounds, phonetic_notations, meaning_audio_url, created_at, updated_at, definition_position
		FROM matched_definitions
		WHERE definition_position <= $5
		ORDER BY word_id ASC, id ASC`

	rows, err := repository.Db.Query(ctx, query, wordlistID, userID, requestedWordIDs, afterDefinitionIDs, MaxDefinitionsPerBatchWordID+1)
	if err != nil {
		return DefinitionBatchPage{}, err
	}
	defer rows.Close()

	// Group rows by word_id
	grouped := make(map[int64]*WordDefinitionsResponse)
	order := make([]int64, 0)
	// Keep a complete cumulative cursor for every requested word. If any word
	// has another page, callers must resend the whole map so words that finish
	// earlier do not reset to definition ID zero on a later request.
	nextCursors := make(map[int64]int64, len(wordIDs))
	for _, wordID := range wordIDs {
		nextCursors[wordID] = definitionCursors[wordID]
	}
	hasMore := false

	for rows.Next() {
		var wordID int64
		var name string
		var def Definition
		var position int
		if scanErr := rows.Scan(
			&wordID, &name,
			&def.ID, &def.Token, &def.Language, &def.PartOfSpeech, &def.PartOfSpeechNormalized, &def.IsVerbType, &def.Meaning,
			&def.Examples, &def.Inflections, &def.Source, &def.SourceID, &def.Sounds, &def.PhoneticNotations,
			&def.MeaningAudioURL, &def.CreatedAt, &def.UpdatedAt, &position,
		); scanErr != nil {
			return DefinitionBatchPage{}, scanErr
		}

		if _, exists := grouped[wordID]; !exists {
			grouped[wordID] = &WordDefinitionsResponse{WordID: wordID, Name: name, Definitions: make([]Definition, 0)}
			order = append(order, wordID)
		}
		if position > MaxDefinitionsPerBatchWordID {
			hasMore = true
			continue
		}
		grouped[wordID].Definitions = append(grouped[wordID].Definitions, def)
		nextCursors[wordID] = def.ID
	}
	if err := rows.Err(); err != nil {
		return DefinitionBatchPage{}, err
	}

	// Preserve ascending order by first appearance
	results := make([]WordDefinitionsResponse, 0, len(grouped))
	for _, id := range order {
		results = append(results, *grouped[id])
	}
	if !hasMore {
		nextCursors = nil
	}
	return DefinitionBatchPage{Results: results, NextCursors: nextCursors}, nil
}

func (repository *DefinitionRepository) GetDefinitionByID(ctx context.Context, definitionID int64) (*Definition, error) {
	query := `
		SELECT id, token, language, part_of_speech, part_of_speech_normalized, is_verb_type, meaning, examples, inflections, 
			   source, source_id, sounds, phonetic_notations, COALESCE(meaning_audio_url,''), created_at, updated_at
		FROM definitions 
		WHERE id = $1`

	var def Definition
	err := repository.Db.QueryRow(ctx, query, definitionID).Scan(
		&def.ID, &def.Token, &def.Language, &def.PartOfSpeech, &def.PartOfSpeechNormalized, &def.IsVerbType, &def.Meaning,
		&def.Examples, &def.Inflections, &def.Source, &def.SourceID,
		&def.Sounds, &def.PhoneticNotations, &def.MeaningAudioURL, &def.CreatedAt, &def.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NotFoundError{}
		}
		return nil, err
	}

	return &def, nil
}

func (repository *DefinitionRepository) CreateExampleAudio(ctx context.Context, definitionID int64, exampleText, audioURL, inflectionType string) error {
	// Generate hash for the example text
	exampleHash := fmt.Sprintf("%x", sha256.Sum256([]byte(exampleText)))

	query := `INSERT INTO definition_example_audio (definition_id, example_text, example_hash, audio_url, inflection_type) 
			  VALUES ($1, $2, $3, $4, NULLIF($5, ''))`

	_, err := repository.Db.Exec(ctx, query, definitionID, exampleText, exampleHash, audioURL, inflectionType)
	return err
}

func (repository *DefinitionRepository) SetMeaningAudioURLIfEmpty(ctx context.Context, definitionID int64, audioURL string) (bool, error) {
	command, err := repository.Db.Exec(ctx, `
		UPDATE definitions
		SET meaning_audio_url = $1, updated_at = NOW()
		WHERE id = $2 AND COALESCE(meaning_audio_url, '') = ''`, audioURL, definitionID)
	if err != nil {
		return false, err
	}
	return command.RowsAffected() == 1, nil
}

func (repository *DefinitionRepository) GetLeastUsedExampleAudio(ctx context.Context, definitionID int64) (*model.DefinitionExampleAudio, error) {
	query := `
		SELECT dea.id, dea.definition_id, dea.example_text, dea.audio_url, COALESCE(dea.inflection_type,''), dea.created_at
		FROM definition_example_audio dea
		LEFT JOIN example_audio_usage eau ON dea.id = eau.example_audio_id
		WHERE dea.definition_id = $1
		ORDER BY COALESCE(eau.usage_count, 0), COALESCE(eau.last_used_at, '1970-01-01')
		LIMIT 1`

	var audio model.DefinitionExampleAudio
	err := repository.Db.QueryRow(ctx, query, definitionID).Scan(
		&audio.ID, &audio.DefinitionID, &audio.ExampleText, &audio.AudioURL,
		&audio.InflectionType, &audio.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NotFoundError{}
		}
		return nil, err
	}

	// Update usage tracking
	updateQuery := `
		INSERT INTO example_audio_usage (definition_id, example_audio_id, last_used_at, usage_count)
		VALUES ($1, $2, NOW(), 1)
		ON CONFLICT (definition_id, example_audio_id)
		DO UPDATE SET last_used_at = NOW(), usage_count = example_audio_usage.usage_count + 1`

	_, err = repository.Db.Exec(ctx, updateQuery, definitionID, audio.ID)
	if err != nil {
		return nil, err
	}

	return &audio, nil
}

func (repository *DefinitionRepository) GetExampleAudioByDefinitionID(ctx context.Context, definitionID int64) ([]model.DefinitionExampleAudio, error) {
	query := `
		SELECT id, definition_id, example_text, audio_url, COALESCE(inflection_type,''), created_at
		FROM definition_example_audio
		WHERE definition_id = $1
		ORDER BY created_at DESC`

	rows, err := repository.Db.Query(ctx, query, definitionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var audioFiles []model.DefinitionExampleAudio
	for rows.Next() {
		var audio model.DefinitionExampleAudio
		err := rows.Scan(&audio.ID, &audio.DefinitionID, &audio.ExampleText,
			&audio.AudioURL, &audio.InflectionType, &audio.CreatedAt)
		if err != nil {
			return nil, err
		}
		audioFiles = append(audioFiles, audio)
	}

	return audioFiles, rows.Err()
}
