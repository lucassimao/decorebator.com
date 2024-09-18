package api

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

	// first we check if there are definitions for this word already
	definitions, err := FindDefinitionsByName(word.Name)

	if err == nil || len(definitions) == 0 {
		// if no existing defintions, then fall back to ai
		definitions, err = openai.GetDefinition(word.Name)
		if err != nil || len(definitions) == 0 {
			logger.Error("failed to fetch definitions using openai", "error", err)
			return err
		}
	}

	defs, err := SaveDefinition(word.Name, word.ID, definitions)

	if err != nil {
		logger.Error("failed to save definitions", "error", err)
		return err
	}

	algorithm := LeitnerSystemStrategy{}
	algorithm.IncludeDefinitions(word.UserID, defs)

	for _, definition := range defs {
		_, err = TriggerImageGenerator(definition.ID, "")

		if err != nil {
			logger.Error("failed to trigger image generator", "definitionId", definition.ID, "error", err)
		}
	}

	logger.Info("definitions fetched", "count", len(defs))
	return nil
}
