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

// TestErrorReporting_ContentSnapshot_VerifyStructure tests that content snapshots
// contain all expected fields and proper data types for different error types
func TestErrorReporting_ContentSnapshot_VerifyStructure(t *testing.T) {
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
			"name":         "Content Test Wordlist",
			"languageCode": "es",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "agua",
		}).
		Expect().
		Status(201)

	wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

	// Create comprehensive definition with all possible fields
	testDefinition := &model.Definition{
		Token:        "agua",
		Language:     "es",
		PartOfSpeech: "sustantivo",
		Meaning:      "Líquido transparente e incoloro que forma mares, lagos y ríos",
		Examples:     []string{"Bebe mucha agua", "El agua del río", "Agua mineral"},
		Source:       "test_comprehensive",
	}

	savedDefinitions, err := definitionService.SaveDefinition(context.Background(), wordID, []*model.Definition{testDefinition}, nil)
	require.NoError(t, err)
	require.Len(t, savedDefinitions, 1)

	definitionID := savedDefinitions[0].ID

	// Submit error report for unrelated meaning (destructive operation)
	server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"wordId":       wordID,
			"definitionId": definitionID,
			"errorType":    "_unrelated_meaning",
		}).
		Expect().
		Status(200)

	// Verify comprehensive content snapshot structure
	var snapshot map[string]interface{}
	err = server.DB.QueryRow(ctx, `
		SELECT content_snapshot FROM error_reports 
		WHERE word_id = $1 ORDER BY reported_at DESC LIMIT 1
	`, wordID).Scan(&snapshot)
	require.NoError(t, err)
	require.NotNil(t, snapshot, "Content snapshot should not be null")

	// Verify all required fields are present
	expectedFields := []string{
		"definition_id", "meaning", "part_of_speech", "source", "language", "token", "examples", "captured_at",
	}
	for _, field := range expectedFields {
		value, exists := snapshot[field]
		assert.True(t, exists, "Snapshot should contain field: %s", field)
		assert.NotNil(t, value, "Snapshot field %s should not be null", field)
	}

	// Verify field types and values
	assert.Equal(t, float64(definitionID), snapshot["definition_id"], "Definition ID should match")
	assert.Equal(t, "Líquido transparente e incoloro que forma mares, lagos y ríos", snapshot["meaning"])
	assert.Equal(t, "sustantivo", snapshot["part_of_speech"])
	assert.Equal(t, "es", snapshot["language"])
	assert.Equal(t, "agua", snapshot["token"])
	assert.Equal(t, "test_comprehensive", snapshot["source"])

	// Verify examples array structure
	examples, ok := snapshot["examples"].([]interface{})
	require.True(t, ok, "Examples should be an array")
	require.Len(t, examples, 3, "Should have 3 examples")
	assert.Equal(t, "Bebe mucha agua", examples[0])
	assert.Equal(t, "El agua del río", examples[1])
	assert.Equal(t, "Agua mineral", examples[2])

	// Verify captured_at timestamp format
	capturedAt, exists := snapshot["captured_at"]
	assert.True(t, exists, "Should have captured_at timestamp")
	assert.IsType(t, "", capturedAt, "captured_at should be a string")

	// Verify it's a valid RFC3339 timestamp
	_, err = time.Parse(time.RFC3339, capturedAt.(string))
	assert.NoError(t, err, "captured_at should be valid RFC3339 timestamp")
}

