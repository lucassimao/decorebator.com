package service

import (
	"context"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/mail"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"
)

const ImageGeneratorQueue = "image_generator"
const TextToSpeechQueue = "text_to_speech"
const DefinitionFetcherQueue = "definition_fetcher"
const SubscriptionReminderQueue = "subscription_reminder"
const ExampleAudioQueue = "example_audio"
const MeaningAudioQueue = "meaning_audio"
const PushNotificationQueue = "push_notification"

// WorkerDatabaseMaxConnections is the worker process's database budget. Keep
// definition work below this value and preserve the combined definition plus
// meaning-audio provider budget until production queue and pool data justify an
// increase.
const WorkerDatabaseMaxConnections int32 = 10
const WorkerDatabaseMinConnections int32 = 2
const DefinitionFetcherQueueMaxWorkers = 2
const MeaningAudioQueueMaxWorkers = 2

func workerQueueConfig(legacyProviderSurfaceEnabled bool) map[string]river.QueueConfig {
	queues := map[string]river.QueueConfig{
		river.QueueDefault:              {MaxWorkers: 100},
		ImageGeneratorQueue:             {MaxWorkers: 5},
		TextToSpeechQueue:               {MaxWorkers: 30},
		DefinitionFetcherQueue:          {MaxWorkers: DefinitionFetcherQueueMaxWorkers},
		SubscriptionReminderQueue:       {MaxWorkers: 10},
		ExampleAudioQueue:               {MaxWorkers: 20},
		MeaningAudioQueue:               {MaxWorkers: MeaningAudioQueueMaxWorkers},
		PushNotificationQueue:           {MaxWorkers: 5},
		GoogleAcknowledgementRetryQueue: {MaxWorkers: 5},
	}
	if legacyProviderSurfaceEnabled {
		queues["revenuecat-webhook"] = river.QueueConfig{MaxWorkers: 5}
		queues["stripe-webhook"] = river.QueueConfig{MaxWorkers: 5}
	}
	return queues
}

// NoOpJobArgs is a no-op job used for periodic jobs that execute inline
type NoOpJobArgs struct{}

func (NoOpJobArgs) Kind() string { return "noop" }

// NoOpWorker is a worker that does nothing
type NoOpWorker struct {
	river.WorkerDefaults[NoOpJobArgs]
}

func (w *NoOpWorker) Work(ctx context.Context, job *river.Job[NoOpJobArgs]) error {
	return nil
}

