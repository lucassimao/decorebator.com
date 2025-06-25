package mocks

import (
	"context"

	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
)

// RevenueCatServiceMock is a mock implementation of the RevenueCatService interface.
type RevenueCatServiceMock struct {
	GetCustomerInfoFunc                          func(ctx context.Context, appUserID string, platform model.PlatformType) (*service.CustomerInfo, error)
	CreateOrUpdateSubscriptionFromRevenueCatFunc func(ctx context.Context, userID int64, customerInfo *service.CustomerInfo, platform model.PlatformType) error
	HandleWebhookFunc                            func(ctx context.Context, payload []byte, authHeader string) error
	RestorePurchasesFunc                         func(ctx context.Context, userID int64, appUserID string, platform model.PlatformType) error
	GetPlanFromProductIDFunc                     func(productID string) model.SubscriptionPlan
}

func (m *RevenueCatServiceMock) GetCustomerInfo(ctx context.Context, appUserID string, platform model.PlatformType) (*service.CustomerInfo, error) {
	return m.GetCustomerInfoFunc(ctx, appUserID, platform)
}

func (m *RevenueCatServiceMock) CreateOrUpdateSubscriptionFromRevenueCat(ctx context.Context, userID int64, customerInfo *service.CustomerInfo, platform model.PlatformType) error {
	return m.CreateOrUpdateSubscriptionFromRevenueCatFunc(ctx, userID, customerInfo, platform)
}

func (m *RevenueCatServiceMock) HandleWebhook(ctx context.Context, payload []byte, authHeader string) error {
	return m.HandleWebhookFunc(ctx, payload, authHeader)
}

func (m *RevenueCatServiceMock) RestorePurchases(ctx context.Context, userID int64, appUserID string, platform model.PlatformType) error {
	return m.RestorePurchasesFunc(ctx, userID, appUserID, platform)
}

func (m *RevenueCatServiceMock) GetPlanFromProductID(productID string) model.SubscriptionPlan {
	return m.GetPlanFromProductIDFunc(productID)
}
