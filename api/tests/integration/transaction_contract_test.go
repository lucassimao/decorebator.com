package integration

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingDefinitionJobService struct {
	service.JobService
	err error
}

func (s failingDefinitionJobService) ScheduleDefinitionJob(
	context.Context,
	int64,
	*int64,
	*service.ErrorReport,
	*pgx.Tx,
) (int64, error) {
	return 0, s.err
}

func TestDefinitionSaveReturnsCommitFailureAndRetryDoesNotDuplicateSet(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	ctx := context.Background()
	token := server.WithPremiumUser(t)

	var userID int64
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT id FROM users ORDER BY created_at DESC LIMIT 1`).Scan(&userID))
	wordlistID := createTestWordlist(server, token, "Definition transaction", "Commit failure")
	word := saveAttributionWord(t, server, userID, wordlistID, "tx-save")

	installDefinitionInsertCommitFailure(t, server)
	defer dropDefinitionInsertCommitFailure(t, server)
	definitions := []*model.Definition{
		definitionFixture("tx-save", "first transactional meaning"),
		definitionFixture("tx-save", "second transactional meaning"),
	}
	_, err := server.AppContext.DefinitionService.SaveDefinition(ctx, word.ID, definitions, nil)
	require.ErrorContains(t, err, "failed to save definitions transaction")

	assertDefinitionSetCounts(t, server, word.ID, "tx-save", 0, 0)
	dropDefinitionInsertCommitFailure(t, server)

	saved, err := server.AppContext.DefinitionService.SaveDefinition(ctx, word.ID, definitions, nil)
	require.NoError(t, err)
	require.Len(t, saved, 2)
	assertDefinitionSetCounts(t, server, word.ID, "tx-save", 2, 2)
}

func TestDefinitionDeletionHandlesMixedSharingAndSchedulesAtomically(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	ctx := context.Background()
	token := server.WithPremiumUser(t)

	var userID int64
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT id FROM users ORDER BY created_at DESC LIMIT 1`).Scan(&userID))
	wordlistID := createTestWordlist(server, token, "Definition deletion", "Mixed sharing")
	targetWord := saveAttributionWord(t, server, userID, wordlistID, "delete-target")
	otherWord := saveAttributionWord(t, server, userID, wordlistID, "delete-other")

	definitions, err := server.AppContext.DefinitionService.SaveDefinition(ctx, targetWord.ID, []*model.Definition{
		definitionFixture("delete-target", "exclusive meaning"),
		definitionFixture("delete-target", "shared meaning"),
	}, nil)
	require.NoError(t, err)
	require.Len(t, definitions, 2)
	exclusiveID, sharedID := definitions[0].ID, definitions[1].ID

	tx, err := server.DB.Begin(ctx)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback(ctx) }()
	_, err = tx.Exec(ctx, `INSERT INTO word_definitions (word_id, definition_id) VALUES ($1, $2)`, otherWord.ID, sharedID)
	require.NoError(t, err)
	require.NoError(t, server.AppContext.LeitnerTrackingService.IncludeDefinitions(
		ctx, targetWord.ID, userID, []int64{exclusiveID, sharedID}, tx,
	))
	require.NoError(t, server.AppContext.LeitnerTrackingService.IncludeDefinitions(
		ctx, otherWord.ID, userID, []int64{sharedID}, tx,
	))
	require.NoError(t, tx.Commit(ctx))
	initialDefinitionJobs := definitionJobCount(t, server, targetWord.ID)

	installWordDefinitionDeleteCommitFailure(t, server, targetWord.ID)
	defer dropWordDefinitionDeleteCommitFailure(t, server)
	err = server.AppContext.DefinitionService.DeleteWordDefinitions(ctx, targetWord.ID, nil)
	require.ErrorContains(t, err, "failed to delete word definitions transaction")
	assertMixedDefinitionState(t, server, targetWord.ID, otherWord.ID, exclusiveID, sharedID, 2, 2)
	_, err = server.AppContext.DefinitionService.ScheduleDefinitionRegeneration(
		ctx, targetWord.ID, server.AppContext.JobService,
	)
	require.ErrorContains(t, err, "failed to schedule definition regeneration")
	assertMixedDefinitionState(t, server, targetWord.ID, otherWord.ID, exclusiveID, sharedID, 2, 2)
	assert.Equal(t, initialDefinitionJobs, definitionJobCount(t, server, targetWord.ID))
	dropWordDefinitionDeleteCommitFailure(t, server)

	_, err = server.AppContext.DefinitionService.ScheduleDefinitionRegeneration(
		ctx, targetWord.ID, failingDefinitionJobService{err: errors.New("forced enqueue failure")},
	)
	require.ErrorContains(t, err, "forced enqueue failure")
	assertMixedDefinitionState(t, server, targetWord.ID, otherWord.ID, exclusiveID, sharedID, 2, 2)

	jobID, err := server.AppContext.DefinitionService.ScheduleDefinitionRegeneration(
		ctx, targetWord.ID, server.AppContext.JobService,
	)
	require.NoError(t, err)
	assert.Positive(t, jobID)
	assertMixedDefinitionState(t, server, targetWord.ID, otherWord.ID, exclusiveID, sharedID, 0, 0)

	var exclusiveDefinitions, sharedDefinitions, jobs int
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM definitions WHERE id=$1`, exclusiveID).Scan(&exclusiveDefinitions))
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM definitions WHERE id=$1`, sharedID).Scan(&sharedDefinitions))
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM river_job
		WHERE id=$1 AND kind='DefinitionFetcher' AND (args->>'wordId')::bigint=$2
	`, jobID, targetWord.ID).Scan(&jobs))
	assert.Zero(t, exclusiveDefinitions)
	assert.Equal(t, 1, sharedDefinitions)
	assert.Equal(t, 1, jobs)
}

func TestLeitnerManagedTransactionsReturnDeferredCommitFailures(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	ctx := context.Background()
	token := server.WithPremiumUser(t)

	var userID int64
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT id FROM users ORDER BY created_at DESC LIMIT 1`).Scan(&userID))
	wordlistID := createTestWordlist(server, token, "Leitner transaction", "Commit failure")
	word := saveAttributionWord(t, server, userID, wordlistID, "tx-leitner")
	definitionID, trackingID := saveAttributionDefinition(t, server, word.ID, userID, "transaction meaning")
	_, err := server.DB.Exec(ctx, `UPDATE leitner_system_tracking SET box_id=5 WHERE id=$1`, trackingID)
	require.NoError(t, err)

	installTrackingCommitFailure(t, server, trackingID)
	defer dropTrackingCommitFailure(t, server)
	err = server.AppContext.LeitnerTrackingService.UpdateQuizProgress(ctx, trackingID, true, nil)
	require.Error(t, err)
	assertTrackingBox(t, server, trackingID, 5)
	dropTrackingCommitFailure(t, server)

	installQuizPerformanceCommitFailure(t, server, trackingID)
	defer dropQuizPerformanceCommitFailure(t, server)
	err = server.AppContext.LeitnerSystemStrategy.SaveQuizResult(ctx, service.QuizResult{
		UserID: userID, WordlistID: wordlistID, WordID: word.ID,
		DefinitionID: definitionID, LeitnerSystemTrackingID: trackingID,
		QuizType: model.WriteWordFromDefinition, IsCorrect: true, ResponseTimeMs: 500,
	}, true, nil)
	require.ErrorContains(t, err, "failed to commit quiz result")
	assertTrackingBox(t, server, trackingID, 5)

	var performanceRows int
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM quiz_performance WHERE leitner_system_tracking_id=$1
	`, trackingID).Scan(&performanceRows))
	assert.Zero(t, performanceRows)
}

func definitionFixture(token, meaning string) *model.Definition {
	return &model.Definition{
		Token: token, Language: "en", PartOfSpeech: "noun", Meaning: meaning,
		Examples: []string{"A valid [example]."}, Source: model.ChatGPT,
	}
}

func assertDefinitionSetCounts(t *testing.T, server *setup.TestServer, wordID int64, token string, definitions, links int) {
	t.Helper()
	ctx := context.Background()
	var definitionCount, linkCount int
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM definitions WHERE token=$1`, token).Scan(&definitionCount))
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM word_definitions wd JOIN definitions d ON d.id=wd.definition_id
		WHERE wd.word_id=$1 AND d.token=$2
	`, wordID, token).Scan(&linkCount))
	assert.Equal(t, definitions, definitionCount)
	assert.Equal(t, links, linkCount)
}

func assertMixedDefinitionState(
	t *testing.T,
	server *setup.TestServer,
	targetWordID, otherWordID, exclusiveID, sharedID int64,
	targetLinks, targetTracking int,
) {
	t.Helper()
	ctx := context.Background()
	var links, targetRows, otherRows int
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM word_definitions WHERE word_id=$1`, targetWordID).Scan(&links))
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM leitner_system_tracking
		WHERE word_id=$1 AND definition_id IN ($2, $3)
	`, targetWordID, exclusiveID, sharedID).Scan(&targetRows))
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM leitner_system_tracking WHERE word_id=$1 AND definition_id=$2
	`, otherWordID, sharedID).Scan(&otherRows))
	assert.Equal(t, targetLinks, links)
	assert.Equal(t, targetTracking, targetRows)
	assert.Equal(t, 1, otherRows)
}

