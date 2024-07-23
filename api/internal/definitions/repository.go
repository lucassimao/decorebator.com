package definitions

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"decorebator.com/internal/common"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Inflection struct {
	Inflection string   `json:"inflection"`
	Tense      string   `json:"tense"`
	Examples   []string `json:"examples"`
}

type Accent string

const (
	US     Accent = "US"
	CANADA Accent = "CA"
	UK     Accent = "UK"
)

type Sound struct {
	Accent Accent `json:"accent"`
	Link   string `json:"link"`
}

type PhoneticNotation struct {
	Ipa    string `json:"ipa"`
	Accent Accent `json:"accent"`
}

type DefinitionSource string

const (
	ChatGPT    DefinitionSource = "ChatGPT"
	Wiktionary DefinitionSource = "wiktionary"
)

type Definition struct {
	ID                int64
	Token             string
	Language          string
	Meaning           string       `json:"meaning"`
	PartOfSpeech      string       `json:"part_of_speech"`
	Examples          []string     `json:"examples"`
	Inflections       []Inflection `json:"inflections"`
	Source            DefinitionSource
	SourceId          string
	Sounds            []Sound
	PhoneticNotations []PhoneticNotation
	ImageUrl          string

	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
}

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
	query := `
		WITH options AS (
			SELECT DISTINCT def.meaning 
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
		)
		SELECT meaning FROM options ORDER BY random() LIMIT $2;
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

func (repository *DefinitionRepository) getRandomExamples(filterOutIds []int, partOfSpeech string, limit int) ([]string, error) {
	query := `
		WITH random_item AS (
			SELECT
			id,
			examples,
			floor(random() * array_length(examples, 1))::int + 1 AS random_index
			FROM definitions
			where array_length(examples, 1) > 0 AND part_of_speech=$1 AND id != ALL($2)
			ORDER BY random()
			LIMIT $3
		)
		SELECT
			id,
			examples[random_index] AS random_example,
			random_index
		FROM random_item;
	`
	rows, err := repository.db.Query(context.Background(), query, partOfSpeech, filterOutIds, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var examples []string
	for rows.Next() {
		var example string
		err = rows.Scan(&example)
		if err != nil {
			return nil, err
		}
		examples = append(examples, example)
	}
	return examples, nil
}

func (repository *DefinitionRepository) getRandomTokens(definitionIdsToIgnore []int, partOfSpeech string, limit int) ([]string, error) {
	// using CTE as DISTINCT couldn't be applied directly to the main query that's using ORDER BY random()
	query := `
		WITH tokens AS (
			SELECT 
				DISTINCT token
			FROM 
				definitions def
			JOIN 
				word_definitions wd ON wd.definition_id = def.id
			WHERE 
				part_of_speech=$1 
				AND wd.word_id NOT IN (select word_id FROM word_definitions WHERE definition_id = ANY($2)) 
		)
		SELECT token FROM tokens ORDER BY random() LIMIT $3;
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

	if err == sql.ErrNoRows {
		return nil, common.NotFoundError{ID: id, Entity: "definition"}
	}

	if err != nil {
		return nil, err
	}

	return &def, nil
}

func (repository *DefinitionRepository) setImage(id int64, imageUrl string) error {

	query := `UPDATE definitions SET image_url = $1 WHERE id=$2`

	_, err := repository.db.Exec(context.Background(), query, imageUrl, id)
	return err
}
