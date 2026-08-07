package errorreporting

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"decorebator.com/internal/app"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingReportedDefinitionJobService struct {
	service.JobService
}

func (f failingReportedDefinitionJobService) ScheduleDefinitionJob(_ context.Context, _ int64, _ *int64, report *service.ErrorReport, _ *pgx.Tx) (int64, error) {
	if report != nil {
		return 0, errors.New("forced definition job insertion failure")
	}
	return 1, nil
}

func (f failingReportedDefinitionJobService) ScheduleAudioJob(_ context.Context, _ int64, _ *int64, _ *service.ErrorReport, _ *pgx.Tx) (int64, error) {
	return 1, nil
}

func (f failingReportedDefinitionJobService) ScheduleResetPasswordEmailJob(context.Context, string, ...pgx.Tx) error {
	return nil
}

func TestErrorRegenerationSubmissionPreservesLearningAndUnrelatedDefinitions(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	ctx := context.Background()
	token := server.WithTestUser(t)
	wordlistID := createRegenerationWordlist(t, server, token)
	reportedWordID := createRegenerationWord(t, server, token, wordlistID, "reported")
	sharedWordID := createRegenerationWord(t, server, token, wordlistID, "shared-owner")

	reportedDefinition := saveRegenerationDefinition(t, server, reportedWordID, "reported meaning")
	unrelatedDefinition := saveRegenerationDefinition(t, server, reportedWordID, "unrelated meaning")
	sharedDefinition := saveRegenerationDefinition(t, server, reportedWordID, "shared meaning")
	_, err := server.DB.Exec(ctx, `
		INSERT INTO word_definitions (word_id, definition_id)
		VALUES ($1, $2)
	`, sharedWordID, sharedDefinition.ID)
	require.NoError(t, err)

	tx, err := server.DB.Begin(ctx)
	require.NoError(t, err)
	require.NoError(t, server.AppContext.LeitnerTrackingService.IncludeDefinitions(
		ctx,
		reportedWordID,
		1,
		[]int64{reportedDefinition.ID, unrelatedDefinition.ID, sharedDefinition.ID},
		tx,
	))
	require.NoError(t, tx.Commit(ctx))

	var trackingID int64
	require.NoError(t, server.DB.QueryRow(ctx, `
		UPDATE leitner_system_tracking
		SET box_id=5, next_review_at=NOW() + INTERVAL '7 days'
		WHERE word_id=$1 AND definition_id=$2
		RETURNING id
	`, reportedWordID, reportedDefinition.ID).Scan(&trackingID))
	_, err = server.DB.Exec(ctx, `
		INSERT INTO quiz_performance (
			user_id, wordlist_id, word_id, definition_id,
			leitner_system_tracking_id, quiz_type, box_id, is_correct, response_time_ms
		) VALUES (1, $1, $2, $3, $4, 'WRITE_WORD_FROM_DEFINITION', 5, TRUE, 900)
	`, wordlistID, reportedWordID, reportedDefinition.ID, trackingID)
	require.NoError(t, err)

	err = server.AppContext.ErrorReportService.ReportError(
		ctx,
		service.UnrelatedMeaning,
		reportedWordID,
		&reportedDefinition.ID,
		1,
		nil,
	)
	require.NoError(t, err)

	for _, definitionID := range []int64{reportedDefinition.ID, unrelatedDefinition.ID, sharedDefinition.ID} {
		var linkCount int
		require.NoError(t, server.DB.QueryRow(ctx, `
			SELECT COUNT(*) FROM word_definitions
			WHERE word_id=$1 AND definition_id=$2
		`, reportedWordID, definitionID).Scan(&linkCount))
		assert.Equal(t, 1, linkCount, "definition %d must remain linked until replacement commits", definitionID)
	}

	var preservedBox, trackingCount, quizCount int
	var skipped bool
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*), MAX(box_id), BOOL_OR(temporarily_skipped_until IS NOT NULL)
		FROM leitner_system_tracking
		WHERE id=$1 AND word_id=$2 AND definition_id=$3
	`, trackingID, reportedWordID, reportedDefinition.ID).Scan(&trackingCount, &preservedBox, &skipped))
	assert.Equal(t, 1, trackingCount)
	assert.Equal(t, 5, preservedBox)
	assert.True(t, skipped)
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM quiz_performance
		WHERE id IS NOT NULL AND leitner_system_tracking_id=$1
	`, trackingID).Scan(&quizCount))
	assert.Equal(t, 1, quizCount)

	var reportDefinitionID *int64
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT definition_id FROM error_reports
		WHERE user_id=1 AND word_id=$1 AND status='pending'
	`, reportedWordID).Scan(&reportDefinitionID))
	require.NotNil(t, reportDefinitionID)
	assert.Equal(t, reportedDefinition.ID, *reportDefinitionID)

	var definitionJobs int
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM river_job
		WHERE kind='DefinitionFetcher'
		  AND (args->>'wordId')::bigint=$1
		  AND args->'errorReport' <> 'null'::jsonb
	`, reportedWordID).Scan(&definitionJobs))
	assert.Equal(t, 1, definitionJobs)

	regenerationTx, err := server.DB.Begin(ctx)
	require.NoError(t, err)
	regenerated, err := server.AppContext.DefinitionService.RegenerateDefinitions(
		ctx,
		reportedWordID,
		1,
		&reportedDefinition.ID,
		[]*model.Definition{
			{
				Token:        "reported",
				Language:     "en",
				PartOfSpeech: "verb",
				Meaning:      "provider surplus meaning",
				Examples:     []string{"This result must not be appended."},
				Source:       model.ChatGPT,
			},
			{
				Token:        "reported",
				Language:     "en",
				PartOfSpeech: "noun",
				Meaning:      "unrelated meaning",
				Examples:     []string{"This duplicates an existing meaning."},
				Source:       model.ChatGPT,
			},
			{
				Token:        "reported",
				Language:     "en",
				PartOfSpeech: "noun",
				Meaning:      "replacement meaning",
				Examples:     []string{"A replacement example."},
				Source:       model.ChatGPT,
			},
		},
		regenerationTx,
	)
	require.NoError(t, err)
	require.Len(t, regenerated, 1)
	require.Equal(t, reportedDefinition.ID, regenerated[0].ID, "exclusive regeneration must preserve definition identity")
	require.NoError(t, server.AppContext.LeitnerTrackingService.IncludeDefinitions(
		ctx, reportedWordID, 1, []int64{regenerated[0].ID}, regenerationTx,
	))
	originalReport := service.ErrorReport{DefinitionID: &reportedDefinition.ID, WordID: &reportedWordID, UserID: 1, ErrorType: string(service.UnrelatedMeaning)}
	replacementDefinitionID := regenerated[0].ID
	replacementReport := service.ErrorReport{DefinitionID: &replacementDefinitionID, WordID: &reportedWordID, UserID: 1, ErrorType: string(service.UnrelatedMeaning)}
	require.NoError(t, server.AppContext.LeitnerSystemStrategy.MarkErrorResolvedTx(
		ctx, originalReport, replacementReport, regenerationTx,
	))
	require.NoError(t, regenerationTx.Commit(ctx))

	var regeneratedMeaning string
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT meaning FROM definitions WHERE id=$1`, reportedDefinition.ID).Scan(&regeneratedMeaning))
	assert.Equal(t, "replacement meaning", regeneratedMeaning)
	var linkedDefinitionCount, surplusDefinitionCount int
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM word_definitions WHERE word_id=$1`, reportedWordID).Scan(&linkedDefinitionCount))
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM definitions WHERE meaning='provider surplus meaning'`).Scan(&surplusDefinitionCount))
	assert.Equal(t, 3, linkedDefinitionCount, "targeted regeneration must not append every provider result")
	assert.Zero(t, surplusDefinitionCount)
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*), MAX(box_id), BOOL_OR(temporarily_skipped_until IS NOT NULL)
		FROM leitner_system_tracking WHERE id=$1
	`, trackingID).Scan(&trackingCount, &preservedBox, &skipped))
	assert.Equal(t, 1, trackingCount)
	assert.Equal(t, 5, preservedBox)
	assert.False(t, skipped)
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM quiz_performance
		WHERE leitner_system_tracking_id=$1 AND definition_id=$2
	`, trackingID, reportedDefinition.ID).Scan(&quizCount))
	assert.Equal(t, 1, quizCount)
	var resolvedReports int
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM error_reports
		WHERE user_id=1 AND word_id=$1 AND definition_id=$2 AND status='resolved'
	`, reportedWordID, reportedDefinition.ID).Scan(&resolvedReports))
	assert.Equal(t, 1, resolvedReports)
	pending, err := server.AppContext.LeitnerSystemStrategy.IsErrorPending(ctx, originalReport)
	require.NoError(t, err)
	assert.False(t, pending, "a post-commit retry must stop before provider regeneration")
}

