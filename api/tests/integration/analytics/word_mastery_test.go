package analytics

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWordMasteryEndpoint_CalculationAccuracy tests word mastery analytics calculations
func TestWordMasteryEndpoint_CalculationAccuracy(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Setup test data using API calls
	testData := setupWordMasteryTestData(t, server, token)

	// Test the word mastery endpoint
	response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/word-mastery", testData.WordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	json.HasValue("wordlist_id", testData.WordlistID)
	json.ContainsKey("stats")

	stats := json.Value("stats").Array()
	assert.Equal(t, len(testData.ExpectedWords), int(stats.Length().Raw()),
		"Should return mastery stats for all words")

	// Verify each word's mastery data matches expected calculations
	for i, stat := range stats.Iter() {
		statObj := stat.Object()
		expectedWord := testData.ExpectedWords[i]

		statObj.HasValue("wordId", expectedWord.WordID)
		statObj.HasValue("word", expectedWord.WordName)
		statObj.Value("masteryLevel").Number().InDelta(expectedWord.MasteryLevel, 0.01)
		statObj.Value("accuracy").Number().InDelta(expectedWord.Accuracy, 0.01)
		statObj.HasValue("streakCount", expectedWord.StreakCount)
		statObj.HasValue("highestBox", expectedWord.HighestBox)
		statObj.ContainsKey("lastSeenAt")
	}
}

// TestWordMasteryEndpoint_EmptyWordlist tests word mastery with no words
func TestWordMasteryEndpoint_EmptyWordlist(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Create empty wordlist via API with unique name to avoid analytics cache conflicts
	emptyWordlistID := createTestWordlist(t, server, token, "Empty Wordlist Analytics", "Test empty wordlist analytics")

	// Clear any potential Redis cache for this specific wordlist before testing
	flushRedisCache(t)

	response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/word-mastery", emptyWordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	json.Value("stats").Array().IsEmpty()
}

// TestWordMasteryEndpoint_MultipleBoxLevels tests words at different Leitner box levels
func TestWordMasteryEndpoint_MultipleBoxLevels(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Setup test data with words in different boxes using API calls
	testData := setupMultiBoxTestData(t, server, token)

	response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/word-mastery", testData.WordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	stats := json.Value("stats").Array()

	// Verify words are ordered by mastery level DESC (repository default order)
	var previousMastery = 1.0
	for _, stat := range stats.Iter() {
		currentMastery := stat.Object().Value("masteryLevel").Number().Raw()
		assert.True(t, currentMastery <= previousMastery,
			"Words should be ordered by mastery level descending")
		previousMastery = currentMastery
	}

	// Verify that highest box values match our test setup
	expectedBoxes := []int64{7, 5, 3, 2, 1} // From highest mastery to lowest
	for i, stat := range stats.Iter() {
		statObj := stat.Object()
		statObj.Value("highestBox").Number().IsEqual(expectedBoxes[i])
	}
}

