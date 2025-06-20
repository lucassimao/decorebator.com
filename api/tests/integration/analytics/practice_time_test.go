package analytics

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPracticeTimeEndpoint_TimeCalculations tests practice time analytics calculations
func TestPracticeTimeEndpoint_TimeCalculations(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Setup test data with known practice time metrics using API calls
	testData := setupPracticeTimeTestData(t, server, token)

	// Test the practice time endpoint with default days (7)
	response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/practice-time", testData.WordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	json.HasValue("wordlistId", testData.WordlistID)
	json.HasValue("days", 7) // Default is 7 days for practice time
	json.ContainsKey("practiceTime")

	practiceTime := json.Value("practiceTime").Array()
	assert.Equal(t, len(testData.ExpectedPracticeTime), int(practiceTime.Length().Raw()),
		"Should return practice time data for all test dates")

	// Verify each day's practice time calculations
	for i, dayData := range practiceTime.Iter() {
		expectedDay := testData.ExpectedPracticeTime[i]
		dayObj := dayData.Object()

		dayObj.ContainsKey("date")
		dayObj.HasValue("practiceTimeMs", expectedDay.PracticeTimeMs)
		dayObj.HasValue("quizCount", expectedDay.QuizCount)

		// Verify practiceTimeMinutes = practiceTimeMs / 60000 (with 1 decimal precision)
		expectedMinutes := float64(expectedDay.PracticeTimeMs) / 60000.0
		dayObj.ContainsKey("practiceTimeMinutes").Value("practiceTimeMinutes").Number().InDelta(expectedMinutes, 0.1)
	}
}

// TestPracticeTimeEndpoint_ResponseTimeFiltering tests response time outlier filtering
func TestPracticeTimeEndpoint_ResponseTimeFiltering(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Setup test data with response time outliers using API calls
	testData := setupPracticeTimeFilteringTestData(t, server, token)

	response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/practice-time", testData.WordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	practiceTime := json.Value("practiceTime").Array()

	// Should have filtered practice time data
	assert.True(t, practiceTime.Length().Raw() > 0, "Should return filtered practice time data")

	// Verify that outliers were filtered
	dayData := practiceTime.Value(0).Object()

	// Our test data includes: 100ms (excluded), 1500ms (included), 45000ms (excluded), 2000ms (included), 300ms (included)
	// Expected total: 1500 + 2000 + 300 = 3800ms
	expectedFilteredTime := 3800
	dayData.Value("practiceTimeMs").Number().IsEqual(expectedFilteredTime)

	// All 5 quiz attempts should be counted, even if some response times are excluded from time calculation
	dayData.Value("quizCount").Number().IsEqual(5)

	// Expected minutes: 3800 / 60000 = 0.063... rounded to 0.1
	expectedMinutes := 0.1
	dayData.Value("practiceTimeMinutes").Number().IsEqual(expectedMinutes)
}

// TestPracticeTimeEndpoint_CustomDateRange tests custom date range parameters
func TestPracticeTimeEndpoint_CustomDateRange(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Setup test data using API calls
	testData := setupPracticeTimeTestData(t, server, token)

	// Test with custom 30-day range (maximum allowed)
	response30Days := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/practice-time", testData.WordlistID)).
		WithQuery("days", "30").
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json30Days := response30Days.JSON().Object()
	json30Days.Value("days").IsEqual(30) // Should accept valid 30-day range

	// Test with custom 1-day range
	response1Day := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/practice-time", testData.WordlistID)).
		WithQuery("days", "1").
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json1Day := response1Day.JSON().Object()
	json1Day.Value("days").IsEqual(1)
	practiceTime1Day := json1Day.Value("practiceTime").Array()

	// Should return at most 1 day
	assert.True(t, practiceTime1Day.Length().Raw() <= 1, "1-day query should return at most 1 day")
}

// TestPracticeTimeEndpoint_EmptyData tests practice time with no quiz data
func TestPracticeTimeEndpoint_EmptyData(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Create wordlist with no practice time data via API
	emptyWordlistID := createTestWordlist(t, server, token, "Empty Practice Time", "No practice data")

	// Clear any potential Redis cache for this wordlist
	flushRedisCache(t)

	response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/practice-time", emptyWordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	json.Value("practiceTime").Array().IsEmpty()
}

