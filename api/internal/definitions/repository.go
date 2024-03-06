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
        INSERT INTO definitions (token, language, part_of_speech, meaning, examples, inflections, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, now())
        RETURNING id, created_at, updated_at`

	wordDefinitionsInsert := `INSERT INTO word_definitions (word_id, definition_id) VALUES ($1, $2)`

	for i, def := range definitions {
		var createdAt pgtype.Timestamp
		var updatedAt pgtype.Timestamp

		// Execute the query within the transaction
		err := tx.QueryRow(context.Background(), definitionsInsert, def.Token, def.Language, def.PartOfSpeech, def.Meaning, def.Examples, def.Inflections).Scan(&def.ID, &createdAt, &updatedAt)
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
