package analytics

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
)

// TestQuizPerformanceEndpoint_MetricCalculations tests quiz performance analytics calculations
func TestQuizPerformanceEndpoint_MetricCalculations(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Setup test data with known quiz performance metrics using API calls
	testData := setupQuizPerformanceTestData(t, server, token)

	// Test the quiz performance endpoint
	response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/quiz-type-performance", testData.WordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	json.HasValue("wordlistId", testData.WordlistID)
	json.ContainsKey("quizPerformance")

	performance := json.Value("quizPerformance").Array()
	assert.Equal(t, len(testData.ExpectedPerformance), int(performance.Length().Raw()),
		"Should return performance for all quiz types in test data")

	// Convert to map for easy verification
	performanceMap := make(map[string]map[string]interface{})
	for _, perf := range performance.Iter() {
		perfObj := perf.Object()
		quizType := perfObj.Value("quizType").String().Raw()
		totalAttempts := perfObj.Value("totalAttempts").Number().Raw()
		successRate := perfObj.Value("successRate").Number().Raw()
		avgResponseMs := perfObj.Value("avgResponseMs").Number().Raw()

		performanceMap[quizType] = map[string]interface{}{
			"totalAttempts": totalAttempts,
			"successRate":   successRate,
			"avgResponseMs": avgResponseMs,
		}
	}

	// Verify each expected quiz type performance
	for _, expected := range testData.ExpectedPerformance {
		actual, exists := performanceMap[expected.QuizType]
		assert.True(t, exists, fmt.Sprintf("Quiz type %s should be present", expected.QuizType))

		assert.Equal(t, float64(expected.TotalAttempts), actual["totalAttempts"],
			fmt.Sprintf("Total attempts for %s", expected.QuizType))
		assert.InDelta(t, expected.SuccessRate, actual["successRate"], 0.1,
			fmt.Sprintf("Success rate for %s", expected.QuizType))
		assert.InDelta(t, expected.AvgResponseMs, actual["avgResponseMs"], 1.0,
			fmt.Sprintf("Average response time for %s", expected.QuizType))
	}
}

// TestQuizPerformanceEndpoint_ResponseTimeFiltering tests response time outlier filtering
func TestQuizPerformanceEndpoint_ResponseTimeFiltering(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Setup test data with response time outliers using API calls
	testData := setupResponseTimeFilteringTestData(t, server, token)

	response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/quiz-type-performance", testData.WordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	performance := json.Value("quizPerformance").Array()

	// Find the test quiz type in results
	var foundTestQuizType bool
	for _, perf := range performance.Iter() {
		perfObj := perf.Object()
		if perfObj.Value("quizType").String().Raw() == "guess_meaning" {
			foundTestQuizType = true
			// Should only include valid response times (200ms-30000ms)
			// Our test data: 100ms (excluded), 1500ms (included), 45000ms (excluded), 2000ms (included)
			// Expected average: (1500 + 2000) / 2 = 1750ms
			expectedAvg := 1750.0
			actualAvg := perfObj.Value("avgResponseMs").Number().Raw()
			assert.InDelta(t, expectedAvg, actualAvg, 1.0,
				"Should filter out response times outside 200ms-30000ms range")
			break
		}
	}
	assert.True(t, foundTestQuizType, "Should find test quiz type in results")
}

// TestQuizPerformanceEndpoint_EmptyData tests quiz performance with no quiz data
func TestQuizPerformanceEndpoint_EmptyData(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Create wordlist with no quiz performance data via API
	emptyWordlistID := createTestWordlist(t, server, token, "Empty Quiz Performance", "No quiz data")

	flushRedisCache(t)

	response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/quiz-type-performance", emptyWordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	json.Value("quizPerformance").Array().IsEmpty()
}

// TestQuizPerformanceEndpoint_ResultOrdering tests that results are ordered by success rate DESC
func TestQuizPerformanceEndpoint_ResultOrdering(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Setup test data with varied success rates using API calls
	testData := setupOrderingTestData(t, server, token)

	response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/quiz-type-performance", testData.WordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	performance := json.Value("quizPerformance").Array()

	// Verify ordering by success rate DESC
	var previousSuccessRate = 100.0
	for _, perf := range performance.Iter() {
		currentSuccessRate := perf.Object().Value("successRate").Number().Raw()
		assert.True(t, currentSuccessRate <= previousSuccessRate,
			"Quiz types should be ordered by success rate descending")
		previousSuccessRate = currentSuccessRate
	}
}

// QuizPerformanceTestData contains expected quiz performance data
type QuizPerformanceTestData struct {
	UserID              int64
	WordlistID          int64
	ExpectedPerformance []ExpectedQuizPerformance
}

type ExpectedQuizPerformance struct {
	QuizType      string
	TotalAttempts int
	SuccessRate   float64
	AvgResponseMs float64
}

