package integration

import (
	"context"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestErrorReporting_UnrelatedMeaning_BasicFlow tests the basic error reporting flow
// for unrelated meaning errors, which should delete definitions and create content snapshots
func TestErrorReporting_UnrelatedMeaning_BasicFlow(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)
	ctx := context.Background()

	// Create wordlist via API
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Test Wordlist",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	// Add word via API
	addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "water",
		}).
		Expect().
		Status(201)

	wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

	// Create definition with examples using service layer
	testDefinition := &model.Definition{
		Token:        "water",
		Language:     "en",
		PartOfSpeech: "noun",
		Meaning:      "A transparent, tasteless, odorless chemical substance",
		Examples:     []string{"Drink plenty of water", "Water the plants", "Still water"},
		Source:       "test",
	}

	// Use SaveDefinition service function to create the definition
	savedDefinitions, err := service.SaveDefinition(wordID, []*model.Definition{testDefinition}, nil)
	require.NoError(t, err, "Failed to create test definition")
	require.Len(t, savedDefinitions, 1, "Should save exactly one definition")
	
	definitionID := savedDefinitions[0].ID

	// Verify definition exists before error report
	var definitionExists bool
	err = server.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM definitions WHERE id = $1)", definitionID).Scan(&definitionExists)
	require.NoError(t, err)
	assert.True(t, definitionExists, "Definition should exist before error report")

	// Submit error report for unrelated meaning
	server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"wordId":       wordID,
			"definitionId": definitionID,
			"errorType":    "_unrelated_meaning",
		}).
		Expect().
		Status(200)

	// Verify error report was created
	assertErrorReportExists(t, server.DB, wordID, definitionID, "_unrelated_meaning")

	// Verify content snapshot was created and contains definition data
	assertContentSnapshotContains(t, server.DB, wordID, map[string]interface{}{
		"definition_id":  float64(definitionID), // JSON numbers are float64
		"meaning":        "A transparent, tasteless, odorless chemical substance",
		"part_of_speech": "noun",
		"language":       "en",
	})

	// Verify definition was deleted (destructive operation)
	assertDefinitionDeleted(t, server.DB, definitionID)

	// Verify foreign key was nullified for destructive operation
	assertDefinitionIDNullified(t, server.DB, wordID)
}

// Helper functions for assertions

func assertErrorReportExists(t *testing.T, db *pgxpool.Pool, wordID, definitionID int64, errorType string) {
	ctx := context.Background()
	var count int
	err := db.QueryRow(ctx, `
		SELECT COUNT(*) FROM error_reports 
		WHERE word_id = $1 AND error_type = $2
	`, wordID, errorType).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Error report should exist")
}

func assertContentSnapshotContains(t *testing.T, db *pgxpool.Pool, wordID int64, expectedFields map[string]interface{}) {
	ctx := context.Background()
	var snapshot map[string]interface{}
	err := db.QueryRow(ctx, `
		SELECT content_snapshot FROM error_reports 
		WHERE word_id = $1 ORDER BY reported_at DESC LIMIT 1
	`, wordID).Scan(&snapshot)
	require.NoError(t, err)
	require.NotNil(t, snapshot, "Content snapshot should not be null")

	for key, expectedValue := range expectedFields {
		actualValue, exists := snapshot[key]
		assert.True(t, exists, "Snapshot should contain key: %s", key)
		assert.Equal(t, expectedValue, actualValue, "Snapshot field %s should match expected value", key)
	}

	// Verify captured_at timestamp exists
	_, exists := snapshot["captured_at"]
	assert.True(t, exists, "Snapshot should contain captured_at timestamp")
}

func assertDefinitionDeleted(t *testing.T, db *pgxpool.Pool, definitionID int64) {
	ctx := context.Background()
	var count int
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM definitions WHERE id = $1", definitionID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Definition should be deleted")
}

func assertDefinitionIDNullified(t *testing.T, db *pgxpool.Pool, wordID int64) {
	ctx := context.Background()
	var definitionID *int64
	err := db.QueryRow(ctx, `
		SELECT definition_id FROM error_reports 
		WHERE word_id = $1 ORDER BY reported_at DESC LIMIT 1
	`, wordID).Scan(&definitionID)
	require.NoError(t, err)
	assert.Nil(t, definitionID, "Definition ID should be NULL for destructive operations")
}

