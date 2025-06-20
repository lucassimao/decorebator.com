package analytics

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCurrentBoxDistributionEndpoint_AccurateDistribution tests current box distribution calculations
func TestCurrentBoxDistributionEndpoint_AccurateDistribution(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Setup test data with known box distribution
	testData := setupBoxDistributionTestData(t, server, token)

	// Test the current box distribution endpoint
	response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/current-box-distribution", testData.WordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	json.HasValue("wordlistId", testData.WordlistID)
	json.HasValue("totalWords", testData.TotalWords)
	json.ContainsKey("distribution")

	distribution := json.Value("distribution").Object()

	// Verify each box count matches our test data
	distribution.Value("box1").Number().IsEqual(testData.ExpectedDistribution[1])
	distribution.Value("box2").Number().IsEqual(testData.ExpectedDistribution[2])
	distribution.Value("box3").Number().IsEqual(testData.ExpectedDistribution[3])
	distribution.Value("box4").Number().IsEqual(testData.ExpectedDistribution[4])
	distribution.Value("box5").Number().IsEqual(testData.ExpectedDistribution[5])
	distribution.Value("box6").Number().IsEqual(testData.ExpectedDistribution[6])
	distribution.Value("box7").Number().IsEqual(testData.ExpectedDistribution[7])
	distribution.Value("totalWords").Number().IsEqual(testData.TotalWords)

	// Verify the sum of all boxes equals total words
	calculatedTotal := 0
	for boxID := 1; boxID <= 7; boxID++ {
		calculatedTotal += testData.ExpectedDistribution[boxID]
	}
	assert.Equal(t, calculatedTotal, testData.TotalWords, "Box distribution sum should equal total words")
}

// TestCurrentBoxDistributionEndpoint_MaxBoxLogic tests that MAX(box_id) logic is used correctly
func TestCurrentBoxDistributionEndpoint_MaxBoxLogic(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Setup test data where words have multiple definitions at different box levels
	testData := setupMaxBoxLogicTestData(t, server, token)

	response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/current-box-distribution", testData.WordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	distribution := json.Value("distribution").Object()

	// Should count each word by its HIGHEST box level, not all definitions
	// We created 2 words: one with max box 5, one with max box 7
	distribution.Value("box1").Number().IsEqual(0)
	distribution.Value("box2").Number().IsEqual(0)
	distribution.Value("box3").Number().IsEqual(0)
	distribution.Value("box4").Number().IsEqual(0)
	distribution.Value("box5").Number().IsEqual(1) // One word at max box 5
	distribution.Value("box6").Number().IsEqual(0)
	distribution.Value("box7").Number().IsEqual(1) // One word at max box 7
	distribution.Value("totalWords").Number().IsEqual(2)
}

// TestBoxDistributionHistoryEndpoint_HistoricalData tests box distribution history
func TestBoxDistributionHistoryEndpoint_HistoricalData(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Setup test data with historical snapshots
	testData := setupHistoricalSnapshotTestData(t, server, token)

	// Test with default days (30)
	response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/box-distribution-history", testData.WordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	json.HasValue("wordlistId", testData.WordlistID)
	json.HasValue("days", 30)
	json.ContainsKey("distribution")

	distribution := json.Value("distribution").Array()

	// Debug: Print the actual length received
	actualLength := int(distribution.Length().Raw())
	t.Logf("Expected %d snapshots, got %d", len(testData.ExpectedSnapshots), actualLength)
	t.Logf("Response body: %v", response.Body().Raw())

	// If we got no snapshots, it might be a data setup issue
	if actualLength == 0 {
		t.Logf("No snapshots returned. UserID: %d, WordlistID: %d", testData.UserID, testData.WordlistID)
	}

	assert.Equal(t, len(testData.ExpectedSnapshots), actualLength,
		"Should return snapshots for all test dates")

	// Verify snapshot structure and ordering (should be DESC by date)
	for i, snapshot := range distribution.Iter() {
		snapshotObj := snapshot.Object()
		expectedSnapshot := testData.ExpectedSnapshots[i]

		snapshotObj.ContainsKey("date")
		snapshotObj.ContainsKey("boxes")

		// The box distribution is nested inside "boxes" field
		boxesObj := snapshotObj.Value("boxes").Object()
		boxesObj.HasValue("box1", expectedSnapshot.Box1Count)
		boxesObj.HasValue("box2", expectedSnapshot.Box2Count)
		boxesObj.HasValue("box3", expectedSnapshot.Box3Count)
		boxesObj.HasValue("box4", expectedSnapshot.Box4Count)
		boxesObj.HasValue("box5", expectedSnapshot.Box5Count)
		boxesObj.HasValue("box6", expectedSnapshot.Box6Count)
		boxesObj.HasValue("box7", expectedSnapshot.Box7Count)
		boxesObj.HasValue("totalWords", expectedSnapshot.TotalWords)
	}
}

