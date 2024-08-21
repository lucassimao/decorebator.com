package api

import (
	"context"
	"encoding/json"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Definition = model.Definition

type DefinitionRepository struct {
	db *pgxpool.Pool
}

func (repository *DefinitionRepository) save(tokenId int64, definitions []Definition) ([]Definition, error) {
	// Start a transaction
	tx, err := repository.db.Begin(context.Background())
	if err != nil {
		return nil, err
	}

	// Prepare the definitions insert
	definitionsInsert := `
        INSERT INTO 
			definitions (token, language, part_of_speech, meaning, 	examples, inflections, source, 
						source_id,sounds,phonetic_notations)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9,$10)
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
		err := tx.QueryRow(context.Background(), definitionsInsert, def.Token,
			def.Language, def.PartOfSpeech, meaning, def.Examples, def.Inflections,
			def.Source, def.SourceId, def.Sounds, def.PhoneticNotations).Scan(&def.ID, &createdAt, &updatedAt)

		if err != nil {
			jsonString, _ := json.Marshal(def)
			common.Logger.Error("failed to insert definition", "definition", jsonString, "tokenId", tokenId)
			tx.Rollback(context.Background())
			return nil, err
		}

		// Update the definition object with the returned values
		def.CreatedAt = createdAt
		def.UpdatedAt = updatedAt
		definitions[i] = def

		_, err = tx.Exec(context.Background(), wordDefinitionsInsert, tokenId, def.ID)

		if err != nil {
			common.Logger.Error("failed to insert word_definition", "def.ID", def.ID, "tokenId", tokenId)
			tx.Rollback(context.Background())
			return nil, err
		}

	}

	if err := tx.Commit(context.Background()); err != nil {
		return nil, err
	}

	return definitions, nil
}

func (repository *DefinitionRepository) getRandomMeanings(definitionIdsToIgnore []int, limit int) ([]string, error) {
	// all other defitions for the same word defined by the the records which ids are in definitionIdsToIgnore will be ignored too
	// selecting random() to avoid the error 'SELECT DISTINCT, ORDER BY expressions must appear in select list (SQLSTATE 42P10)'
	query := `
		WITH options AS (
			SELECT DISTINCT def.meaning, random()
			FROM definitions def
			JOIN word_definitions wd ON wd.definition_id = def.id
			WHERE wd.word_id NOT IN (
				SELECT word_id 
				FROM word_definitions 
				WHERE definition_id = ANY($1)
			)
			AND length(def.meaning) < 50
			AND def.meaning NOT IN (
				SELECT meaning 
				FROM definitions 
				WHERE id = ANY($1)
			)
			AND part_of_speech IN (
				SELECT part_of_speech 
				FROM definitions 
				WHERE id = ANY($1)
			)
			ORDER BY random() LIMIT $2
		)
		SELECT meaning FROM options;
	`

	rows, err := repository.db.Query(context.Background(), query, definitionIdsToIgnore, limit)
	if err != nil {
		return nil, err
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

func (repository *DefinitionRepository) getRandomTokens(definitionIdsToIgnore []int, partOfSpeech string, limit int) ([]string, error) {
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
			ORDER BY random() LIMIT $3
		)
		SELECT token FROM tokens ;
	`
	rows, err := repository.db.Query(context.Background(), query, partOfSpeech, definitionIdsToIgnore, limit)
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

func (repository *DefinitionRepository) getById(id int64) (*Definition, error) {

	var query = `SELECT token, language, part_of_speech, meaning, examples, inflections, source, 
						source_id,sounds,phonetic_notations, created_at, updated_at
		FROM definitions WHERE id = $1`

	var def Definition
	def.ID = id

	err := repository.db.QueryRow(context.Background(), query, id).Scan(
		&def.Token, &def.Language, &def.PartOfSpeech, &def.Meaning, &def.Examples, &def.Inflections,
		&def.Source, &def.SourceId, &def.Sounds, &def.PhoneticNotations, &def.CreatedAt, &def.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, common.NotFoundError{ID: id, Entity: "definition"}
	}

	if err != nil {
		return nil, err
	}

	return &def, nil
}