func assertTrackingBox(t *testing.T, server *setup.TestServer, trackingID int64, expected int) {
	t.Helper()
	var box int
	require.NoError(t, server.DB.QueryRow(context.Background(), `
		SELECT box_id FROM leitner_system_tracking WHERE id=$1
	`, trackingID).Scan(&box))
	assert.Equal(t, expected, box)
}

func definitionJobCount(t *testing.T, server *setup.TestServer, wordID int64) int {
	t.Helper()
	var jobs int
	require.NoError(t, server.DB.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM river_job
		WHERE kind='DefinitionFetcher' AND (args->>'wordId')::bigint=$1
	`, wordID).Scan(&jobs))
	return jobs
}

func installDefinitionInsertCommitFailure(t *testing.T, server *setup.TestServer) {
	t.Helper()
	_, err := server.DB.Exec(context.Background(), `
		CREATE OR REPLACE FUNCTION test_fail_definition_insert_commit()
		RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN
			IF NEW.token='tx-save' THEN
				RAISE EXCEPTION 'forced deferred definition insert failure';
			END IF;
			RETURN NEW;
		END;
		$$;
		CREATE CONSTRAINT TRIGGER test_definition_insert_commit_failure
		AFTER INSERT ON definitions
		DEFERRABLE INITIALLY DEFERRED
		FOR EACH ROW EXECUTE FUNCTION test_fail_definition_insert_commit();
	`)
	require.NoError(t, err)
}

func dropDefinitionInsertCommitFailure(t *testing.T, server *setup.TestServer) {
	t.Helper()
	_, err := server.DB.Exec(context.Background(), `
		DROP TRIGGER IF EXISTS test_definition_insert_commit_failure ON definitions;
		DROP FUNCTION IF EXISTS test_fail_definition_insert_commit();
	`)
	require.NoError(t, err)
}

func installWordDefinitionDeleteCommitFailure(t *testing.T, server *setup.TestServer, wordID int64) {
	t.Helper()
	_, err := server.DB.Exec(context.Background(), fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION test_fail_word_definition_delete_commit()
		RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN
			IF OLD.word_id=%d THEN
				RAISE EXCEPTION 'forced deferred word definition delete failure';
			END IF;
			RETURN OLD;
		END;
		$$;
		CREATE CONSTRAINT TRIGGER test_word_definition_delete_commit_failure
		AFTER DELETE ON word_definitions
		DEFERRABLE INITIALLY DEFERRED
		FOR EACH ROW EXECUTE FUNCTION test_fail_word_definition_delete_commit();
	`, wordID))
	require.NoError(t, err)
}

