package api

import (
	"context"
	"encoding/json"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateDefinitionImageDTO = model.CreateDefinitionImageDTO
type DefinitionImage = model.DefinitionImage

type DefinitionImageRepository struct {
	db *pgxpool.Pool
}

func (repository *DefinitionImageRepository) save(dto CreateDefinitionImageDTO) (DefinitionImage, error) {

	definitionsInsert := `
        INSERT INTO 
			definition_images (api,description, model, prompt,is_visible,definition_id, url)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at
	`

	var createdAt pgtype.Timestamp
	var id int64

	var def DefinitionImage

	// Start a transaction
	tx, err := repository.db.Begin(context.Background())
	if err != nil {
		return DefinitionImage{}, err
	}

	// existing images will be hidden
	_, err = tx.Exec(context.Background(), "UPDATE definition_images SET is_visible=$1 WHERE definition_id=$2", false, dto.DefinitionId)

	if err != nil {
		common.Logger.Error("failed to hide existing images", "definitionId", def.ID, "error", err)
		tx.Rollback(context.Background())
		return DefinitionImage{}, err
	}

	// Execute the query within the transaction
	err = tx.QueryRow(context.Background(), definitionsInsert, dto.Api, dto.Description, dto.Model,
		dto.Prompt, true, dto.DefinitionId, dto.URL).Scan(&def.ID, &createdAt)

	if err != nil {
		jsonString, _ := json.Marshal(dto)
		common.Logger.Error("failed to insert definition", "definitionImage", jsonString, "error", err)
		tx.Rollback(context.Background())
		return DefinitionImage{}, err
	}

	if err := tx.Commit(context.Background()); err != nil {
		return DefinitionImage{}, err
	}

	def.CreatedAt = createdAt
	def.ID = id

	return def, nil
}
