package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"decorebator.com/tests/integration/setup"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// TestLeitnerQuizLoop_ReproducesProductionIssue reproduces the quiz loop issue from production
// where the user lsimaocosta+bs1@gmail.com gets stuck in a loop with the same quiz
func TestLeitnerQuizLoop_ReproducesProductionIssue(t *testing.T) {
	server := setup.NewTestServer(t)

	// Seed production data
	seedProductionData(t, server.DB)

	// Login as the production user
	loginResp := server.Expect.POST("/login").
		WithJSON(map[string]interface{}{
			"email":    "lsimaocosta+bs1@gmail.com",
			"password": "test123", // We'll use a test password
		}).
		Expect().
		Status(http.StatusOK)

	token := loginResp.Header("authorization").NotEmpty().Raw()

	// Request quizzes 100 times, answering correctly each time
	runQuizLoopTest(t, server, token, 100)
}

// TestLeitnerQuizLoop_50Iterations tests the algorithm with fewer iterations
func TestLeitnerQuizLoop_50Iterations(t *testing.T) {
	server := setup.NewTestServer(t)

	seedProductionData(t, server.DB)

	loginResp := server.Expect.POST("/login").
		WithJSON(map[string]interface{}{
			"email":    "lsimaocosta+bs1@gmail.com",
			"password": "test123",
		}).
		Expect().
		Status(http.StatusOK)

	token := loginResp.Header("authorization").NotEmpty().Raw()
	runQuizLoopTest(t, server, token, 50)
}

// TestLeitnerQuizLoop_200Iterations tests the algorithm with more iterations
func TestLeitnerQuizLoop_200Iterations(t *testing.T) {
	server := setup.NewTestServer(t)

	seedProductionData(t, server.DB)

	loginResp := server.Expect.POST("/login").
		WithJSON(map[string]interface{}{
			"email":    "lsimaocosta+bs1@gmail.com",
			"password": "test123",
		}).
		Expect().
		Status(http.StatusOK)

	token := loginResp.Header("authorization").NotEmpty().Raw()
	runQuizLoopTest(t, server, token, 200)
}

// TestLeitnerQuizLoop_500Iterations tests the algorithm with many iterations
func TestLeitnerQuizLoop_500Iterations(t *testing.T) {
	server := setup.NewTestServer(t)

	seedProductionData(t, server.DB)

	loginResp := server.Expect.POST("/login").
		WithJSON(map[string]interface{}{
			"email":    "lsimaocosta+bs1@gmail.com",
			"password": "test123",
		}).
		Expect().
		Status(http.StatusOK)

	token := loginResp.Header("authorization").NotEmpty().Raw()
	runQuizLoopTest(t, server, token, 500)
}