// TestErrorReporting_ContentSnapshot_VerifyStructure tests that content snapshots
// contain all expected fields and proper data types for different error types
func TestErrorReporting_ContentSnapshot_VerifyStructure(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

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

	savedDefinitions, err := service.SaveDefinition(wordID, []*model.Definition{testDefinition}, nil)
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

// TestErrorReporting_FlaggedContentIndex_ExampleError tests that flaggedContentIndex
// is properly captured in content snapshots for unrelated example errors
func TestErrorReporting_FlaggedContentIndex_ExampleError(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)
	ctx := context.Background()

	// Create wordlist and word
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Flagged Content Test",
			"languageCode": "fr",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "maison",
		}).
		Expect().
		Status(201)

	wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

	// Create definition with multiple examples
	testDefinition := &model.Definition{
		Token:        "maison",
		Language:     "fr",
		PartOfSpeech: "nom",
		Meaning:      "Bâtiment destiné à l'habitation",
		Examples:     []string{"J'habite dans une maison", "Cette maison est belle", "La maison de mes parents"},
		Source:       "test_flagged",
	}

	savedDefinitions, err := service.SaveDefinition(wordID, []*model.Definition{testDefinition}, nil)
	require.NoError(t, err)
	require.Len(t, savedDefinitions, 1)
	
	definitionID := savedDefinitions[0].ID

	// Submit error report for unrelated example with flagged content index = 1 (second example)
	server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"wordId":              wordID,
			"definitionId":        definitionID,
			"errorType":           "_unrelated_example",
			"flaggedContentIndex": 1, // Flag the second example: "Cette maison est belle"
		}).
		Expect().
		Status(200)

	// Verify content snapshot contains flagged_content_index
	var snapshot map[string]interface{}
	err = server.DB.QueryRow(ctx, `
		SELECT content_snapshot FROM error_reports 
		WHERE word_id = $1 ORDER BY reported_at DESC LIMIT 1
	`, wordID).Scan(&snapshot)
	require.NoError(t, err)
	require.NotNil(t, snapshot, "Content snapshot should not be null")

	// Verify flagged_content_index is present and correct
	flaggedIndex, exists := snapshot["flagged_content_index"]
	assert.True(t, exists, "Content snapshot should contain flagged_content_index")
	assert.Equal(t, float64(1), flaggedIndex, "Flagged content index should be 1")

	// Verify examples are still present in snapshot
	examples, ok := snapshot["examples"].([]interface{})
	require.True(t, ok, "Examples should be an array")
	require.Len(t, examples, 3, "Should have 3 examples in snapshot")
	
	// Verify the flagged example matches index 1
	assert.Equal(t, "Cette maison est belle", examples[1], "Flagged example should match index 1")

	// Verify other snapshot fields are present
	assert.Equal(t, float64(definitionID), snapshot["definition_id"])
	assert.Equal(t, "Bâtiment destiné à l'habitation", snapshot["meaning"])
	assert.Equal(t, "nom", snapshot["part_of_speech"])
	assert.Equal(t, "fr", snapshot["language"])
	assert.Equal(t, "maison", snapshot["token"])

	// Verify definition was deleted (destructive operation for unrelated example)
	assertDefinitionDeleted(t, server.DB, definitionID)

	// Verify foreign key was nullified for destructive operation
	assertDefinitionIDNullified(t, server.DB, wordID)
}

