package service

import (
	"context"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

const IMAGE_GENERATOR_QUEUE = "image_generator"
const TEXT_TO_SPEECH_QUEUE = "text_to_speech"
const DEFINITION_FETCHER_QUEUE = "definition_fetcher"
const SUBSCRIPTION_REMINDER_QUEUE = "subscription_reminder"
const EXAMPLE_AUDIO_QUEUE = "example_audio"

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

func GetRiverClient() (*river.Client[pgx.Tx], error) {
	db, err := common.GetDBConnection()

	if err != nil {
		return nil, err
	}

	// Create word service for workers
	wordService := NewWordService(db, NewOpenAIModerationService())

	riverWorkers := river.NewWorkers()
	river.AddWorker(riverWorkers, &ImageGeneratorWorker{})
	river.AddWorker(riverWorkers, NewTextToSpeechWorker(wordService))
	river.AddWorker(riverWorkers, NewDefinitionFetcherWorker(wordService))
	river.AddWorker(riverWorkers, &ExampleAudioWorker{})
	river.AddWorker(riverWorkers, &SubscriptionReminderWorker{
		db:       db,
		subRepo:  repository.NewSubscriptionRepository(db),
		userRepo: &repository.UserRepository{Db: db},
	})
	river.AddWorker(riverWorkers, &NoOpWorker{})
	apiClient := NewRevenueCatAPIClient()
	river.AddWorker(riverWorkers, NewRevenueCatWebhookWorker(NewRevenueCatService(db, apiClient)))

	// Create periodic jobs for renewal reminders
	periodicJobs := []*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(24*time.Hour), // Run daily
			func() (river.JobArgs, *river.InsertOpts) {
				// Schedule renewal reminders by checking subscriptions
				go func() {
					ctx := context.Background()
					if err := ScheduleRenewalReminders(ctx, db); err != nil {
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
			river.QueueDefault:          {MaxWorkers: 100},
			IMAGE_GENERATOR_QUEUE:       {MaxWorkers: 5},
			TEXT_TO_SPEECH_QUEUE:        {MaxWorkers: 30}, //max of 50 per openai docs
			DEFINITION_FETCHER_QUEUE:    {MaxWorkers: 50},
			SUBSCRIPTION_REMINDER_QUEUE: {MaxWorkers: 10},
			EXAMPLE_AUDIO_QUEUE:         {MaxWorkers: 20},
			"revenuecat-webhook":        {MaxWorkers: 5},
		},
		Workers:      riverWorkers,
		Logger:       common.Logger,
		PeriodicJobs: periodicJobs,
	})
	if err != nil {
		return nil, err
	}

	return riverClient, nil
}
