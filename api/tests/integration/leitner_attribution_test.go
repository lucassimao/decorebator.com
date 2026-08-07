package integration

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type lockedLogBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (b *lockedLogBuffer) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(data)
}

func (b *lockedLogBuffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.b.Reset()
}

func (b *lockedLogBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.String()
}

func TestLeitnerAttributionUsesExactWordDefinitionPair(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	ctx := context.Background()
	token := server.WithPremiumUser(t)

	var userID int64
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT id FROM users ORDER BY created_at DESC LIMIT 1`).Scan(&userID))
	foreignWordlistID := createTestWordlist(server, token, "Attribution foreign", "Foreign tracking")
	targetWordlistID := createTestWordlist(server, token, "Attribution target", "Target tracking")

	currentWord := saveAttributionWord(t, server, userID, targetWordlistID, "target-current")
	skippedWord := saveAttributionWord(t, server, userID, targetWordlistID, "target-skipped")
	currentDefinitionID, currentTrackingID := saveAttributionDefinition(t, server, currentWord.ID, userID, "target current")
	_, skippedTrackingID := saveAttributionDefinition(t, server, skippedWord.ID, userID, "target skipped")

	_, err := server.DB.Exec(ctx, `
		UPDATE leitner_system_tracking
		SET box_id=5, next_review_at=NOW(), temporarily_skipped_until=NULL
		WHERE id=$1
	`, currentTrackingID)
	require.NoError(t, err)
	_, err = server.DB.Exec(ctx, `
		UPDATE leitner_system_tracking
		SET box_id=5, next_review_at=NOW(), temporarily_skipped_until=NOW() + INTERVAL '30 minutes'
		WHERE id=$1
	`, skippedTrackingID)
	require.NoError(t, err)

	for i := 0; i < 9; i++ {
		foreignWord := saveAttributionWord(t, server, userID, foreignWordlistID, "foreign-"+string(rune('a'+i)))
		foreignDefinitionID, foreignTrackingID := saveAttributionDefinition(
			t, server, foreignWord.ID, userID, "foreign meaning "+string(rune('a'+i)),
		)
		_, err = server.DB.Exec(ctx, `
			UPDATE leitner_system_tracking
			SET box_id=7, next_review_at=NOW(), temporarily_skipped_until=NULL
			WHERE id=$1
		`, foreignTrackingID)
		require.NoError(t, err)
		_, err = server.DB.Exec(ctx, `
			INSERT INTO word_definitions (word_id, definition_id) VALUES ($1, $2)
		`, currentWord.ID, foreignDefinitionID)
		require.NoError(t, err)
	}

	var logs lockedLogBuffer
	originalLogger := common.Logger
	loggerRestored := false
	// This package intentionally has no parallel tests. Keep this test serial
	// while it temporarily replaces the process-wide logger.
	common.Logger = slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	t.Cleanup(func() {
		if !loggerRestored {
			common.Logger = originalLogger
		}
	})

	quiz, err := server.AppContext.LeitnerSystemStrategy.CreateQuiz(
		ctx, targetWordlistID, userID, []model.QuizType{model.WriteWordFromDefinition},
	)
	require.NoError(t, err)
	require.NotNil(t, quiz)
	assert.NotContains(t, logs.String(), "high_box_7_concentration",
		"foreign tracking rows cross-linked by definition must not inflate target box telemetry")

	logs.Reset()
	_, err = server.DB.Exec(ctx, `
		UPDATE leitner_system_tracking SET next_review_at=NULL
		WHERE word_id IN ($1, $2)
	`, currentWord.ID, skippedWord.ID)
	require.NoError(t, err)
	_, err = server.AppContext.LeitnerSystemStrategy.CreateQuiz(
		ctx, targetWordlistID, userID, []model.QuizType{model.WriteWordFromDefinition},
	)
	require.Error(t, err)
	assert.Contains(t, logs.String(), "totalWords=2",
		"no-due diagnostics must count only tracking rows whose exact word-definition relation exists")
	assert.NotContains(t, logs.String(), "totalWords=11")

	// SaveQuizResult starts a detached analytics update. Restore the process-wide
	// logger before that goroutine can be launched so test cleanup never races it.
	common.Logger = originalLogger
	loggerRestored = true

	_, err = server.DB.Exec(ctx, `
		UPDATE leitner_system_tracking SET next_review_at=NOW() WHERE id=$1;
	`, currentTrackingID)
	require.NoError(t, err)
	_, err = server.DB.Exec(ctx, `
		UPDATE leitner_system_tracking SET next_review_at=NOW(),
			temporarily_skipped_until=NOW() + INTERVAL '30 minutes'
		WHERE id=$1
	`, skippedTrackingID)
	require.NoError(t, err)

	err = server.AppContext.LeitnerSystemStrategy.SaveQuizResult(ctx, common.QuizResult{
		UserID: userID, WordlistID: targetWordlistID, WordID: currentWord.ID,
		DefinitionID: currentDefinitionID, LeitnerSystemTrackingID: currentTrackingID,
		QuizType: model.WriteWordFromDefinition, IsCorrect: false, ResponseTimeMs: 800,
	}, true, nil)
	require.NoError(t, err)

	var skippedUntil *time.Time
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT temporarily_skipped_until FROM leitner_system_tracking WHERE id=$1
	`, skippedTrackingID).Scan(&skippedUntil))
	assert.Nil(t, skippedUntil,
		"foreign available rows must not prevent rescue from clearing the target wordlist's earliest skip")
}

func saveAttributionWord(t *testing.T, server *setup.TestServer, userID, wordlistID int64, name string) *service.Word {
	t.Helper()
	word, err := server.AppContext.WordService.SaveWord(context.Background(), &service.Word{
		Name: name, UserID: userID, WordlistID: wordlistID,
	})
	require.NoError(t, err)
	_, err = server.DB.Exec(context.Background(), `
		UPDATE words SET processing_status='completed', processing_completed_at=NOW() WHERE id=$1
	`, word.ID)
	require.NoError(t, err)
	return word
}

func saveAttributionDefinition(t *testing.T, server *setup.TestServer, wordID, userID int64, meaning string) (int64, int64) {
	t.Helper()
	ctx := context.Background()
	tx, err := server.DB.Begin(ctx)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback(ctx) }()
	definitions, err := server.AppContext.DefinitionService.SaveDefinition(ctx, wordID, []*model.Definition{{
		Token: strings.ReplaceAll(meaning, " ", "-"), Language: "en", PartOfSpeech: "noun",
		Meaning: meaning, Examples: []string{"A valid [example]."}, Source: model.ChatGPT,
	}}, &tx)
	require.NoError(t, err)
	require.Len(t, definitions, 1)
	require.NoError(t, server.AppContext.LeitnerTrackingService.IncludeDefinitions(
		ctx, wordID, userID, []int64{definitions[0].ID}, tx,
	))
	var trackingID int64
	require.NoError(t, tx.QueryRow(ctx, `
		SELECT id FROM leitner_system_tracking
		WHERE user_id=$1 AND word_id=$2 AND definition_id=$3
	`, userID, wordID, definitions[0].ID).Scan(&trackingID))
	require.NoError(t, tx.Commit(ctx))
	return definitions[0].ID, trackingID
}
