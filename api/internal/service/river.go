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
) (*river.Client[pgx.Tx], error) {
	riverWorkers := river.NewWorkers()
	river.AddWorker(riverWorkers, NewImageGeneratorWorker(definitionService, definitionImageService, userService))
	river.AddWorker(riverWorkers, NewTextToSpeechWorker(wordService, definitionService, leitnerSystemStrategy, userService))
	river.AddWorker(riverWorkers, NewDefinitionFetcherWorker(db, wordService, definitionService, jobService, leitnerSystemStrategy, leitnerSystemStrategy.leitnerTrackingService, userService))
	river.AddWorker(riverWorkers, NewExampleAudioWorker(definitionService, wordService, userService))
	river.AddWorker(riverWorkers, &SubscriptionReminderWorker{
		db:          db,
		subRepo:     repository.NewSubscriptionRepository(db),
		userRepo:    &repository.UserRepository{Db: db},
		mailService: mailService,
	})
	river.AddWorker(riverWorkers, &NoOpWorker{})
	river.AddWorker(riverWorkers, NewRevenueCatWebhookWorker(revenueCatService))
	river.AddWorker(riverWorkers, NewStripeWebhookWorker(subscriptionService))

	// Create periodic jobs for renewal reminders
	periodicJobs := []*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(24*time.Hour), // Run daily
			func() (river.JobArgs, *river.InsertOpts) {
				// Schedule renewal reminders by checking subscriptions
				go func() {
					ctx := context.Background()
					if err := ScheduleRenewalReminders(ctx, db, mailService); err != nil {
						common.Logger.Error("Failed to schedule renewal reminders", "error", err)
					}
				}()
				// Return a no-op job
				return NoOpJobArgs{}, nil
			},
			&river.PeriodicJobOpts{
				RunOnStart: true, // Run immediately on startup
			},
		),
		// Materialized view refresh job removed - now using Redis caching for analytics
	}

	riverClient, err := river.NewClient(riverpgxv5.New(db), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault:        {MaxWorkers: 100},
			ImageGeneratorQueue:       {MaxWorkers: 5},
			TextToSpeechQueue:         {MaxWorkers: 30}, //max of 50 per openai docs
			DefinitionFetcherQueue:    {MaxWorkers: 50},
			SubscriptionReminderQueue: {MaxWorkers: 10},
			ExampleAudioQueue:         {MaxWorkers: 20},
			"revenuecat-webhook":      {MaxWorkers: 5},
			"stripe-webhook":          {MaxWorkers: 5},
		},
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