// TestBoxDistributionEndpoint_CustomDateRange tests custom date range parameter
func TestBoxDistributionEndpoint_CustomDateRange(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	testData := setupHistoricalSnapshotTestData(t, server, token)

	// Test with custom days parameter
	response7Days := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/box-distribution-history", testData.WordlistID)).
		WithQuery("days", "7").
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json7Days := response7Days.JSON().Object()
	json7Days.Value("days").IsEqual(7)
	distribution7Days := json7Days.Value("distribution").Array()

	// Should return fewer or equal snapshots for 7 days vs 30 days
	assert.True(t, distribution7Days.Length().Raw() <= float64(len(testData.ExpectedSnapshots)),
		"7-day query should return same or fewer snapshots")
}

// TestBoxDistributionEndpoint_EmptyData tests box distribution with no data
func TestBoxDistributionEndpoint_EmptyData(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithPremiumUser(t)

	// Create empty wordlist via API with unique name
	emptyWordlistID := createTestWordlist(t, server, token, "Empty Box Distribution Analytics", "No words for box distribution")

	// Clear any potential Redis cache for this wordlist
	flushRedisCache(t)

	// Test current distribution
	response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/current-box-distribution", emptyWordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	json := response.JSON().Object()
	distribution := json.Value("distribution").Object()
	distribution.Value("totalWords").Number().IsEqual(0)
	distribution.Value("box1").Number().IsEqual(0)
	distribution.Value("box7").Number().IsEqual(0)

	// Test historical distribution
	responseHistory := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/box-distribution-history", emptyWordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)

	jsonHistory := responseHistory.JSON().Object()
	jsonHistory.Value("distribution").Array().IsEmpty()
}

// BoxDistributionTestData contains expected box distribution data
type BoxDistributionTestData struct {
	UserID               int64
	WordlistID           int64
	TotalWords           int
	ExpectedDistribution map[int]int // box_id -> count
}

type HistoricalSnapshotTestData struct {
	UserID            int64
	WordlistID        int64
	ExpectedSnapshots []ExpectedSnapshot
}

type ExpectedSnapshot struct {
	Date       string
	Box1Count  int
	Box2Count  int
	Box3Count  int
	Box4Count  int
	Box5Count  int
	Box6Count  int
	Box7Count  int
	TotalWords int
}

// setupBoxDistributionTestData creates test data with known box distribution using API calls
func setupBoxDistributionTestData(t *testing.T, server *setup.TestServer, token string) *BoxDistributionTestData {
	ctx := context.Background()

	userID := getUserID(ctx, t, server.DB)

	// Create wordlist via API with unique name
	wordlistID := createTestWordlist(t, server, token, "Box Distribution Accurate Test", "Test wordlist for accurate box distribution")

	// Clear any potential Redis cache for this wordlist
	flushRedisCache(t)

	// Define exact box distribution for testing
	boxDistribution := map[int]int{
		1: 3, // 3 words in box 1
		2: 2, // 2 words in box 2
		3: 0, // 0 words in box 3
		4: 1, // 1 word in box 4
		5: 2, // 2 words in box 5
		6: 1, // 1 word in box 6
		7: 4, // 4 words in box 7
	}

	totalWords := 0
	for boxID, count := range boxDistribution {
		for i := 0; i < count; i++ {
			totalWords++
			wordName := fmt.Sprintf("word_box%d_%d", boxID, i+1)

			// Create word structure via API
			wordID, definitionID, _ := createBasicWordStructure(t, server, token, wordlistID, wordName)

			// Update Leitner system tracking with specific box
			_, err := server.DB.Exec(ctx,
				`UPDATE leitner_system_tracking SET box_id = $1
				 WHERE word_id = $2 AND definition_id = $3`,
				boxID, wordID, definitionID)
			require.NoError(t, err)
		}
	}

	return &BoxDistributionTestData{
		UserID:               userID,
		WordlistID:           wordlistID,
		TotalWords:           totalWords,
		ExpectedDistribution: boxDistribution,
	}
}

