package service

import (
	"context"
	"errors"

	"decorebator.com/internal/common"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// JobService provides abstraction for job insertion operations
type JobService interface {
	ScheduleImageJob(definitionID int64, userID *int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error)
	ScheduleAudioJob(wordID int64, userID *int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error)
	ScheduleDefinitionJob(wordID int64, userID *int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error)
	ScheduleExampleAudioJob(definitionID int64, wordID int64, userID *int64, tx *pgx.Tx) error
	ScheduleStripeWebhookJob(eventID, eventType string, eventData []byte) (int64, error)
	ScheduleRevenueCatWebhookJob(eventType string, eventData []byte) (int64, error)
	RetryJob(jobID int64) error
}

// JobServiceImpl implements JobService using River client
type JobServiceImpl struct {
	riverClient *river.Client[pgx.Tx]
}

// NewJobService creates a new job service with River client
func NewJobService(riverClient *river.Client[pgx.Tx]) *JobServiceImpl {
	return &JobServiceImpl{
		riverClient: riverClient,
	}
}

func (js *JobServiceImpl) ScheduleImageJob(definitionID int64, userID *int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error) {
	opts := river.InsertOpts{
		Queue: IMAGE_GENERATOR_QUEUE,
	}

	return js.enqueueJob(&opts, ImageGeneratorArgs{
		DefinitionId: definitionID,
		UserID:       userID,
		ErrorReport:  errorReport,
	}, tx)
}

func (js *JobServiceImpl) ScheduleAudioJob(wordID int64, userID *int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error) {
	opts := river.InsertOpts{
		Queue: TEXT_TO_SPEECH_QUEUE,
	}

	return js.enqueueJob(&opts, TextToSpeechArgs{
		WordId:      wordID,
		UserID:      userID,
		ErrorReport: errorReport,
	}, tx)
}

func (js *JobServiceImpl) ScheduleDefinitionJob(wordID int64, userID *int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error) {
	opts := river.InsertOpts{
		Queue: DEFINITION_FETCHER_QUEUE,
	}

	return js.enqueueJob(&opts, DefinitionFetcherArgs{
		WordId:      wordID,
		UserID:      userID,
		ErrorReport: errorReport,
	}, tx)
}

func (js *JobServiceImpl) ScheduleExampleAudioJob(definitionID int64, wordID int64, userID *int64, tx *pgx.Tx) error {
	args := ExampleAudioArgs{
		DefinitionID: definitionID,
		WordID:       wordID,
		UserID:       userID,
	}

	opts := &river.InsertOpts{
		Queue: EXAMPLE_AUDIO_QUEUE,
	}

	if tx != nil {
		_, err := js.riverClient.InsertTx(context.Background(), *tx, args, opts)
		return err
	}
	_, err := js.riverClient.Insert(context.Background(), args, opts)
	return err
}

func (js *JobServiceImpl) enqueueJob(opts *river.InsertOpts, args river.JobArgs, tx *pgx.Tx) (int64, error) {
	logger := common.Logger.With("func", "enqueueJob", "Kind", args.Kind())

	var result *rivertype.JobInsertResult
	var err error

	if tx != nil {
		result, err = js.riverClient.InsertTx(context.Background(), *tx, args, opts)
	} else {
		result, err = js.riverClient.Insert(context.Background(), args, opts)
	}

	if err != nil {
		logger.Error("failed to trigger river job", "error", err)
		return -1, errors.New("could not insert job")
	}

	return result.Job.ID, nil
}

func (js *JobServiceImpl) ScheduleStripeWebhookJob(eventID, eventType string, eventData []byte) (int64, error) {
	args := StripeWebhookArgs{
		EventID:   eventID,
		EventType: eventType,
		EventData: eventData,
	}
	opts := &river.InsertOpts{
		Queue: "stripe-webhook",
	}
	return js.enqueueJob(opts, args, nil)
}

func (js *JobServiceImpl) ScheduleRevenueCatWebhookJob(_ string, eventData []byte) (int64, error) {
	args := RevenueCatWebhookArgs{
		Payload: eventData,
	}
	opts := &river.InsertOpts{
		Queue: "revenuecat-webhook",
	}
	return js.enqueueJob(opts, args, nil)
}

func (js *JobServiceImpl) RetryJob(jobID int64) error {
	_, err := js.riverClient.JobRetry(context.Background(), jobID)
	return err
}
