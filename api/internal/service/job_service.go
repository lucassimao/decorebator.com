package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"decorebator.com/internal/common"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// JobService provides abstraction for job insertion operations
type JobService interface {
	ScheduleImageJob(ctx context.Context, definitionID int64, userID *int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error)
	ScheduleAudioJob(ctx context.Context, wordID int64, userID *int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error)
	ScheduleDefinitionJob(ctx context.Context, wordID int64, userID *int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error)
	ScheduleExampleAudioJob(ctx context.Context, definitionID int64, wordID int64, userID *int64, tx *pgx.Tx) error
	ScheduleMeaningAudioJob(ctx context.Context, definitionID int64, wordID int64, userID *int64, tx *pgx.Tx) error
	ScheduleAccountCleanupJob(ctx context.Context, profileObjectPrefix string, tx *pgx.Tx) error
	ScheduleProfileUploadReconciliationJob(ctx context.Context, args ProfileUploadReconciliationArgs) (int64, error)
	FinalizeProfileUploadReconciliationJob(ctx context.Context, jobID int64, args ProfileUploadReconciliationArgs, transactions ...pgx.Tx) error
	ScheduleResetPasswordEmailJob(ctx context.Context, email string, transactions ...pgx.Tx) error
	ScheduleStripeWebhookJob(ctx context.Context, eventID, eventType string, eventData []byte) (int64, error)
	ScheduleRevenueCatWebhookJob(ctx context.Context, eventType string, eventData []byte) (int64, error)
	RetryJob(ctx context.Context, jobID int64) error
}

func (js *JobServiceImpl) ScheduleResetPasswordEmailJob(
	ctx context.Context,
	email string,
	transactions ...pgx.Tx,
) error {
	var tx *pgx.Tx
	if len(transactions) > 0 {
		tx = &transactions[0]
	}
	_, err := js.enqueueJob(ctx, &river.InsertOpts{
		Queue:       ResetPasswordEmailQueue,
		MaxAttempts: 5,
	}, ResetPasswordEmailArgs{Email: email}, tx)
	return err
}

func (js *JobServiceImpl) ScheduleProfileUploadReconciliationJob(ctx context.Context, args ProfileUploadReconciliationArgs) (int64, error) {
	return js.enqueueJob(ctx, &river.InsertOpts{
		Queue:       AccountCleanupQueue,
		MaxAttempts: 25,
		ScheduledAt: time.Now().Add(5 * time.Minute),
	}, args, nil)
}

func (js *JobServiceImpl) FinalizeProfileUploadReconciliationJob(ctx context.Context, jobID int64, args ProfileUploadReconciliationArgs, transactions ...pgx.Tx) error {
	encoded, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("encode profile upload reconciliation: %w", err)
	}
	var updatedID int64
	query := func() interface{ Scan(...any) error } {
		if len(transactions) > 0 {
			return transactions[0].QueryRow(ctx, `
				UPDATE river_job SET args=$2::jsonb
				WHERE id=$1 AND kind='profile_upload_reconciliation_v1'
				  AND state='scheduled'
				  AND (args->>'user_id')::bigint=$3
				  AND args->>'bucket'=$4
				  AND args->>'object_name'=$5
				  AND args->>'object_url'=$6
				  AND COALESCE((args->>'delete_all_versions')::boolean, false)=true
				  AND COALESCE(args->>'object_version_id', '')=''
				RETURNING id
			`, jobID, encoded, args.UserID, args.Bucket, args.ObjectName, args.ObjectURL)
		}
		return js.riverClient.Driver().GetExecutor().QueryRow(ctx, `
			UPDATE river_job SET args=$2::jsonb
			WHERE id=$1 AND kind='profile_upload_reconciliation_v1'
			  AND state='scheduled'
			  AND (args->>'user_id')::bigint=$3
			  AND args->>'bucket'=$4
			  AND args->>'object_name'=$5
			  AND args->>'object_url'=$6
			  AND COALESCE((args->>'delete_all_versions')::boolean, false)=true
			  AND COALESCE(args->>'object_version_id', '')=''
			RETURNING id
		`, jobID, encoded, args.UserID, args.Bucket, args.ObjectName, args.ObjectURL)
	}
	err = query().Scan(&updatedID)
	if err != nil {
		return fmt.Errorf("finalize profile upload reconciliation intent: %w", err)
	}
	if updatedID != jobID {
		return errors.New("finalized an unexpected profile upload reconciliation intent")
	}
	return nil
}

