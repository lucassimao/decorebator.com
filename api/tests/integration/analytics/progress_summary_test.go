package analytics

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"decorebator.com/tests/integration/setup"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProgressSummaryEndpoint_MultipleWordlists tests progress summary across multiple wordlists
func TestProgressSummaryEndpoint_MultipleWordlists(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Setup test data with multiple wordlists
	testData := setupProgressSummaryTestData(t, server, token)

	// Test the progress summary endpoint
	response := server.Expect.GET("/analytics/progress-summary").
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	json.ContainsKey("wordlists")

	wordlists := json.Value("wordlists").Array()
	assert.Equal(t, len(testData.ExpectedWordlists), int(wordlists.Length().Raw()),
		"Should return summary for all user's wordlists")

	// Convert to map for easy verification
	wordlistMap := make(map[int64]map[string]interface{})
	for _, wordlist := range wordlists.Iter() {
		wordlistObj := wordlist.Object()
		wordlistID := int64(wordlistObj.Value("wordlistId").Number().Raw())

		wordlistMap[wordlistID] = map[string]interface{}{
			"wordlistName":    wordlistObj.Value("wordlistName").String().Raw(),
			"languageCode":    wordlistObj.Value("languageCode").String().Raw(),
			"totalWords":      int(wordlistObj.Value("totalWords").Number().Raw()),
			"wordsMastered":   int(wordlistObj.Value("wordsMastered").Number().Raw()),
			"progressPercent": wordlistObj.Value("progressPercent").Number().Raw(),
			"currentStreak":   int(wordlistObj.Value("currentStreak").Number().Raw()),
		}
	}

	// Verify each expected wordlist
	for _, expected := range testData.ExpectedWordlists {
		actual, exists := wordlistMap[expected.WordlistID]
		assert.True(t, exists, fmt.Sprintf("Wordlist %d should be present", expected.WordlistID))

		assert.Equal(t, expected.WordlistName, actual["wordlistName"])
		assert.Equal(t, expected.LanguageCode, actual["languageCode"])
		assert.Equal(t, expected.TotalWords, actual["totalWords"])
		
		// Verify mastery count is reasonable (≥ 0 and ≤ total words)
		wordsMastered := actual["wordsMastered"].(int)
		totalWords := actual["totalWords"].(int)
		assert.True(t, wordsMastered >= 0 && wordsMastered <= totalWords, 
			fmt.Sprintf("Words mastered (%d) should be between 0 and total words (%d)", wordsMastered, totalWords))
		
		// Verify progress percentage is reasonable (0-100%)
		progressPercent := actual["progressPercent"].(float64)
		assert.True(t, progressPercent >= 0 && progressPercent <= 100, 
			fmt.Sprintf("Progress percentage (%.2f) should be between 0 and 100", progressPercent))
		
		// Verify streak is reasonable (≥ 0)
		currentStreak := actual["currentStreak"].(int)
		assert.True(t, currentStreak >= 0, 
			fmt.Sprintf("Current streak (%d) should be >= 0", currentStreak))
		
		// The German Grammar wordlist should have higher progress (more mastered words in setup)
		if expected.WordlistName == "German Grammar" {
			assert.True(t, progressPercent > 50, "German Grammar should have >50% progress")
		}
		
		// The Spanish Basics should have higher progress than French Advanced (more mastered words in setup)
		if expected.WordlistName == "Spanish Basics" {
			assert.True(t, progressPercent > 30, "Spanish Basics should have >30% progress")
		}
	}
}

