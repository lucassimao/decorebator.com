package integration

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/stretchr/testify/require"
)

func TestMeaningAudioURLCompareAndSetAndNotFoundContract(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	db := setup.CreateTestDB(t)
	defer db.Close()
	require.NoError(t, setup.RunMigrations(db))
	ctx := context.Background()
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	var definitionID int64
	require.NoError(t, db.QueryRow(ctx, `INSERT INTO definitions
		(token,language,part_of_speech,part_of_speech_normalized,meaning,examples,inflections,source)
		VALUES ($1,'en','noun','noun','meaning',ARRAY['example'],'[]'::jsonb,'test') RETURNING id`,
		"meaning-audio-"+suffix).Scan(&definitionID))

	definitionService := service.NewDefinitionService(db)
	_, err := definitionService.GetDefinitionByID(ctx, definitionID+1_000_000)
	require.Error(t, err)
	require.True(t, errors.Is(err, common.NotFoundError{}))

	results := make(chan bool, 2)
	errorsCh := make(chan error, 2)
	var wg sync.WaitGroup
	for _, url := range []string{"https://objects.test/first.mp3", "https://objects.test/second.mp3"} {
		wg.Add(1)
		go func(candidate string) {
			defer wg.Done()
			updated, updateErr := definitionService.SetMeaningAudioURLIfEmpty(ctx, definitionID, candidate)
			results <- updated
			errorsCh <- updateErr
		}(url)
	}
	wg.Wait()
	close(results)
	close(errorsCh)

	updates := 0
	for updated := range results {
		if updated {
			updates++
		}
	}
	for updateErr := range errorsCh {
		require.NoError(t, updateErr)
	}
	require.Equal(t, 1, updates)

	definition, err := definitionService.GetDefinitionByID(ctx, definitionID)
	require.NoError(t, err)
	require.Contains(t, []string{"https://objects.test/first.mp3", "https://objects.test/second.mp3"}, definition.MeaningAudioURL)
}

func TestMeaningAudioJobIsTransactionalAndUniqueByDefinition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	db := setup.CreateTestDB(t)
	defer db.Close()
	require.NoError(t, setup.RunMigrations(db))
	ctx := context.Background()
	riverClient, err := river.NewClient(riverpgxv5.New(db), &river.Config{})
	require.NoError(t, err)
	jobService := service.NewJobService(riverClient)
	definitionID := time.Now().UnixNano()
	firstUserID := int64(101)
	secondUserID := int64(202)

	tx, err := db.Begin(ctx)
	require.NoError(t, err)
	require.NoError(t, jobService.ScheduleMeaningAudioJob(ctx, definitionID, 11, &firstUserID, &tx))
	require.NoError(t, tx.Rollback(ctx))
	require.Equal(t, 0, meaningAudioJobCount(t, db, definitionID))

	require.NoError(t, jobService.ScheduleMeaningAudioJob(ctx, definitionID, 11, &firstUserID, nil))
	require.NoError(t, jobService.ScheduleMeaningAudioJob(ctx, definitionID, 99, &secondUserID, nil))
	require.Equal(t, 1, meaningAudioJobCount(t, db, definitionID), "word and user changes must not bypass per-definition uniqueness")
}

func meaningAudioJobCount(t *testing.T, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, definitionID int64) int {
	t.Helper()
	var count int
	require.NoError(t, db.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM river_job
		WHERE kind = 'DefinitionMeaningAudio' AND args->>'definitionId' = $1`,
		strconv.FormatInt(definitionID, 10)).Scan(&count))
	return count
}
