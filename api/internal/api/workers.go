package api

import (
	"context"
	"errors"

	"decorebator.com/internal/common"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

const IMAGE_GENERATOR_QUEUE = "image_generator"
const TEXT_TO_SPEECH_QUEUE = "text_to_speech"
const DEFINITION_FETCHER_QUEUE = "definition_fetcher"

func GetRiverClient() (*river.Client[pgx.Tx], error) {
	db, err := common.GetDBConnection()

	if err != nil {
		return nil, err
	}

	riverWorkers := river.NewWorkers()
	river.AddWorker(riverWorkers, &ImageGeneratorWorker{})
	river.AddWorker(riverWorkers, &TextToSpeechWorker{})

	riverClient, err := river.NewClient(riverpgxv5.New(db), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault:       {MaxWorkers: 100},
			IMAGE_GENERATOR_QUEUE:    {MaxWorkers: 5},
			TEXT_TO_SPEECH_QUEUE:     {MaxWorkers: 5},
			DEFINITION_FETCHER_QUEUE: {MaxWorkers: 50},
		},
		Workers: riverWorkers,
		Logger:  common.Logger,
	})
	if err != nil {
		return nil, err
	}

	return riverClient, nil
}

func TriggerImageGenerator(definitionId int64, customPrompt string) (int64, error) {
	opts := river.InsertOpts{
		Queue: IMAGE_GENERATOR_QUEUE,
	}

	return triggerWorker(&opts, ImageGeneratorArgs{
		DefinitionId: definitionId,
		CustomPrompt: customPrompt,
	})
}

func TriggerTextToSpeech(wordId int64) (int64, error) {
	opts := river.InsertOpts{
		Queue: TEXT_TO_SPEECH_QUEUE,
	}

	return triggerWorker(&opts, TextToSpeechArgs{
		WordId: wordId,
	})
}

func TriggerDefinitionFetcher(wordId int64) (int64, error) {
	opts := river.InsertOpts{
		Queue: DEFINITION_FETCHER_QUEUE,
	}

	return triggerWorker(&opts, DefinitionFetcherArgs{
		WordId: wordId,
	})
}

func triggerWorker(opts *river.InsertOpts, args river.JobArgs) (int64, error) {
	logger := common.Logger.With("func", "triggerWorker", "Kind", args.Kind())

	riverClient, err := GetRiverClient()
	if err != nil {
		logger.Error("failed to open river connection", "error", err)
		return -1, errors.New("could not open river client")
	}

	result, err := riverClient.Insert(context.Background(), args, opts)

	if err != nil {
		logger.Error("failed to trigger river job", "error", err)
		return -1, errors.New("could not insert job")
	}

	return result.Job.ID, nil
}