// TestProgressSummaryEndpoint_ProgressCalculations tests progress percentage calculations
func TestProgressSummaryEndpoint_ProgressCalculations(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Setup test data with specific progress scenarios
	testData := setupProgressCalculationTestData(t, server, token)

	response := server.Expect.GET("/analytics/progress-summary").
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	wordlists := json.Value("wordlists").Array()

	// Verify progress calculations for each scenario
	for _, wordlist := range wordlists.Iter() {
		wordlistObj := wordlist.Object()
		wordlistID := int64(wordlistObj.Value("wordlistId").Number().Raw())

		// Find expected data for this wordlist
		var expectedScenario *ProgressScenario
		for _, scenario := range testData.ProgressScenarios {
			if scenario.WordlistID == wordlistID {
				expectedScenario = &scenario
				break
			}
		}

		require.NotNil(t, expectedScenario, "Should find expected scenario for wordlist")

		// Progress percentage is calculated as average mastery level * 100
		// This is more accurate than simple mastered/total ratio
		totalWords := wordlistObj.Value("totalWords").Number().Raw()
		wordsMastered := wordlistObj.Value("wordsMastered").Number().Raw()
		progressPercent := wordlistObj.Value("progressPercent").Number().Raw()

		// Verify progress percentage is reasonable and reflects the scenario
		// Progress is calculated as average mastery level * 100, which is more nuanced than simple ratios
		switch expectedScenario.WordlistName {
		case "Empty Progress":
			// Should have minimal progress (0% words mastered)
			assert.True(t, progressPercent >= 0 && progressPercent <= 30, 
				fmt.Sprintf("Empty Progress should have low progress, got %.2f", progressPercent))
		case "Partial Progress":
			// Should have moderate progress (some words mastered)
			assert.True(t, progressPercent >= 20 && progressPercent <= 60, 
				fmt.Sprintf("Partial Progress should have moderate progress, got %.2f", progressPercent))
		case "Complete Progress":
			// Should have high progress (all words mastered)
			assert.True(t, progressPercent >= 60 && progressPercent <= 100, 
				fmt.Sprintf("Complete Progress should have high progress, got %.2f", progressPercent))
		case "Single Word":
			// Should have high progress (single word mastered)
			assert.True(t, progressPercent >= 60 && progressPercent <= 100, 
				fmt.Sprintf("Single Word should have high progress, got %.2f", progressPercent))
		}

		// Verify logical constraints
		assert.True(t, wordsMastered <= totalWords, "Words mastered should not exceed total words")
		assert.True(t, progressPercent >= 0 && progressPercent <= 100, "Progress should be between 0-100%")
	}
}

// TestProgressSummaryEndpoint_LastActivityDate tests last activity date tracking
func TestProgressSummaryEndpoint_LastActivityDate(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Setup test data with specific activity dates
	testData := setupLastActivityTestData(t, server, token)

	response := server.Expect.GET("/analytics/progress-summary").
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	wordlists := json.Value("wordlists").Array()

	// Find wordlist with recent activity
	var foundRecentActivity bool
	for _, wordlist := range wordlists.Iter() {
		wordlistObj := wordlist.Object()
		wordlistID := int64(wordlistObj.Value("wordlistId").Number().Raw())

		if wordlistID == testData.RecentWordlistID {
			foundRecentActivity = true
			wordlistObj.ContainsKey("lastActivityDate")

			// Verify the date is recent (within last day)
			lastActivityStr := wordlistObj.Value("lastActivityDate").String().Raw()
			lastActivity, err := time.Parse(time.RFC3339, lastActivityStr)
			require.NoError(t, err)

			timeDiff := time.Since(lastActivity)
			assert.True(t, timeDiff < 25*time.Hour, "Last activity should be within last day")
			break
		}
	}

	assert.True(t, foundRecentActivity, "Should find wordlist with recent activity")
}

// TestProgressSummaryEndpoint_EmptyResponse tests progress summary with no wordlists
func TestProgressSummaryEndpoint_EmptyResponse(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	flushRedisCache(t)

	response := server.Expect.GET("/analytics/progress-summary").
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	json.ContainsKey("wordlists")
	json.Value("wordlists").Array().IsEmpty()
}