func TestSharedDefinitionRegenerationMovesOnlyReportedRelationshipAndHistory(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	ctx := context.Background()
	token := server.WithTestUser(t)
	wordlistID := createRegenerationWordlist(t, server, token)
	reportedWordID := createRegenerationWord(t, server, token, wordlistID, "shared-report")
	otherWordID := createRegenerationWord(t, server, token, wordlistID, "shared-other")
	sharedDefinition := saveRegenerationDefinition(t, server, reportedWordID, "original shared meaning")
	_, err := server.DB.Exec(ctx, `INSERT INTO word_definitions (word_id, definition_id) VALUES ($1, $2)`, otherWordID, sharedDefinition.ID)
	require.NoError(t, err)

	tx, err := server.DB.Begin(ctx)
	require.NoError(t, err)
	require.NoError(t, server.AppContext.LeitnerTrackingService.IncludeDefinitions(ctx, reportedWordID, 1, []int64{sharedDefinition.ID}, tx))
	require.NoError(t, tx.Commit(ctx))
	var trackingID int64
	require.NoError(t, server.DB.QueryRow(ctx, `
		UPDATE leitner_system_tracking SET box_id=6
		WHERE word_id=$1 AND definition_id=$2 RETURNING id
	`, reportedWordID, sharedDefinition.ID).Scan(&trackingID))
	_, err = server.DB.Exec(ctx, `
		INSERT INTO quiz_performance (
			user_id, wordlist_id, word_id, definition_id,
			leitner_system_tracking_id, quiz_type, box_id, is_correct, response_time_ms
		) VALUES (1, $1, $2, $3, $4, 'WRITE_WORD_FROM_DEFINITION', 6, TRUE, 700)
	`, wordlistID, reportedWordID, sharedDefinition.ID, trackingID)
	require.NoError(t, err)

	require.NoError(t, server.AppContext.ErrorReportService.ReportError(
		ctx, service.UnrelatedMeaning, reportedWordID, &sharedDefinition.ID, 1, nil,
	))

	tx, err = server.DB.Begin(ctx)
	require.NoError(t, err)
	regenerated, err := server.AppContext.DefinitionService.RegenerateDefinitions(
		ctx, reportedWordID, 1, &sharedDefinition.ID,
		[]*model.Definition{{
			Token: "shared-report", Language: "en", PartOfSpeech: "noun",
			Meaning: "private replacement meaning", Examples: []string{"Replacement."}, Source: model.ChatGPT,
		}}, tx,
	)
	require.NoError(t, err)
	require.Len(t, regenerated, 1)
	replacementID := regenerated[0].ID
	require.NotEqual(t, sharedDefinition.ID, replacementID, "shared content must be copied before the reported relationship moves")
	require.NoError(t, server.AppContext.LeitnerTrackingService.IncludeDefinitions(ctx, reportedWordID, 1, []int64{replacementID}, tx))
	originalReport := service.ErrorReport{DefinitionID: &sharedDefinition.ID, WordID: &reportedWordID, UserID: 1, ErrorType: string(service.UnrelatedMeaning)}
	replacementReport := service.ErrorReport{DefinitionID: &replacementID, WordID: &reportedWordID, UserID: 1, ErrorType: string(service.UnrelatedMeaning)}
	require.NoError(t, server.AppContext.LeitnerSystemStrategy.MarkErrorResolvedTx(ctx, originalReport, replacementReport, tx))
	require.NoError(t, tx.Commit(ctx))

	var reportedOldLinks, reportedNewLinks, otherOldLinks int
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM word_definitions WHERE word_id=$1 AND definition_id=$2`, reportedWordID, sharedDefinition.ID).Scan(&reportedOldLinks))
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM word_definitions WHERE word_id=$1 AND definition_id=$2`, reportedWordID, replacementID).Scan(&reportedNewLinks))
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM word_definitions WHERE word_id=$1 AND definition_id=$2`, otherWordID, sharedDefinition.ID).Scan(&otherOldLinks))
	assert.Zero(t, reportedOldLinks)
	assert.Equal(t, 1, reportedNewLinks)
	assert.Equal(t, 1, otherOldLinks)

	var trackingDefinitionID int64
	var boxID int
	var skipped bool
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT definition_id, box_id, temporarily_skipped_until IS NOT NULL
		FROM leitner_system_tracking WHERE id=$1
	`, trackingID).Scan(&trackingDefinitionID, &boxID, &skipped))
	assert.Equal(t, replacementID, trackingDefinitionID)
	assert.Equal(t, 6, boxID)
	assert.False(t, skipped)
	var quizDefinitionID int64
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT definition_id FROM quiz_performance WHERE leitner_system_tracking_id=$1
	`, trackingID).Scan(&quizDefinitionID))
	assert.Equal(t, replacementID, quizDefinitionID)

	var originalMeaning, replacementMeaning string
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT meaning FROM definitions WHERE id=$1`, sharedDefinition.ID).Scan(&originalMeaning))
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT meaning FROM definitions WHERE id=$1`, replacementID).Scan(&replacementMeaning))
	assert.Equal(t, "original shared meaning", originalMeaning)
	assert.Equal(t, "private replacement meaning", replacementMeaning)
}

func TestProcessingFailedRegenerationAddsContentWithoutReplacingExistingLearning(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	ctx := context.Background()
	token := server.WithTestUser(t)
	wordlistID := createRegenerationWordlist(t, server, token)
	wordID := createRegenerationWord(t, server, token, wordlistID, "retry")
	existingDefinition := saveRegenerationDefinition(t, server, wordID, "existing meaning")
	tx, err := server.DB.Begin(ctx)
	require.NoError(t, err)
	require.NoError(t, server.AppContext.LeitnerTrackingService.IncludeDefinitions(ctx, wordID, 1, []int64{existingDefinition.ID}, tx))
	require.NoError(t, tx.Commit(ctx))
	var trackingID int64
	require.NoError(t, server.DB.QueryRow(ctx, `
		UPDATE leitner_system_tracking SET box_id=4
		WHERE word_id=$1 AND definition_id=$2 RETURNING id
	`, wordID, existingDefinition.ID).Scan(&trackingID))

	require.NoError(t, server.AppContext.ErrorReportService.ReportError(
		ctx, service.ProcessingFailed, wordID, nil, 1, nil,
	))
	tx, err = server.DB.Begin(ctx)
	require.NoError(t, err)
	regenerated, err := server.AppContext.DefinitionService.RegenerateDefinitions(
		ctx, wordID, 1, nil,
		[]*model.Definition{{
			Token: "retry", Language: "en", PartOfSpeech: "verb",
			Meaning: "newly generated meaning", Examples: []string{"Generated."}, Source: model.ChatGPT,
		}}, tx,
	)
	require.NoError(t, err)
	require.Len(t, regenerated, 1)
	require.NoError(t, server.AppContext.LeitnerTrackingService.IncludeDefinitions(ctx, wordID, 1, []int64{regenerated[0].ID}, tx))
	report := service.ErrorReport{WordID: &wordID, UserID: 1, ErrorType: string(service.ProcessingFailed)}
	require.NoError(t, server.AppContext.LeitnerSystemStrategy.MarkErrorResolvedTx(ctx, report, report, tx))
	require.NoError(t, tx.Commit(ctx))

	var oldLinkCount, oldBox, newLinkCount, newBox, resolvedCount int
	var oldSkipped bool
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM word_definitions WHERE word_id=$1 AND definition_id=$2
	`, wordID, existingDefinition.ID).Scan(&oldLinkCount))
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT box_id, temporarily_skipped_until IS NOT NULL
		FROM leitner_system_tracking WHERE id=$1
	`, trackingID).Scan(&oldBox, &oldSkipped))
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM word_definitions WHERE word_id=$1 AND definition_id=$2
	`, wordID, regenerated[0].ID).Scan(&newLinkCount))
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT box_id FROM leitner_system_tracking WHERE word_id=$1 AND definition_id=$2
	`, wordID, regenerated[0].ID).Scan(&newBox))
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM error_reports WHERE word_id=$1 AND status='resolved'
	`, wordID).Scan(&resolvedCount))
	assert.Equal(t, 1, oldLinkCount)
	assert.Equal(t, 4, oldBox)
	assert.False(t, oldSkipped)
	assert.Equal(t, 1, newLinkCount)
	assert.Equal(t, 1, newBox)
	assert.Equal(t, 1, resolvedCount)
}

