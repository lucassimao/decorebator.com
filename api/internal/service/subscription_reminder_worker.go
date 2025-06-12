package service

import (
	"context"
	"fmt"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/mail"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

// SubscriptionReminderArgs defines the args for subscription reminder job
type SubscriptionReminderArgs struct {
	SubscriptionID int64 `json:"subscription_id"`
}

// Kind returns the kind string for this job
func (SubscriptionReminderArgs) Kind() string { return "subscription_reminder" }

// SubscriptionReminderWorker sends renewal reminder emails
type SubscriptionReminderWorker struct {
	river.WorkerDefaults[SubscriptionReminderArgs]
	db       *pgxpool.Pool
	subRepo  *repository.SubscriptionRepository
	userRepo *repository.UserRepository
}

// Work processes the subscription reminder job
func (w *SubscriptionReminderWorker) Work(ctx context.Context, job *river.Job[SubscriptionReminderArgs]) error {
	logger := common.Logger.With("worker", "SubscriptionReminderWorker", "subscription_id", job.Args.SubscriptionID)
	logger.Info("Processing subscription reminder")

	// Get subscription
	sub, err := w.subRepo.GetSubscriptionByID(ctx, job.Args.SubscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}
	if sub == nil {
		return fmt.Errorf("subscription not found: %d", job.Args.SubscriptionID)
	}

	// Skip if subscription is not active
	if sub.Status != model.StatusActive {
		logger.Info("Skipping reminder for non-active subscription")
		return nil
	}

	// Get user
	users, err := w.userRepo.Find(repository.FindUserArgs{ID: &sub.UserID})
	if err != nil || len(users) == 0 {
		return fmt.Errorf("failed to get user: %w", err)
	}
	user := &users[0]

	// Send reminder email
	emailData := mail.SubscriptionEmailData{
		PlanName:        string(sub.Plan),
		AmountCents:     sub.AmountCents,
		Currency:        sub.Currency,
		NextBillingDate: sub.CurrentPeriodEnd,
		SubscriptionID:  sub.StripeSubscriptionID,
	}

	if err := mail.SendRenewalReminderEmail(user, emailData); err != nil {
		return fmt.Errorf("failed to send renewal reminder: %w", err)
	}

	logger.Info("Renewal reminder sent successfully")
	return nil
}

// ScheduleRenewalReminders schedules reminder emails for subscriptions renewing soon
func ScheduleRenewalReminders(ctx context.Context, db *pgxpool.Pool) error {
	logger := common.Logger.With("func", "ScheduleRenewalReminders")
	logger.Info("Running subscription renewal scheduler")

	subRepo := repository.NewSubscriptionRepository(db)
	userRepo := &repository.UserRepository{Db: db}

	// Get subscriptions renewing in the next 3-4 days
	reminderWindow := time.Now().AddDate(0, 0, 3)
	reminderWindowEnd := time.Now().AddDate(0, 0, 4)

	subscriptions, err := subRepo.GetSubscriptionsRenewingBetween(ctx, reminderWindow, reminderWindowEnd)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions: %w", err)
	}

	// Process reminders directly
	for _, sub := range subscriptions {
		// Check context cancellation
		select {
		case <-ctx.Done():
			logger.Warn("Context cancelled, stopping reminder processing")
			return ctx.Err()
		default:
		}

		// Check if reminder already sent
		alreadySent, err := subRepo.HasSentRenewalReminder(ctx, sub.ID, reminderWindow)
		if err != nil {
			logger.Error("Failed to check reminder status", "error", err, "subscription_id", sub.ID)
			continue
		}
		if alreadySent {
			continue
		}

		// Get user
		users, err := userRepo.Find(repository.FindUserArgs{ID: &sub.UserID})
		if err != nil || len(users) == 0 {
			logger.Error("Failed to get user", "error", err, "subscription_id", sub.ID)
			continue
		}
		user := &users[0]

		// Send reminder email
		emailData := mail.SubscriptionEmailData{
			PlanName:        string(sub.Plan),
			AmountCents:     sub.AmountCents,
			Currency:        sub.Currency,
			NextBillingDate: sub.CurrentPeriodEnd,
			SubscriptionID:  sub.StripeSubscriptionID,
		}

		if err := mail.SendRenewalReminderEmail(user, emailData); err != nil {
			logger.Error("Failed to send renewal reminder", "error", err, "subscription_id", sub.ID)
			continue
		}

		// Mark reminder as sent
		if err := subRepo.MarkRenewalReminderSent(ctx, sub.ID); err != nil {
			logger.Error("Failed to mark reminder sent", "error", err, "subscription_id", sub.ID)
			continue
		}

		logger.Info("Sent renewal reminder", "subscription_id", sub.ID)
	}

	logger.Info("Subscription renewal scheduler completed", "reminders_sent", len(subscriptions))
	return nil
}
