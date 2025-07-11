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
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Definition = model.Definition

type DefinitionRepository struct {
	Db *pgxpool.Pool
}

func (repository *DefinitionRepository) Save(tokenID int64, definitions []*Definition, tx *pgx.Tx) ([]*Definition, error) {
	var managedTx pgx.Tx
	var err error
	ctx := context.Background()

	// If no transaction provided, create and manage our own
	if tx == nil {
		managedTx, err = repository.Db.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			if err == nil {
				if commitErr := managedTx.Commit(ctx); commitErr != nil {
					common.Logger.Error("failed to commit transaction in definition repository", "error", commitErr)
				}
			} else {
				if rollbackErr := managedTx.Rollback(ctx); rollbackErr != nil {
					common.Logger.Error("failed to rollback transaction in definition repository", "error", rollbackErr)
				}
			}
		}()
	} else {
		managedTx = *tx
	}

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
		err := managedTx.QueryRow(context.Background(), definitionsInsert, def.Token,
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

		_, err = managedTx.Exec(context.Background(), wordDefinitionsInsert, tokenID, def.ID)

		if err != nil {
			common.Logger.Error("failed to insert word_definition", "def.ID", def.ID, "tokenID", tokenID)
			return nil, err
		}
	}

	return definitions, nil
}

// all other defitions for the same word defined by the the records which ids are in definitionIdsToIgnore will be ignored too
func (repository *DefinitionRepository) GetRandomMeanings(definitionIDsToIgnore []int, limit int) ([]string, error) {
	// Handle empty array case
	if len(definitionIDsToIgnore) == 0 {
		return []string{}, nil
	}

	query := `
        WITH target_definitions AS (
			SELECT DISTINCT part_of_speech 
			FROM definitions 
			WHERE id = ANY($1)
		),
        excluded_words AS (
			SELECT DISTINCT wd.word_id
			FROM word_definitions wd
			WHERE wd.definition_id = ANY($1)
		),
        candidates AS (
			SELECT DISTINCT def.meaning
			FROM definitions def
			JOIN word_definitions wd ON wd.definition_id = def.id
			WHERE def.part_of_speech IN (SELECT part_of_speech FROM target_definitions)
			AND wd.word_id NOT IN (SELECT word_id FROM excluded_words)
			AND def.id <> ALL($1)
		)
        SELECT meaning
        FROM candidates
        ORDER BY length(meaning) ASC, RANDOM() 
        LIMIT $2;
	`

	rows, err := repository.Db.Query(context.Background(), query, definitionIDsToIgnore, limit)
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

func (repository *DefinitionRepository) GetRandomTokens(definitionIDsToIgnore []int, partOfSpeech string, limit int) ([]string, error) {
	// selecting random() to avoid the error 'SELECT DISTINCT, ORDER BY expressions must appear in select list (SQLSTATE 42P10)'
	query := `
		WITH tokens AS (
			SELECT 
				DISTINCT token, random()
			FROM 
				definitions def
			JOIN 
				word_definitions wd ON wd.definition_id = def.id
			WHERE 
				part_of_speech=$1 
				AND wd.word_id NOT IN (select word_id FROM word_definitions WHERE definition_id = ANY($2))
				AND token NOT IN (SELECT token FROM definitions WHERE id = ANY($2))
			ORDER BY random() LIMIT $3
		)
		SELECT token FROM tokens ;
	`
	rows, err := repository.Db.Query(context.Background(), query, partOfSpeech, definitionIDsToIgnore, limit)
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

func (repository *DefinitionRepository) Find(args FindArgs) ([]*Definition, error) {
	var builder strings.Builder

	builder.WriteString(`SELECT id, token, language, part_of_speech, part_of_speech_normalized, is_verb_type, meaning, examples, inflections, source, 
						source_id,sounds,phonetic_notations, created_at, updated_at
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
		index++
		queryArgs = append(queryArgs, *args.Name)
	}

	if len(filters) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(filters, " AND "))
	}

	query := builder.String()
	rows, err := repository.Db.Query(context.Background(), query, queryArgs...)

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
			&def.Source, &def.SourceID, &def.Sounds, &def.PhoneticNotations,
			&def.CreatedAt, &def.UpdatedAt)

		if err != nil {
			return nil, err
		}

		results = append(results, &def)
	}

	return results, nil
}

// Delete definitions and word_definitions only if they are associated to a single word.
// If associated to more than one word, then just dele word_definitions entries.
// If tx is nil, a new transaction is created and managed internally.
func (repository *DefinitionRepository) DeleteWordDefinitions(wordID int64, tx *pgx.Tx) error {
	var managedTx pgx.Tx
	var err error
	ctx := context.Background()

	// If no transaction provided, create and manage our own
	if tx == nil {
		managedTx, err = repository.Db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			if err == nil {
				if commitErr := managedTx.Commit(ctx); commitErr != nil {
					common.Logger.Error("failed to commit transaction in definition repository", "error", commitErr)
				}
			} else {
				if rollbackErr := managedTx.Rollback(ctx); rollbackErr != nil {
					common.Logger.Error("failed to rollback transaction in definition repository", "error", rollbackErr)
				}
			}
		}()
	} else {
		managedTx = *tx
	}

	query := `SELECT count(distinct word_id) as count
				FROM word_definitions 
				WHERE definition_id IN (SELECT definition_id from word_definitions WHERE word_id=$1)`

	row := managedTx.QueryRow(ctx, query, wordID)
	var count int

	err = row.Scan(&count)

	if err != nil {
		common.Logger.Error("failed to query how many words use the same definition", "error", err, "wordID", wordID)
		return err
	}

	// just one word use these definitions. Drop them'll
	if count == 1 {
		// deleting from definitions table cascades to word_definitions and definition_images
		_, err = managedTx.Exec(ctx, "DELETE FROM definitions WHERE id in (SELECT definition_id from word_definitions WHERE word_id=$1)", wordID)
		if err != nil {
			return errors.New("failed to delete definitions")
		}
	} else {
		// definitions are shared. will only delete entries in the many to many table
		_, err = managedTx.Exec(ctx, "DELETE from word_definitions WHERE word_id=$1", wordID)
		if err != nil {
			return errors.New("failed to delete word_definitions")
		}
	}

	return nil
}

func (repository *DefinitionRepository) DidUserCreateWord(wordID, userID int64) (bool, error) {
	query := `SELECT count(*) FROM words w WHERE w.id=$1 and user_id=$2`
	row := repository.Db.QueryRow(context.Background(), query, wordID, userID)
	var count int
	err := row.Scan(&count)

	if err != nil {
		return false, err
	}

	return count == 1, nil
}

func (repository *DefinitionRepository) GetDefinitionsByWordID(wordID, userID int64) ([]*Definition, error) {
	query := `
		SELECT d.id, d.token, d.language, d.part_of_speech, d.part_of_speech_normalized, d.is_verb_type, d.meaning, d.examples, d.inflections, 
			   d.source, d.source_id, d.sounds, d.phonetic_notations, d.created_at, d.updated_at
		FROM definitions d
		JOIN word_definitions wd ON wd.definition_id = d.id
		JOIN words w ON w.id = wd.word_id
		WHERE w.id = $1 AND w.user_id = $2
		ORDER BY d.id ASC`

	rows, err := repository.Db.Query(context.Background(), query, wordID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var definitions []*Definition
	for rows.Next() {
		var def Definition

		err = rows.Scan(&def.ID,
			&def.Token, &def.Language, &def.PartOfSpeech, &def.PartOfSpeechNormalized, &def.IsVerbType, &def.Meaning, &def.Examples, &def.Inflections,
			&def.Source, &def.SourceID, &def.Sounds, &def.PhoneticNotations,
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

func NewDefinitionRepository(db *pgxpool.Pool) *DefinitionRepository {
	return &DefinitionRepository{Db: db}
}

func (repository *DefinitionRepository) GetDefinitionByID(definitionID int64) (*Definition, error) {
	query := `
		SELECT id, token, language, part_of_speech, part_of_speech_normalized, is_verb_type, meaning, examples, inflections, 
			   source, source_id, sounds, phonetic_notations, created_at, updated_at
		FROM definitions 
		WHERE id = $1`

	var def Definition
	err := repository.Db.QueryRow(context.Background(), query, definitionID).Scan(
		&def.ID, &def.Token, &def.Language, &def.PartOfSpeech, &def.PartOfSpeechNormalized, &def.IsVerbType, &def.Meaning,
		&def.Examples, &def.Inflections, &def.Source, &def.SourceID,
		&def.Sounds, &def.PhoneticNotations, &def.CreatedAt, &def.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NotFoundError{}
		}
		return nil, err
	}

	return &def, nil
}

func (repository *DefinitionRepository) CreateExampleAudio(definitionID int64, exampleText, audioURL, inflectionType string) error {
	// Generate hash for the example text
	exampleHash := fmt.Sprintf("%x", sha256.Sum256([]byte(exampleText)))

	query := `INSERT INTO definition_example_audio (definition_id, example_text, example_hash, audio_url, inflection_type) 
			  VALUES ($1, $2, $3, $4, NULLIF($5, ''))`

	_, err := repository.Db.Exec(context.Background(), query, definitionID, exampleText, exampleHash, audioURL, inflectionType)
	return err
}

func (repository *DefinitionRepository) GetLeastUsedExampleAudio(definitionID int64) (*model.DefinitionExampleAudio, error) {
	query := `
		SELECT dea.id, dea.definition_id, dea.example_text, dea.audio_url, COALESCE(dea.inflection_type,''), dea.created_at
		FROM definition_example_audio dea
		LEFT JOIN example_audio_usage eau ON dea.id = eau.example_audio_id
		WHERE dea.definition_id = $1
		ORDER BY COALESCE(eau.usage_count, 0), COALESCE(eau.last_used_at, '1970-01-01')
		LIMIT 1`

	var audio model.DefinitionExampleAudio
	err := repository.Db.QueryRow(context.Background(), query, definitionID).Scan(
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

	_, err = repository.Db.Exec(context.Background(), updateQuery, definitionID, audio.ID)
	if err != nil {
		return nil, err
	}

	return &audio, nil
}

func (repository *DefinitionRepository) GetExampleAudioByDefinitionID(definitionID int64) ([]model.DefinitionExampleAudio, error) {
	query := `
		SELECT id, definition_id, example_text, audio_url, COALESCE(inflection_type,''), created_at
		FROM definition_example_audio
		WHERE definition_id = $1
		ORDER BY created_at DESC`

	rows, err := repository.Db.Query(context.Background(), query, definitionID)
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
