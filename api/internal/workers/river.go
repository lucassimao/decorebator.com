package workers

import (
	"decorebator.com/internal/common"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

func GetRiverClient() (*river.Client[pgx.Tx], error) {
	db, err := common.GetDBConnection()

	if err != nil {
		return nil, err
	}

	riverWorkers := river.NewWorkers()
	river.AddWorker(riverWorkers, &ImageGeneratorWorker{})

	riverClient, err := river.NewClient(riverpgxv5.New(db), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 5},
		},
		Workers: riverWorkers,
		Logger:  common.Logger,
	})
	if err != nil {
		return nil, err
	}

	return riverClient, nil
}