// TestErrorReporting_SnapshotComparison_DestructiveVsNonDestructive compares snapshots
// between destructive and non-destructive operations to ensure proper foreign key handling
func TestErrorReporting_SnapshotComparison_DestructiveVsNonDestructive(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	// Create services for definition operations
	definitionService := service.NewDefinitionService(server.DB)

	token := server.WithPremiumUser(t)

	// Test destructive operation (unrelated meaning)
	createResp1 := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Destructive Test",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID1 := int(createResp1.JSON().Object().Value("id").Number().Raw())

	addWordResp1 := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID1).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "destructive",
		}).
		Expect().
		Status(201)

	wordID1 := int64(addWordResp1.JSON().Object().Value("id").Number().Raw())

	testDefinition1 := &model.Definition{
		Token:        "destructive",
		Language:     "en",
		PartOfSpeech: "adjective",
		Meaning:      "Causing damage or harm",
		Examples:     []string{"A destructive force", "Destructive behavior"},
		Source:       "test_destructive",
	}

	savedDefinitions1, err := definitionService.SaveDefinition(context.Background(), wordID1, []*model.Definition{testDefinition1}, nil)
	require.NoError(t, err)
	definitionID1 := savedDefinitions1[0].ID

	// Submit destructive error report
	server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"wordId":       wordID1,
			"definitionId": definitionID1,
			"errorType":    "_unrelated_meaning",
		}).
		Expect().
		Status(200)

	// Test non-destructive operation (unrelated image)
	createResp2 := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Non-Destructive Test",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID2 := int(createResp2.JSON().Object().Value("id").Number().Raw())

	addWordResp2 := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID2).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "nondestructive",
		}).
		Expect().
		Status(201)

	wordID2 := int64(addWordResp2.JSON().Object().Value("id").Number().Raw())

	testDefinition2 := &model.Definition{
		Token:        "nondestructive",
		Language:     "en",
		PartOfSpeech: "adjective",
		Meaning:      "Not causing damage or harm",
		Examples:     []string{"A nondestructive test", "Nondestructive evaluation"},
		Source:       "test_nondestructive",
	}

	savedDefinitions2, err := definitionService.SaveDefinition(context.Background(), wordID2, []*model.Definition{testDefinition2}, nil)
	require.NoError(t, err)
	definitionID2 := savedDefinitions2[0].ID

	// Submit non-destructive error report
	server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"wordId":       wordID2,
			"definitionId": definitionID2,
			"errorType":    "_unrelated_image",
		}).
		Expect().
		Status(200)

	// Verify both snapshots have similar structure but different foreign key handling
	ctx := context.Background()

	var snapshot1, snapshot2 map[string]interface{}
	err = server.DB.QueryRow(ctx, `
		SELECT content_snapshot FROM error_reports 
		WHERE word_id = $1 ORDER BY reported_at DESC LIMIT 1
	`, wordID1).Scan(&snapshot1)
	require.NoError(t, err)
	require.NotNil(t, snapshot1, "Destructive snapshot should not be null")

	err = server.DB.QueryRow(ctx, `
		SELECT content_snapshot FROM error_reports 
		WHERE word_id = $1 ORDER BY reported_at DESC LIMIT 1
	`, wordID2).Scan(&snapshot2)
	require.NoError(t, err)
	require.NotNil(t, snapshot2, "Non-destructive snapshot should not be null")

	// Both should have the same structure
	for _, field := range []string{"definition_id", "meaning", "part_of_speech", "language", "token", "examples", "captured_at"} {
		assert.Contains(t, snapshot1, field, "Destructive snapshot should contain %s", field)
		assert.Contains(t, snapshot2, field, "Non-destructive snapshot should contain %s", field)
	}

	// Verify foreign key handling differences
	assertDefinitionIDNullified(t, server.DB, wordID1)                // Destructive - NULL
	assertDefinitionIDPreserved(t, server.DB, wordID2, definitionID2) // Non-destructive - preserved

	// Verify definition existence differences
	assertDefinitionDeleted(t, server.DB, definitionID1)    // Destructive - deleted
	assertDefinitionNotDeleted(t, server.DB, definitionID2) // Non-destructive - preserved
}

// TestErrorReporting_QuizContext_SnapshotIncludesQuizDetails tests that quiz context
// is properly included in error report snapshots when provided
func TestErrorReporting_QuizContext_SnapshotIncludesQuizDetails(t *testing.T) {
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
			"name":         "Quiz Context Test",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "example",
		}).
		Expect().
		Status(201)

	wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

	testDefinition := &model.Definition{
		Token:        "example",
		Language:     "en",
		PartOfSpeech: "noun",
		Meaning:      "A thing characteristic of its kind or illustrating a general rule",
		Examples:     []string{"An example of good behavior", "For example"},
		Source:       "test_quiz_context",
	}

	savedDefinitions, err := definitionService.SaveDefinition(context.Background(), wordID, []*model.Definition{testDefinition}, nil)
	require.NoError(t, err)
	require.Len(t, savedDefinitions, 1)

	definitionID := savedDefinitions[0].ID

	// Submit error report with quiz context
	server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"wordId":       wordID,
			"definitionId": definitionID,
			"errorType":    "_unrelated_meaning",
			"quizDetails": map[string]interface{}{
				"quizType":    "GUESS_MEANING",
				"value":       "example",
				"options":     []string{"sample", "instance", "model", "illustration"},
				"answerIndex": 1,
				"context":     "quiz",
			},
		}).
		Expect().
		Status(200)

	// Verify quiz context is included in content snapshot
	var snapshot map[string]interface{}
	err = server.DB.QueryRow(ctx, `
		SELECT content_snapshot FROM error_reports 
		WHERE word_id = $1 ORDER BY reported_at DESC LIMIT 1
	`, wordID).Scan(&snapshot)
	require.NoError(t, err)
	require.NotNil(t, snapshot, "Content snapshot should not be null")

	// Verify quiz context is present in snapshot
	quizContext, exists := snapshot["quiz_context"]
	assert.True(t, exists, "Snapshot should contain quiz_context")
	require.NotNil(t, quizContext, "Quiz context should not be null")

	quizContextMap, ok := quizContext.(map[string]interface{})
	require.True(t, ok, "Quiz context should be a map")

	// Verify all quiz context fields
	assert.Equal(t, "GUESS_MEANING", quizContextMap["quiz_type"], "Quiz type should match")
	assert.Equal(t, "example", quizContextMap["value"], "Quiz value should match")
	assert.Equal(t, "quiz", quizContextMap["context"], "Context should match")
	assert.Equal(t, float64(1), quizContextMap["answer_index"], "Answer index should match")

	// Verify options array
	options, optionsOk := quizContextMap["options"].([]interface{})
	require.True(t, optionsOk, "Options should be an array")
	require.Len(t, options, 4, "Should have 4 options")
	assert.Equal(t, "sample", options[0])
	assert.Equal(t, "instance", options[1])
	assert.Equal(t, "model", options[2])
	assert.Equal(t, "illustration", options[3])

	// Verify captured_at timestamp in quiz context
	capturedAt, exists := quizContextMap["captured_at"]
	assert.True(t, exists, "Quiz context should have captured_at timestamp")
	assert.IsType(t, "", capturedAt, "captured_at should be a string")

	// Verify it's a valid RFC3339 timestamp
	_, err = time.Parse(time.RFC3339, capturedAt.(string))
	assert.NoError(t, err, "captured_at should be valid RFC3339 timestamp")
}

