package errorreporting

import (
	"context"
	"errors"
	"fmt"
	"sync"
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

		savedDefinitions, err := definitionService.SaveDefinition(context.Background(), wordID, []*model.Definition{testDefinition}, nil)
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
	errorResp.Header("Retry-After").NotEmpty()
	require.Positive(t, int(errorResp.JSON().Object().Value("retryAfter").Number().Raw()))
	require.Equal(t,
		errorResp.Header("Retry-After").Raw(),
		formatRetryAfter(int(errorResp.JSON().Object().Value("retryAfter").Number().Raw())),
	)
}

func TestErrorReporting_ConcurrentSubmissionsUseCommittedQuotaEvents(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	ctx := context.Background()
	token := server.WithTestUser(t)
	var userID int64
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT id FROM users ORDER BY id DESC LIMIT 1`).Scan(&userID))
	wordlistID := createTestWordlist(server, token, "Concurrent quota", "en")
	definitionService := service.NewDefinitionService(server.DB)

	type target struct {
		wordID       int64
		definitionID int64
	}
	targets := make([]target, service.FreeHourlyLimit+1)
	for i := range targets {
		wordID := createTestWord(server, token, wordlistID, fmt.Sprintf("quota-word-%d", i))
		definitionID := createTestDefinition(t, definitionService, wordID, &model.Definition{
			Token: "quota", Language: "en", PartOfSpeech: "noun",
			Meaning: "Concurrent quota meaning", Examples: []string{"A concurrent quota example."}, Source: "api_rate_1",
		})
		targets[i] = target{wordID: wordID, definitionID: definitionID}
	}

	start := make(chan struct{})
	errs := make([]error, len(targets))
	var wg sync.WaitGroup
	for index, reportTarget := range targets {
		wg.Add(1)
		go func(index int, reportTarget target) {
			defer wg.Done()
			<-start
			errs[index] = server.AppContext.ErrorReportService.ReportError(
				context.Background(), service.UnrelatedMeaning, reportTarget.wordID, &reportTarget.definitionID, userID, nil,
			)
		}(index, reportTarget)
	}
	close(start)
	wg.Wait()

	succeeded := 0
	limited := 0
	var rejectedTarget target
	for index, err := range errs {
		if err == nil {
			succeeded++
			continue
		}
		var rateLimitErr service.RateLimitError
		if errors.As(err, &rateLimitErr) {
			limited++
			rejectedTarget = targets[index]
			continue
		}
		require.NoError(t, err)
	}
	require.Equal(t, service.FreeHourlyLimit, succeeded)
	require.Equal(t, 1, limited)

	var quotaEvents, reports, hourlyCount, dailyCount int
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT count(*) FROM error_report_quota_events WHERE user_id=$1`, userID).Scan(&quotaEvents))
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT count(*) FROM error_reports WHERE user_id=$1`, userID).Scan(&reports))
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT hourly_count, daily_count FROM error_report_limits WHERE user_id=$1`, userID).Scan(&hourlyCount, &dailyCount))
	require.Equal(t, service.FreeHourlyLimit, quotaEvents)
	require.Equal(t, service.FreeHourlyLimit, reports)
	require.Equal(t, service.FreeHourlyLimit, hourlyCount)
	require.Equal(t, service.FreeHourlyLimit, dailyCount)

	err := server.AppContext.ErrorReportService.ReportError(
		ctx, service.UnrelatedMeaning, rejectedTarget.wordID, &rejectedTarget.definitionID, userID, nil,
	)
	var rateLimitErr service.RateLimitError
	require.ErrorAs(t, err, &rateLimitErr)
	require.Positive(t, rateLimitErr.RetryAfter)
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT count(*) FROM error_report_quota_events WHERE user_id=$1`, userID).Scan(&quotaEvents))
	require.Equal(t, service.FreeHourlyLimit, quotaEvents)
}

func TestErrorReporting_UpsertedSubmissionAddsCommittedQuotaEvent(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	ctx := context.Background()
	token := server.WithTestUser(t)
	var userID int64
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT id FROM users ORDER BY id DESC LIMIT 1`).Scan(&userID))
	wordlistID := createTestWordlist(server, token, "Upsert quota", "en")
	wordID := createTestWord(server, token, wordlistID, "upsert-quota")
	definitionID := createTestDefinition(t, service.NewDefinitionService(server.DB), wordID, &model.Definition{
		Token: "upsert-quota", Language: "en", PartOfSpeech: "noun",
		Meaning: "A report whose pending row is updated.", Examples: []string{"The retry is counted."}, Source: "api_rate_1",
	})

	require.NoError(t, server.AppContext.ErrorReportService.ReportError(
		ctx, service.UnrelatedMeaning, wordID, &definitionID, userID, nil,
	))
	_, err := server.DB.Exec(ctx, `
		UPDATE error_report_cooldowns
		SET cooldown_until=NOW() - INTERVAL '1 second'
		WHERE user_id=$1 AND word_id=$2 AND definition_id=$3 AND error_type=$4
	`, userID, wordID, definitionID, service.UnrelatedMeaning)
	require.NoError(t, err)
	require.NoError(t, server.AppContext.ErrorReportService.ReportError(
		ctx, service.UnrelatedMeaning, wordID, &definitionID, userID, nil,
	))

	var quotaEvents, pendingReports, hourlyCount, dailyCount int
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT count(*) FROM error_report_quota_events WHERE user_id=$1`, userID).Scan(&quotaEvents))
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT count(*) FROM error_reports WHERE user_id=$1 AND status='pending'`, userID).Scan(&pendingReports))
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT hourly_count, daily_count FROM error_report_limits WHERE user_id=$1`, userID).Scan(&hourlyCount, &dailyCount))
	require.Equal(t, 2, quotaEvents)
	require.Equal(t, 1, pendingReports, "the existing pending report is updated in place")
	require.Equal(t, 2, hourlyCount)
	require.Equal(t, 2, dailyCount)
}

func formatRetryAfter(value int) string {
	return fmt.Sprintf("%d", value)
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

		savedDefinitions, err := definitionService.SaveDefinition(context.Background(), wordID, []*model.Definition{testDefinition}, nil)
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

	savedDefinitions, err := definitionService.SaveDefinition(context.Background(), wordID, []*model.Definition{testDefinition}, nil)
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
