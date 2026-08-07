package integration

import (
	"context"
	"testing"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/require"
)

// TestLeitnerNullUpdatedAtAssumption verifies our core assumption that updated_at is never NULL
func TestLeitnerNullUpdatedAtAssumption(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	ctx := context.Background()
	_, err := server.AppContext.UserService.SaveUser(ctx, "Boundary", "Sentinel", "testpass123", "boundary-sentinel@example.com", nil, nil)
	require.NoError(t, err)

	// Create basic test data using services
	userID, wordlistID := createBasicLeitnerDataUsingServices(ctx, t, server)
	require.NotEqual(t, userID, wordlistID, "fixture IDs must differ to detect swapped scope fields")

	t.Run("VerifyNoNullUpdatedAtInProduction", func(t *testing.T) {
		var nullCount int
		err := server.DB.QueryRow(ctx,
			"SELECT COUNT(*) FROM leitner_system_tracking WHERE updated_at IS NULL").Scan(&nullCount)
		require.NoError(t, err)
		require.Equal(t, 0, nullCount, "Found NULL updated_at values - this breaks our algorithm assumptions")
	})

	t.Run("QuizGenerationWorksAfterDeadCodeCleanup", func(t *testing.T) {
		// Test that quiz generation still works with our existing data patterns
		var analyticsWriter service.LeitnerAnalyticsWriter = server.AppContext.AnalyticsService
		strategy := service.NewLeitnerSystemStrategy(
			server.DB,
			server.AppContext.WordService,
			server.AppContext.DefinitionService,
			&analyticsWriter,
			service.NewLeitnerTrackingService(server.DB),
		)

		quiz, err := strategy.CreateQuiz(ctx, wordlistID, userID, []model.QuizType{model.GuessMeaning})
		require.NoError(t, err)
		require.NotNil(t, quiz)
		require.Equal(t, model.GuessMeaning, quiz.Type)
		require.Len(t, quiz.Options, 3, "two scoped distractors plus the answer are required")
	})

	t.Run("Box1WordsHaveCorrectPriority", func(t *testing.T) {
		// Verify Box 1 words get highest priority weight
		query := `
			SELECT 
				CASE 
					WHEN lst.box_id = 1 THEN 1.0  -- Box 1 always 100% ready
					ELSE (EXTRACT(EPOCH FROM (NOW() - lst.updated_at))/3600) / NULLIF(CASE lst.box_id
						WHEN 2 THEN 6
						WHEN 3 THEN 24
						WHEN 4 THEN 72
						WHEN 5 THEN 168
						WHEN 6 THEN 336
						WHEN 7 THEN 720
					END, 0)
				END as progress_ratio
			FROM leitner_system_tracking lst 
			WHERE lst.user_id = $1 AND lst.box_id = 1
			LIMIT 1`

		var progressRatio float64
		err := server.DB.QueryRow(ctx, query, userID).Scan(&progressRatio)
		require.NoError(t, err)
		require.Equal(t, 1.0, progressRatio, "Box 1 words should have progress_ratio = 1.0")
	})
}