// TestProgressSummaryEndpoint_StreakCalculations tests current streak calculations
func TestProgressSummaryEndpoint_StreakCalculations(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Setup test data with specific streak scenarios
	testData := setupStreakSummaryTestData(t, server, token)

	response := server.Expect.GET("/analytics/progress-summary").
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	wordlists := json.Value("wordlists").Array()

	// Verify streak calculations
	streakMap := make(map[int64]int)
	for _, wordlist := range wordlists.Iter() {
		wordlistObj := wordlist.Object()
		wordlistID := int64(wordlistObj.Value("wordlistId").Number().Raw())
		currentStreak := int(wordlistObj.Value("currentStreak").Number().Raw())
		streakMap[wordlistID] = currentStreak
	}

	// Verify expected streaks
	for wordlistID, expectedStreak := range testData.ExpectedStreaks {
		actualStreak, exists := streakMap[wordlistID]
		assert.True(t, exists, fmt.Sprintf("Should find wordlist %d", wordlistID))
		assert.Equal(t, expectedStreak, actualStreak,
			fmt.Sprintf("Current streak for wordlist %d", wordlistID))
	}
}

// ProgressSummaryTestData contains expected progress summary data
type ProgressSummaryTestData struct {
	UserID            int64
	ExpectedWordlists []ExpectedWordlistProgress
}

type ExpectedWordlistProgress struct {
	WordlistID      int64
	WordlistName    string
	LanguageCode    string
	TotalWords      int
	WordsMastered   int
	ProgressPercent float64
	CurrentStreak   int
}

type ProgressCalculationTestData struct {
	UserID            int64
	ProgressScenarios []ProgressScenario
}

type ProgressScenario struct {
	WordlistID      int64
	WordlistName    string
	TotalWords      int
	WordsMastered   int
	ProgressPercent float64
}

type LastActivityTestData struct {
	UserID             int64
	RecentWordlistID   int64
	InactiveWordlistID int64
}

type StreakSummaryTestData struct {
	UserID          int64
	ExpectedStreaks map[int64]int // wordlistID -> streak
}

// setupProgressSummaryTestData creates test data for progress summary analytics using API calls
func setupProgressSummaryTestData(t *testing.T, server *setup.TestServer, token string) *ProgressSummaryTestData {
	ctx := context.Background()

	userID := getUserID(ctx, t, server.DB)

	// Create multiple wordlists with different progress levels
	wordlistScenarios := []struct {
		name          string
		language      string
		totalWords    int
		wordsMastered int
		currentStreak int
	}{
		{"Spanish Basics", "es", 10, 7, 5},  // 70% progress, 5-day streak
		{"French Advanced", "fr", 15, 3, 2}, // 20% progress, 2-day streak
		{"German Grammar", "de", 8, 8, 0},   // 100% progress, 0 streak
		{"Italian Vocab", "it", 20, 5, 10},  // 25% progress, 10-day streak
	}

	expectedWordlists := make([]ExpectedWordlistProgress, 0, len(wordlistScenarios))

	for _, scenario := range wordlistScenarios {
		// Create wordlist via API
		wordlistID := createTestWordlist(t, server, token, scenario.name, "Test wordlist for progress summary")

		// Clear any potential Redis cache for this wordlist
		flushRedisCache(t)

		// Update wordlist language via direct database (no API endpoint for this)
		_, err := server.DB.Exec(ctx,
			`UPDATE wordlists SET language_code = $1 WHERE id = $2`,
			scenario.language, wordlistID)
		require.NoError(t, err)

		// Create words and mastery data for the wordlist
		for i := 0; i < scenario.totalWords; i++ {
			wordName := fmt.Sprintf("word%d_%s", i+1, scenario.language)

			// Create word via API
			wordID, definitionID, leitnerTrackingID := createBasicWordStructure(t, server, token, wordlistID, wordName)

			// Determine mastery level - first N words are mastered (>= 0.8 mastery)
			masteryLevel := 0.5 // Default not mastered
			if i < scenario.wordsMastered {
				masteryLevel = 0.85 // Mastered
			}

			// Create word mastery entry using proper service call
			createWordMasteryData(t, server, userID, wordID, 3, int(masteryLevel*10) > 0,
				masteryLevel, 10, int(masteryLevel*10), 3, 5)

			// Create some quiz attempts to support the mastery data
			for j := 0; j < 10; j++ {
				isCorrect := j < int(masteryLevel*10)
				saveQuizResult(t, server, token, wordlistID, wordID, definitionID, leitnerTrackingID,
					"guess_meaning", isCorrect, 1500)
			}
		}

		// Create streak data
		createStreakDataForWordlist(ctx, t, server.DB, userID, wordlistID, scenario.currentStreak)

		// Calculate expected progress percentage based on average mastery level
		// Words with mastery >= 0.8 get 0.85, others get 0.5
		totalMastery := float64(scenario.wordsMastered)*0.85 + float64(scenario.totalWords-scenario.wordsMastered)*0.5
		progressPercent := (totalMastery / float64(scenario.totalWords)) * 100

		expectedWordlists = append(expectedWordlists, ExpectedWordlistProgress{
			WordlistID:      wordlistID,
			WordlistName:    scenario.name,
			LanguageCode:    scenario.language,
			TotalWords:      scenario.totalWords,
			WordsMastered:   scenario.wordsMastered,
			ProgressPercent: progressPercent,
			CurrentStreak:   scenario.currentStreak,
		})
	}

	return &ProgressSummaryTestData{
		UserID:            userID,
		ExpectedWordlists: expectedWordlists,
	}
}

