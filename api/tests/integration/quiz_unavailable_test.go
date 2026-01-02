package integration

import (
	"context"
	"testing"

	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/require"
)

func TestQuizCreate_WordlistEmptyReturnsConflict(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)

	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Empty Wordlist",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	resp := server.Expect.POST("/wordlists/{wordlistId}/quizzes", wordlistID).
		WithHeader("Authorization", token).
		Expect().
		Status(409).
		JSON().Object()

	resp.Value("status").String().IsEqual("wordlist_empty")
	resp.Value("error").String().NotEmpty()
}

func TestQuizCreate_WordlistProcessingReturnsConflict(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)

	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Processing Wordlist",
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

	resp := server.Expect.POST("/wordlists/{wordlistId}/quizzes", wordlistID).
		WithHeader("Authorization", token).
		Expect().
		Status(409).
		JSON().Object()

	resp.Value("status").String().IsEqual("wordlist_processing")
	resp.Value("error").String().NotEmpty()
}

func TestQuizCreate_AllLearnedReturnsConflict(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)

	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Learned Wordlist",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	addResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "hello",
		}).
		Expect().
		Status(201)

	wordID := int(addResp.JSON().Object().Value("id").Number().Raw())

	// Mark word as learned and completed so availability check passes processing but fails unlearned
	_, err := server.DB.Exec(context.Background(), `
		UPDATE words
		SET learned = TRUE, processing_status = 'completed'
		WHERE id = $1
	`, wordID)
	require.NoError(t, err)

	resp := server.Expect.POST("/wordlists/{wordlistId}/quizzes", wordlistID).
		WithHeader("Authorization", token).
		Expect().
		Status(409).
		JSON().Object()

	resp.Value("status").String().IsEqual("no_unlearned_words")
	resp.Value("error").String().NotEmpty()
}
