package service

import (
	"context"
	"errors"

	"decorebator.com/internal/common"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

func TriggerGenerateImageWorker(definitionId int64, customPrompt string, errorReport *ErrorReport, tx *pgx.Tx) (int64, error) {
	opts := river.InsertOpts{
		Queue: IMAGE_GENERATOR_QUEUE,
	}

	return triggerWorker(&opts, ImageGeneratorArgs{
		DefinitionId: definitionId,
		CustomPrompt: customPrompt,
		ErrorReport:  errorReport,
	}, tx)
}

func TriggerTextToSpeechWorker(wordId int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error) {
	opts := river.InsertOpts{
		Queue: TEXT_TO_SPEECH_QUEUE,
	}

	return triggerWorker(&opts, TextToSpeechArgs{
		WordId:      wordId,
		ErrorReport: errorReport,
	}, tx)
}

func TriggerFetchDefinitionWorker(wordId int64, errorReport *ErrorReport, tx *pgx.Tx) (int64, error) {
	opts := river.InsertOpts{
		Queue: DEFINITION_FETCHER_QUEUE,
	}

	return triggerWorker(&opts, DefinitionFetcherArgs{
		WordId:      wordId,
		ErrorReport: errorReport,
	}, tx)
}

func triggerWorker(opts *river.InsertOpts, args river.JobArgs, tx *pgx.Tx) (int64, error) {
	logger := common.Logger.With("func", "triggerWorker", "Kind", args.Kind())

	riverClient, err := GetRiverClient()
	if err != nil {
		logger.Error("failed to open river connection", "error", err)
		return -1, errors.New("could not open river client")
	}

	var result *rivertype.JobInsertResult
	if tx != nil {
		result, err = riverClient.InsertTx(context.Background(), *tx, args, opts)
	} else {
		result, err = riverClient.Insert(context.Background(), args, opts)
	}

	if err != nil {
		logger.Error("failed to trigger river job", "error", err)
		return -1, errors.New("could not insert job")
	}

	return result.Job.ID, nil
}
