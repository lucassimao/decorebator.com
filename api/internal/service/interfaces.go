package service

import (
	"context"

	"decorebator.com/internal/model"
)

// RevenueCatAPIClient defines the interface for making external API calls to RevenueCat.
// This interface is extracted to allow for easy mocking in tests while keeping business logic testable.
type RevenueCatAPIClient interface {
	GetCustomerInfo(ctx context.Context, appUserID string) (*CustomerInfo, error)
}

// RevenueCatService defines the interface for interacting with the RevenueCat API.
type RevenueCatService interface {
	RestorePurchases(ctx context.Context, userID int64, appUserID string, platform model.PlatformType) error
	ProcessRevenueCatEvent(ctx context.Context, event model.RevenueCatEvent, userID int64) error
	CheckEventExists(ctx context.Context, eventID string) (bool, error)
	StoreRevenueCatEvent(ctx context.Context, event *model.RevenueCatEvent, userID *int64) error
	HasPremiumSubscription(ctx context.Context, appUserID string) (bool, error)
}