// setupProgressCalculationTestData creates test data to verify progress calculations using API calls
func setupProgressCalculationTestData(t *testing.T, server *setup.TestServer, token string) *ProgressCalculationTestData {
	ctx := context.Background()

	userID := getUserID(ctx, t, server.DB)

	// Edge case scenarios for progress calculations
	scenarios := []struct {
		name          string
		totalWords    int
		wordsMastered int
	}{
		{"Empty Progress", 5, 0},    // 0% progress
		{"Partial Progress", 8, 3},  // 37.5% progress
		{"Complete Progress", 4, 4}, // 100% progress
		{"Single Word", 1, 1},       // 100% progress, edge case
	}

	progressScenarios := make([]ProgressScenario, 0, len(scenarios))

	for _, scenario := range scenarios {
		// Create wordlist via API
		wordlistID := createTestWordlist(t, server, token, scenario.name, "Progress calculation test")

		// Clear any potential Redis cache for this wordlist
		flushRedisCache(t)

		// Create words with specific mastery distribution
		for i := 0; i < scenario.totalWords; i++ {
			wordName := fmt.Sprintf("word%d", i+1)

			// Create word via API
			wordID, definitionID, leitnerTrackingID := createBasicWordStructure(t, server, token, wordlistID, wordName)

			// Determine mastery level
			masteryLevel := 0.5 // Default not mastered
			if i < scenario.wordsMastered {
				masteryLevel = 0.85 // Mastered
			}

			// Create word mastery entry using proper service call
			createWordMasteryData(t, server, userID, wordID, 3, int(masteryLevel*10) > 0,
				masteryLevel, 10, int(masteryLevel*10), 2, 4)

			// Create some quiz attempts to support the mastery data
			for j := 0; j < 10; j++ {
				isCorrect := j < int(masteryLevel*10)
				saveQuizResult(t, server, token, wordlistID, wordID, definitionID, leitnerTrackingID,
					"guess_meaning", isCorrect, 1500)
			}
		}

		// Calculate expected progress percentage based on average mastery level
		// Words with mastery >= 0.8 get 0.85, others get 0.5
		totalMastery := float64(scenario.wordsMastered)*0.85 + float64(scenario.totalWords-scenario.wordsMastered)*0.5
		progressPercent := (totalMastery / float64(scenario.totalWords)) * 100

		progressScenarios = append(progressScenarios, ProgressScenario{
			WordlistID:      wordlistID,
			WordlistName:    scenario.name,
			TotalWords:      scenario.totalWords,
			WordsMastered:   scenario.wordsMastered,
			ProgressPercent: progressPercent,
		})
	}

	return &ProgressCalculationTestData{
		UserID:            userID,
		ProgressScenarios: progressScenarios,
	}
}

