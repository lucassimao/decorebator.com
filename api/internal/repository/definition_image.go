package repository

import (
	"context"
	"fmt"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateDefinitionImageDTO = model.CreateDefinitionImageDTO
type DefinitionImage = model.DefinitionImage

type DefinitionImageRepository struct {
	db *pgxpool.Pool
}

func NewDefinitionImageRepository(db *pgxpool.Pool) *DefinitionImageRepository {
	return &DefinitionImageRepository{db}
}

func (repository *DefinitionImageRepository) Save(ctx context.Context, dto CreateDefinitionImageDTO) (*DefinitionImage, error) {
	var def DefinitionImage

	// Start a transaction
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		err = &common.DatabaseError{
			Msg: "failed to open transaction",
			Err: err,
		}
		return nil, err
	}

	// handle rollback if needed
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				common.Logger.Error("failed to rollback transaction in definition image repository", "error", rollbackErr)
			}
		}
	}()

	// existing images will be hidden
	_, err = tx.Exec(ctx, "UPDATE definition_images SET is_visible=$1 WHERE definition_id=$2", false, dto.DefinitionId)

	if err != nil {
		err = &common.DatabaseError{
			Msg: fmt.Sprintf("failed to hide existing images for definition #%d", def.ID),
			Err: err,
		}
		return nil, err
	}

	insert := `
		INSERT INTO 
			definition_images (api,description, model, prompt,is_visible,definition_id, url)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	// Execute the query within the transaction
	err = tx.QueryRow(ctx, insert, dto.Api, dto.Description, dto.Model,
		dto.Prompt, true, dto.DefinitionId, dto.URL).
		Scan(&def.ID, &def.CreatedAt)

	if err != nil {
		err = &common.DatabaseError{
			Msg: "failed to insert definition",
			Err: err,
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		err = &common.DatabaseError{
			Msg: "failed to commit transaction",
			Err: err,
		}
		return nil, err
	}

	return &def, nil
}