// TestPracticeTimeEndpoint_ParameterValidation tests query parameter validation
func TestPracticeTimeEndpoint_ParameterValidation(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Create basic wordlist for testing via API
	wordlistID := createTestWordlist(t, server, token, "Practice Time Parameter Test", "Test parameter validation")

	// Clear any potential Redis cache for this wordlist
	flushRedisCache(t)

	tests := []struct {
		name         string
		daysParam    string
		expectedDays int
		description  string
	}{
		{"InvalidDays_String", "invalid", 7, "Invalid string should default to 7"},
		{"InvalidDays_Negative", "-5", 7, "Negative days should default to 7"},
		{"InvalidDays_Zero", "0", 7, "Zero days should default to 7"},
		{"InvalidDays_TooLarge", "50", 7, "Days > 30 should default to 7"},
		{"ValidDays_Small", "1", 1, "Valid small number should be accepted"},
		{"ValidDays_Medium", "15", 15, "Valid medium number should be accepted"},
		{"ValidDays_Maximum", "30", 30, "Valid maximum should be accepted"},
	}

	for _, test := range tests {
		t.Run(test.name, func(_ *testing.T) {
			response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/practice-time", wordlistID)).
				WithQuery("days", test.daysParam).
				WithHeader("Authorization", token).
				Expect().
				Status(http.StatusOK)

			json := response.JSON().Object()
			json.Value("days").Number().IsEqual(test.expectedDays)
		})
	}
}

// TestPracticeTimeEndpoint_ResultOrdering tests that results are ordered by date DESC
func TestPracticeTimeEndpoint_ResultOrdering(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Setup test data with multiple dates using API calls
	testData := setupMultipleDatesTestData(t, server, token)

	response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/practice-time", testData.WordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	practiceTime := json.Value("practiceTime").Array()

	// Verify ordering by date DESC (most recent first)
	var previousDate *time.Time
	for _, dayData := range practiceTime.Iter() {
		dateStr := dayData.Object().Value("date").String().Raw()
		currentDate, err := time.Parse("2006-01-02", dateStr[:10]) // Extract date part
		require.NoError(t, err)

		if previousDate != nil {
			assert.True(t, currentDate.Before(*previousDate) || currentDate.Equal(*previousDate),
				"Dates should be in descending order (most recent first)")
		}
		previousDate = &currentDate
	}
}

// PracticeTimeTestData contains expected practice time data
type PracticeTimeTestData struct {
	UserID               int64
	WordlistID           int64
	ExpectedPracticeTime []ExpectedPracticeTime
}

type ExpectedPracticeTime struct {
	Date           string
	PracticeTimeMs int
	QuizCount      int
}

// setupPracticeTimeTestData creates test data for practice time analytics using API calls
func setupPracticeTimeTestData(t *testing.T, server *setup.TestServer, token string) *PracticeTimeTestData {
	ctx := context.Background()

	userID := getUserID(ctx, t, server.DB)

	// Create wordlist via API
	wordlistID := createTestWordlist(t, server, token, "Practice Time Test", "Test wordlist for practice time")

	// Clear any potential Redis cache for this wordlist
	flushRedisCache(t)

	// Create basic word structure via API
	wordID, definitionID, leitnerTrackingID := createBasicWordStructure(t, server, token, wordlistID, "practiceword")

	// Create quiz results with various response times - all will be aggregated into today's date
	// since the API creates quiz results with current timestamp
	responseTimesMs := []int{
		1500, 2000, 1800, 2200, 1600, // Total: 9100ms, Count: 5
		1200, 1800, 2500, // Total: 5500ms, Count: 3
		2200, 1400, 1700, 1900, 2000, 1600, // Total: 10800ms, Count: 6
	}

	totalMs := 0
	for _, responseTime := range responseTimesMs {
		totalMs += responseTime
		// Submit quiz result via API - all will be created with today's timestamp
		saveQuizResult(t, server, token, wordlistID, wordID, definitionID, leitnerTrackingID,
			"guess_meaning", true, responseTime)
	}

	// Since all quizzes are created today, we expect one result for today's date
	expectedPracticeTime := []ExpectedPracticeTime{
		{
			Date:           time.Now().UTC().Format("2006-01-02"),
			PracticeTimeMs: totalMs,              // Total: 25400ms (9100+5500+10800)
			QuizCount:      len(responseTimesMs), // Total: 14 quizzes
		},
	}

	return &PracticeTimeTestData{
		UserID:               userID,
		WordlistID:           wordlistID,
		ExpectedPracticeTime: expectedPracticeTime,
	}
}