// TestWordMasteryEndpoint_ErrorCases tests error handling
func TestWordMasteryEndpoint_ErrorCases(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	t.Run("NonexistentWordlist", func(_ *testing.T) {
		server.Expect.GET("/analytics/wordlists/999999/word-mastery").
			WithHeader("Authorization", token).
			Expect().
			Status(http.StatusNotFound)
	})

	t.Run("InvalidWordlistID", func(_ *testing.T) {
		server.Expect.GET("/analytics/wordlists/invalid/word-mastery").
			WithHeader("Authorization", token).
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("UnauthorizedAccess", func(_ *testing.T) {
		server.Expect.GET("/analytics/wordlists/1/word-mastery").
			Expect().
			Status(http.StatusUnauthorized)
	})
}

// WordMasteryTestData contains expected word mastery data
type WordMasteryTestData struct {
	UserID        int64
	WordlistID    int64
	ExpectedWords []ExpectedWordMastery
}

type ExpectedWordMastery struct {
	WordID       int64
	WordName     string
	MasteryLevel float64
	Accuracy     float64
	StreakCount  int
	HighestBox   int64
}

// setupWordMasteryTestData creates test data for word mastery analytics using API calls
func setupWordMasteryTestData(t *testing.T, server *setup.TestServer, token string) *WordMasteryTestData {
	ctx := context.Background()

	// Get user ID
	userID := getUserID(ctx, t, server.DB)

	// Create test wordlist via API with unique name to avoid conflicts
	wordlistID := createTestWordlist(t, server, token, "Word Mastery Calculation Test", "Test wordlist for word mastery calculations")

	// Clear any potential Redis cache for this wordlist
	flushRedisCache(t)

	// Test scenarios with known outcomes
	testWords := []struct {
		name            string
		masteryLevel    float64
		totalAttempts   int
		correctAttempts int
		streakCount     int
		maxStreak       int
		boxLevel        int64
	}{
		{"excellent", 0.95, 20, 19, 8, 10, 7}, // Near perfect
		{"good", 0.75, 16, 12, 4, 6, 5},       // Good performance
		{"average", 0.50, 10, 5, 2, 3, 3},     // Average performance
		{"struggling", 0.25, 12, 3, 0, 2, 2},  // Poor performance
	}

	expectedWords := make([]ExpectedWordMastery, 0, len(testWords))

	for _, wordData := range testWords {
		// Create word via API
		wordID, definitionID, _ := createBasicWordStructure(t, server, token, wordlistID, wordData.name)

		// Update Leitner tracking box level
		_, err := server.DB.Exec(ctx,
			`UPDATE leitner_system_tracking SET box_id = $1 
			 WHERE word_id = $2 AND definition_id = $3`,
			wordData.boxLevel, wordID, definitionID)
		require.NoError(t, err)

		// Create word mastery entry using the proper repository call
		createWordMasteryData(t, server, userID, wordID, wordData.boxLevel, wordData.correctAttempts > 0,
			wordData.masteryLevel, wordData.totalAttempts, wordData.correctAttempts,
			wordData.streakCount, wordData.maxStreak)

		// Calculate expected accuracy
		expectedAccuracy := float64(wordData.correctAttempts) / float64(wordData.totalAttempts)

		expectedWords = append(expectedWords, ExpectedWordMastery{
			WordID:       wordID,
			WordName:     wordData.name,
			MasteryLevel: wordData.masteryLevel,
			Accuracy:     expectedAccuracy,
			StreakCount:  wordData.streakCount,
			HighestBox:   wordData.boxLevel,
		})
	}

	return &WordMasteryTestData{
		UserID:        userID,
		WordlistID:    wordlistID,
		ExpectedWords: expectedWords,
	}
}

// setupMultiBoxTestData creates test data with words distributed across different Leitner boxes using API calls
func setupMultiBoxTestData(t *testing.T, server *setup.TestServer, token string) *WordMasteryTestData {
	ctx := context.Background()

	userID := getUserID(ctx, t, server.DB)

	// Create wordlist via API with unique name
	wordlistID := createTestWordlist(t, server, token, "Multi-Box Distribution Test", "Test wordlist for multi-box distribution")

	// Clear any potential Redis cache for this wordlist
	flushRedisCache(t)

	// Create words with decreasing mastery levels (to test ordering)
	boxLevels := []struct {
		boxID        int64
		masteryLevel float64
		wordName     string
	}{
		{7, 0.95, "expert"},       // Highest mastery
		{5, 0.80, "advanced"},     // High mastery
		{3, 0.60, "intermediate"}, // Medium mastery
		{2, 0.35, "beginner"},     // Low mastery
		{1, 0.15, "struggling"},   // Lowest mastery
	}

	expectedWords := make([]ExpectedWordMastery, 0, len(boxLevels))

	for _, boxData := range boxLevels {
		// Create word via API
		wordID, definitionID, _ := createBasicWordStructure(t, server, token, wordlistID, boxData.wordName)

		// Update Leitner tracking to specific box
		_, err := server.DB.Exec(ctx,
			`UPDATE leitner_system_tracking SET box_id = $1 
			 WHERE word_id = $2 AND definition_id = $3`,
			boxData.boxID, wordID, definitionID)
		require.NoError(t, err)

		// Create mastery data that matches the box level
		totalAttempts := int(boxData.boxID * 3) // More attempts for higher boxes
		correctAttempts := int(float64(totalAttempts) * boxData.masteryLevel)

		// Create word mastery entry using the proper repository call
		createWordMasteryData(t, server, userID, wordID, boxData.boxID, correctAttempts > 0,
			boxData.masteryLevel, totalAttempts, correctAttempts,
			int(boxData.masteryLevel*5), int(boxData.masteryLevel*8))

		expectedWords = append(expectedWords, ExpectedWordMastery{
			WordID:       wordID,
			WordName:     boxData.wordName,
			MasteryLevel: boxData.masteryLevel,
			Accuracy:     float64(correctAttempts) / float64(totalAttempts),
			StreakCount:  int(boxData.masteryLevel * 5),
			HighestBox:   boxData.boxID,
		})
	}

	return &WordMasteryTestData{
		UserID:        userID,
		WordlistID:    wordlistID,
		ExpectedWords: expectedWords,
	}
}
