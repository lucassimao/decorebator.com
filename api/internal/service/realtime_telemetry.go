package service

import (
	"context"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RealtimeTelemetryService struct {
	repo *repository.RealtimeTelemetryRepository
}

func NewRealtimeTelemetryService(db *pgxpool.Pool) *RealtimeTelemetryService {
	return &RealtimeTelemetryService{
		repo: repository.NewRealtimeTelemetryRepository(db),
	}
}

func (s *RealtimeTelemetryService) StoreRealtimeTelemetry(ctx context.Context, telemetry *model.RealtimeChatTelemetry) error {
	if telemetry == nil {
		return nil
	}
	return s.repo.InsertRealtimeChatTelemetry(ctx, telemetry)
}
