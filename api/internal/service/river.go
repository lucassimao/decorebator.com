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
const BACKFILL_INFLECTIONS_QUEUE = "backfill_inflections"

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

	riverWorkers := river.NewWorkers()
	river.AddWorker(riverWorkers, &ImageGeneratorWorker{})
	river.AddWorker(riverWorkers, &TextToSpeechWorker{})
	river.AddWorker(riverWorkers, &DefinitionFetcherWorker{})
	river.AddWorker(riverWorkers, &SubscriptionReminderWorker{
		db:       db,
		subRepo:  repository.NewSubscriptionRepository(db),
		userRepo: &repository.UserRepository{Db: db},
	})
	river.AddWorker(riverWorkers, &NoOpWorker{})

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
		// refresh materialized views
		river.NewPeriodicJob(
			river.PeriodicInterval(1*time.Hour), // Run hourly
			func() (river.JobArgs, *river.InsertOpts) {

				db, err := common.GetDBConnection()

				if err == nil {
					ctx := context.Background()
					db.Exec(ctx, `REFRESH MATERIALIZED VIEW CONCURRENTLY mv_word_mastery_current`)
					db.Exec(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_quiz_type_performance")
				} else {
					common.Logger.Error("Could not refresh materialized views", "error", err)
				}

				// Return a no-op job
				return NoOpJobArgs{}, nil
			},
			&river.PeriodicJobOpts{
				RunOnStart: false,
			},
		),
	}

	riverClient, err := river.NewClient(riverpgxv5.New(db), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault:          {MaxWorkers: 100},
			IMAGE_GENERATOR_QUEUE:       {MaxWorkers: 5},
			TEXT_TO_SPEECH_QUEUE:        {MaxWorkers: 30}, //max of 50 per openai docs
			DEFINITION_FETCHER_QUEUE:    {MaxWorkers: 50},
			SUBSCRIPTION_REMINDER_QUEUE: {MaxWorkers: 10},
			BACKFILL_INFLECTIONS_QUEUE:  {MaxWorkers: 1}, // Single worker to respect API rate limits
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