// NewWorkerRiverClient creates a River client for worker processes using individual services
// This eliminates circular dependencies by accepting services as parameters
func NewWorkerRiverClient(
	db *pgxpool.Pool,
	definitionService *DefinitionService,
	definitionImageService *DefinitionImageService,
	wordService *WordService,
	userService *UserService,
	leitnerSystemStrategy *LeitnerSystemStrategy,
	jobService JobService,
	revenueCatService RevenueCatService,
	subscriptionService *SubscriptionService,
	mailService *mail.MailService,
	providerInbox *repository.ProviderEventInboxRepository,
	googleAcknowledgementWorker *GoogleAcknowledgementRetryWorker,
	googleAcknowledgementSweep *GoogleAcknowledgementSweepWorker,
	legacyProviderSurfaceEnabled bool,
) (*river.Client[pgx.Tx], error) {
	riverWorkers := river.NewWorkers()
	river.AddWorker(riverWorkers, NewImageGeneratorWorker(
		definitionService, definitionImageService, userService, leitnerSystemStrategy,
	))
	river.AddWorker(riverWorkers, NewTextToSpeechWorker(wordService, definitionService, leitnerSystemStrategy, userService))
	river.AddWorker(riverWorkers, NewDefinitionFetcherWorker(db, wordService, definitionService, jobService, leitnerSystemStrategy, leitnerSystemStrategy.leitnerTrackingService, userService))
	river.AddWorker(riverWorkers, NewExampleAudioWorker(definitionService, wordService, userService))
	river.AddWorker(riverWorkers, NewMeaningAudioWorker(definitionService, wordService, userService))
	river.AddWorker(riverWorkers, &SubscriptionReminderWorker{
		db:          db,
		subRepo:     repository.NewSubscriptionRepository(db),
		userRepo:    &repository.UserRepository{Db: db},
		mailService: mailService,
	})
	pushService := NewPushNotificationService(
		&repository.PushTokenRepository{Db: db},
		&repository.PushNotificationRepository{Db: db},
		&repository.PushNotificationEventRepository{Db: db},
		&repository.PushReceiptRepository{Db: db},
	)
	// Register worker handlers; periodic jobs enqueue args separately.
	river.AddWorker(riverWorkers, NewDueItemsReminderWorker(pushService))
	river.AddWorker(riverWorkers, NewDailyPracticeReminderWorker(pushService))
	river.AddWorker(riverWorkers, NewPushReceiptWorker(pushService))
	river.AddWorker(riverWorkers, &NoOpWorker{})
	if legacyProviderSurfaceEnabled {
		river.AddWorker(riverWorkers, NewRevenueCatWebhookWorker(revenueCatService))
		river.AddWorker(riverWorkers, NewStripeWebhookWorker(subscriptionService))
	}
	if providerInbox != nil {
		healthWorker, err := NewProviderEventInboxHealthWorker(
			providerInbox, nil, providerInboxHealthGracePeriod,
		)
		if err != nil {
			return nil, err
		}
		river.AddWorker(riverWorkers, healthWorker)
	}
	if googleAcknowledgementWorker != nil {
		river.AddWorker(riverWorkers, googleAcknowledgementWorker)
	}
	if googleAcknowledgementSweep != nil {
		river.AddWorker(riverWorkers, googleAcknowledgementSweep)
	}

	// Create periodic jobs for renewal reminders
	periodicJobs := []*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(24*time.Hour), // Run daily
			func() (river.JobArgs, *river.InsertOpts) {
				// Schedule renewal reminders by checking subscriptions
				go func() {
					ctx := context.Background()
					if err := ScheduleRenewalReminders(ctx, db, mailService); err != nil {
						common.Logger.ErrorContext(ctx, "Failed to schedule renewal reminders", "error", err)
					}
				}()
				// Return a no-op job
				return NoOpJobArgs{}, nil
			},
			&river.PeriodicJobOpts{
				RunOnStart: true, // Run immediately on startup
			},
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(15*time.Minute),
			func() (river.JobArgs, *river.InsertOpts) {
				return DueItemsReminderArgs{}, &river.InsertOpts{Queue: PushNotificationQueue}
			},
			&river.PeriodicJobOpts{
				RunOnStart: true,
			},
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(15*time.Minute),
			func() (river.JobArgs, *river.InsertOpts) {
				return DailyPracticeReminderArgs{}, &river.InsertOpts{Queue: PushNotificationQueue}
			},
			&river.PeriodicJobOpts{
				RunOnStart: true,
			},
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(1*time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return PushReceiptArgs{}, &river.InsertOpts{Queue: PushNotificationQueue}
			},
			&river.PeriodicJobOpts{
				RunOnStart: true,
			},
		),
		// Materialized view refresh job removed - now using Redis caching for analytics
	}
	if providerInbox != nil {
		periodicJobs = append(periodicJobs, river.NewPeriodicJob(
			river.PeriodicInterval(5*time.Minute),
			func() (river.JobArgs, *river.InsertOpts) {
				return ProviderEventInboxHealthArgs{}, nil
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		))
	}
	if googleAcknowledgementSweep != nil {
		periodicJobs = append(periodicJobs, river.NewPeriodicJob(
			river.PeriodicInterval(5*time.Minute),
			func() (river.JobArgs, *river.InsertOpts) {
				return GoogleAcknowledgementSweepArgs{}, &river.InsertOpts{Queue: GoogleAcknowledgementRetryQueue}
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		))
	}

	riverClient, err := river.NewClient(riverpgxv5.New(db), &river.Config{
		Queues:       workerQueueConfig(legacyProviderSurfaceEnabled),
		Workers:      riverWorkers,
		Logger:       common.Logger,
		PeriodicJobs: periodicJobs,
		ErrorHandler: NewRiverErrorHandler(),
		Middleware: []rivertype.Middleware{
			common.NewWorkerContextMiddleware(),
		},
	})
	if err != nil {
		return nil, err
	}

	return riverClient, nil
}
