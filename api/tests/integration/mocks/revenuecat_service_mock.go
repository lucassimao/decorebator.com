package mocks

import (
	"context"

	"decorebator.com/internal/model"
)

// MockRevenueCatService is a mock implementation of the RevenueCatService interface
type MockRevenueCatService struct {
	RestorePurchasesFunc              func(ctx context.Context, userID int64, appUserID string, platform model.PlatformType) error
	ProcessRevenueCatEventFunc        func(ctx context.Context, event model.RevenueCatEvent, userID int64) error
	CheckEventExistsFunc              func(ctx context.Context, eventID string) (bool, error)
	StoreRevenueCatEventFunc          func(ctx context.Context, event *model.RevenueCatEvent, userID *int64) error
	GetUserByRevenueCatCustomerIDFunc func(ctx context.Context, appUserID string) (*model.User, error)
}

func (m *MockRevenueCatService) RestorePurchases(ctx context.Context, userID int64, appUserID string, platform model.PlatformType) error {
	if m.RestorePurchasesFunc != nil {
		return m.RestorePurchasesFunc(ctx, userID, appUserID, platform)
	}
	return nil
}

func (m *MockRevenueCatService) ProcessRevenueCatEvent(ctx context.Context, event model.RevenueCatEvent, userID int64) error {
	if m.ProcessRevenueCatEventFunc != nil {
		return m.ProcessRevenueCatEventFunc(ctx, event, userID)
	}
	return nil
}

func (m *MockRevenueCatService) CheckEventExists(ctx context.Context, eventID string) (bool, error) {
	if m.CheckEventExistsFunc != nil {
		return m.CheckEventExistsFunc(ctx, eventID)
	}
	return false, nil
}

func (m *MockRevenueCatService) StoreRevenueCatEvent(ctx context.Context, event *model.RevenueCatEvent, userID *int64) error {
	if m.StoreRevenueCatEventFunc != nil {
		return m.StoreRevenueCatEventFunc(ctx, event, userID)
	}
	return nil
}

func (m *MockRevenueCatService) GetUserByRevenueCatCustomerID(ctx context.Context, appUserID string) (*model.User, error) {
	if m.GetUserByRevenueCatCustomerIDFunc != nil {
		return m.GetUserByRevenueCatCustomerIDFunc(ctx, appUserID)
	}
	return nil, nil
}
