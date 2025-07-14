package errorreporting

import (
	"context"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestErrorReporting_SideEffects_TimestampUpdates tests that last_regenerated_at timestamps
// are properly updated for different error types
func TestErrorReporting_SideEffects_TimestampUpdates(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	// Create services for definition operations
	definitionService := service.NewDefinitionService(server.DB)

	token := server.WithTestUser(t)
	ctx := context.Background()

	// Create shared wordlist for both subtests (free plan allows only 1 wordlist)
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Timestamp Test",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	// Test definition timestamp update for image errors
	t.Run("DefinitionTimestampUpdate", func(t *testing.T) {
		// Create word in shared wordlist
		addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
			WithHeader("Authorization", token).
			WithJSON(map[string]interface{}{
				"name": "timestamp",
			}).
			Expect().
			Status(201)

		wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

		// Create definition
		testDefinition := &model.Definition{
			Token:        "timestamp",
			Language:     "en",
			PartOfSpeech: "noun",
			Meaning:      "A record of when something occurred",
			Examples:     []string{"The timestamp shows when the error occurred"},
			Source:       "test_timestamp",
		}

		savedDefinitions, err := definitionService.SaveDefinition(wordID, []*model.Definition{testDefinition}, nil)
		require.NoError(t, err)
		require.Len(t, savedDefinitions, 1)

		definitionID := savedDefinitions[0].ID

		// Check timestamp before error report
		var timestampBefore *time.Time
		err = server.DB.QueryRow(ctx, "SELECT last_regenerated_at FROM definitions WHERE id = $1", definitionID).Scan(&timestampBefore)
		require.NoError(t, err)

		// Submit error report for unrelated image (should update definition timestamp)
		server.Expect.POST("/errorReports").
			WithHeader("Authorization", token).
			WithJSON(map[string]interface{}{
				"wordId":       wordID,
				"definitionId": definitionID,
				"errorType":    "_unrelated_image",
			}).
			Expect().
			Status(200)

		// Check timestamp after error report
		var timestampAfter *time.Time
		err = server.DB.QueryRow(ctx, "SELECT last_regenerated_at FROM definitions WHERE id = $1", definitionID).Scan(&timestampAfter)
		require.NoError(t, err)

		// Verify timestamp was updated
		if timestampBefore == nil {
			assert.NotNil(t, timestampAfter, "Timestamp should be set after error report")
		} else {
			assert.True(t, timestampAfter.After(*timestampBefore), "Timestamp should be updated after error report")
		}

		// Verify that River job was triggered for image regeneration
		var jobCount int
		err = server.DB.QueryRow(ctx, `
			SELECT COUNT(*) FROM river_job 
			WHERE kind = 'ImageGenerator' 
			AND queue = 'image_generator'
			AND (args->>'definitionId')::bigint = $1
		`, definitionID).Scan(&jobCount)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, jobCount, 1, "Image generator job should be triggered for _unrelated_image error")
	})

	// Test word timestamp update for audio errors
	t.Run("WordTimestampUpdate", func(t *testing.T) {
		// Create another word in same shared wordlist (reusing wordlist due to free plan limit)
		addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
			WithHeader("Authorization", token).
			WithJSON(map[string]interface{}{
				"name": "audiotimestamp",
			}).
			Expect().
			Status(201)

		wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

		// Check timestamp before error report
		var timestampBefore *time.Time
		err := server.DB.QueryRow(ctx, "SELECT audio_last_regenerated_at FROM words WHERE id = $1", wordID).Scan(&timestampBefore)
		require.NoError(t, err)

		// Submit error report for sound not playing (should update word audio timestamp)
		server.Expect.POST("/errorReports").
			WithHeader("Authorization", token).
			WithJSON(map[string]interface{}{
				"wordId":    wordID,
				"errorType": "_sound_not_playing",
			}).
			Expect().
			Status(200)

		// Check timestamp after error report
		var timestampAfter *time.Time
		err = server.DB.QueryRow(ctx, "SELECT audio_last_regenerated_at FROM words WHERE id = $1", wordID).Scan(&timestampAfter)
		require.NoError(t, err)

		// Verify timestamp was updated
		if timestampBefore == nil {
			assert.NotNil(t, timestampAfter, "Audio timestamp should be set after error report")
		} else {
			assert.True(t, timestampAfter.After(*timestampBefore), "Audio timestamp should be updated after error report")
		}

		// Verify that River job was triggered for text-to-speech regeneration
		var jobCount int
		err = server.DB.QueryRow(ctx, `
			SELECT COUNT(*) FROM river_job 
			WHERE kind = 'TextToSpeech' 
			AND queue = 'text_to_speech'
			AND (args->>'wordId')::bigint = $1
		`, wordID).Scan(&jobCount)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, jobCount, 1, "Text-to-speech job should be triggered for _sound_not_playing error")
	})
}