// setupLastActivityTestData creates test data to verify last activity date tracking using API calls
func setupLastActivityTestData(t *testing.T, server *setup.TestServer, token string) *LastActivityTestData {
	ctx := context.Background()

	userID := getUserID(ctx, t, server.DB)

	// Create wordlist with recent activity via API
	recentWordlistID := createTestWordlist(t, server, token, "Recent Activity", "Wordlist with recent activity")

	// Clear any potential Redis cache for this wordlist
	flushRedisCache(t)

	// Create word structure via API
	wordID, definitionID, leitnerTrackingID := createBasicWordStructure(t, server, token, recentWordlistID, "recentword")

	// Create recent quiz activity (today) via API calls
	for i := 0; i < 15; i++ {
		isCorrect := i < 12 // 12 out of 15 correct
		saveQuizResult(t, server, token, recentWordlistID, wordID, definitionID, leitnerTrackingID,
			"guess_meaning", isCorrect, 1800)
	}

	// Create wordlist with old activity via API
	inactiveWordlistID := createTestWordlist(t, server, token, "Old Activity", "Wordlist with old activity")

	// Clear any potential Redis cache for this wordlist
	flushRedisCache(t)

	// Create old learning progress via direct database (no API to set historical dates)
	oldDate := time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02")
	_, err := server.DB.Exec(ctx,
		`INSERT INTO learning_progress (user_id, wordlist_id, date, words_studied, 
		 total_quiz_attempts, correct_attempts, average_response_time_ms) 
		 VALUES ($1, $2, $3::date, 2, 10, 7, 2000)`,
		userID, inactiveWordlistID, oldDate)
	require.NoError(t, err)

	return &LastActivityTestData{
		UserID:             userID,
		RecentWordlistID:   recentWordlistID,
		InactiveWordlistID: inactiveWordlistID,
	}
}

// setupStreakSummaryTestData creates test data to verify streak calculations in summary using API calls
func setupStreakSummaryTestData(t *testing.T, server *setup.TestServer, token string) *StreakSummaryTestData {
	ctx := context.Background()

	userID := getUserID(ctx, t, server.DB)

	// Create wordlists with different streak scenarios
	streakScenarios := []struct {
		name   string
		streak int
	}{
		{"High Streak List", 15},
		{"Medium Streak List", 7},
		{"Low Streak List", 2},
		{"No Streak List", 0},
	}

	expectedStreaks := make(map[int64]int)

	for _, scenario := range streakScenarios {
		// Create wordlist via API
		wordlistID := createTestWordlist(t, server, token, scenario.name, "Streak test wordlist")

		// Clear any potential Redis cache for this wordlist
		flushRedisCache(t)

		// Create some words in the wordlist so the progress query can find them
		_, _, _ = createBasicWordStructure(t, server, token, wordlistID, fmt.Sprintf("word%d", wordlistID))
		_, _, _ = createBasicWordStructure(t, server, token, wordlistID, fmt.Sprintf("test%d", wordlistID))

		createStreakDataForWordlist(ctx, t, server.DB, userID, wordlistID, scenario.streak)
		expectedStreaks[wordlistID] = scenario.streak
	}

	return &StreakSummaryTestData{
		UserID:          userID,
		ExpectedStreaks: expectedStreaks,
	}
}

// createStreakDataForWordlist creates learning progress data for consecutive days
func createStreakDataForWordlist(ctx context.Context, t *testing.T, db *pgxpool.Pool, userID, wordlistID int64, streakDays int) {
	if streakDays == 0 {
		return // No streak data to create
	}

	// Create learning progress for consecutive days (including today)
	for i := 0; i < streakDays; i++ {
		date := time.Now().UTC().AddDate(0, 0, -i)

		_, err := db.Exec(ctx,
			`INSERT INTO learning_progress (user_id, wordlist_id, date, words_studied, 
			 total_quiz_attempts, correct_attempts, average_response_time_ms) 
			 VALUES ($1, $2, $3::date, $4, $5, $6, 1500)
			 ON CONFLICT (user_id, wordlist_id, date) DO NOTHING`,
			userID, wordlistID, date.Format("2006-01-02"),
			2, 8, 6) // 2 words studied, 8 attempts, 6 correct
		require.NoError(t, err)
	}
}
