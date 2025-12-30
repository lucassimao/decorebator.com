package service

import (
	"context"
	"time"

	"decorebator.com/internal/common"
	"github.com/riverqueue/river"
)

type DailyPracticeReminderArgs struct{}

func (DailyPracticeReminderArgs) Kind() string { return "daily_practice_reminder" }

type DailyPracticeReminderWorker struct {
	river.WorkerDefaults[DailyPracticeReminderArgs]
	pushService *PushNotificationService
}

func NewDailyPracticeReminderWorker(pushService *PushNotificationService) *DailyPracticeReminderWorker {
	return &DailyPracticeReminderWorker{pushService: pushService}
}

func (w *DailyPracticeReminderWorker) Work(ctx context.Context, _ *river.Job[DailyPracticeReminderArgs]) error {
	if w.pushService == nil {
		return nil
	}
	if err := w.pushService.SendDailyPracticeReminders(ctx, time.Now().UTC()); err != nil {
		common.Logger.Error("failed to send daily practice reminders", "error", err)
		return err
	}
	return nil
}