func TestErrorRegenerationRejectsDefinitionFromAnotherWord(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	ctx := context.Background()
	token := server.WithTestUser(t)
	wordlistID := createRegenerationWordlist(t, server, token)
	firstWordID := createRegenerationWord(t, server, token, wordlistID, "first")
	secondWordID := createRegenerationWord(t, server, token, wordlistID, "second")
	foreignDefinition := saveRegenerationDefinition(t, server, secondWordID, "foreign meaning")

	err := server.AppContext.ErrorReportService.ReportError(
		ctx, service.UnrelatedMeaning, firstWordID, &foreignDefinition.ID, 1, nil,
	)
	require.EqualError(t, err, "validation failed")

	var reportCount, cooldownCount int
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM error_reports`).Scan(&reportCount))
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM error_report_cooldowns`).Scan(&cooldownCount))
	assert.Zero(t, reportCount)
	assert.Zero(t, cooldownCount)
}

func TestErrorRegenerationRollsBackEverySideEffectWhenJobInsertionFails(t *testing.T) {
	server := setup.NewTestServer(t, func(builder *app.ContextBuilder) *app.ContextBuilder {
		return builder.WithJobService(failingReportedDefinitionJobService{})
	})
	defer server.Cleanup()

	ctx := context.Background()
	token := server.WithTestUser(t)
	wordlistID := createRegenerationWordlist(t, server, token)
	wordID := createRegenerationWord(t, server, token, wordlistID, "rollback")
	definition := saveRegenerationDefinition(t, server, wordID, "stable meaning")

	tx, err := server.DB.Begin(ctx)
	require.NoError(t, err)
	require.NoError(t, server.AppContext.LeitnerTrackingService.IncludeDefinitions(
		ctx, wordID, 1, []int64{definition.ID}, tx,
	))
	require.NoError(t, tx.Commit(ctx))

	err = server.AppContext.ErrorReportService.ReportError(
		ctx, service.UnrelatedMeaning, wordID, &definition.ID, 1, nil,
	)
	require.ErrorContains(t, err, "forced definition job insertion failure")

	var linkCount, reportCount, cooldownCount, skippedCount, regeneratedCount int
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM word_definitions WHERE word_id=$1 AND definition_id=$2
	`, wordID, definition.ID).Scan(&linkCount))
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM error_reports`).Scan(&reportCount))
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM error_report_cooldowns`).Scan(&cooldownCount))
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM leitner_system_tracking
		WHERE word_id=$1 AND definition_id=$2 AND temporarily_skipped_until IS NOT NULL
	`, wordID, definition.ID).Scan(&skippedCount))
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM definitions WHERE id=$1 AND last_regenerated_at IS NOT NULL
	`, definition.ID).Scan(&regeneratedCount))
	assert.Equal(t, 1, linkCount)
	assert.Zero(t, reportCount)
	assert.Zero(t, cooldownCount)
	assert.Zero(t, skippedCount)
	assert.Zero(t, regeneratedCount)
}

func TestConcurrentDuplicateErrorReportsScheduleOneRegeneration(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	ctx := context.Background()
	token := server.WithTestUser(t)
	wordlistID := createRegenerationWordlist(t, server, token)
	wordID := createRegenerationWord(t, server, token, wordlistID, "concurrent")
	definition := saveRegenerationDefinition(t, server, wordID, "concurrent meaning")

	start := make(chan struct{})
	errorsByRequest := make([]error, 2)
	var wg sync.WaitGroup
	for i := range errorsByRequest {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-start
			errorsByRequest[index] = server.AppContext.ErrorReportService.ReportError(
				ctx, service.UnrelatedMeaning, wordID, &definition.ID, 1, nil,
			)
		}(i)
	}
	close(start)
	wg.Wait()

	var successes, cooldowns int
	for _, reportErr := range errorsByRequest {
		if reportErr == nil {
			successes++
			continue
		}
		var cooldownErr service.CooldownError
		if errors.As(reportErr, &cooldownErr) {
			cooldowns++
			continue
		}
		t.Fatalf("unexpected concurrent report error: %v", reportErr)
	}
	assert.Equal(t, 1, successes)
	assert.Equal(t, 1, cooldowns)

	var jobs, pendingReports int
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM river_job
		WHERE kind='DefinitionFetcher'
		  AND (args->>'wordId')::bigint=$1
		  AND args->'errorReport' <> 'null'::jsonb
	`, wordID).Scan(&jobs))
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM error_reports
		WHERE user_id=1 AND word_id=$1 AND definition_id=$2 AND status='pending'
	`, wordID, definition.ID).Scan(&pendingReports))
	assert.Equal(t, 1, jobs, fmt.Sprintf("errors: %v", errorsByRequest))
	assert.Equal(t, 1, pendingReports)
}

func TestSupersededRegenerationJobStopsBeforeProviderWork(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	ctx := context.Background()
	token := server.WithTestUser(t)
	wordlistID := createRegenerationWordlist(t, server, token)
	wordID := createRegenerationWord(t, server, token, wordlistID, "superseded")
	definition := saveRegenerationDefinition(t, server, wordID, "superseded meaning")
	tx, err := server.DB.Begin(ctx)
	require.NoError(t, err)
	require.NoError(t, server.AppContext.LeitnerTrackingService.IncludeDefinitions(ctx, wordID, 1, []int64{definition.ID}, tx))
	require.NoError(t, tx.Commit(ctx))

	require.NoError(t, server.AppContext.ErrorReportService.ReportError(
		ctx, service.UnrelatedMeaning, wordID, &definition.ID, 1, nil,
	))
	require.NoError(t, server.AppContext.ErrorReportService.ReportError(
		ctx, service.UnrelatedExample, wordID, &definition.ID, 1, nil,
	))

	meaningReport := service.ErrorReport{
		DefinitionID: &definition.ID, WordID: &wordID, UserID: 1,
		ErrorType: string(service.UnrelatedMeaning),
	}
	exampleReport := service.ErrorReport{
		DefinitionID: &definition.ID, WordID: &wordID, UserID: 1,
		ErrorType: string(service.UnrelatedExample),
	}
	meaningPending, err := server.AppContext.LeitnerSystemStrategy.IsErrorPending(ctx, meaningReport)
	require.NoError(t, err)
	examplePending, err := server.AppContext.LeitnerSystemStrategy.IsErrorPending(ctx, exampleReport)
	require.NoError(t, err)
	assert.False(t, meaningPending, "the superseded job must stop before paid generation")
	assert.True(t, examplePending)

	staleTx, err := server.DB.Begin(ctx)
	require.NoError(t, err)
	err = server.AppContext.LeitnerSystemStrategy.MarkErrorResolvedTx(ctx, meaningReport, meaningReport, staleTx)
	require.ErrorIs(t, err, service.ErrErrorReportNotPending)
	require.NoError(t, staleTx.Rollback(ctx))
	var stillSkipped bool
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT temporarily_skipped_until IS NOT NULL
		FROM leitner_system_tracking WHERE word_id=$1 AND definition_id=$2
	`, wordID, definition.ID).Scan(&stillSkipped))
	assert.True(t, stillSkipped, "a stale job must not clear the newer report's temporary skip")

	var pendingRows int
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM error_reports
		WHERE user_id=1 AND word_id=$1 AND definition_id=$2
		  AND error_type=$3 AND status='pending'
	`, wordID, definition.ID, string(service.UnrelatedExample)).Scan(&pendingRows))
	assert.Equal(t, 1, pendingRows)
}

func TestDifferentRegenerationKindsRemainIndependentlyPending(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	ctx := context.Background()
	token := server.WithTestUser(t)
	wordlistID := createRegenerationWordlist(t, server, token)
	wordID := createRegenerationWord(t, server, token, wordlistID, "parallel-kinds")
	definition := saveRegenerationDefinition(t, server, wordID, "parallel meaning")

	require.NoError(t, server.AppContext.ErrorReportService.ReportError(
		ctx, service.UnrelatedMeaning, wordID, &definition.ID, 1, nil,
	))
	require.NoError(t, server.AppContext.ErrorReportService.ReportError(
		ctx, service.MissingImage, wordID, &definition.ID, 1, nil,
	))

	meaningReport := service.ErrorReport{
		DefinitionID: &definition.ID, WordID: &wordID, UserID: 1,
		ErrorType: string(service.UnrelatedMeaning),
	}
	imageReport := service.ErrorReport{
		DefinitionID: &definition.ID, WordID: &wordID, UserID: 1,
		ErrorType: string(service.MissingImage),
	}
	meaningPending, err := server.AppContext.LeitnerSystemStrategy.IsErrorPending(ctx, meaningReport)
	require.NoError(t, err)
	imagePending, err := server.AppContext.LeitnerSystemStrategy.IsErrorPending(ctx, imageReport)
	require.NoError(t, err)
	assert.True(t, meaningPending)
	assert.True(t, imagePending)

	var pendingRows, definitionJobs, imageJobs int
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM error_reports
		WHERE user_id=1 AND word_id=$1 AND definition_id=$2 AND status='pending'
	`, wordID, definition.ID).Scan(&pendingRows))
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM river_job
		WHERE kind='DefinitionFetcher' AND (args->>'wordId')::bigint=$1
		  AND args->'errorReport'->>'errorType'=$2
	`, wordID, string(service.UnrelatedMeaning)).Scan(&definitionJobs))
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM river_job
		WHERE kind='ImageGenerator' AND (args->>'definitionId')::bigint=$1
		  AND args->'errorReport'->>'errorType'=$2
	`, definition.ID, string(service.MissingImage)).Scan(&imageJobs))
	assert.Equal(t, 2, pendingRows)
	assert.Equal(t, 1, definitionJobs)
	assert.Equal(t, 1, imageJobs)
}

