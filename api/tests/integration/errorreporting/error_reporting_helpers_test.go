package errorreporting

import (
	"context"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/gavv/httpexpect/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper functions for error reporting tests

func assertErrorReportExists(t *testing.T, db *pgxpool.Pool, wordID int64, errorType string) {
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

func assertDefinitionNotDeleted(t *testing.T, db *pgxpool.Pool, definitionID int64) {
	ctx := context.Background()
	var count int
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM definitions WHERE id = $1", definitionID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Definition should still exist for non-destructive operations")
}

func assertDefinitionIDPreserved(t *testing.T, db *pgxpool.Pool, wordID, expectedDefinitionID int64) {
	ctx := context.Background()
	var definitionID *int64
	err := db.QueryRow(ctx, `
		SELECT definition_id FROM error_reports 
		WHERE word_id = $1 ORDER BY reported_at DESC LIMIT 1
	`, wordID).Scan(&definitionID)
	require.NoError(t, err)
	require.NotNil(t, definitionID, "Definition ID should not be NULL for non-destructive operations")
	assert.Equal(t, expectedDefinitionID, *definitionID, "Definition ID should be preserved for non-destructive operations")
}

func assertWordSnapshotStructure(t *testing.T, db *pgxpool.Pool, wordID int64, expectedToken, expectedLanguage string) {
	ctx := context.Background()
	var snapshot map[string]interface{}
	err := db.QueryRow(ctx, `
		SELECT content_snapshot FROM error_reports 
		WHERE word_id = $1 ORDER BY reported_at DESC LIMIT 1
	`, wordID).Scan(&snapshot)
	require.NoError(t, err)
	require.NotNil(t, snapshot, "Content snapshot should not be null")

	// Verify word snapshot contains expected fields
	expectedFields := []string{"word_id", "token", "language", "captured_at"}
	for _, field := range expectedFields {
		value, exists := snapshot[field]
		assert.True(t, exists, "Word snapshot should contain field: %s", field)
		assert.NotNil(t, value, "Word snapshot field %s should not be null", field)
	}

	// Verify specific word snapshot values
	assert.Equal(t, float64(wordID), snapshot["word_id"], "Word ID should match")
	assert.Equal(t, expectedToken, snapshot["token"], "Token should match")
	assert.Equal(t, expectedLanguage, snapshot["language"], "Language should match")

	// Verify captured_at timestamp format
	capturedAt, exists := snapshot["captured_at"]
	assert.True(t, exists, "Should have captured_at timestamp")
	assert.IsType(t, "", capturedAt, "captured_at should be a string")

	// Verify it's a valid RFC3339 timestamp
	_, err = time.Parse(time.RFC3339, capturedAt.(string))
	assert.NoError(t, err, "captured_at should be valid RFC3339 timestamp")
}

// Setup Helpers - Reduce code duplication across test files

// createTestWordlist creates a wordlist via API and returns the wordlist ID
func createTestWordlist(server *setup.TestServer, token, name, languageCode string) int {
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         name,
			"languageCode": languageCode,
		}).
		Expect().
		Status(201)

	return int(createResp.JSON().Object().Value("id").Number().Raw())
}

// createTestWord creates a word via API and returns the word ID
func createTestWord(server *setup.TestServer, token string, wordlistID int, wordName string) int64 {
	addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": wordName,
		}).
		Expect().
		Status(201)

	return int64(addWordResp.JSON().Object().Value("id").Number().Raw())
}

// createTestDefinition creates a definition using the service layer and returns the definition ID
func createTestDefinition(t *testing.T, definitionService *service.DefinitionService, wordID int64, definition *model.Definition) int64 {
	savedDefinitions, err := definitionService.SaveDefinition(context.Background(), wordID, []*model.Definition{definition}, nil)
	require.NoError(t, err, "Failed to create test definition")
	require.Len(t, savedDefinitions, 1, "Should save exactly one definition")

	return savedDefinitions[0].ID
}

// submitErrorReport submits an error report via API and returns the response
func submitErrorReport(server *setup.TestServer, token string, wordID, definitionID int64, errorType string) *httpexpect.Response {
	payload := setup.CreateErrorReportPayload(wordID, definitionID, errorType, nil)
	return server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(payload).
		Expect()
}

// createWordWithDefinition creates both word and definition in one call and returns both IDs
func createWordWithDefinition(t *testing.T, server *setup.TestServer, token string, wordlistID int, wordName string, definition *model.Definition, definitionService *service.DefinitionService) (int64, int64) {
	wordID := createTestWord(server, token, wordlistID, wordName)
	definitionID := createTestDefinition(t, definitionService, wordID, definition)
	return wordID, definitionID
}

// River Job Verification Helpers - Fix Current Broken Tests

// assertRiverJobTriggered verifies that a River job was triggered for the expected worker type
func assertRiverJobTriggered(t *testing.T, db *pgxpool.Pool, jobKind, queue string, resourceID int64, resourceField string) {
	ctx := context.Background()
	var jobCount int

	query := `
		SELECT COUNT(*) FROM river_job 
		WHERE kind = $1 
		AND queue = $2
		AND (args->>$3)::bigint = $4
	`

	err := db.QueryRow(ctx, query, jobKind, queue, resourceField, resourceID).Scan(&jobCount)
	require.NoError(t, err, "Failed to query River jobs")

	assert.GreaterOrEqual(t, jobCount, 1, "River job should be triggered for %s", jobKind)
}

// assertImageGeneratorJobTriggered verifies that an ImageGenerator job was triggered for a definition
func assertImageGeneratorJobTriggered(t *testing.T, db *pgxpool.Pool, definitionID int64) {
	assertRiverJobTriggered(t, db, "ImageGenerator", "image_generator", definitionID, "definitionId")
}

// assertDefinitionFetcherJobTriggered verifies that a DefinitionFetcher job was triggered for a word
func assertDefinitionFetcherJobTriggered(t *testing.T, db *pgxpool.Pool, wordID int64) {
	assertRiverJobTriggered(t, db, "DefinitionFetcher", "definition_fetcher", wordID, "wordId")
}

// assertTextToSpeechJobTriggered verifies that a TextToSpeech job was triggered for a word
func assertTextToSpeechJobTriggered(t *testing.T, db *pgxpool.Pool, wordID int64) {
	assertRiverJobTriggered(t, db, "TextToSpeech", "text_to_speech", wordID, "wordId")
}

// Database State Validation Helpers

// assertRecordCount verifies the expected number of records in a table matching a condition
func assertRecordCount(t *testing.T, db *pgxpool.Pool, table, condition string, expectedCount int, params ...interface{}) {
	ctx := context.Background()
	var count int

	query := "SELECT COUNT(*) FROM " + table
	if condition != "" {
		query += " WHERE " + condition
	}

	err := db.QueryRow(ctx, query, params...).Scan(&count)
	require.NoError(t, err, "Failed to count records in %s", table)
	assert.Equal(t, expectedCount, count, "Record count mismatch in table %s", table)
}
