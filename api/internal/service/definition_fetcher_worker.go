package service

import (
	"context"
	"errors"

	"decorebator.com/internal/common"
	"decorebator.com/internal/openai"
	"github.com/riverqueue/river"
)

type DefinitionFetcherArgs struct {
	WordId      int64        `json:"wordId"`
	ErrorReport *ErrorReport `json:"errorReport"`
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

	definitionData, err := openai.GetDefinition(word.Name)
	if err != nil {
		logger.Error("failed to fetch definitions using openai", "error", err)
		return err
	}

	if len(definitionData.Definitions) == 0 {
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

	definitions, err := SaveDefinition(word.Name, word.ID, definitionData.Definitions, tx)

	if err != nil {
		logger.Error("failed to save definitions", "error", err)
		return err
	}

	// Save pronunciation to word table
	if definitionData.Pronunciation != "" {
		_, err = tx.Exec(ctx, "UPDATE words SET pronunciation = $1 WHERE id = $2", definitionData.Pronunciation, word.ID)
		if err != nil {
			logger.Error("failed to save pronunciation", "error", err)
			return err
		}
		logger.Info("pronunciation saved", "pronunciation", definitionData.Pronunciation, "wordId", word.ID)
	}

	definitionIds := []int64{}
	for _, definition := range definitions {
		definitionIds = append(definitionIds, definition.ID)

		_, err = TriggerGenerateImageWorker(definition.ID, "", nil, &tx)

		if err != nil {
			logger.Error("failed to trigger image generator", "definitionId", definition.ID, "error", err)
		}
	}
	strategy := LeitnerSystemStrategy{}
	strategy.IncludeDefinitions(word.ID, word.UserID, definitionIds, tx)

	// if this job was triggered by an error report, then mark the issue as solved
	if job.Args.ErrorReport != nil {
		return strategy.MarkErrorResolved(*job.Args.ErrorReport)
	}

	logger.Info("definitions fetched", "count", len(definitions))
	return nil
}
