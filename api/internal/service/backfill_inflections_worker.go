package service

import (
	"context"
	"errors"

	"decorebator.com/internal/common"
	"decorebator.com/internal/openai"
	"github.com/riverqueue/river"
)

type BackfillInflectionsArgs struct {
	BatchSize int `json:"batchSize"`
}

func (BackfillInflectionsArgs) Kind() string { return "BackfillInflections" }

type BackfillInflectionsWorker struct {
	river.WorkerDefaults[BackfillInflectionsArgs]
}

func (w *BackfillInflectionsWorker) Work(ctx context.Context, job *river.Job[BackfillInflectionsArgs]) error {
	logger := common.Logger.With("worker", "BackfillInflections")
	batchSize := job.Args.BatchSize
	
	if batchSize <= 0 || batchSize > 20 {
		batchSize = 5 // Default safe batch size for API rate limits
	}

	logger.Info("starting inflections backfill", "batchSize", batchSize)

	db, err := common.GetDBConnection()
	if err != nil {
		return err
	}

	// Find verb definitions that don't have inflections
	query := `
		SELECT id, token, part_of_speech
		FROM definitions
		WHERE part_of_speech IN ('verb', 'phrasal verb')
		AND (inflections IS NULL OR jsonb_array_length(inflections) = 0)
		ORDER BY id
		LIMIT $1
	`

	rows, err := db.Query(ctx, query, batchSize)
	if err != nil {
		return err
	}
	defer rows.Close()

	var definitionsToProcess []struct {
		ID           int64
		Token        string
		PartOfSpeech string
	}

	for rows.Next() {
		var id int64
		var token, pos string
		err = rows.Scan(&id, &token, &pos)
		if err != nil {
			logger.Error("error scanning definition row", "error", err)
			continue
		}
		definitionsToProcess = append(definitionsToProcess, struct {
			ID           int64
			Token        string
			PartOfSpeech string
		}{id, token, pos})
	}

	if len(definitionsToProcess) == 0 {
		logger.Info("no verb definitions need inflection backfill")
		return nil
	}

	logger.Info("found definitions needing backfill", "count", len(definitionsToProcess))

	// Group definitions by token to avoid duplicate API calls
	tokenMap := make(map[string][]int64)
	for _, def := range definitionsToProcess {
		tokenMap[def.Token] = append(tokenMap[def.Token], def.ID)
	}

	successCount := 0
	errorCount := 0

	// Process each unique token
	for token, definitionIds := range tokenMap {
		logger.Info("processing token", "token", token, "definitionCount", len(definitionIds))

		// Get new definition with inflections from OpenAI
		definitionData, err := openai.GetDefinition(token)
		if err != nil {
			logger.Error("failed to get definition from OpenAI", "token", token, "error", err)
			errorCount += len(definitionIds)
			continue
		}

		// Find verb definitions with inflections in the response
		var verbInflections []struct {
			Inflection string   `json:"inflection"`
			Tense      string   `json:"tense"`
			Examples   []string `json:"examples"`
		}

		for _, def := range definitionData.Definitions {
			if (def.PartOfSpeech == "verb" || def.PartOfSpeech == "phrasal verb") && len(def.Inflections) > 0 {
				verbInflections = def.Inflections
				break // Use the first verb definition's inflections
			}
		}

		if len(verbInflections) == 0 {
			logger.Warn("no verb inflections found in OpenAI response", "token", token)
			errorCount += len(definitionIds)
			continue
		}

		// Update existing definitions with inflections
		for _, defId := range definitionIds {
			updateQuery := `
				UPDATE definitions 
				SET inflections = $1,
				    updated_at = NOW()
				WHERE id = $2
			`

			_, err = db.Exec(ctx, updateQuery, verbInflections, defId)
			if err != nil {
				logger.Error("failed to update definition", "definitionId", defId, "error", err)
				errorCount++
			} else {
				logger.Info("updated definition with inflections", "definitionId", defId, "inflectionCount", len(verbInflections))
				successCount++
			}
		}
	}

	logger.Info("backfill batch completed", "successCount", successCount, "errorCount", errorCount)

	// If there are more definitions to process, queue another job
	if len(definitionsToProcess) == batchSize {
		// Check if there are more definitions
		var remainingCount int
		err = db.QueryRow(ctx, `
			SELECT COUNT(*) 
			FROM definitions
			WHERE part_of_speech IN ('verb', 'phrasal verb')
			AND (inflections IS NULL OR jsonb_array_length(inflections) = 0)
		`).Scan(&remainingCount)

		if err == nil && remainingCount > 0 {
			logger.Info("queueing next backfill batch", "remaining", remainingCount)
			_, err = TriggerBackfillInflectionsWorker(batchSize)
			if err != nil {
				logger.Error("failed to queue next backfill batch", "error", err)
			}
		}
	}

	if errorCount > 0 && successCount == 0 {
		return errors.New("all definitions in batch failed to process")
	}

	return nil
}

// TriggerBackfillInflectionsWorker queues a backfill job
func TriggerBackfillInflectionsWorker(batchSize int) (*river.JobInsertResult, error) {
	client, err := GetRiverClient()
	if err != nil {
		return nil, err
	}

	opts := &river.InsertOpts{
		Queue: BACKFILL_INFLECTIONS_QUEUE,
	}

	return client.Insert(context.Background(), BackfillInflectionsArgs{
		BatchSize: batchSize,
	}, opts)
}