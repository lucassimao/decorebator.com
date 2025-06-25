package service

import (
	"context"

	"decorebator.com/internal/model"
)

// RevenueCatService defines the interface for interacting with the RevenueCat API.
type RevenueCatService interface {
	GetCustomerInfo(ctx context.Context, appUserID string) (*CustomerInfo, error)
	CreateOrUpdateSubscriptionFromRevenueCat(ctx context.Context, userID int64, customerInfo *CustomerInfo, platform model.PlatformType) error
	HandleWebhook(ctx context.Context, payload []byte) error
	RestorePurchases(ctx context.Context, userID int64, appUserID string, platform model.PlatformType) error
	GetPlanFromProductID(productID string) model.SubscriptionPlan
}
