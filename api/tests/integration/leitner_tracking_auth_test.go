package integration

import (
	"context"
	"fmt"
	"testing"

	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/require"
)

func TestQuizAnswerValidatesTrackingOwnership(t *testing.T) {
	server := setup.NewTestServer(t)
	ctx := context.Background()

	token := server.WithTestUser(t)

	// Resolve user ID for the token we just created.
	var userID int64
	err := server.DB.QueryRow(ctx, "SELECT id FROM users ORDER BY created_at DESC LIMIT 1").Scan(&userID)
	require.NoError(t, err)

	// Create a wordlist and word via API.
	wordlistID := createTestWordlist(server, token, "Ownership Test", "Tracking ownership")
	wordResp := server.Expect.POST(fmt.Sprintf("/wordlists/%d/words", wordlistID)).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":  "ownership-word",
			"notes": "test",
		}).
		Expect().
		Status(201)

	wordID := int64(wordResp.JSON().Object().Value("id").Number().Raw())

	// Create definition and tracking row using services (background jobs are not running in tests).
	definitionService := service.NewDefinitionService(server.DB)
	leitnerTrackingService := service.NewLeitnerTrackingService(server.DB)

	tx, err := server.DB.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	definition := &model.Definition{
		Token:                  "ownership-word",
		Language:               "en",
		Meaning:                "ownership test meaning",
		PartOfSpeech:           "noun",
		Examples:               []string{"This is a [ownership-word] example."},
		Source:                 "test",
		PartOfSpeechNormalized: "noun",
	}

	definitions, err := definitionService.SaveDefinition(ctx, wordID, []*model.Definition{definition}, &tx)
	require.NoError(t, err)
	require.Len(t, definitions, 1)

	definitionID := definitions[0].ID

	err = leitnerTrackingService.IncludeDefinitions(ctx, wordID, userID, []int64{definitionID}, tx)
	require.NoError(t, err)

	var trackingID int64
	err = tx.QueryRow(ctx,
		`SELECT id FROM leitner_system_tracking WHERE word_id = $1 AND definition_id = $2 AND user_id = $3`,
		wordID, definitionID, userID).Scan(&trackingID)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	// Control case: correct identifiers succeed.
	server.Expect.PATCH(fmt.Sprintf("/wordlists/%d/quizzes", wordlistID)).
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithJSON(map[string]interface{}{
			"wordlistID":              wordlistID,
			"wordID":                  wordID,
			"definitionID":            definitionID,
			"leitnerSystemTrackingID": trackingID,
			"quizType":                "GUESS_MEANING",
			"isCorrect":               true,
			"responseTimeMs":          1000,
		}).
		Expect().
		Status(204)

	// Wrong wordlist ID should be rejected.
	otherWordlistID := createTestWordlist(server, token, "Other List", "Different list")
	server.Expect.PATCH(fmt.Sprintf("/wordlists/%d/quizzes", otherWordlistID)).
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithJSON(map[string]interface{}{
			"wordlistID":              otherWordlistID,
			"wordID":                  wordID,
			"definitionID":            definitionID,
			"leitnerSystemTrackingID": trackingID,
			"quizType":                "GUESS_MEANING",
			"isCorrect":               true,
			"responseTimeMs":          1000,
		}).
		Expect().
		Status(404)

	// Wrong definition ID should be rejected.
	server.Expect.PATCH(fmt.Sprintf("/wordlists/%d/quizzes", wordlistID)).
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithJSON(map[string]interface{}{
			"wordlistID":              wordlistID,
			"wordID":                  wordID,
			"definitionID":            definitionID + 999,
			"leitnerSystemTrackingID": trackingID,
			"quizType":                "GUESS_MEANING",
			"isCorrect":               true,
			"responseTimeMs":          1000,
		}).
		Expect().
		Status(404)
}

func createTestWordlist(server *setup.TestServer, token string, name, description string) int64 {
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
