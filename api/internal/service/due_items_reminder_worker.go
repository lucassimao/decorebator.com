package service

import (
	"context"
	"time"

	"decorebator.com/internal/common"
	"github.com/riverqueue/river"
)

type DueItemsReminderArgs struct{}

func (DueItemsReminderArgs) Kind() string { return "due_items_reminder" }

type DueItemsReminderWorker struct {
	river.WorkerDefaults[DueItemsReminderArgs]
	pushService *PushNotificationService
}

func NewDueItemsReminderWorker(pushService *PushNotificationService) *DueItemsReminderWorker {
	return &DueItemsReminderWorker{pushService: pushService}
}

func (w *DueItemsReminderWorker) Work(ctx context.Context, _ *river.Job[DueItemsReminderArgs]) error {
	if w.pushService == nil {
		return nil
	}
	if err := w.pushService.SendDueItemsReminders(ctx, time.Now().UTC()); err != nil {
		common.Logger.ErrorContext(ctx, "failed to send due items reminders", "error", err)
		return err
	}
	return nil
}