func dropWordDefinitionDeleteCommitFailure(t *testing.T, server *setup.TestServer) {
	t.Helper()
	_, err := server.DB.Exec(context.Background(), `
		DROP TRIGGER IF EXISTS test_word_definition_delete_commit_failure ON word_definitions;
		DROP FUNCTION IF EXISTS test_fail_word_definition_delete_commit();
	`)
	require.NoError(t, err)
}

func installTrackingCommitFailure(t *testing.T, server *setup.TestServer, trackingID int64) {
	t.Helper()
	_, err := server.DB.Exec(context.Background(), fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION test_fail_tracking_commit()
		RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN
			IF NEW.id=%d THEN RAISE EXCEPTION 'forced deferred tracking failure'; END IF;
			RETURN NEW;
		END;
		$$;
		CREATE CONSTRAINT TRIGGER test_tracking_commit_failure
		AFTER UPDATE ON leitner_system_tracking
		DEFERRABLE INITIALLY DEFERRED
		FOR EACH ROW EXECUTE FUNCTION test_fail_tracking_commit();
	`, trackingID))
	require.NoError(t, err)
}

func dropTrackingCommitFailure(t *testing.T, server *setup.TestServer) {
	t.Helper()
	_, err := server.DB.Exec(context.Background(), `
		DROP TRIGGER IF EXISTS test_tracking_commit_failure ON leitner_system_tracking;
		DROP FUNCTION IF EXISTS test_fail_tracking_commit();
	`)
	require.NoError(t, err)
}

func installQuizPerformanceCommitFailure(t *testing.T, server *setup.TestServer, trackingID int64) {
	t.Helper()
	_, err := server.DB.Exec(context.Background(), fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION test_fail_quiz_performance_commit()
		RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN
			IF NEW.leitner_system_tracking_id=%d THEN
				RAISE EXCEPTION 'forced deferred quiz performance failure';
			END IF;
			RETURN NEW;
		END;
		$$;
		CREATE CONSTRAINT TRIGGER test_quiz_performance_commit_failure
		AFTER INSERT ON quiz_performance
		DEFERRABLE INITIALLY DEFERRED
		FOR EACH ROW EXECUTE FUNCTION test_fail_quiz_performance_commit();
	`, trackingID))
	require.NoError(t, err)
}

func dropQuizPerformanceCommitFailure(t *testing.T, server *setup.TestServer) {
	t.Helper()
	_, err := server.DB.Exec(context.Background(), `
		DROP TRIGGER IF EXISTS test_quiz_performance_commit_failure ON quiz_performance;
		DROP FUNCTION IF EXISTS test_fail_quiz_performance_commit();
	`)
	require.NoError(t, err)
}
