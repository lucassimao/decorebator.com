package service

import (
	"context"
	"encoding/json"
	"strconv"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"github.com/riverqueue/river"
)

type RevenueCatWebhookArgs struct {
	Payload []byte `json:"payload"`
}

func (RevenueCatWebhookArgs) Kind() string { return "revenuecat-webhook" }

type RevenueCatWebhookWorker struct {
	river.WorkerDefaults[RevenueCatWebhookArgs]
	rcService RevenueCatService
}

func NewRevenueCatWebhookWorker(rcService RevenueCatService) *RevenueCatWebhookWorker {
	return &RevenueCatWebhookWorker{
		rcService: rcService,
	}
}

func (w *RevenueCatWebhookWorker) Work(ctx context.Context, job *river.Job[RevenueCatWebhookArgs]) error {
	logger := common.Logger.With("worker", "revenuecat-webhook")

	// Parse webhook payload
	var webhook model.RevenueCatWebhook
	if err := json.Unmarshal(job.Args.Payload, &webhook); err != nil {
		logger.Error("failed to unmarshal webhook", "error", err)
		return err
	}

	// Log webhook event
	logger.Info("RevenueCat webhook received",
		"event_type", webhook.Event.Type,
		"app_user_id", webhook.Event.AppUserID,
		"product_id", webhook.Event.ProductID,
		"entitlement_ids", webhook.Event.EntitlementIDs,
	)

	// Check if we've already processed this event
	exists, err := w.rcService.CheckEventExists(ctx, webhook.Event.ID)
	if err != nil {
		logger.Error("failed to check event existence", "error", err)
		return err
	}
	if exists {
		logger.Info("RevenueCat event already processed", "event_id", webhook.Event.ID)
		return nil
	}

	// Parse app_user_id as user ID (app_user_id is stringified user ID)
	userID, err := strconv.ParseInt(webhook.Event.AppUserID, 10, 64)
	if err != nil {
		logger.Error("failed to parse app_user_id as user ID",
			"app_user_id", webhook.Event.AppUserID,
			"error", err)
		// Store event without user ID for debugging
		if storeErr := w.rcService.StoreRevenueCatEvent(ctx, &webhook.Event, nil); storeErr != nil {
			logger.Error("failed to store event", "error", storeErr)
			return storeErr
		}
		return err
	}

	// Process event based on type
	if err := w.rcService.ProcessRevenueCatEvent(ctx, webhook.Event, userID); err != nil {
		logger.Error("failed to process event", "error", err)
		return err
	}

	// Store the event
	if err := w.rcService.StoreRevenueCatEvent(ctx, &webhook.Event, &userID); err != nil {
		logger.Error("failed to store event", "error", err)
		return err
	}

	return nil
}