// setupQuizPerformanceTestData creates test data for quiz performance analytics using API calls
func setupQuizPerformanceTestData(t *testing.T, server *setup.TestServer, token string) *QuizPerformanceTestData {
	ctx := context.Background()

	userID := getUserID(ctx, t, server.DB)

	// Create wordlist via API
	wordlistID := createTestWordlist(t, server, token, "Quiz Performance Test", "Test wordlist for quiz performance")

	// Clear any potential Redis cache for this wordlist
	flushRedisCache(t)

	// Create basic word structure via API
	wordID, definitionID, leitnerTrackingID := createBasicWordStructure(t, server, token, wordlistID, "testword")

	// Define quiz performance test scenarios
	quizScenarios := []struct {
		quizType string
		attempts []QuizAttempt
	}{
		{
			quizType: "guess_meaning",
			attempts: []QuizAttempt{
				{true, 1500}, {false, 2200}, {true, 1800}, // 2/3 correct, avg: 1833ms
			},
		},
		{
			quizType: "word_from_meaning",
			attempts: []QuizAttempt{
				{true, 2000}, {true, 1600}, // 2/2 correct, avg: 1800ms
			},
		},
		{
			quizType: "word_from_image",
			attempts: []QuizAttempt{
				{false, 3000}, {true, 2500}, {false, 2800}, {true, 2200}, // 2/4 correct, avg: 2625ms
			},
		},
	}

	expectedPerformance := make([]ExpectedQuizPerformance, 0, len(quizScenarios))

	for _, scenario := range quizScenarios {
		totalAttempts := len(scenario.attempts)
		correctAttempts := 0
		totalResponseTime := 0

		for _, attempt := range scenario.attempts {
			if attempt.IsCorrect {
				correctAttempts++
			}
			totalResponseTime += attempt.ResponseTimeMs

			// Submit quiz result via API instead of direct database insert
			saveQuizResult(t, server, token, wordlistID, wordID, definitionID, leitnerTrackingID,
				scenario.quizType, attempt.IsCorrect, attempt.ResponseTimeMs)
		}

		successRate := float64(correctAttempts) / float64(totalAttempts) * 100
		avgResponseMs := float64(totalResponseTime) / float64(totalAttempts)

		expectedPerformance = append(expectedPerformance, ExpectedQuizPerformance{
			QuizType:      scenario.quizType,
			TotalAttempts: totalAttempts,
			SuccessRate:   successRate,
			AvgResponseMs: avgResponseMs,
		})
	}

	return &QuizPerformanceTestData{
		UserID:              userID,
		WordlistID:          wordlistID,
		ExpectedPerformance: expectedPerformance,
	}
}

// setupResponseTimeFilteringTestData creates test data to verify response time filtering using API calls
func setupResponseTimeFilteringTestData(t *testing.T, server *setup.TestServer, token string) *QuizPerformanceTestData {
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
		isCorrect      bool
		responseTimeMs int
		shouldInclude  bool // Whether this should be included in average calculation
	}{
		{true, 100, false},    // Too fast (< 200ms) - should be excluded
		{true, 1500, true},    // Valid response time
		{false, 45000, false}, // Too slow (> 30000ms) - should be excluded
		{true, 2000, true},    // Valid response time
	}

	for _, data := range responseTimeData {
		// Submit quiz result via API
		saveQuizResult(t, server, token, wordlistID, wordID, definitionID, leitnerTrackingID,
			"guess_meaning", data.isCorrect, data.responseTimeMs)
	}

	return &QuizPerformanceTestData{
		UserID:     userID,
		WordlistID: wordlistID,
	}
}

// setupOrderingTestData creates test data to verify result ordering using API calls
func setupOrderingTestData(t *testing.T, server *setup.TestServer, token string) *QuizPerformanceTestData {
	ctx := context.Background()

	userID := getUserID(ctx, t, server.DB)

	// Create wordlist via API
	wordlistID := createTestWordlist(t, server, token, "Ordering Test", "Test result ordering")

	// Clear any potential Redis cache for this wordlist
	flushRedisCache(t)

	// Create word structure via API
	wordID, definitionID, leitnerTrackingID := createBasicWordStructure(t, server, token, wordlistID, "ordertest")

	// Create quiz types with different success rates to test ordering
	quizTypesWithSuccessRates := []struct {
		quizType    string
		successRate float64
	}{
		{"perfect_quiz", 1.0}, // 100% success rate
		{"good_quiz", 0.8},    // 80% success rate
		{"average_quiz", 0.5}, // 50% success rate
		{"poor_quiz", 0.2},    // 20% success rate
	}

	for _, quizData := range quizTypesWithSuccessRates {
		attemptsPerType := 5
		correctAttempts := int(float64(attemptsPerType) * quizData.successRate)

		for i := 0; i < attemptsPerType; i++ {
			isCorrect := i < correctAttempts
			// Submit quiz result via API
			saveQuizResult(t, server, token, wordlistID, wordID, definitionID, leitnerTrackingID,
				quizData.quizType, isCorrect, 1500)
		}
	}

	return &QuizPerformanceTestData{
		UserID:     userID,
		WordlistID: wordlistID,
	}
}

// QuizAttempt represents a single quiz attempt
type QuizAttempt struct {
	IsCorrect      bool
	ResponseTimeMs int
}
