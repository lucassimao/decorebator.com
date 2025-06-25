package service

import (
	"context"

	"decorebator.com/internal/common"
	"github.com/riverqueue/river"
)

type RevenueCatWebhookArgs struct {
	Payload    []byte `json:"payload"`
	AuthHeader string `json:"auth_header"`
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

	err := w.rcService.HandleWebhook(ctx, job.Args.Payload, job.Args.AuthHeader)
	if err != nil {
		logger.Error("failed to process revenuecat webhook", "error", err)
		return err
	}

	return nil
}