// setupPracticeTimeFilteringTestData creates test data to verify response time filtering using API calls
func setupPracticeTimeFilteringTestData(t *testing.T, server *setup.TestServer, token string) *PracticeTimeTestData {
	ctx := context.Background()

	userID := getUserID(ctx, t, server.DB)

	// Create wordlist via API
	wordlistID := createTestWordlist(t, server, token, "Response Time Filtering Test", "Test response time filtering")

	// Clear any potential Redis cache for this wordlist
	flushRedisCache(t)

	// Create word structure via API
	wordID, definitionID, leitnerTrackingID := createBasicWordStructure(t, server, token, wordlistID, "filtertest")

	// Insert quiz attempts with response times including outliers via API
	responseTimeData := []struct {
		responseTimeMs int
		shouldInclude  bool // Whether this should be included in practice time calculation
	}{
		{100, false},   // Too fast (< 200ms) - should be excluded
		{1500, true},   // Valid response time
		{45000, false}, // Too slow (> 30000ms) - should be excluded
		{2000, true},   // Valid response time
		{300, true},    // Valid response time (just above minimum)
	}

	for _, data := range responseTimeData {
		// Submit quiz result via API
		saveQuizResult(t, server, token, wordlistID, wordID, definitionID, leitnerTrackingID,
			"guess_meaning", true, data.responseTimeMs)
	}

	return &PracticeTimeTestData{
		UserID:     userID,
		WordlistID: wordlistID,
	}
}

// setupMultipleDatesTestData creates test data across multiple dates to test ordering using API calls
func setupMultipleDatesTestData(t *testing.T, server *setup.TestServer, token string) *PracticeTimeTestData {
	ctx := context.Background()

	userID := getUserID(ctx, t, server.DB)

	// Create wordlist via API
	wordlistID := createTestWordlist(t, server, token, "Multiple Dates Test", "Test multiple dates ordering")

	// Clear any potential Redis cache for this wordlist
	flushRedisCache(t)

	// Create word structure via API
	wordID, definitionID, leitnerTrackingID := createBasicWordStructure(t, server, token, wordlistID, "ordertest")

	// Create quiz data for today via API (API uses current timestamp)
	todayAttempts := 3
	for j := 0; j < todayAttempts; j++ {
		responseTime := 1500 + (j * 200) // 1500, 1700, 1900 ms

		// Submit quiz result via API
		saveQuizResult(t, server, token, wordlistID, wordID, definitionID, leitnerTrackingID,
			"guess_meaning", true, responseTime)
	}

	// Create historical quiz_performance data directly in database for past dates
	// This is necessary because the API always uses current timestamp
	for i := 1; i < 5; i++ { // Skip i=0 (today, already created via API)
		practiceDate := time.Now().UTC().AddDate(0, 0, -i).Format("2006-01-02")
		attemptsPerDay := 2 + (i % 2) // 2 or 3 attempts per day

		for j := 0; j < attemptsPerDay; j++ {
			responseTime := 1500 + (j * 200) // Vary response times slightly

			// Insert historical quiz_performance data directly
			_, err := server.DB.Exec(ctx,
				`INSERT INTO quiz_performance (
					user_id, wordlist_id, word_id, definition_id, leitner_system_tracking_id,
					quiz_type, box_id, is_correct, response_time_ms, created_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::date + INTERVAL '12 hours')`,
				userID, wordlistID, wordID, definitionID, leitnerTrackingID,
				"guess_meaning", 1, true, responseTime, practiceDate,
			)
			require.NoError(t, err)
		}
	}

	return &PracticeTimeTestData{
		UserID:     userID,
		WordlistID: wordlistID,
	}
}
