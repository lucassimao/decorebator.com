package integration

import (
	"context"
	"testing"

	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWordProcessingStatus_InitialState verifies that new words start with "pending" status
func TestWordProcessingStatus_InitialState(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	// Create a test user
	token := server.WithTestUser(t)

	// Create a wordlist
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Test Wordlist",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	// Get processing status
	statusResp := server.Expect.GET("/wordlists/{wordlistId}/processing-status", wordlistID).
		WithHeader("Authorization", token).
		Expect().
		Status(200).
		JSON().Object()

	// Verify response structure
	statusResp.ContainsKey("words")
	statusResp.Value("words").Array().Length().IsEqual(0)

	// Add a word
	server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "hello",
		}).
		Expect().
		Status(201)

	// Get words and verify initial status
	wordsResp := server.Expect.GET("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		Expect().
		Status(200).
		JSON().Array()

	wordsResp.Length().IsEqual(1)
	word := wordsResp.Value(0).Object()
	word.Value("processingStatus").String().IsEqual("pending")
	word.Value("processingError").String().IsEmpty()
}

// TestWordProcessingStatus_GetEndpoint tests the processing status endpoint
func TestWordProcessingStatus_GetEndpoint(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)

	// Create a wordlist
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Test Wordlist",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	// Add multiple words
	words := []string{"hello", "world", "test"}
	for _, word := range words {
		server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
			WithHeader("Authorization", token).
			WithJSON(map[string]interface{}{
				"name": word,
			}).
			Expect().
			Status(201)
	}

	// Get processing status
	statusResp := server.Expect.GET("/wordlists/{wordlistId}/processing-status", wordlistID).
		WithHeader("Authorization", token).
		Expect().
		Status(200).
		JSON().Object()

	// Verify response structure
	statusResp.ContainsKey("words")
	statusResp.ContainsKey("summary")

	// Check words array
	wordsArray := statusResp.Value("words").Array()
	wordsArray.Length().IsEqual(3)

	// Check each word has required fields
	for i := 0; i < 3; i++ {
		word := wordsArray.Value(i).Object()
		word.ContainsKey("id")
		word.ContainsKey("name")
		word.ContainsKey("processingStatus")
		word.Value("processingStatus").String().IsEqual("pending")
	}

	// Check summary
	summary := statusResp.Value("summary").Object()
	summary.Value("total").Number().IsEqual(3)
	summary.Value("pending").Number().IsEqual(3)
	summary.Value("processing").Number().IsEqual(0)
	summary.Value("completed").Number().IsEqual(0)
	summary.Value("failed").Number().IsEqual(0)
}

// TestWordProcessingStatus_WorkerUpdates simulates worker status updates
func TestWordProcessingStatus_WorkerUpdates(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)

	// Create wordlist and add word
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Test Wordlist",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "hello",
		}).
		Expect().
		Status(201)

	// Get the word ID from database
	ctx := context.Background()
	var wordID int64
	err := server.DB.QueryRow(ctx,
		"SELECT id FROM words WHERE wordlist_id = $1 AND name = $2",
		wordlistID, "hello").Scan(&wordID)
	require.NoError(t, err)

	// Simulate worker updating status to "processing"
	err = updateWordProcessingStatus(server, wordID, "processing", "")
	require.NoError(t, err)

	// Verify status changed
	statusResp := server.Expect.GET("/wordlists/{wordlistId}/processing-status", wordlistID).
		WithHeader("Authorization", token).
		Expect().
		Status(200).
		JSON().Object()

	word := statusResp.Value("words").Array().Value(0).Object()
	word.Value("processingStatus").String().IsEqual("processing")
	word.Value("processingStartedAt").String().NotEmpty()

	summary := statusResp.Value("summary").Object()
	summary.Value("processing").Number().IsEqual(1)
	summary.Value("pending").Number().IsEqual(0)

	// Simulate worker completing successfully
	err = updateWordProcessingStatus(server, wordID, "completed", "")
	require.NoError(t, err)

	// Verify final status
	statusResp2 := server.Expect.GET("/wordlists/{wordlistId}/processing-status", wordlistID).
		WithHeader("Authorization", token).
		Expect().
		Status(200).
		JSON().Object()

	word2 := statusResp2.Value("words").Array().Value(0).Object()
	word2.Value("processingStatus").String().IsEqual("completed")
	word2.Value("processingCompletedAt").String().NotEmpty()

	summary2 := statusResp2.Value("summary").Object()
	summary2.Value("completed").Number().IsEqual(1)
	summary2.Value("processing").Number().IsEqual(0)
}

// TestWordProcessingStatus_FailedProcessing tests failed word processing
func TestWordProcessingStatus_FailedProcessing(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)

	// Create wordlist and add word
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Test Wordlist",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "invalidword123",
		}).
		Expect().
		Status(201)

	// Get the word ID
	ctx := context.Background()
	var wordID int64
	err := server.DB.QueryRow(ctx,
		"SELECT id FROM words WHERE wordlist_id = $1 AND name = $2",
		wordlistID, "invalidword123").Scan(&wordID)
	require.NoError(t, err)

	// Simulate worker failing with error
	errorMsg := "No definitions found for this word"
	err = updateWordProcessingStatus(server, wordID, "failed", errorMsg)
	require.NoError(t, err)

	// Verify failed status
	statusResp := server.Expect.GET("/wordlists/{wordlistId}/processing-status", wordlistID).
		WithHeader("Authorization", token).
		Expect().
		Status(200).
		JSON().Object()

	word := statusResp.Value("words").Array().Value(0).Object()
	word.Value("processingStatus").String().IsEqual("failed")
	word.Value("processingError").String().IsEqual(errorMsg)
	word.Value("processingCompletedAt").String().NotEmpty()

	summary := statusResp.Value("summary").Object()
	summary.Value("failed").Number().IsEqual(1)
}

