package errorreporting

import (
	"testing"

	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/require"
)

// TestErrorReporting_RateLimit_FreeUser tests that free users are properly rate limited
func TestErrorReporting_RateLimit_FreeUser(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	// Create services for definition operations
	definitionService := service.NewDefinitionService(server.DB)

	token := server.WithTestUser(t)

	// Create wordlist
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Rate Limit Test",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	// Create multiple words and definitions to avoid cooldown conflicts
	words := []struct {
		name     string
		meaning  string
		examples []string
	}{
		{"test1", "First test word", []string{"Example 1"}},
		{"test2", "Second test word", []string{"Example 2"}},
		{"test3", "Third test word", []string{"Example 3"}},
		{"test4", "Fourth test word", []string{"Example 4"}},
	}

	var errorReports []map[string]interface{}

	for i, word := range words {
		// Add word
		addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
			WithHeader("Authorization", token).
			WithJSON(map[string]interface{}{
				"name": word.name,
			}).
			Expect().
			Status(201)

		wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

		// Create definition
		testDefinition := &model.Definition{
			Token:        word.name,
			Language:     "en",
			PartOfSpeech: "noun",
			Meaning:      word.meaning,
			Examples:     word.examples,
			Source:       "test_rate_limit",
		}

		savedDefinitions, err := definitionService.SaveDefinition(wordID, []*model.Definition{testDefinition}, nil)
		require.NoError(t, err)
		require.Len(t, savedDefinitions, 1)

		definitionID := savedDefinitions[0].ID

		errorReports = append(errorReports, map[string]interface{}{
			"wordId":       wordID,
			"definitionId": definitionID,
			"errorType":    "_unrelated_meaning",
		})

		// Submit error reports for first 3 words - should succeed
		if i < 3 {
			server.Expect.POST("/errorReports").
				WithHeader("Authorization", token).
				WithJSON(errorReports[i]).
				Expect().
				Status(200)
		}
	}

	// Submit fourth error report - should be rate limited (exceeds free user hourly limit of 3)
	errorResp := server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(errorReports[3]).
		Expect().
		Status(429)

	// Verify rate limit error message
	errorResp.JSON().Object().HasValue("error", "Hourly limit exceeded. You can report 3 errors per hour.")
}

// TestErrorReporting_RateLimit_PremiumUser tests that premium users have higher rate limits
func TestErrorReporting_RateLimit_PremiumUser(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	// Create services for definition operations
	definitionService := service.NewDefinitionService(server.DB)

	token := server.WithPremiumUser(t)

	// Create wordlist
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Premium Rate Test",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	// Create multiple words and definitions to avoid cooldown conflicts
	// Premium users have 10 reports per hour, so test 10 successful + 1 blocked to verify limit
	words := []struct {
		name     string
		meaning  string
		examples []string
	}{
		{"premium1", "First premium word", []string{"Premium example 1"}},
		{"premium2", "Second premium word", []string{"Premium example 2"}},
		{"premium3", "Third premium word", []string{"Premium example 3"}},
		{"premium4", "Fourth premium word", []string{"Premium example 4"}},
		{"premium5", "Fifth premium word", []string{"Premium example 5"}},
		{"premium6", "Sixth premium word", []string{"Premium example 6"}},
		{"premium7", "Seventh premium word", []string{"Premium example 7"}},
		{"premium8", "Eighth premium word", []string{"Premium example 8"}},
		{"premium9", "Ninth premium word", []string{"Premium example 9"}},
		{"premium10", "Tenth premium word", []string{"Premium example 10"}},
		{"premium11", "Eleventh premium word", []string{"Premium example 11"}},
	}

	var errorReports []map[string]interface{}

	for _, word := range words {
		// Add word
		addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
			WithHeader("Authorization", token).
			WithJSON(map[string]interface{}{
				"name": word.name,
			}).
			Expect().
			Status(201)

		wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

		// Create definition
		testDefinition := &model.Definition{
			Token:        word.name,
			Language:     "en",
			PartOfSpeech: "adjective",
			Meaning:      word.meaning,
			Examples:     word.examples,
			Source:       "test_premium_rate",
		}

		savedDefinitions, err := definitionService.SaveDefinition(wordID, []*model.Definition{testDefinition}, nil)
		require.NoError(t, err)
		require.Len(t, savedDefinitions, 1)

		definitionID := savedDefinitions[0].ID

		errorReports = append(errorReports, map[string]interface{}{
			"wordId":       wordID,
			"definitionId": definitionID,
			"errorType":    "_unrelated_meaning",
		})
	}

	// Submit first 10 error reports - should all succeed for premium users (limit is 10/hour)
	for i := 0; i < 10; i++ {
		server.Expect.POST("/errorReports").
			WithHeader("Authorization", token).
			WithJSON(errorReports[i]).
			Expect().
			Status(200).
			JSON().Object().
			NotContainsKey("error") // Should not get rate limit error

		// Log progress for verification
		t.Logf("Premium user successfully submitted error report %d/10", i+1)
	}

	// Submit 11th error report - should be rate limited (exceeds premium user hourly limit of 10)
	errorResp := server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(errorReports[10]). // 11th report
		Expect().
		Status(429)

	// Verify rate limit error message for premium users
	errorResp.JSON().Object().HasValue("error", "Hourly limit exceeded. You can report 10 errors per hour.")

	t.Log("Premium user rate limiting test completed successfully - 10 reports succeeded, 11th was properly blocked")
}

// TestErrorReporting_CooldownMechanism tests that cooldown prevents duplicate reports
func TestErrorReporting_CooldownMechanism(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	// Create services for definition operations
	definitionService := service.NewDefinitionService(server.DB)

	token := server.WithTestUser(t)

	// Create wordlist and word
	createResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name":         "Cooldown Test",
			"languageCode": "en",
		}).
		Expect().
		Status(201)

	wordlistID := int(createResp.JSON().Object().Value("id").Number().Raw())

	addWordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
		WithHeader("Authorization", token).
		WithJSON(map[string]interface{}{
			"name": "cooldown",
		}).
		Expect().
		Status(201)

	wordID := int64(addWordResp.JSON().Object().Value("id").Number().Raw())

	// Create definition
	testDefinition := &model.Definition{
		Token:        "cooldown",
		Language:     "en",
		PartOfSpeech: "noun",
		Meaning:      "A period of time to wait",
		Examples:     []string{"The cooldown period is one hour"},
		Source:       "test_cooldown",
	}

	savedDefinitions, err := definitionService.SaveDefinition(wordID, []*model.Definition{testDefinition}, nil)
	require.NoError(t, err)
	require.Len(t, savedDefinitions, 1)

	definitionID := savedDefinitions[0].ID

	errorReport := map[string]interface{}{
		"wordId":       wordID,
		"definitionId": definitionID,
		"errorType":    "_unrelated_meaning",
	}

	// Submit first error report - should succeed
	server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(errorReport).
		Expect().
		Status(200)

	// Submit same error report immediately - should be blocked by cooldown
	errorResp := server.Expect.POST("/errorReports").
		WithHeader("Authorization", token).
		WithJSON(errorReport).
		Expect().
		Status(429)

	// Verify cooldown error message contains retry information
	responseObj := errorResp.JSON().Object()
	responseObj.ContainsKey("error")
	responseObj.ContainsKey("cooldownUntil")
	responseObj.ContainsKey("retryAfter")
	responseObj.HasValue("windowType", "cooldown")

	// Error message should mention waiting period
	errorMsg := responseObj.Value("error").String().Raw()
	require.Contains(t, errorMsg, "Please wait before reporting this error again", "Cooldown message should mention waiting period")
}
