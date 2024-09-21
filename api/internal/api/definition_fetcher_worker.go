package api

import (
	"context"
	"errors"

	"decorebator.com/internal/common"
	"decorebator.com/internal/openai"
	"github.com/riverqueue/river"
	"go.uber.org/dig"
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

	// first we check if there are definitions for this word already
	definitions, _ := FindDefinitionsByName(word.Name)

	if len(definitions) > 0 {
		ids := []int64{}
		for _, def := range definitions {
			ids = append(ids, def.ID)
		}
		LinkDefinitions(word.ID, ids)
	} else {
		// if no existing defintions, then fall back to ai
		openAiDefinitions, err := openai.GetDefinition(word.Name)
		if err != nil || len(openAiDefinitions) == 0 {
			logger.Error("failed to fetch definitions using openai", "error", err)
			return err
		}
		definitions, err = SaveDefinition(word.Name, word.ID, openAiDefinitions)

		if err != nil {
			logger.Error("failed to save definitions", "error", err)
			return err
		}
	}

	algorithm := LeitnerSystemStrategy{}
	algorithm.IncludeDefinitions(word.UserID, definitions)

	for _, definition := range definitions {
		container := dig.New()
		err := container.Invoke(func(trigger common.WorkerTrigger) error {
			_, err = trigger.TriggerImageGenerator(definition.ID, "")
			return err
		})

		if err != nil {
			logger.Error("failed to trigger image generator", "definitionId", definition.ID, "error", err)
		}
	}

	logger.Info("definitions fetched", "count", len(definitions))
	return nil
}
