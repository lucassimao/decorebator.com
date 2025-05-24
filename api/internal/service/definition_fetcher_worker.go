package service

import (
	"context"
	"errors"

	"decorebator.com/internal/common"
	"decorebator.com/internal/openai"
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
	word, err := GetWordById(wordID)

	if err != nil {
		if errors.Is(err, common.NotFoundError{}) {
			return river.JobCancel(errors.New("word not found"))
		}
		return err
	}

	openAiDefinitions, err := openai.GetDefinition(word.Name)
	if err != nil {
		logger.Error("failed to fetch definitions using openai", "error", err)
		return err
	}

	if len(openAiDefinitions) == 0 {
		return river.JobCancel(errors.New("no definition found"))
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

	definitions, err := SaveDefinition(word.Name, word.ID, openAiDefinitions, tx)

	if err != nil {
		logger.Error("failed to save definitions", "error", err)
		return err
	}

	definitionIds := []int64{}
	for _, definition := range definitions {
		definitionIds = append(definitionIds, definition.ID)

		_, err = TriggerGenerateImageWorker(definition.ID, "", &tx)

		if err != nil {
			logger.Error("failed to trigger image generator", "definitionId", definition.ID, "error", err)
		}
	}
	quizStrategy := LeitnerSystemStrategy{}
	quizStrategy.IncludeDefinitions(word.ID, word.UserID, definitionIds, tx)

	logger.Info("definitions fetched", "count", len(definitions))
	return nil
}