// setupMaxBoxLogicTestData creates test data to verify MAX(box_id) logic using service layer calls
func setupMaxBoxLogicTestData(t *testing.T, server *setup.TestServer, token string) *BoxDistributionTestData {
	ctx := context.Background()

	userID := getUserID(ctx, t, server.DB)

	// Create wordlist via API with unique name
	wordlistID := createTestWordlist(t, server, token, "Max Box Logic Analytics Test", "Test MAX(box_id) logic for analytics")

	// Clear any potential Redis cache for this wordlist
	flushRedisCache(t)

	// Debug: Log the wordlist ID we're working with
	t.Logf("MaxBoxLogic test - UserID: %d, WordlistID: %d", userID, wordlistID)

	// Create first word with multiple definitions at different box levels
	wordID1, _, leitnerID1 := createBasicWordStructure(t, server, token, wordlistID, "word1")

	// Update the first definition to box 5 (will be the max)
	_, err := server.DB.Exec(ctx,
		`UPDATE leitner_system_tracking SET box_id = $1 WHERE id = $2`,
		5, leitnerID1)
	require.NoError(t, err)

	// Create additional definitions for word1 using service layer calls
	for i, boxLevel := range []int{2, 3} {
		// Create additional definitions using SaveDefinition service call
		tx, err := server.DB.Begin(ctx)
		require.NoError(t, err)
		defer func() {
			if err != nil {
				_ = tx.Rollback(ctx)
			} else {
				_ = tx.Commit(ctx)
			}
		}()

		definition := &model.Definition{
			Token:                  fmt.Sprintf("word1_def%d", i+2),
			Language:               "en",
			Meaning:                fmt.Sprintf("Meaning %d for word1", i+2),
			PartOfSpeech:           "noun",
			Examples:               []string{fmt.Sprintf("Example %d", i+2)},
			Source:                 "test",
			PartOfSpeechNormalized: "noun",
		}

		definitions, err := service.SaveDefinition(fmt.Sprintf("word1_def%d", i+2), wordID1, []*model.Definition{definition}, tx)
		require.NoError(t, err)
		require.Len(t, definitions, 1)

		definitionID := definitions[0].ID

		// Add to Leitner system using service call
		strategy := service.LeitnerSystemStrategy{}
		err = strategy.IncludeDefinitions(wordID1, userID, []int64{definitionID}, tx)
		require.NoError(t, err)

		// Update box level for this definition
		_, err = tx.Exec(ctx,
			`UPDATE leitner_system_tracking SET box_id = $1 
			 WHERE word_id = $2 AND definition_id = $3`,
			boxLevel, wordID1, definitionID)
		require.NoError(t, err)
	}

	// Create second word with multiple definitions at different box levels
	wordID2, _, leitnerID2 := createBasicWordStructure(t, server, token, wordlistID, "word2")

	// Update the first definition to box 7 (will be the max)
	_, err = server.DB.Exec(ctx,
		`UPDATE leitner_system_tracking SET box_id = $1 WHERE id = $2`,
		7, leitnerID2)
	require.NoError(t, err)

	// Create additional definitions for word2 using service layer calls
	for i, boxLevel := range []int{1, 4} {
		// Create additional definitions using SaveDefinition service call
		tx, err := server.DB.Begin(ctx)
		require.NoError(t, err)
		defer func() {
			if err != nil {
				_ = tx.Rollback(ctx)
			} else {
				_ = tx.Commit(ctx)
			}
		}()

		definition := &model.Definition{
			Token:                  fmt.Sprintf("word2_def%d", i+2),
			Language:               "en",
			Meaning:                fmt.Sprintf("Meaning %d for word2", i+2),
			PartOfSpeech:           "noun",
			Examples:               []string{fmt.Sprintf("Example %d", i+2)},
			Source:                 "test",
			PartOfSpeechNormalized: "noun",
		}

		definitions, err := service.SaveDefinition(fmt.Sprintf("word2_def%d", i+2), wordID2, []*model.Definition{definition}, tx)
		require.NoError(t, err)
		require.Len(t, definitions, 1)

		definitionID := definitions[0].ID

		// Add to Leitner system using service call
		strategy := service.LeitnerSystemStrategy{}
		err = strategy.IncludeDefinitions(wordID2, userID, []int64{definitionID}, tx)
		require.NoError(t, err)

		// Update box level for this definition
		_, err = tx.Exec(ctx,
			`UPDATE leitner_system_tracking SET box_id = $1 
			 WHERE word_id = $2 AND definition_id = $3`,
			boxLevel, wordID2, definitionID)
		require.NoError(t, err)
	}

	// Debug: Verify what we've created
	var wordCount int
	err = server.DB.QueryRow(ctx,
		"SELECT COUNT(DISTINCT word_id) FROM leitner_system_tracking WHERE user_id = $1",
		userID).Scan(&wordCount)
	require.NoError(t, err)
	t.Logf("MaxBoxLogic test - Created %d words for user %d", wordCount, userID)

	return &BoxDistributionTestData{
		UserID:     userID,
		WordlistID: wordlistID,
		TotalWords: 2, // Only 2 words total, regardless of number of definitions
	}
}