// TestErrorReporting_SideEffects_LeitnerSystemUpdate tests that error reports
// properly interact with the Leitner system for temporary skipping
func TestErrorReporting_SideEffects_LeitnerSystemUpdate(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	// Create services for definition operations
	definitionService := service.NewDefinitionService(server.DB)

	token := server.WithTestUser(t)
	ctx := context.Background()

	// Create wordlist and word
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Leitner Test",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "leitner",
		}).
		Expect().
		Status(201)

	wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

	// Create definition
	testDefinition := &model.Definition{
		Token:        "leitner",
		Language:     "en",
		PartOfSpeech: "noun",
		Meaning:      "A spaced repetition system",
		Examples:     []string{"The Leitner system helps with memorization"},
		Source:       "test_leitner",
	}

	savedDefinitions, err := definitionService.SaveDefinition(wordID, []*model.Definition{testDefinition}, nil)
	require.NoError(t, err)
	require.Len(t, savedDefinitions, 1)

	definitionID := savedDefinitions[0].ID

	// Add definitions to Leitner system tracking (normally done by word creation flow)
	// Create strategy with proper dependencies
	moderationService := service.NewOpenAIModerationService()
	wordService := service.NewWordService(server.DB, moderationService)
	leitnerStrategy := service.NewLeitnerSystemStrategy(wordService, definitionService)
	tx, txErr := server.DB.Begin(ctx)
	require.NoError(t, txErr)
	defer tx.Rollback(ctx)

	err = leitnerStrategy.IncludeDefinitions(wordID, 1, []int64{definitionID}, tx) // userID = 1 from test setup
	require.NoError(t, err)
	err = tx.Commit(ctx)
	require.NoError(t, err)

	// Verify Leitner tracking exists before error report
	var trackingExistsBefore bool
	err = server.DB.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM leitner_system_tracking 
			WHERE user_id = (SELECT user_id FROM words WHERE id = $1) 
			AND word_id = $1
		)
	`, wordID).Scan(&trackingExistsBefore)
	require.NoError(t, err)
	assert.True(t, trackingExistsBefore, "Leitner system tracking should exist before error report")

	// Submit error report (destructive type that deletes definitions)
	server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"wordId":       wordID,
			"definitionId": definitionID,
			"errorType":    "_unrelated_meaning",
		}).
		Expect().
		Status(200)

	// Verify that Leitner system tracking was cleaned up (CASCADE delete when definitions are deleted)
	var trackingExistsAfter bool
	err = server.DB.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM leitner_system_tracking 
			WHERE user_id = (SELECT user_id FROM words WHERE id = $1) 
			AND word_id = $1
		)
	`, wordID).Scan(&trackingExistsAfter)
	require.NoError(t, err)
	assert.False(t, trackingExistsAfter, "Leitner system tracking should be deleted after destructive error report due to CASCADE constraint")

	// Verify that River job was triggered for definition fetching (to regenerate deleted definitions)
	var jobCount int
	err = server.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM river_job 
		WHERE kind = 'DefinitionFetcher' 
		AND queue = 'definition_fetcher'
		AND (args->>'wordId')::bigint = $1
	`, wordID).Scan(&jobCount)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, jobCount, 1, "Definition fetcher job should be triggered for _unrelated_meaning error")
}

// TestErrorReporting_SideEffects_CooldownCreation tests that cooldown periods
// are properly set after error reports
func TestErrorReporting_SideEffects_CooldownCreation(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	// Create services for definition operations
	definitionService := service.NewDefinitionService(server.DB)

	token := server.WithTestUser(t)
	ctx := context.Background()

	// Create wordlist and word
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Cooldown Side Effect Test",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "cooldown",
		}).
		Expect().
		Status(201)

	wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

	// Create definition
	testDefinition := &model.Definition{
		Token:        "cooldown",
		Language:     "en",
		PartOfSpeech: "noun",
		Meaning:      "Testing cooldown side effects",
		Examples:     []string{"Cooldown prevents spam"},
		Source:       "test_cooldown_side_effect",
	}

	savedDefinitions, err := definitionService.SaveDefinition(wordID, []*model.Definition{testDefinition}, nil)
	require.NoError(t, err)
	require.Len(t, savedDefinitions, 1)

	definitionID := savedDefinitions[0].ID

	// Verify no cooldown exists before error report
	var cooldownBefore *time.Time
	err = server.DB.QueryRow(ctx, `
		SELECT cooldown_until FROM error_report_cooldowns 
		WHERE user_id = (SELECT user_id FROM words WHERE id = $1) 
		AND word_id = $1 AND definition_id = $2 AND error_type = '_unrelated_meaning'
	`, wordID, definitionID).Scan(&cooldownBefore)
	assert.Error(t, err, "No cooldown should exist before error report")

	// Submit error report
	server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"wordId":       wordID,
			"definitionId": definitionID,
			"errorType":    "_unrelated_meaning",
		}).
		Expect().
		Status(200)

	// Verify cooldown was created
	var cooldownAfter time.Time
	err = server.DB.QueryRow(ctx, `
		SELECT cooldown_until FROM error_report_cooldowns 
		WHERE user_id = (SELECT user_id FROM words WHERE id = $1) 
		AND word_id = $1 AND definition_id = $2 AND error_type = '_unrelated_meaning'
	`, wordID, definitionID).Scan(&cooldownAfter)
	require.NoError(t, err, "Cooldown should exist after error report")

	// Verify cooldown is set for approximately 1 hour in the future
	expectedCooldownTime := time.Now().Add(time.Hour)
	timeDiff := cooldownAfter.Sub(expectedCooldownTime).Abs()
	assert.Less(t, timeDiff, time.Minute*5, "Cooldown should be set for approximately 1 hour from now")
}

// TestErrorReporting_SideEffects_DatabaseConsistency tests that all database
// operations in error reporting maintain consistency
func TestErrorReporting_SideEffects_DatabaseConsistency(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	// Create services for definition operations
	definitionService := service.NewDefinitionService(server.DB)

	token := server.WithTestUser(t)
	ctx := context.Background()

	// Create wordlist and word for destructive operation test
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Consistency Test",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "consistency",
		}).
		Expect().
		Status(201)

	wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

	// Create definition
	testDefinition := &model.Definition{
		Token:        "consistency",
		Language:     "en",
		PartOfSpeech: "noun",
		Meaning:      "The quality of being consistent",
		Examples:     []string{"Database consistency is important"},
		Source:       "test_consistency",
	}

	savedDefinitions, err := definitionService.SaveDefinition(wordID, []*model.Definition{testDefinition}, nil)
	require.NoError(t, err)
	require.Len(t, savedDefinitions, 1)

	definitionID := savedDefinitions[0].ID

	// Count related records before error report
	var definitionCount, wordDefinitionCount, errorReportCount int
	err = server.DB.QueryRow(ctx, "SELECT COUNT(*) FROM definitions WHERE id = $1", definitionID).Scan(&definitionCount)
	require.NoError(t, err)
	err = server.DB.QueryRow(ctx, "SELECT COUNT(*) FROM word_definitions WHERE definition_id = $1", definitionID).Scan(&wordDefinitionCount)
	require.NoError(t, err)
	err = server.DB.QueryRow(ctx, "SELECT COUNT(*) FROM error_reports WHERE word_id = $1", wordID).Scan(&errorReportCount)
	require.NoError(t, err)

	// Verify initial state
	assert.Equal(t, 1, definitionCount, "Should have 1 definition initially")
	assert.Equal(t, 1, wordDefinitionCount, "Should have 1 word_definition initially")
	assert.Equal(t, 0, errorReportCount, "Should have 0 error reports initially")

	// Submit destructive error report
	server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"wordId":       wordID,
			"definitionId": definitionID,
			"errorType":    "_unrelated_meaning",
		}).
		Expect().
		Status(200)

	// Count related records after error report
	err = server.DB.QueryRow(ctx, "SELECT COUNT(*) FROM definitions WHERE id = $1", definitionID).Scan(&definitionCount)
	require.NoError(t, err)
	err = server.DB.QueryRow(ctx, "SELECT COUNT(*) FROM word_definitions WHERE definition_id = $1", definitionID).Scan(&wordDefinitionCount)
	require.NoError(t, err)
	err = server.DB.QueryRow(ctx, "SELECT COUNT(*) FROM error_reports WHERE word_id = $1", wordID).Scan(&errorReportCount)
	require.NoError(t, err)

	// Verify final state after destructive operation
	assert.Equal(t, 0, definitionCount, "Definition should be deleted after destructive error report")
	assert.Equal(t, 0, wordDefinitionCount, "Word_definition should be deleted after destructive error report")
	assert.Equal(t, 1, errorReportCount, "Should have 1 error report after submission")

	// Verify error report has content snapshot even though definition is deleted
	var hasSnapshot bool
	err = server.DB.QueryRow(ctx, "SELECT content_snapshot IS NOT NULL FROM error_reports WHERE word_id = $1", wordID).Scan(&hasSnapshot)
	require.NoError(t, err)
	assert.True(t, hasSnapshot, "Error report should have content snapshot even after definition deletion")
}