func TestErrorRegenerationReturnsCommitFailureAndRollsBack(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	ctx := context.Background()
	token := server.WithTestUser(t)
	wordlistID := createRegenerationWordlist(t, server, token)
	wordID := createRegenerationWord(t, server, token, wordlistID, "commitfail")
	definition := saveRegenerationDefinition(t, server, wordID, "commit failure meaning")

	_, err := server.DB.Exec(ctx, `
		CREATE OR REPLACE FUNCTION test_reject_error_report_commit()
		RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN
			IF EXISTS (SELECT 1 FROM words WHERE id=NEW.word_id AND name='commitfail') THEN
				RAISE EXCEPTION 'forced deferred error-report commit failure';
			END IF;
			RETURN NEW;
		END;
		$$;
		CREATE CONSTRAINT TRIGGER test_error_report_commit_failure
		AFTER INSERT OR UPDATE ON error_reports
		DEFERRABLE INITIALLY DEFERRED
		FOR EACH ROW EXECUTE FUNCTION test_reject_error_report_commit();
	`)
	require.NoError(t, err)
	defer func() {
		_, dropErr := server.DB.Exec(context.Background(), `
			DROP TRIGGER IF EXISTS test_error_report_commit_failure ON error_reports;
			DROP FUNCTION IF EXISTS test_reject_error_report_commit();
		`)
		require.NoError(t, dropErr)
	}()

	err = server.AppContext.ErrorReportService.ReportError(
		ctx, service.UnrelatedMeaning, wordID, &definition.ID, 1, nil,
	)
	require.ErrorContains(t, err, "failed to commit error report transaction")

	var reports, cooldowns, reportedJobs, skipped, regenerated int
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM error_reports WHERE word_id=$1`, wordID).Scan(&reports))
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM error_report_cooldowns WHERE word_id=$1`, wordID).Scan(&cooldowns))
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM river_job
		WHERE kind='DefinitionFetcher' AND (args->>'wordId')::bigint=$1
		  AND args->'errorReport' <> 'null'::jsonb
	`, wordID).Scan(&reportedJobs))
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM leitner_system_tracking
		WHERE word_id=$1 AND temporarily_skipped_until IS NOT NULL
	`, wordID).Scan(&skipped))
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM definitions WHERE id=$1 AND last_regenerated_at IS NOT NULL
	`, definition.ID).Scan(&regenerated))
	assert.Zero(t, reports)
	assert.Zero(t, cooldowns)
	assert.Zero(t, reportedJobs)
	assert.Zero(t, skipped)
	assert.Zero(t, regenerated)
}

func createRegenerationWordlist(t *testing.T, server *setup.TestServer, token string) int64 {
	t.Helper()
	response := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]any{"name": "Regeneration", "languageCode": "en"}).
		Expect().
		Status(http.StatusCreated)
	return int64(response.JSON().Object().Value("id").Number().Raw())
}

func createRegenerationWord(t *testing.T, server *setup.TestServer, token string, wordlistID int64, name string) int64 {
	t.Helper()
	response := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]any{"name": name}).
		Expect().
		Status(http.StatusCreated)
	return int64(response.JSON().Object().Value("id").Number().Raw())
}

func saveRegenerationDefinition(t *testing.T, server *setup.TestServer, wordID int64, meaning string) *model.Definition {
	t.Helper()
	definitions, err := server.AppContext.DefinitionService.SaveDefinition(context.Background(), wordID, []*model.Definition{{
		Token:        "reported",
		Language:     "en",
		PartOfSpeech: "noun",
		Meaning:      meaning,
		Examples:     []string{"A stable example."},
		Source:       model.ChatGPT,
	}}, nil)
	require.NoError(t, err)
	require.Len(t, definitions, 1)
	return definitions[0]
}