// runQuizLoopTest executes the quiz loop test with the specified number of iterations
func runQuizLoopTest(t *testing.T, server *setup.TestServer, token string, iterations int) {
	// Track quiz types received
	quizTypesReceived := make(map[string]int)
	quizDefinitionIDs := make(map[int64]int)

	// Request quizzes, answering correctly each time
	for i := 0; i < iterations; i++ {
		// Create quiz
		quizResp := server.Expect.POST("/wordlists/7/quizzes").
			WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
			Expect()

		// Check if we got an error response
		if quizResp.Raw().StatusCode != http.StatusOK {
			t.Fatalf("Failed to get quiz at iteration %d - status code: %d, response: %s",
				i+1, quizResp.Raw().StatusCode, quizResp.Body().Raw())
		}

		// The quiz is returned directly, not wrapped in an object
		quiz := quizResp.JSON().Object()
		quizType := quiz.Value("type").String().Raw()
		quizID := int64(quiz.Value("id").Number().Raw())
		definitionID := int64(quiz.Value("definitionId").Number().Raw())
		wordID := int64(quiz.Value("wordId").Number().Raw())

		// Track quiz types and definitions
		quizTypesReceived[quizType]++
		quizDefinitionIDs[definitionID]++

		// Log every 10th iteration or every 50th for large tests
		logInterval := 10
		if iterations > 200 {
			logInterval = 50
		}
		if (i+1)%logInterval == 0 {
			t.Logf("Iteration %d: Quiz type=%s, definitionID=%d", i+1, quizType, definitionID)
		}

		// Save quiz result (answer correctly)
		server.Expect.PATCH("/wordlists/7/quizzes").
			WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
			WithJSON(map[string]interface{}{
				"wordlistID":              7,
				"wordID":                  wordID,
				"definitionID":            definitionID,
				"leitnerSystemTrackingID": quizID,
				"quizType":                quizType,
				"isCorrect":               true,
				"responseTimeMs":          1000,
			}).
			Expect().
			Status(http.StatusNoContent)

		// Small delay to simulate real usage
		time.Sleep(5 * time.Millisecond)
	}

	// Verify results
	t.Logf("\n=== Quiz Type Distribution (%d iterations) ===", iterations)
	for quizType, count := range quizTypesReceived {
		percentage := float64(count) / float64(iterations) * 100
		t.Logf("%s: %d times (%.1f%%)", quizType, count, percentage)
	}

	t.Logf("\n=== Definition Distribution ===")
	for defID, count := range quizDefinitionIDs {
		percentage := float64(count) / float64(iterations) * 100
		t.Logf("Definition %d: %d times (%.1f%%)", defID, count, percentage)
	}

	// Check that we got variety in quiz types
	require.Greater(t, len(quizTypesReceived), 1, "Should receive multiple quiz types")

	// Check that all quiz types that should be available were used
	expectedQuizTypes := []string{
		"GUESS_MEANING", "WORD_FROM_MEANING", "WORD_FROM_IMAGE", "COMPLETE_SENTENCE",
		"WRITE_WORD_FROM_DEFINITION", "WORD_FROM_AUDIO", "WORD_FROM_EXAMPLE_AUDIO", "MEANING_FROM_AUDIO",
	}

	// Log which quiz types were never received
	t.Logf("\n=== Missing Quiz Types ===")
	for _, expectedType := range expectedQuizTypes {
		if _, found := quizTypesReceived[expectedType]; !found {
			t.Logf("Never received: %s", expectedType)
		}
	}

	// Check for the stuck quiz problem - if one definition dominates
	maxCount := 0
	var dominantDefID int64
	for defID, count := range quizDefinitionIDs {
		if count > maxCount {
			maxCount = count
			dominantDefID = defID
		}
	}

	dominanceRatio := float64(maxCount) / float64(iterations)
	t.Logf("\n=== Dominance Analysis ===")
	t.Logf("Most frequent definition: %d appeared %d times (%.1f%%)", dominantDefID, maxCount, dominanceRatio*100)

	// If one definition appears more than 80% of the time, we have the stuck quiz problem
	if dominanceRatio > 0.8 {
		t.Errorf("Quiz stuck problem detected! Definition %d dominated %.1f%% of quizzes", dominantDefID, dominanceRatio*100)
	}

	// Additional checks for balanced quiz types
	for quizType, count := range quizTypesReceived {
		percentage := float64(count) / float64(iterations) * 100
		// Flag if any single quiz type dominates too much (>40%)
		if percentage > 40 {
			t.Logf("WARNING: Quiz type %s appears %.1f%% which may be too frequent", quizType, percentage)
		}
	}
}

