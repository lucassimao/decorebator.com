package service

import (
	"context"
	"fmt"
	"time"

	"decorebator.com/internal/model"
)

type StorePremiumAccessReader interface {
	HasEffectivePremiumAccess(context.Context, int64, model.StoreEnvironment, time.Time) (bool, error)
}

// EffectiveAccessService keeps store access additive while legacy providers
// coexist. Revoked store state cannot revoke a still-valid legacy plan.
type EffectiveAccessService struct {
	reader      StorePremiumAccessReader
	environment model.StoreEnvironment
	now         func() time.Time
}

func NewEffectiveAccessService(
	reader StorePremiumAccessReader,
	environment model.StoreEnvironment,
	now func() time.Time,
) (*EffectiveAccessService, error) {
	if reader == nil || !environment.Valid() {
		return nil, fmt.Errorf("effective store access dependencies are required")
	}
	if now == nil {
		now = time.Now
	}
	return &EffectiveAccessService{reader: reader, environment: environment, now: now}, nil
}

func (s *EffectiveAccessService) Plan(
	ctx context.Context,
	userID int64,
	legacy model.SubscriptionPlan,
) (model.SubscriptionPlan, error) {
	if legacy == model.PlanMonthly || legacy == model.PlanAnnual {
		return legacy, nil
	}
	grants, err := s.reader.HasEffectivePremiumAccess(ctx, userID, s.environment, s.now().UTC())
	if err != nil {
		return model.PlanFree, fmt.Errorf("read effective store access: %w", err)
	}
	if grants {
		// The legacy enum has no store-neutral premium value. This is a
		// request/response-only compatibility projection and is never persisted.
		return model.PlanMonthly, nil
	}
	return model.PlanFree, nil
}

func (s *EffectiveAccessService) HasPremium(ctx context.Context, userID int64) (bool, error) {
	plan, err := s.Plan(ctx, userID, model.PlanFree)
	return plan != model.PlanFree, err
}
