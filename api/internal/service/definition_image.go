package service

import (
	"context"
	"fmt"
	"runtime/debug"

	"decorebator.com/internal/common"
	rep "decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DefinitionImageService handles definition image operations with dependency injection
type DefinitionImageService struct {
	definitionImageRepository *rep.DefinitionImageRepository
}

// NewDefinitionImageService creates a new DefinitionImageService with injected dependencies
func NewDefinitionImageService(db *pgxpool.Pool) *DefinitionImageService {
	return &DefinitionImageService{
		definitionImageRepository: rep.NewDefinitionImageRepository(db),
	}
}

func (s *DefinitionImageService) SaveDefinitionImage(ctx context.Context, dto rep.CreateDefinitionImageDTO) (*rep.DefinitionImage, error) {
	definitionImage, err := s.definitionImageRepository.Save(ctx, dto)
	if err != nil {
		msg := "failed to save definition image"
		common.Logger.Error(msg, "error", err, "stacktrace", string(debug.Stack()))
		return nil, fmt.Errorf(msg+": %w", err)
	}

	return definitionImage, nil
}

func (s *DefinitionImageService) SaveDefinitionImageTx(ctx context.Context, dto rep.CreateDefinitionImageDTO, tx pgx.Tx) (*rep.DefinitionImage, error) {
	definitionImage, err := s.definitionImageRepository.SaveTx(ctx, dto, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to save definition image: %w", err)
	}
	return definitionImage, nil
}
