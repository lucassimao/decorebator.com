package errorreporting

import (
	"context"
	"testing"
	"time"

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
