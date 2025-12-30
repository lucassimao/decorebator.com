package service

import (
	"context"
	"time"

	"decorebator.com/internal/common"
	"github.com/riverqueue/river"
)

type PushReceiptArgs struct{}

func (PushReceiptArgs) Kind() string { return "push_receipt_check" }

type PushReceiptWorker struct {
	river.WorkerDefaults[PushReceiptArgs]
	pushService *PushNotificationService
}

func NewPushReceiptWorker(pushService *PushNotificationService) *PushReceiptWorker {
	return &PushReceiptWorker{pushService: pushService}
}

func (w *PushReceiptWorker) Work(ctx context.Context, _ *river.Job[PushReceiptArgs]) error {
	if w.pushService == nil {
		return nil
	}
	if err := w.pushService.CheckReceipts(ctx, time.Now().UTC()); err != nil {
		common.Logger.Error("failed to check push receipts", "error", err)
		return err
	}
	return nil
}