func (js *JobServiceImpl) ScheduleAccountCleanupJob(ctx context.Context, profileObjectPrefix string, tx *pgx.Tx) error {
	_, err := js.enqueueJob(ctx, &river.InsertOpts{
		Queue:       AccountCleanupQueue,
		MaxAttempts: 25,
		ScheduledAt: time.Now().Add(5 * time.Minute),
	}, AccountCleanupArgs{ProfileBuckets: common.MinIOCleanupBuckets(), ProfileObjectPrefix: profileObjectPrefix}, tx)
	return err
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

func (js *JobServiceImpl) ScheduleImageJob(ctx context.Context, definitionID int64, userID *int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error) {
	opts := river.InsertOpts{Queue: ImageGeneratorQueue}

	return js.enqueueJob(ctx, &opts, ImageGeneratorArgs{
		DefinitionID: definitionID,
		UserID:       userID,
		ErrorReport:  errorReport,
	}, tx)
}

func (js *JobServiceImpl) ScheduleAudioJob(ctx context.Context, wordID int64, userID *int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error) {
	opts := river.InsertOpts{Queue: TextToSpeechQueue}

	return js.enqueueJob(ctx, &opts, TextToSpeechArgs{
		WordID:      wordID,
		UserID:      userID,
		ErrorReport: errorReport,
	}, tx)
}

func (js *JobServiceImpl) ScheduleDefinitionJob(ctx context.Context, wordID int64, userID *int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error) {
	opts := river.InsertOpts{Queue: DefinitionFetcherQueue}

	return js.enqueueJob(ctx, &opts, DefinitionFetcherArgs{
		WordID:      wordID,
		UserID:      userID,
		ErrorReport: errorReport,
	}, tx)
}

func (js *JobServiceImpl) ScheduleExampleAudioJob(ctx context.Context, definitionID int64, wordID int64, userID *int64, tx *pgx.Tx) error {
	args := ExampleAudioArgs{
		DefinitionID: definitionID,
		WordID:       wordID,
		UserID:       userID,
	}

	opts := &river.InsertOpts{Queue: ExampleAudioQueue}

	if tx != nil {
		_, err := js.riverClient.InsertTx(ctx, *tx, args, opts)
		return err
	}
	_, err := js.riverClient.Insert(ctx, args, opts)
	return err
}

func meaningAudioInsertOpts() *river.InsertOpts {
	return &river.InsertOpts{
		Queue:       MeaningAudioQueue,
		MaxAttempts: 5,
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByQueue: true,
		},
	}
}

func (js *JobServiceImpl) ScheduleMeaningAudioJob(ctx context.Context, definitionID int64, wordID int64, userID *int64, tx *pgx.Tx) error {
	args := MeaningAudioArgs{
		DefinitionID: definitionID,
		WordID:       wordID,
		UserID:       userID,
	}

	if tx != nil {
		_, err := js.riverClient.InsertTx(ctx, *tx, args, meaningAudioInsertOpts())
		return err
	}
	_, err := js.riverClient.Insert(ctx, args, meaningAudioInsertOpts())
	return err
}

func (js *JobServiceImpl) enqueueJob(ctx context.Context, opts *river.InsertOpts, args river.JobArgs, tx *pgx.Tx) (int64, error) {
	logger := common.Logger.With("func", "enqueueJob", "Kind", args.Kind())

	var result *rivertype.JobInsertResult
	var err error

	if tx != nil {
		result, err = js.riverClient.InsertTx(ctx, *tx, args, opts)
	} else {
		result, err = js.riverClient.Insert(ctx, args, opts)
	}

	if err != nil {
		logger.Error("failed to trigger river job", "error", err)
		return -1, errors.New("could not insert job")
	}

	return result.Job.ID, nil
}

func (js *JobServiceImpl) ScheduleStripeWebhookJob(ctx context.Context, eventID, eventType string, eventData []byte) (int64, error) {
	args := StripeWebhookArgs{
		EventID:   eventID,
		EventType: eventType,
		EventData: eventData,
	}
	opts := &river.InsertOpts{
		Queue: "stripe-webhook",
	}
	return js.enqueueJob(ctx, opts, args, nil)
}

func (js *JobServiceImpl) ScheduleRevenueCatWebhookJob(ctx context.Context, _ string, eventData []byte) (int64, error) {
	args := RevenueCatWebhookArgs{
		Payload: eventData,
	}
	opts := &river.InsertOpts{
		Queue: "revenuecat-webhook",
	}
	return js.enqueueJob(ctx, opts, args, nil)
}

func (js *JobServiceImpl) RetryJob(ctx context.Context, jobID int64) error {
	_, err := js.riverClient.JobRetry(ctx, jobID)
	return err
}

func (js *JobServiceImpl) ScheduleGoogleAcknowledgementRetry(ctx context.Context, bindingID int64) error {
	if bindingID <= 0 {
		return errors.New("valid Google purchase binding is required")
	}
	_, err := js.enqueueJob(ctx, &river.InsertOpts{
		Queue:       GoogleAcknowledgementRetryQueue,
		MaxAttempts: 18,
		UniqueOpts:  river.UniqueOpts{ByArgs: true, ByQueue: true},
	}, GoogleAcknowledgementRetryArgs{BindingID: bindingID}, nil)
	return err
}
