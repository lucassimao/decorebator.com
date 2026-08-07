package repository

import (
	"context"
	"fmt"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5"
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
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return nil, &common.DatabaseError{
			Msg: "failed to open transaction",
			Err: err,
		}
	}
	defer common.RollbackTx(ctx, tx, "definition image save")

	definitionImage, err := repository.SaveTx(ctx, dto, tx)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, &common.DatabaseError{Msg: "failed to commit transaction", Err: err}
	}
	return definitionImage, nil
}

// SaveTx hides prior images and inserts the replacement in the caller's
// transaction so report resolution can commit atomically with the new image.
func (repository *DefinitionImageRepository) SaveTx(ctx context.Context, dto CreateDefinitionImageDTO, tx pgx.Tx) (*DefinitionImage, error) {
	var def DefinitionImage

	// existing images will be hidden
	_, err := tx.Exec(ctx, "UPDATE definition_images SET is_visible=$1 WHERE definition_id=$2", false, dto.DefinitionID)

	if err != nil {
		err = &common.DatabaseError{
			Msg: fmt.Sprintf("failed to hide existing images for definition #%d", dto.DefinitionID),
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
	err = tx.QueryRow(ctx, insert, dto.API, dto.Description, dto.Model,
		dto.Prompt, true, dto.DefinitionID, dto.URL).
		Scan(&def.ID, &def.CreatedAt)

	if err != nil {
		err = &common.DatabaseError{
			Msg: "failed to insert definition",
			Err: err,
		}
		return nil, err
	}

	return &def, nil
}
