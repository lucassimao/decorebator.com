package errorreporting

import (
	"context"
	"testing"

	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
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
	assertErrorReportExists(t, server.DB, wordID, "_unrelated_meaning")

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

// TestErrorReporting_UnrelatedImage_ImageRegeneration tests that unrelated image errors
// preserve definition snapshots and trigger image regeneration workers
func TestErrorReporting_UnrelatedImage_ImageRegeneration(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)
	ctx := context.Background()

	// Create wordlist and word
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Image Error Test",
			"languageCode": "de",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "Hund",
		}).
		Expect().
		Status(201)

	wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

	// Create definition
	testDefinition := &model.Definition{
		Token:        "Hund",
		Language:     "de",
		PartOfSpeech: "Substantiv",
		Meaning:      "Ein vierbeiniges Haustier",
		Examples:     []string{"Der Hund bellt", "Mein Hund ist freundlich"},
		Source:       "test_image_error",
	}

	savedDefinitions, err := service.SaveDefinition(wordID, []*model.Definition{testDefinition}, nil)
	require.NoError(t, err)
	require.Len(t, savedDefinitions, 1)

	definitionID := savedDefinitions[0].ID

	// Verify definition exists before error report
	var definitionExists bool
	err = server.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM definitions WHERE id = $1)", definitionID).Scan(&definitionExists)
	require.NoError(t, err)
	assert.True(t, definitionExists, "Definition should exist before error report")

	// Submit error report for unrelated image (non-destructive operation)
	server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"wordId":       wordID,
			"definitionId": definitionID,
			"errorType":    "_unrelated_image",
		}).
		Expect().
		Status(200)

	// Verify error report was created
	assertErrorReportExists(t, server.DB, wordID, "_unrelated_image")

	// Verify content snapshot was created and contains definition data
	assertContentSnapshotContains(t, server.DB, wordID, map[string]interface{}{
		"definition_id":  float64(definitionID),
		"meaning":        "Ein vierbeiniges Haustier",
		"part_of_speech": "Substantiv",
		"language":       "de",
		"token":          "Hund",
	})

	// Verify definition was NOT deleted (non-destructive operation)
	assertDefinitionNotDeleted(t, server.DB, definitionID)

	// Verify foreign key was preserved for non-destructive operation
	assertDefinitionIDPreserved(t, server.DB, wordID, definitionID)
}

// TestErrorReporting_MissingImage_ImageRegeneration tests that missing image errors
// preserve definition snapshots and trigger image regeneration workers
func TestErrorReporting_MissingImage_ImageRegeneration(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)

	// Create wordlist and word
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Missing Image Test",
			"languageCode": "it",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "gatto",
		}).
		Expect().
		Status(201)

	wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

	// Create definition
	testDefinition := &model.Definition{
		Token:        "gatto",
		Language:     "it",
		PartOfSpeech: "sostantivo",
		Meaning:      "Un piccolo mammifero carnivoro domestico",
		Examples:     []string{"Il gatto dorme", "Ho un gatto nero"},
		Source:       "test_missing_image",
	}

	savedDefinitions, err := service.SaveDefinition(wordID, []*model.Definition{testDefinition}, nil)
	require.NoError(t, err)
	require.Len(t, savedDefinitions, 1)

	definitionID := savedDefinitions[0].ID

	// Submit error report for missing image (non-destructive operation)
	server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"wordId":       wordID,
			"definitionId": definitionID,
			"errorType":    "_missing_image",
		}).
		Expect().
		Status(200)

	// Verify error report was created
	assertErrorReportExists(t, server.DB, wordID, "_missing_image")

	// Verify content snapshot was created and contains definition data
	assertContentSnapshotContains(t, server.DB, wordID, map[string]interface{}{
		"definition_id":  float64(definitionID),
		"meaning":        "Un piccolo mammifero carnivoro domestico",
		"part_of_speech": "sostantivo",
		"language":       "it",
		"token":          "gatto",
	})

	// Verify definition was NOT deleted (non-destructive operation)
	assertDefinitionNotDeleted(t, server.DB, definitionID)

	// Verify foreign key was preserved for non-destructive operation
	assertDefinitionIDPreserved(t, server.DB, wordID, definitionID)
}

// TestErrorReporting_SoundNotPlaying_AudioRegeneration tests that sound not playing errors
// capture word snapshots and trigger audio regeneration workers
func TestErrorReporting_SoundNotPlaying_AudioRegeneration(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)

	// Create wordlist and word
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Audio Error Test",
			"languageCode": "ja",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "猫",
		}).
		Expect().
		Status(201)

	wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

	// Submit error report for sound not playing (word-level error, no definition required)
	server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"wordId":    wordID,
			"errorType": "_sound_not_playing",
		}).
		Expect().
		Status(200)

	// Verify error report was created (no definitionID required for audio errors)
	var count int
	err := server.DB.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM error_reports 
		WHERE word_id = $1 AND error_type = $2
	`, wordID, "_sound_not_playing").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Error report should exist")

	// Verify word snapshot structure and content
	assertWordSnapshotStructure(t, server.DB, wordID, "猫", "ja")

	// Verify definition_id is null for word-level errors (no specific definition involved)
	var definitionID *int64
	err = server.DB.QueryRow(context.Background(), `
		SELECT definition_id FROM error_reports 
		WHERE word_id = $1 ORDER BY reported_at DESC LIMIT 1
	`, wordID).Scan(&definitionID)
	require.NoError(t, err)
	assert.Nil(t, definitionID, "Definition ID should be NULL for word-level audio errors")
}

// TestErrorReporting_ProcessingFailed_DefinitionRetry tests that processing failed errors
// trigger definition regeneration without content snapshots
func TestErrorReporting_ProcessingFailed_DefinitionRetry(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)

	// Create wordlist and word
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Processing Failed Test",
			"languageCode": "pt",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "casa",
		}).
		Expect().
		Status(201)

	wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

	// Submit error report for processing failed (no definition required)
	server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"wordId":    wordID,
			"errorType": "_processing_failed",
		}).
		Expect().
		Status(200)

	// Verify error report was created
	var count int
	err := server.DB.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM error_reports 
		WHERE word_id = $1 AND error_type = $2
	`, wordID, "_processing_failed").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Error report should exist")

	// Verify no content snapshot for processing failed errors (it's a retry, not content quality issue)
	var snapshot map[string]interface{}
	err = server.DB.QueryRow(context.Background(), `
		SELECT content_snapshot FROM error_reports 
		WHERE word_id = $1 ORDER BY reported_at DESC LIMIT 1
	`, wordID).Scan(&snapshot)
	require.NoError(t, err)
	assert.Nil(t, snapshot, "Content snapshot should be NULL for processing failed errors")

	// Verify definition_id is null for processing retries
	var definitionID *int64
	err = server.DB.QueryRow(context.Background(), `
		SELECT definition_id FROM error_reports 
		WHERE word_id = $1 ORDER BY reported_at DESC LIMIT 1
	`, wordID).Scan(&definitionID)
	require.NoError(t, err)
	assert.Nil(t, definitionID, "Definition ID should be NULL for processing failed errors")
}