// TestWordProcessingStatus_MultipleWords tests status with multiple words in different states
func TestWordProcessingStatus_MultipleWords(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)

	// Create wordlist
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Test Wordlist",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	// Add multiple words
	words := []string{"pending1", "processing1", "completed1", "failed1"}
	for _, word := range words {
		server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
			WithHeader("Authorization", token).
			WithJSON(map[string]interface{}{
				"name": word,
			}).
			Expect().
			Status(201)
	}

	ctx := context.Background()

	// Update each word to different status
	// Leave "pending1" as is

	// Update "processing1" to processing
	var processingID int64
	err := server.DB.QueryRow(ctx, "SELECT id FROM words WHERE name = $1", "processing1").Scan(&processingID)
	require.NoError(t, err)
	require.NoError(t, err)
	err = updateWordProcessingStatus(server, processingID, "processing", "")
	require.NoError(t, err)

	// Update "completed1" to completed
	var completedID int64
	err = server.DB.QueryRow(ctx, "SELECT id FROM words WHERE name = $1", "completed1").Scan(&completedID)
	require.NoError(t, err)
	err = updateWordProcessingStatus(server, completedID, "completed", "")
	require.NoError(t, err)

	// Update "failed1" to failed
	var failedID int64
	err = server.DB.QueryRow(ctx, "SELECT id FROM words WHERE name = $1", "failed1").Scan(&failedID)
	require.NoError(t, err)
	err = updateWordProcessingStatus(server, failedID, "failed", "Test error")
	require.NoError(t, err)

	// Get processing status
	statusResp := server.Expect.GET("/wordlists/{wordlistId}/processing-status", wordlistID).
		WithHeader("Authorization", token).
		Expect().
		Status(200).
		JSON().Object()

	// Verify summary counts
	summary := statusResp.Value("summary").Object()
	summary.Value("total").Number().IsEqual(4)
	summary.Value("pending").Number().IsEqual(1)
	summary.Value("processing").Number().IsEqual(1)
	summary.Value("completed").Number().IsEqual(1)
	summary.Value("failed").Number().IsEqual(1)

	// Verify each word status
	wordsArray := statusResp.Value("words").Array()
	statusMap := make(map[string]string)
	for i := 0; i < 4; i++ {
		word := wordsArray.Value(i).Object()
		name := word.Value("name").String().Raw()
		status := word.Value("processingStatus").String().Raw()
		statusMap[name] = status
	}

	assert.Equal(t, "pending", statusMap["pending1"])
	assert.Equal(t, "processing", statusMap["processing1"])
	assert.Equal(t, "completed", statusMap["completed1"])
	assert.Equal(t, "failed", statusMap["failed1"])
}

// TestWordProcessingStatus_Authorization tests that users can only see their own wordlist status
func TestWordProcessingStatus_Authorization(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	// Create two users
	token1 := server.WithTestUser(t)
	token2 := server.WithTestUser(t)

	// User 1 creates a wordlist
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token1).
		WithJSON(map[string]interface{}{
			"name":         "User 1 Wordlist",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	// User 1 adds a word
	server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token1).
		WithJSON(map[string]interface{}{
			"name": "secret",
		}).
		Expect().
		Status(201)

	// User 1 can see processing status
	server.Expect.GET("/wordlists/{wordlistId}/processing-status", wordlistID).
		WithHeader("Authorization", token1).
		Expect().
		Status(200)

	// User 2 cannot see processing status (gets 404)
	server.Expect.GET("/wordlists/{wordlistId}/processing-status", wordlistID).
		WithHeader("Authorization", token2).
		Expect().
		Status(404)
}

// TestWordProcessingStatus_ErrorReporting tests error reporting with nullable definition IDs
func TestWordProcessingStatus_ErrorReporting(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)

	// Create wordlist and add word
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Test Wordlist",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "testword",
		}).
		Expect().
		Status(201)

	wordID := int(addWordResp.JSON().Object().Value("id").Number().Raw())

	// Test word-level error report (null definition ID)
	server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"wordId":       wordID,
			"definitionId": nil, // null definition ID for word-level errors
			"errorType":    "_processing_failed",
		}).
		Expect().
		Status(200)

	// Test cooldown mechanism with null definition ID
	server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"wordId":       wordID,
			"definitionId": nil,
			"errorType":    "_processing_failed",
		}).
		Expect().
		Status(429) // Should be rate limited

	// Test invalid request (definition ID required for image errors)
	server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"wordId":       wordID,
			"definitionId": nil, // null not allowed for image errors
			"errorType":    "_unrelated_image",
		}).
		Expect().
		Status(500) // Should fail because definition ID is required
}

// Helper function to update word processing status (simulates worker behavior)
func updateWordProcessingStatus(ts *setup.TestServer, wordID int64, status, errorMsg string) error {
	return ts.AppContext.WordService.UpdateProcessingStatus(wordID, status, errorMsg, nil)
}