func seedProductionData(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	// Start transaction
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	// 1. Insert user (with test password)
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, first_name, last_name, email, password_hash, subscription_plan)
		VALUES (5, 'lucas', 'Simao', 'lsimaocosta+bs1@gmail.com', 
			'$2a$10$.TeyMV.FtW8wJUKFkIGJXe3E4dmAsnxwe83HhlRBdQms2QcIj3vN6', 'monthly')
		ON CONFLICT (id) DO UPDATE SET subscription_plan = EXCLUDED.subscription_plan
	`) // password is "test123"
	require.NoError(t, err)

	// 2. Insert wordlist
	_, err = tx.Exec(ctx, `
		INSERT INTO wordlists (id, name, description, user_id, language_code, words_count, pronunciation_system)
		VALUES (7, 'Test browser stack1', '', 5, 'en', 2, 'ipa')
		ON CONFLICT (id) DO NOTHING
	`)
	require.NoError(t, err)

	// 3. Insert words
	_, err = tx.Exec(ctx, `
		INSERT INTO words (id, name, notes, learned, user_id, wordlist_id, pronunciation, audio_url)
		VALUES 
			(168, 'shrill', '', false, 5, 7, '', 'https://decorebator.nyc3.digitaloceanspaces.com/audio/audio-138-shrill.mp3'),
			(169, 'crests', '', false, 5, 7, '', 'https://decorebator.nyc3.digitaloceanspaces.com/audio/audio-131-crests.mp3')
		ON CONFLICT (id) DO NOTHING
	`)
	require.NoError(t, err)

	// 4. Insert definitions
	_, err = tx.Exec(ctx, `
		INSERT INTO definitions (id, token, language, part_of_speech, meaning, examples, inflections, source, sounds, phonetic_notations, part_of_speech_normalized)
		VALUES 
			(259, 'crests', 'en', 'noun', 'The top or highest part of something, especially a mountain or wave.', 
				'{"The mountain has two prominent [crests].","Waves form high [crests] during storms."}', 
				'[]'::jsonb, 'ChatGPT', null, null, 'noun'),
			(260, 'crests', 'en', 'verb', 'To reach the top or peak of something.', 
				'{"The eagle [crests] the peak.","She [crests] the hill each morning on her run.","The sun [crests] the horizon at dawn.","The eagle [crested] the peak yesterday.","She [crested] the hill every morning last week.","The sun [crested] the horizon at dawn yesterday.","The eagle has [crested] many mountains.","She had [crested] all the hills before noon.","The sun had [crested] the horizon by the time we arrived."}',
				'[{"tense": "present", "examples": ["The eagle [crests] the peak.", "She [crests] the hill each morning on her run.", "The sun [crests] the horizon at dawn."], "inflection": "crests"}, {"tense": "past", "examples": ["The eagle [crested] the peak yesterday.", "She [crested] the hill every morning last week.", "The sun [crested] the horizon at dawn yesterday."], "inflection": "crested"}, {"tense": "past participle", "examples": ["The eagle has [crested] many mountains.", "She had [crested] all the hills before noon.", "The sun had [crested] the horizon by the time we arrived."], "inflection": "crested"}]'::jsonb,
				'ChatGPT', null, null, 'verb'),
			(273, 'shrill', 'en', 'adjective', 'Having a high-pitched and piercing sound that is often unpleasant to hear.', 
				'{"The [shrill] noise of the alarm was unbearable.","Her [shrill] voice could be heard across the room.","He let out a [shrill] whistle to grab everyone''s attention."}',
				'[]'::jsonb, 'ChatGPT', null, null, 'adjective'),
			(274, 'shrill', 'en', 'verb', 'To make a high-pitched and piercing sound.', 
				'{"The siren continued to [shrill] in the distance.","He would often [shrill] at his pet to stop misbehaving.","The birds began to [shrill] at dawn."}',
				'[{"tense": "present", "examples": ["The siren [shrills] every time there is an emergency.", "He [shrills] angrily at the noisy neighbors.", "As the teacher [shrills] at the students to silence them, they went quiet."], "inflection": "shrills"}, {"tense": "past", "examples": ["The child [shrilled] with delight at the sight of the ice cream.", "He [shrilled] loudly when he stubbed his toe.", "The alarm [shrilled] in the night, waking everyone up."], "inflection": "shrilled"}, {"tense": "past participle", "examples": ["The bird has [shrilled] joyfully in the morning.", "She has [shrilled] at the top of her lungs in anger.", "The device, once activated, has [shrilled] without stopping."], "inflection": "shrilled"}]'::jsonb,
				'ChatGPT', null, null, 'verb')
		ON CONFLICT (id) DO NOTHING
	`)
	require.NoError(t, err)

	// 5. Insert word_definitions mapping
	_, err = tx.Exec(ctx, `
		INSERT INTO word_definitions (word_id, definition_id)
		VALUES 
			(168, 273), (168, 274),
			(169, 259), (169, 260)
		ON CONFLICT DO NOTHING
	`)
	require.NoError(t, err)

	// 6. Insert leitner_system_tracking (with current state from production)
	// Note: Clear temporarily_skipped_until as those are in the past
	_, err = tx.Exec(ctx, `
		INSERT INTO leitner_system_tracking (id, user_id, definition_id, box_id, word_id, updated_at, temporarily_skipped_until)
		VALUES 
			(315, 5, 273, 7, 168, NOW() - INTERVAL '20 minutes', null),
			(316, 5, 274, 1, 168, NOW() - INTERVAL '45 minutes', null),
			(317, 5, 259, 4, 169, NOW() - INTERVAL '18 minutes', null),
			(318, 5, 260, 1, 169, NOW() - INTERVAL '20 minutes', null)
		ON CONFLICT (id) DO UPDATE SET 
			box_id = EXCLUDED.box_id,
			updated_at = EXCLUDED.updated_at,
			temporarily_skipped_until = EXCLUDED.temporarily_skipped_until
	`)
	require.NoError(t, err)

	// 7. Insert definition_images
	_, err = tx.Exec(ctx, `
		INSERT INTO definition_images (id, definition_id, url, description, is_visible, api, model, prompt)
		VALUES 
			(120, 259, 'https://decorebator.nyc3.digitaloceanspaces.com/images/definition-259-1749580933.png', 'The mountain has two prominent [crests].', true, 'openai', 'dall-e-3', 'The mountain has two prominent [crests].'),
			(119, 260, 'https://decorebator.nyc3.digitaloceanspaces.com/images/definition-260-1749580932.png', 'The sun had [crested] the horizon by the time we arrived.', true, 'openai', 'dall-e-3', 'The sun had [crested] the horizon by the time we arrived.'),
			(133, 273, 'https://decorebator.nyc3.digitaloceanspaces.com/images/definition-273-1749650851.png', 'He let out a [shrill] whistle to grab everyone''s attention.', true, 'openai', 'dall-e-3', 'He let out a [shrill] whistle to grab everyone''s attention.'),
			(134, 274, 'https://decorebator.nyc3.digitaloceanspaces.com/images/definition-274-1749650853.png', 'He would often [shrill] at his pet to stop misbehaving.', true, 'openai', 'dall-e-3', 'He would often [shrill] at his pet to stop misbehaving.')
		ON CONFLICT (id) DO NOTHING
	`)
	require.NoError(t, err)

	// 8. Insert definition_example_audio
	_, err = tx.Exec(ctx, `
		INSERT INTO definition_example_audio (id, definition_id, example_text, example_hash, audio_url, inflection_type)
		VALUES 
			(60, 259, 'The mountain has two prominent [crests].', '711c2cee3663b255f83563aea6ac2622', 'https://decorebator.nyc3.digitaloceanspaces.com/audio/example-259-711c2cee.mp3', ''),
			(61, 259, 'Waves form high [crests] during storms.', 'dfbc381b15ee2f89363bc3063167a507', 'https://decorebator.nyc3.digitaloceanspaces.com/audio/example-259-dfbc381b.mp3', ''),
			(80, 260, 'The sun had [crested] the horizon by the time we arrived.', '04729d158fbd5f6e6ad325d13c9c4672', 'https://decorebator.nyc3.digitaloceanspaces.com/audio/example-260-04729d15.mp3', ''),
			(83, 260, 'She [crests] the hill each morning on her run.', '91e971f4d6f1db02eab67070244aeeae', 'https://decorebator.nyc3.digitaloceanspaces.com/audio/example-260-91e971f4.mp3', 'present'),
			(84, 260, 'The sun [crested] the horizon at dawn yesterday.', '89abfb64b660b985e9ee0986d52cc02b', 'https://decorebator.nyc3.digitaloceanspaces.com/audio/example-260-89abfb64.mp3', 'past'),
			(103, 273, 'The [shrill] noise of the alarm was unbearable.', '214abc3c72a18db65c0469c18e11e7b7', 'https://decorebator.nyc3.digitaloceanspaces.com/audio/example-273-214abc3c.mp3', ''),
			(105, 273, 'Her [shrill] voice could be heard across the room.', 'cc851c89af95084fe22176c386e728ee', 'https://decorebator.nyc3.digitaloceanspaces.com/audio/example-273-cc851c89.mp3', ''),
			(108, 273, 'He let out a [shrill] whistle to grab everyone''s attention.', 'd954fb8e1d9e08219b8565fd1f5649c7', 'https://decorebator.nyc3.digitaloceanspaces.com/audio/example-273-d954fb8e.mp3', ''),
			(102, 274, 'He would often [shrill] at his pet to stop misbehaving.', '0ab6316137b6773a0a374296130eafd6', 'https://decorebator.nyc3.digitaloceanspaces.com/audio/example-274-0ab63161.mp3', ''),
			(104, 274, 'As the teacher [shrills] at the students to silence them, they went quiet.', '0d8491a0dc7dd4b3f1bf19ff7d916357', 'https://decorebator.nyc3.digitaloceanspaces.com/audio/example-274-0d8491a0.mp3', 'present'),
			(107, 274, 'The child [shrilled] with delight at the sight of the ice cream.', '2000aea9eff5b518cbd7eb4079373015', 'https://decorebator.nyc3.digitaloceanspaces.com/audio/example-274-2000aea9.mp3', 'past'),
			(109, 274, 'The device, once activated, has [shrilled] without stopping.', 'bbc89cf6c7341fb98939c2b7601818e9', 'https://decorebator.nyc3.digitaloceanspaces.com/audio/example-274-bbc89cf6.mp3', 'past participle')
		ON CONFLICT (id) DO NOTHING
	`)
	require.NoError(t, err)

	// Commit transaction
	err = tx.Commit(ctx)
	require.NoError(t, err)

	t.Log("Production data seeded successfully")
}