// TestErrorReporting_FlashcardContext_SnapshotIncludesFlashcardDetails tests that flashcard context
// is properly included in error report snapshots when provided
func TestErrorReporting_FlashcardContext_SnapshotIncludesFlashcardDetails(t *testing.T) {
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
			"name":         "Flashcard Context Test",
			"languageCode": "de",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "Beispiel",
		}).
		Expect().
		Status(201)

	wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

	testDefinition := &model.Definition{
		Token:        "Beispiel",
		Language:     "de",
		PartOfSpeech: "Substantiv",
		Meaning:      "Ein Muster oder eine Probe von etwas",
		Examples:     []string{"Ein gutes Beispiel", "Zum Beispiel"},
		Source:       "test_flashcard_context",
	}

	savedDefinitions, err := definitionService.SaveDefinition(context.Background(), wordID, []*model.Definition{testDefinition}, nil)
	require.NoError(t, err)
	_ = savedDefinitions[0].ID // Definition ID not needed for audio errors

	// Submit error report with flashcard context (audio error)
	server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"wordId":    wordID,
			"errorType": "_sound_not_playing",
			"quizDetails": map[string]interface{}{
				"quizType":    "FLASHCARD_VIEW",
				"value":       "Beispiel",
				"options":     []string{"Ein Muster oder eine Probe von etwas"},
				"answerIndex": 0,
				"context":     "flashcard",
			},
		}).
		Expect().
		Status(200)

	// Verify flashcard context is included in content snapshot
	var snapshot map[string]interface{}
	err = server.DB.QueryRow(ctx, `
		SELECT content_snapshot FROM error_reports 
		WHERE word_id = $1 ORDER BY reported_at DESC LIMIT 1
	`, wordID).Scan(&snapshot)
	require.NoError(t, err)
	require.NotNil(t, snapshot, "Content snapshot should not be null")

	// Verify flashcard context is present in snapshot
	quizContext, exists := snapshot["quiz_context"]
	assert.True(t, exists, "Snapshot should contain quiz_context")
	require.NotNil(t, quizContext, "Quiz context should not be null")

	quizContextMap, ok := quizContext.(map[string]interface{})
	require.True(t, ok, "Quiz context should be a map")

	// Verify flashcard-specific values
	assert.Equal(t, "FLASHCARD_VIEW", quizContextMap["quiz_type"], "Quiz type should be FLASHCARD_VIEW")
	assert.Equal(t, "Beispiel", quizContextMap["value"], "Value should be word name")
	assert.Equal(t, "flashcard", quizContextMap["context"], "Context should be flashcard")
	assert.Equal(t, float64(0), quizContextMap["answer_index"], "Answer index should be 0 for flashcards")

	// Verify options contain definition meaning
	options, optionsOk := quizContextMap["options"].([]interface{})
	require.True(t, optionsOk, "Options should be an array")
	require.Len(t, options, 1, "Should have 1 option for flashcards")
	assert.Equal(t, "Ein Muster oder eine Probe von etwas", options[0])
}