// TestErrorReporting_RateLimit_FreeUser tests that free users are properly rate limited
func TestErrorReporting_RateLimit_FreeUser(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)

	// Create wordlist
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Rate Limit Test",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	// Create multiple words and definitions to avoid cooldown conflicts
	words := []struct {
		name     string
		meaning  string
		examples []string
	}{
		{"test1", "First test word", []string{"Example 1"}},
		{"test2", "Second test word", []string{"Example 2"}},
		{"test3", "Third test word", []string{"Example 3"}},
		{"test4", "Fourth test word", []string{"Example 4"}},
	}

	var errorReports []map[string]interface{}

	for i, word := range words {
		// Add word
		addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
			WithHeader("Authorization", token).
			WithJSON(map[string]interface{}{
				"name": word.name,
			}).
			Expect().
			Status(201)

		wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

		// Create definition
		testDefinition := &model.Definition{
			Token:        word.name,
			Language:     "en",
			PartOfSpeech: "noun",
			Meaning:      word.meaning,
			Examples:     word.examples,
			Source:       "test_rate_limit",
		}

		savedDefinitions, err := service.SaveDefinition(wordID, []*model.Definition{testDefinition}, nil)
		require.NoError(t, err)
		require.Len(t, savedDefinitions, 1)

		definitionID := savedDefinitions[0].ID

		errorReports = append(errorReports, map[string]interface{}{
			"wordId":       wordID,
			"definitionId": definitionID,
			"errorType":    "_unrelated_meaning",
		})

		// Submit error reports for first 3 words - should succeed
		if i < 3 {
			server.Expect.POST("/errorReports").
				WithHeader("Authorization", token).
				WithJSON(errorReports[i]).
				Expect().
				Status(200)
		}
	}

	// Submit fourth error report - should be rate limited (exceeds free user hourly limit of 3)
	errorResp := server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(errorReports[3]).
		Expect().
		Status(429)

	// Verify rate limit error message
	errorResp.JSON().Object().ValueEqual("error", "Hourly limit exceeded. You can report 3 errors per hour.")
}

// TestErrorReporting_RateLimit_PremiumUser tests that premium users have higher rate limits
func TestErrorReporting_RateLimit_PremiumUser(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithPremiumUser(t)

	// Create wordlist
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Premium Rate Test",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	// Create multiple words and definitions to avoid cooldown conflicts
	// Premium users have 10 reports per hour, so test 10 successful + 1 blocked to verify limit
	words := []struct {
		name     string
		meaning  string
		examples []string
	}{
		{"premium1", "First premium word", []string{"Premium example 1"}},
		{"premium2", "Second premium word", []string{"Premium example 2"}},
		{"premium3", "Third premium word", []string{"Premium example 3"}},
		{"premium4", "Fourth premium word", []string{"Premium example 4"}},
		{"premium5", "Fifth premium word", []string{"Premium example 5"}},
		{"premium6", "Sixth premium word", []string{"Premium example 6"}},
		{"premium7", "Seventh premium word", []string{"Premium example 7"}},
		{"premium8", "Eighth premium word", []string{"Premium example 8"}},
		{"premium9", "Ninth premium word", []string{"Premium example 9"}},
		{"premium10", "Tenth premium word", []string{"Premium example 10"}},
		{"premium11", "Eleventh premium word", []string{"Premium example 11"}},
	}

	var errorReports []map[string]interface{}

	for _, word := range words {
		// Add word
		addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
			WithHeader("Authorization", token).
			WithJSON(map[string]interface{}{
				"name": word.name,
			}).
			Expect().
			Status(201)

		wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

		// Create definition
		testDefinition := &model.Definition{
			Token:        word.name,
			Language:     "en",
			PartOfSpeech: "adjective",
			Meaning:      word.meaning,
			Examples:     word.examples,
			Source:       "test_premium_rate",
		}

		savedDefinitions, err := service.SaveDefinition(wordID, []*model.Definition{testDefinition}, nil)
		require.NoError(t, err)
		require.Len(t, savedDefinitions, 1)

		definitionID := savedDefinitions[0].ID

		errorReports = append(errorReports, map[string]interface{}{
			"wordId":       wordID,
			"definitionId": definitionID,
			"errorType":    "_unrelated_meaning",
		})
	}

	// Submit first 10 error reports - should all succeed for premium users (limit is 10/hour)
	for i := 0; i < 10; i++ {
		server.Expect.POST("/errorReports").
			WithHeader("Authorization", token).
			WithJSON(errorReports[i]).
			Expect().
			Status(200).
			JSON().Object().
			NotContainsKey("error") // Should not get rate limit error

		// Log progress for verification
		t.Logf("Premium user successfully submitted error report %d/10", i+1)
	}

	// Submit 11th error report - should be rate limited (exceeds premium user hourly limit of 10)
	errorResp := server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(errorReports[10]). // 11th report
		Expect().
		Status(429)

	// Verify rate limit error message for premium users
	errorResp.JSON().Object().ValueEqual("error", "Hourly limit exceeded. You can report 10 errors per hour.")
	
	t.Log("Premium user rate limiting test completed successfully - 10 reports succeeded, 11th was properly blocked")
}