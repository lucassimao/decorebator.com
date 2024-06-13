package definitions

import (
	"context"
	"log"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Inflection struct {
	Inflection string   `json:"inflection"`
	Tense      string   `json:"tense"`
	Examples   []string `json:"examples"`
}

type Definition struct {
	ID           int64
	Token        string
	Language     string
	Meaning      string       `json:"meaning"`
	PartOfSpeech string       `json:"part_of_speech"`
	Examples     []string     `json:"examples"`
	Inflections  []Inflection `json:"inflections"`
	Source       string

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
		log.Printf("Error while starting tx %v\n", err)
		return nil, err
	}

	// Prepare the definitions insert
	definitionsInsert := `
        INSERT INTO definitions (token, language, part_of_speech, meaning, examples, inflections, created_at, source)
        VALUES ($1, $2, $3, $4, $5, $6, now(), $7)
        RETURNING id, created_at, updated_at`

	wordDefinitionsInsert := `INSERT INTO word_definitions (word_id, definition_id) VALUES ($1, $2)`

	for i, def := range definitions {
		var createdAt pgtype.Timestamp
		var updatedAt pgtype.Timestamp

		// Execute the query within the transaction
		err := tx.QueryRow(context.Background(), definitionsInsert, def.Token, def.Language, def.PartOfSpeech, def.Meaning, def.Examples, def.Inflections, def.Source).Scan(&def.ID, &createdAt, &updatedAt)
		if err != nil {
			tx.Rollback(context.Background())
			log.Printf("Failed definitions insert: %v\n", err)
			return nil, err
		}

		// Update the definition object with the returned values
		def.CreatedAt = createdAt
		def.UpdatedAt = updatedAt
		definitions[i] = def

		_, err = tx.Exec(context.Background(), wordDefinitionsInsert, tokenId, def.ID)

		if err != nil {
			tx.Rollback(context.Background())
			log.Printf("Failed to insert into word_definition %v %v %v\n", tokenId, def.ID, err)
			return nil, err
		}

	}

	// Commit the transaction
	if err := tx.Commit(context.Background()); err != nil {
		log.Printf("Failed word_definitions insert: %v\n", err)
		return nil, err
	}

	return definitions, nil
}

func (repository *DefinitionRepository) getRandomMeanings(definitionIdsToIgnore []int, limit int) ([]string, error) {
	// all other defitions for the same word defined by the the records which ids are in definitionIdsToIgnore will be ignored too
	query := `
		SELECT meaning FROM definitions def
		JOIN word_definitions wd ON wd.definition_id = def.id
		WHERE wd.word_id NOT IN (select word_id FROM word_definitions WHERE definition_id = ANY($1)) 
		ORDER BY random() 
		LIMIT $2;
	`

	rows, err := repository.db.Query(context.Background(), query, definitionIdsToIgnore, limit)
	defer rows.Close()
	if err != nil {
		log.Printf("Failed to get random meanings: %v\n", err)
		return nil, err
	}

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
		log.Printf("Failed to get random examples: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var examples []string
	for rows.Next() {
		var example string
		err = rows.Scan(&example)
		if err != nil {
			log.Printf("Failed to scan results from random meanings: %v\n", err)
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
				token
			FROM 
				definitions def
			JOIN 
				word_definitions wd ON wd.definition_id = def.id
			WHERE 
				part_of_speech=$1 
				AND wd.word_id NOT IN (select word_id FROM word_definitions WHERE definition_id = ANY($2)) 
			ORDER BY random()
		)
		SELECT DISTINCT token FROM tokens LIMIT $3;
	`
	rows, err := repository.db.Query(context.Background(), query, partOfSpeech, definitionIdsToIgnore, limit)
	if err != nil {
		log.Printf("Failed to get random tokens: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		err = rows.Scan(&token)
		if err != nil {
			log.Printf("Failed to scan results from random tokens: %v\n", err)
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}