// setupHistoricalSnapshotTestData creates test data with historical box distribution snapshots using API calls
func setupHistoricalSnapshotTestData(t *testing.T, server *setup.TestServer, token string) *HistoricalSnapshotTestData {
	ctx := context.Background()

	userID := getUserID(ctx, t, server.DB)

	// Create wordlist via API with unique name
	wordlistID := createTestWordlist(t, server, token, "Box Distribution Historical Test", "Test historical box snapshots")

	// Clear any potential Redis cache for this wordlist
	flushRedisCache(t)

	// Create historical snapshots for different dates (in DESC order for testing)
	snapshots := []ExpectedSnapshot{
		{
			Date:      time.Now().Format("2006-01-02"),
			Box1Count: 2, Box2Count: 1, Box3Count: 0, Box4Count: 1,
			Box5Count: 1, Box6Count: 0, Box7Count: 2, TotalWords: 7,
		},
		{
			Date:      time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			Box1Count: 3, Box2Count: 2, Box3Count: 1, Box4Count: 0,
			Box5Count: 0, Box6Count: 1, Box7Count: 1, TotalWords: 8,
		},
		{
			Date:      time.Now().AddDate(0, 0, -2).Format("2006-01-02"),
			Box1Count: 4, Box2Count: 1, Box3Count: 1, Box4Count: 1,
			Box5Count: 1, Box6Count: 0, Box7Count: 0, TotalWords: 8,
		},
	}

	// Insert historical snapshots (these would be created by background jobs in production)
	for _, snapshot := range snapshots {
		_, err := server.DB.Exec(ctx,
			`INSERT INTO box_distribution_snapshot (user_id, wordlist_id, snapshot_date, 
			 box_1_count, box_2_count, box_3_count, box_4_count, 
			 box_5_count, box_6_count, box_7_count) 
			 VALUES ($1, $2, $3::date, $4, $5, $6, $7, $8, $9, $10)`,
			userID, wordlistID, snapshot.Date,
			snapshot.Box1Count, snapshot.Box2Count, snapshot.Box3Count, snapshot.Box4Count,
			snapshot.Box5Count, snapshot.Box6Count, snapshot.Box7Count)
		require.NoError(t, err)
	}

	return &HistoricalSnapshotTestData{
		UserID:            userID,
		WordlistID:        wordlistID,
		ExpectedSnapshots: snapshots,
	}
}
