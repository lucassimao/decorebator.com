package analytics

import (
	"context"
	"fmt"
	"testing"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository/analytics"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// createBasicWordStructure creates a word, definition, and leitner tracking for testing
// Now uses proper service layer calls instead of direct database inserts
func createBasicWordStructure(t *testing.T, server *setup.TestServer, token string, wordlistID int64, wordName string) (int64, int64, int64) {
	// Create word via API
	wordResp := server.Expect.POST(fmt.Sprintf("/wordlists/%d/words", wordlistID)).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":  wordName,
			"notes": "Test word for analytics",
		}).
		Expect().
		Status(201)

	wordJSON := wordResp.JSON().Object()
	wordID := int64(wordJSON.Value("id").Number().Raw())

	ctx := context.Background()

	// Get user ID first
	userID := getUserID(ctx, t, server.DB)

	// Verify the word exists and belongs to the user
	var wordExists bool
	err := server.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM words WHERE id = $1 AND user_id = $2)",
		wordID, userID).Scan(&wordExists)
	require.NoError(t, err)
	require.True(t, wordExists, "Word should exist in database")

	// Start a transaction for the service calls
	tx, err := server.DB.Begin(ctx)
	require.NoError(t, err)
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	// Create definition using SaveDefinition service call
	definition := &model.Definition{
		Token:                  wordName,
		Language:               "en",
		Meaning:                "Test meaning for " + wordName,
		PartOfSpeech:           "noun",
		Examples:               []string{"Test example with " + wordName},
		Source:                 "test",
		PartOfSpeechNormalized: "noun",
	}

	definitions, err := service.SaveDefinition(wordID, []*model.Definition{definition}, &tx)
	require.NoError(t, err)
	require.Len(t, definitions, 1)

	definitionID := definitions[0].ID

	// Use IncludeDefinitions service call to add to Leitner system
	strategy := service.LeitnerSystemStrategy{}
	err = strategy.IncludeDefinitions(wordID, userID, []int64{definitionID}, tx)
	require.NoError(t, err)

	// Get the leitner tracking ID that was created
	var leitnerTrackingID int64
	err = tx.QueryRow(ctx,
		`SELECT id FROM leitner_system_tracking WHERE word_id = $1 AND definition_id = $2`,
		wordID, definitionID).Scan(&leitnerTrackingID)
	require.NoError(t, err)

	return wordID, definitionID, leitnerTrackingID
}

// getUserID retrieves the most recently created user ID (from WithTestUser)
func getUserID(ctx context.Context, t *testing.T, db *pgxpool.Pool) int64 {
	var userID int64
	err := db.QueryRow(ctx, "SELECT id FROM users ORDER BY created_at DESC LIMIT 1").Scan(&userID)
	require.NoError(t, err)
	return userID
}

// createTestWordlist creates a basic wordlist for testing via API
func createTestWordlist(_ *testing.T, server *setup.TestServer, token string, name, description string) int64 {
	// Create wordlist via API
	wordlistResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":                name,
			"description":         description,
			"languageCode":        "en",
			"pronunciationSystem": "ipa",
		}).
		Expect().
		Status(201)

	wordlistJSON := wordlistResp.JSON().Object()
	return int64(wordlistJSON.Value("id").Number().Raw())
}

// saveQuizResult submits a quiz result via API
func saveQuizResult(_ *testing.T, server *setup.TestServer, token string, wordlistID, wordID, definitionID, leitnerTrackingID int64, quizType string, isCorrect bool, responseTimeMs int) {
	server.Expect.PATCH(fmt.Sprintf("/wordlists/%d/quizzes", wordlistID)).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"wordlistID":              wordlistID,
			"wordID":                  wordID,
			"definitionID":            definitionID,
			"leitnerSystemTrackingID": leitnerTrackingID,
			"quizType":                quizType,
			"isCorrect":               isCorrect,
			"responseTimeMs":          responseTimeMs,
		}).
		Expect().
		Status(204)
}

// flushRedisCache clears all Redis cache to ensure test isolation
func flushRedisCache(t *testing.T) {
	ctx := context.Background()

	// Get Redis client to clear cache
	redisClient, err := common.GetRedisClient()
	if err != nil {
		// Redis not available in test environment, log and continue
		t.Logf("Redis not available for cache flush: %v", err)
		return
	}

	// Use FLUSHALL to clear entire Redis cache for complete test isolation
	err = redisClient.FlushAll(ctx).Err()
	if err != nil {
		t.Logf("Failed to flush Redis cache: %v", err)
	} else {
		t.Logf("Flushed all Redis cache for test isolation")
	}
}

// createWordMasteryData creates word mastery data using the proper repository call
func createWordMasteryData(t *testing.T, server *setup.TestServer, userID, wordID int64, boxID int64, isCorrect bool, masteryLevel float64, totalAttempts, correctAttempts, streakCount, maxStreak int) {
	ctx := context.Background()

	// Start a transaction
	tx, err := server.DB.Begin(ctx)
	require.NoError(t, err)
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	// Create analytics repository to access UpsertWordMastery
	analyticsRepo := analytics.NewWordMasteryRepository(server.DB)

	// Use the repository method to create word mastery data
	err = analyticsRepo.UpsertWordMastery(ctx, tx, userID, wordID, boxID, isCorrect)
	require.NoError(t, err)

	// For tests that need specific mastery levels, we need to update with custom values
	if masteryLevel > 0 {
		_, err = tx.Exec(ctx,
			`UPDATE word_mastery SET 
				mastery_level = $1, 
				total_attempts = $2, 
				correct_attempts = $3, 
				streak_count = $4, 
				max_streak = $5,
				last_seen_at = NOW() - INTERVAL '1 hour'
			WHERE user_id = $6 AND word_id = $7`,
			masteryLevel, totalAttempts, correctAttempts, streakCount, maxStreak, userID, wordID)
		require.NoError(t, err)
	}
}
