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
	TriggerGenerateImageWorker(definitionID int64, userID *int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error)
	TriggerTextToSpeechWorker(wordID int64, userID *int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error)
	TriggerFetchDefinitionWorker(wordID int64, userID *int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error)
	QueueExampleAudioJob(definitionID int64, wordID int64, userID *int64, tx *pgx.Tx) error
	TriggerStripeWebhookWorker(eventID, eventType string, eventData []byte) (int64, error)
	TriggerRevenueCatWebhookWorker(eventType string, eventData []byte) (int64, error)
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

func (js *JobServiceImpl) TriggerGenerateImageWorker(definitionID int64, userID *int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error) {
	opts := river.InsertOpts{
		Queue: IMAGE_GENERATOR_QUEUE,
	}

	return js.triggerWorker(&opts, ImageGeneratorArgs{
		DefinitionId: definitionID,
		UserID:       userID,
		ErrorReport:  errorReport,
	}, tx)
}

func (js *JobServiceImpl) TriggerTextToSpeechWorker(wordID int64, userID *int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error) {
	opts := river.InsertOpts{
		Queue: TEXT_TO_SPEECH_QUEUE,
	}

	return js.triggerWorker(&opts, TextToSpeechArgs{
		WordId:      wordID,
		UserID:      userID,
		ErrorReport: errorReport,
	}, tx)
}

func (js *JobServiceImpl) TriggerFetchDefinitionWorker(wordID int64, userID *int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error) {
	opts := river.InsertOpts{
		Queue: DEFINITION_FETCHER_QUEUE,
	}

	return js.triggerWorker(&opts, DefinitionFetcherArgs{
		WordId:      wordID,
		UserID:      userID,
		ErrorReport: errorReport,
	}, tx)
}

func (js *JobServiceImpl) QueueExampleAudioJob(definitionID int64, wordID int64, userID *int64, tx *pgx.Tx) error {
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

func (js *JobServiceImpl) triggerWorker(opts *river.InsertOpts, args river.JobArgs, tx *pgx.Tx) (int64, error) {
	logger := common.Logger.With("func", "triggerWorker", "Kind", args.Kind())

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

func (js *JobServiceImpl) TriggerStripeWebhookWorker(eventID, eventType string, eventData []byte) (int64, error) {
	args := StripeWebhookArgs{
		EventID:   eventID,
		EventType: eventType,
		EventData: eventData,
	}
	opts := &river.InsertOpts{
		Queue: "stripe-webhook",
	}
	return js.triggerWorker(opts, args, nil)
}

func (js *JobServiceImpl) TriggerRevenueCatWebhookWorker(_ string, eventData []byte) (int64, error) {
	args := RevenueCatWebhookArgs{
		Payload: eventData,
	}
	opts := &river.InsertOpts{
		Queue: "revenuecat-webhook",
	}
	return js.triggerWorker(opts, args, nil)
}

func (js *JobServiceImpl) RetryJob(jobID int64) error {
	_, err := js.riverClient.JobRetry(context.Background(), jobID)
	return err
}
