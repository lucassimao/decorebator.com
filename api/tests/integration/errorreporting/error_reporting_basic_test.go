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

	// Create test data using new helpers
	wordlistID := createTestWordlist(server, token, "Test Wordlist", "en")
	definition := setup.CreateErrorReportDefinition("water", "en", "noun")
	wordID, definitionID := createWordWithDefinition(t, server, token, wordlistID, "water", definition)

	// Verify definition exists before error report
	assertRecordCount(t, server.DB, "definitions", "id = $1", 1, definitionID)

	// Submit error report for unrelated meaning
	resp := submitErrorReport(server, token, wordID, definitionID, "_unrelated_meaning")
	resp.Status(200)

	// Verify error report was created
	assertErrorReportExists(t, server.DB, wordID, "_unrelated_meaning")

	// Verify content snapshot was created and contains definition data
	assertContentSnapshotContains(t, server.DB, wordID, map[string]interface{}{
		"definition_id":  float64(definitionID), // JSON numbers are float64
		"part_of_speech": "noun",
		"language":       "en",
		"token":          "water",
	})

	// Verify definition was deleted (destructive operation)
	assertDefinitionDeleted(t, server.DB, definitionID)

	// Verify foreign key was nullified for destructive operation
	assertDefinitionIDNullified(t, server.DB, wordID)

	// Verify River job was triggered for definition fetching
	assertDefinitionFetcherJobTriggered(t, server.DB, wordID)
}

// TestErrorReporting_UnrelatedImage_ImageRegeneration tests that unrelated image errors
// preserve definition snapshots and trigger image regeneration workers
func TestErrorReporting_UnrelatedImage_ImageRegeneration(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)

	// Create test data using new helpers
	wordlistID := createTestWordlist(server, token, "Image Error Test", "de")
	definition := setup.CreateErrorReportDefinition("Hund", "de", "Substantiv")
	wordID, definitionID := createWordWithDefinition(t, server, token, wordlistID, "Hund", definition)

	// Verify definition exists before error report
	assertRecordCount(t, server.DB, "definitions", "id = $1", 1, definitionID)

	// Submit error report for unrelated image (non-destructive operation)
	resp := submitErrorReport(server, token, wordID, definitionID, "_unrelated_image")
	resp.Status(200)

	// Verify error report was created
	assertErrorReportExists(t, server.DB, wordID, "_unrelated_image")

	// Verify content snapshot was created and contains definition data
	assertContentSnapshotContains(t, server.DB, wordID, map[string]interface{}{
		"definition_id":  float64(definitionID),
		"part_of_speech": "Substantiv",
		"language":       "de",
		"token":          "Hund",
	})

	// Verify definition was NOT deleted (non-destructive operation)
	assertDefinitionNotDeleted(t, server.DB, definitionID)

	// Verify foreign key was preserved for non-destructive operation
	assertDefinitionIDPreserved(t, server.DB, wordID, definitionID)

	// Verify River job was triggered for image generation
	assertImageGeneratorJobTriggered(t, server.DB, definitionID)
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

	// Verify River job was triggered for image generation
	assertImageGeneratorJobTriggered(t, server.DB, definitionID)
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

	// Verify River job was triggered for text-to-speech generation
	assertTextToSpeechJobTriggered(t, server.DB, wordID)
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

	// Verify River job was triggered for definition fetching
	assertDefinitionFetcherJobTriggered(t, server.DB, wordID)
}

// TestErrorReporting_UnrelatedExample_DefinitionRegeneration tests that unrelated example errors
// trigger definition regeneration and delete definitions
func TestErrorReporting_UnrelatedExample_DefinitionRegeneration(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)

	// Create test data using new helpers
	wordlistID := createTestWordlist(server, token, "Unrelated Example Test", "fr")
	definition := setup.CreateErrorReportDefinition("chat", "fr", "nom")
	wordID, definitionID := createWordWithDefinition(t, server, token, wordlistID, "chat", definition)

	// Verify definition exists before error report
	assertRecordCount(t, server.DB, "definitions", "id = $1", 1, definitionID)

	// Submit error report for unrelated example (destructive operation)
	resp := submitErrorReport(server, token, wordID, definitionID, "_unrelated_example")
	resp.Status(200)

	// Verify error report was created
	assertErrorReportExists(t, server.DB, wordID, "_unrelated_example")

	// Verify content snapshot was created and contains definition data
	assertContentSnapshotContains(t, server.DB, wordID, map[string]interface{}{
		"definition_id":  float64(definitionID),
		"part_of_speech": "nom",
		"language":       "fr",
		"token":          "chat",
	})

	// Verify definition was deleted (destructive operation)
	assertDefinitionDeleted(t, server.DB, definitionID)

	// Verify foreign key was nullified for destructive operation
	assertDefinitionIDNullified(t, server.DB, wordID)

	// Verify River job was triggered for definition fetching
	assertDefinitionFetcherJobTriggered(t, server.DB, wordID)
}