func TestCreateQuizHandlesInsufficientScopedDistractors(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	ctx := context.Background()

	user, err := server.AppContext.UserService.SaveUser(ctx, "Solo", "Learner", "testpass123", "solo-learner@example.com", nil, nil)
	require.NoError(t, err)
	wordlist, err := server.AppContext.WordlistService.SaveWordlist(ctx, &service.Wordlist{
		Name:         "Solo list",
		Description:  "No distractor candidates",
		UserID:       user.ID,
		LanguageCode: "en",
	})
	require.NoError(t, err)
	word, err := server.AppContext.WordService.SaveWord(ctx, &service.Word{
		Name:       "solo",
		UserID:     user.ID,
		WordlistID: wordlist.ID,
	})
	require.NoError(t, err)
	_, err = server.DB.Exec(ctx, `UPDATE words SET processing_status = 'completed' WHERE id = $1`, word.ID)
	require.NoError(t, err)

	tx, err := server.DB.Begin(ctx)
	require.NoError(t, err)
	definitions, err := server.AppContext.DefinitionService.SaveDefinition(ctx, word.ID, []*model.Definition{{
		Token:        "solo",
		Language:     "en",
		PartOfSpeech: "noun",
		Meaning:      "done by one person",
		Examples:     []string{"A [solo] attempt."},
		Source:       "ChatGPT",
	}}, &tx)
	require.NoError(t, err)
	require.Len(t, definitions, 1)
	require.NoError(t, server.AppContext.LeitnerTrackingService.IncludeDefinitions(ctx, word.ID, user.ID, []int64{definitions[0].ID}, tx))
	require.NoError(t, tx.Commit(ctx))

	var analyticsWriter service.LeitnerAnalyticsWriter = server.AppContext.AnalyticsService
	strategy := service.NewLeitnerSystemStrategy(
		server.DB,
		server.AppContext.WordService,
		server.AppContext.DefinitionService,
		&analyticsWriter,
		service.NewLeitnerTrackingService(server.DB),
	)

	quiz, err := strategy.CreateQuiz(ctx, wordlist.ID, user.ID, nil)
	require.NoError(t, err)
	require.Equal(t, model.WriteWordFromDefinition, quiz.Type)
	require.Equal(t, []string{"solo"}, quiz.Options)

	_, err = strategy.CreateQuiz(ctx, wordlist.ID, user.ID, []model.QuizType{model.GuessMeaning})
	var unavailable common.QuizUnavailableError
	require.ErrorAs(t, err, &unavailable)
	require.Equal(t, common.QuizUnavailableNoMatchingQuizTypes, unavailable.Reason)
}

