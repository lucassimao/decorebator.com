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

func TestQuizResponseUsesTrackingWordIDForSharedDefinitions(t *testing.T) {
	server := setup.NewTestServer(t)
	ctx := context.Background()

	token := server.WithPremiumUser(t)

	// Resolve user ID for the token we just created.
	var userID int64
	err := server.DB.QueryRow(ctx, "SELECT id FROM users ORDER BY created_at DESC LIMIT 1").Scan(&userID)
	require.NoError(t, err)

	wordlistA := createTestWordlist(server, token, "Shared Defs A", "Wordlist A")
	wordlistB := createTestWordlist(server, token, "Shared Defs B", "Wordlist B")

	wordName := "shared-word"

	// Create word in wordlist A.
	wordAResp := server.Expect.POST(fmt.Sprintf("/wordlists/%d/words", wordlistA)).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":  wordName,
			"notes": "test",
		}).
		Expect().
		Status(201)
	wordAID := int64(wordAResp.JSON().Object().Value("id").Number().Raw())

	// Create a definition for word A and include in tracking.
	definitionService := service.NewDefinitionService(server.DB)
	leitnerTrackingService := service.NewLeitnerTrackingService(server.DB)

	tx, err := server.DB.Begin(ctx)
	require.NoError(t, err)
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	definition := &model.Definition{
		Token:                  wordName,
		Language:               "en",
		Meaning:                "shared meaning",
		PartOfSpeech:           "noun",
		Examples:               []string{"Example with " + wordName},
		Source:                 "test",
		PartOfSpeechNormalized: "noun",
	}

	definitions, err := definitionService.SaveDefinition(ctx, wordAID, []*model.Definition{definition}, &tx)
	require.NoError(t, err)
	require.Len(t, definitions, 1)
	definitionID := definitions[0].ID

	err = leitnerTrackingService.IncludeDefinitions(ctx, wordAID, userID, []int64{definitionID}, tx)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	// Ensure word A is completed (not strictly required for wordlist B, but keeps data consistent).
	_, err = server.DB.Exec(ctx, "UPDATE words SET processing_status = 'completed', processing_completed_at = NOW() WHERE id = $1", wordAID)
	require.NoError(t, err)

	// Create word in wordlist B with the same name to reuse definitions.
	wordBResp := server.Expect.POST(fmt.Sprintf("/wordlists/%d/words", wordlistB)).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":  wordName,
			"notes": "test",
		}).
		Expect().
		Status(201)
	wordBID := int64(wordBResp.JSON().Object().Value("id").Number().Raw())

	// Ensure word B is completed so quizzes are available.
	_, err = server.DB.Exec(ctx, "UPDATE words SET processing_status = 'completed', processing_completed_at = NOW() WHERE id = $1", wordBID)
	require.NoError(t, err)

	// Ensure tracking row exists for word B + shared definition.
	var trackingBID int64
	err = server.DB.QueryRow(ctx,
		`SELECT id FROM leitner_system_tracking WHERE user_id = $1 AND word_id = $2 AND definition_id = $3`,
		userID, wordBID, definitionID).Scan(&trackingBID)
	require.NoError(t, err)

	// Request a quiz for wordlist B.
	quizResp := server.Expect.POST(fmt.Sprintf("/wordlists/%d/quizzes", wordlistB)).
		WithHeader("Authorization", token).
		Expect().
		Status(200)

	quizJSON := quizResp.JSON().Object()
	quizTrackingID := int64(quizJSON.Value("id").Number().Raw())
	quizWordID := int64(quizJSON.Value("wordId").Number().Raw())

	require.Equal(t, wordBID, quizWordID, "quiz should return the word_id from the requested wordlist")

	var trackingWordID int64
	err = server.DB.QueryRow(ctx, "SELECT word_id FROM leitner_system_tracking WHERE id = $1", quizTrackingID).Scan(&trackingWordID)
	require.NoError(t, err)
	require.Equal(t, wordBID, trackingWordID, "quiz tracking id should belong to the same word_id")
}
