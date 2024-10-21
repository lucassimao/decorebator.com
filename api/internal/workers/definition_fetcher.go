package workers

import (
	"context"
	"errors"

	"decorebator.com/internal/common"
	"decorebator.com/internal/openai"
	service "decorebator.com/internal/service"
	"github.com/riverqueue/river"
)

type DefinitionFetcherArgs struct {
	WordId int64 `json:"wordId"`
}

func (DefinitionFetcherArgs) Kind() string { return "DefinitionFetcher" }

type DefinitionFetcherWorker struct {
	river.WorkerDefaults[DefinitionFetcherArgs]
}

func (w *DefinitionFetcherWorker) Work(ctx context.Context, job *river.Job[DefinitionFetcherArgs]) error {
	logger := common.Logger.With("worker", "DefinitionFetcher")
	wordID := job.Args.WordId
	word, err := service.GetWordById(wordID)

	if err != nil {
		if errors.Is(err, common.NotFoundError{}) {
			return river.JobCancel(errors.New("word not found"))
		}
		return err
	}

	openAiDefinitions, err := openai.GetDefinition(word.Name)
	if err != nil || len(openAiDefinitions) == 0 {
		logger.Error("failed to fetch definitions using openai", "error", err)
		return err
	}

	db, err := common.GetDBConnection()
	if err != nil {
		return err
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			tx.Commit(ctx)
		} else {
			tx.Rollback(ctx)
		}
	}()

	definitions, err := service.SaveDefinition(word.Name, word.ID, openAiDefinitions, tx)

	if err != nil {
		logger.Error("failed to save definitions", "error", err)
		return err
	}

	definitionIds := []int64{}
	for _, definition := range definitions {
		definitionIds = append(definitionIds, definition.ID)
	}
	quizStrategy := service.LeitnerSystemStrategy{}
	quizStrategy.IncludeDefinitions(word.ID, word.UserID, definitionIds, tx)

	var container, ok = common.GetDigContainerFromContext(ctx)
	if !ok {
		return errors.New("dig container not found")
	}

	for _, definition := range definitions {
		err := container.Invoke(func(srv common.ContentGenerationService) error {
			_, err = srv.GenerateImage(definition.ID, "", &tx)
			return err
		})

		if err != nil {
			logger.Error("failed to trigger image generator", "definitionId", definition.ID, "error", err)
		}
	}

	logger.Info("definitions fetched", "count", len(definitions))
	return nil
}