// createBasicLeitnerDataUsingServices creates minimal test data using service calls
func createBasicLeitnerDataUsingServices(ctx context.Context, t *testing.T, server *setup.TestServer) (userID, wordlistID int64) {
	// 1. Create test user using UserService
	user, err := server.AppContext.UserService.SaveUser(ctx, "Test", "User", "testpass123", "test@example.com", nil, nil)
	require.NoError(t, err)
	require.NotNil(t, user)
	userID = user.ID

	// 2. Create test wordlist using WordlistService
	wordlist := &service.Wordlist{
		Name:         "Test Wordlist",
		Description:  "For testing",
		UserID:       userID,
		LanguageCode: "en",
	}
	createdWordlist, err := server.AppContext.WordlistService.SaveWordlist(ctx, wordlist)
	require.NoError(t, err)
	require.NotNil(t, createdWordlist)
	wordlistID = createdWordlist.ID

	// 3. Create test words using WordService
	word1 := &service.Word{
		Name:       "test1",
		Notes:      "",
		UserID:     userID,
		WordlistID: wordlistID,
	}
	createdWord1, err := server.AppContext.WordService.SaveWord(ctx, word1)
	require.NoError(t, err)
	require.NotNil(t, createdWord1)

	word2 := &service.Word{
		Name:       "test2",
		Notes:      "",
		UserID:     userID,
		WordlistID: wordlistID,
	}
	createdWord2, err := server.AppContext.WordService.SaveWord(ctx, word2)
	require.NoError(t, err)
	require.NotNil(t, createdWord2)
	word3 := &service.Word{
		Name:       "test3",
		Notes:      "",
		UserID:     userID,
		WordlistID: wordlistID,
	}
	createdWord3, err := server.AppContext.WordService.SaveWord(ctx, word3)
	require.NoError(t, err)
	require.NotNil(t, createdWord3)
	word4 := &service.Word{
		Name:       "test4",
		Notes:      "",
		UserID:     userID,
		WordlistID: wordlistID,
	}
	createdWord4, err := server.AppContext.WordService.SaveWord(ctx, word4)
	require.NoError(t, err)
	require.NotNil(t, createdWord4)

	_, err = server.DB.Exec(ctx, `
		UPDATE words
		SET processing_status = 'completed'
		WHERE id = ANY($1)
	`, []int64{createdWord1.ID, createdWord2.ID, createdWord3.ID, createdWord4.ID})
	require.NoError(t, err)

	// 4. Since background jobs don't run in tests, manually create definitions and Leitner tracking
	tx, err := server.DB.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	// Create definitions for the words
	definitions1 := []*model.Definition{{
		Token:                  "test1",
		Language:               "en",
		PartOfSpeech:           "noun",
		Meaning:                "A test definition 1",
		Examples:               []string{"This is a [test1] example."},
		Source:                 "ChatGPT",
		PartOfSpeechNormalized: "noun",
	}}
	savedDefs1, err := server.AppContext.DefinitionService.SaveDefinition(ctx, createdWord1.ID, definitions1, &tx)
	require.NoError(t, err)
	require.Len(t, savedDefs1, 1)

	definitions2 := []*model.Definition{{
		Token:                  "test2",
		Language:               "en",
		PartOfSpeech:           "adverb",
		Meaning:                "A test definition 2",
		Examples:               []string{"This is a [test2] example."},
		Source:                 "ChatGPT",
		PartOfSpeechNormalized: "adverb",
	}}
	savedDefs2, err := server.AppContext.DefinitionService.SaveDefinition(ctx, createdWord2.ID, definitions2, &tx)
	require.NoError(t, err)
	require.Len(t, savedDefs2, 1)

	definitions3 := []*model.Definition{{
		Token:                  "test3",
		Language:               "en",
		PartOfSpeech:           "noun",
		Meaning:                "A test definition 3",
		Examples:               []string{"This is a [test3] example."},
		Source:                 "ChatGPT",
		PartOfSpeechNormalized: "noun",
	}}
	savedDefs3, err := server.AppContext.DefinitionService.SaveDefinition(ctx, createdWord3.ID, definitions3, &tx)
	require.NoError(t, err)
	require.Len(t, savedDefs3, 1)

	definitions4 := []*model.Definition{{
		Token:                  "test4",
		Language:               "en",
		PartOfSpeech:           "noun",
		Meaning:                "A test definition 4",
		Examples:               []string{"This is a [test4] example."},
		Source:                 "ChatGPT",
		PartOfSpeechNormalized: "noun",
	}}
	savedDefs4, err := server.AppContext.DefinitionService.SaveDefinition(ctx, createdWord4.ID, definitions4, &tx)
	require.NoError(t, err)
	require.Len(t, savedDefs4, 1)

	// Include definitions in Leitner tracking system (this sets updated_at to NOW())
	err = server.AppContext.LeitnerTrackingService.IncludeDefinitions(ctx, createdWord1.ID, userID, []int64{savedDefs1[0].ID}, tx)
	require.NoError(t, err)

	err = server.AppContext.LeitnerTrackingService.IncludeDefinitions(ctx, createdWord2.ID, userID, []int64{savedDefs2[0].ID}, tx)
	require.NoError(t, err)

	err = server.AppContext.LeitnerTrackingService.IncludeDefinitions(ctx, createdWord3.ID, userID, []int64{savedDefs3[0].ID}, tx)
	require.NoError(t, err)

	err = server.AppContext.LeitnerTrackingService.IncludeDefinitions(ctx, createdWord4.ID, userID, []int64{savedDefs4[0].ID}, tx)
	require.NoError(t, err)

	// Manually set one definition to box 2 with earlier timestamp to test prioritization
	_, err = tx.Exec(ctx, `
		UPDATE leitner_system_tracking 
		SET box_id = 2,
			updated_at = NOW() - INTERVAL '10 hours',
			next_review_at = NOW() - INTERVAL '4 hours'
		WHERE definition_id = $1 AND user_id = $2
	`, savedDefs2[0].ID, userID)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	return userID, wordlistID
}
