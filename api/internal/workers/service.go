package workers

import (
	"context"
	"errors"

	"decorebator.com/internal/common"
	internal "decorebator.com/internal/workers/internal"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

const IMAGE_GENERATOR_QUEUE = "image_generator"

func GetRiverClient() (*river.Client[pgx.Tx], error) {
	db, err := common.GetDBConnection()

	if err != nil {
		return nil, err
	}

	riverWorkers := river.NewWorkers()
	river.AddWorker(riverWorkers, &internal.ImageGeneratorWorker{})

	riverClient, err := river.NewClient(riverpgxv5.New(db), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault:    {MaxWorkers: 100},
			IMAGE_GENERATOR_QUEUE: {MaxWorkers: 5},
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
	logger := common.Logger.With("definitionId", definitionId, "func", TriggerImageGenerator)

	riverClient, err := GetRiverClient()
	if err != nil {
		logger.Error("failed to open river connection", "error", err)
		return -1, errors.New("could not open river client")
	}

	opts := river.InsertOpts{
		Queue: IMAGE_GENERATOR_QUEUE,
	}

	result, err := riverClient.Insert(context.Background(), internal.ImageGeneratorArgs{
		DefinitionId: definitionId,
		CustomPrompt: customPrompt,
	}, &opts)

	if err != nil {
		logger.Error("failed to trigger river job", "error", err)
		return -1, errors.New("could not insert job")
	}

	return result.Job.ID, nil
}
