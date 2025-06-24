package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"decorebator.com/internal/common"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestServerWithBurstDetection creates a test server with burst detection enabled
func createTestServerWithBurstDetection(t *testing.T) *setup.TestServer {
	// Create test server with burst detection enabled (not disabled like normal tests)
	config := &setup.TestConfig{
		DisableBurstDetector: false, // Enable burst detection for these tests
		ModerationService:    service.NewMockModerationService(),
	}

	return setup.NewTestServer(t, config)
}

func TestBurstDetectorHTTP(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Get Redis client for test isolation
	redisClient, err := common.GetRedisClient()
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("Free user error report burst detection", func(t *testing.T) {
		redisClient.FlushAll(ctx)
		server := createTestServerWithBurstDetection(t)
		defer server.Cleanup()

		token := server.WithTestUser(t)

		// Create a wordlist and word for error reporting (free users can do this)
		wordlistResp := server.Expect.POST("/wordlists").
			WithHeader("Authorization", "Bearer "+token).
			WithJSON(map[string]interface{}{
				"name":         "Free User Test List",
				"description":  "For error report testing",
				"languageCode": "en",
			}).
			Expect().
			Status(http.StatusCreated)

		wordlistID := wordlistResp.JSON().Object().Value("id").Number().Raw()

		wordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
			WithHeader("Authorization", "Bearer "+token).
			WithJSON(map[string]interface{}{
				"name":  "testword",
				"notes": "test notes",
			}).
			Expect().
			Status(http.StatusCreated)

		wordID := wordResp.JSON().Object().Value("id").Number().Raw()

		// Test normal usage (under threshold of 20 error reports)
		for i := 0; i < 10; i++ {
			server.Expect.POST("/errorReports").
				WithHeader("Authorization", "Bearer "+token).
				WithJSON(map[string]interface{}{
					"wordId":      wordID,
					"errorType":   "_unrelated_example",
					"description": fmt.Sprintf("Error report %d", i),
				}).
				Expect().
				Status(http.StatusOK)
		}

		// Trigger burst detection (exceed threshold of 20 error reports per minute)
		for i := 10; i < 22; i++ {
			resp := server.Expect.POST("/errorReports").
				WithHeader("Authorization", "Bearer "+token).
				WithJSON(map[string]interface{}{
					"wordId":      wordID,
					"errorType":   "_unrelated_example",
					"description": fmt.Sprintf("Burst report %d", i),
				}).
				Expect()

			if i < 20 {
				// Should still succeed
				resp.Status(http.StatusOK)
			} else if i == 20 {
				// 21st request - first violation should return warning
				resp.Status(http.StatusTooManyRequests)
				respBody := resp.JSON().Object()
				respBody.Value("error_code").String().IsEqual("BURST_WARNING")
				respBody.Value("retry_after").Number().IsEqual(60)
				respBody.ContainsKey("error")
				break
			}
		}

		// Reset burst window to simulate time passing
		burstKey := fmt.Sprintf("burst:%d:error_report", getUserIDFromToken(t, server, token))
		redisClient.Del(ctx, burstKey)

		// Second burst should result in blocking
		for i := 0; i < 22; i++ {
			resp := server.Expect.POST("/errorReports").
				WithHeader("Authorization", "Bearer "+token).
				WithJSON(map[string]interface{}{
					"wordId":      wordID,
					"errorType":   "_unrelated_example",
					"description": fmt.Sprintf("Block report %d", i),
				}).
				Expect()

			if i < 20 {
				// Should succeed
				resp.Status(http.StatusOK)
			} else if i == 20 {
				// Should trigger blocking
				resp.Status(http.StatusForbidden)
				respBody := resp.JSON().Object()
				respBody.Value("error_code").String().IsEqual("ACCOUNT_SUSPENDED_BURST")
				respBody.ContainsKey("suspended_until")
				break
			}
		}

		// Verify user is now blocked on all subsequent requests
		server.Expect.POST("/errorReports").
			WithHeader("Authorization", "Bearer "+token).
			WithJSON(map[string]interface{}{
				"wordId":      wordID,
				"errorType":   "_unrelated_example",
				"description": "Should be blocked",
			}).
			Expect().
			Status(http.StatusForbidden).
			JSON().Object().
			Value("error_code").String().IsEqual("ACCOUNT_SUSPENDED_BURST")
	})

	t.Run("Premium user wordlist creation burst protection", func(t *testing.T) {
		redisClient.FlushAll(ctx)
		server := createTestServerWithBurstDetection(t)
		defer server.Cleanup()

		token := server.WithPremiumUser(t)

		// Premium users should have same burst detection (no bypass)
		// Test wordlist creation burst (threshold: 10) since premium users have unlimited wordlists

		// Test normal usage (under threshold of 10)
		for i := 0; i < 5; i++ {
			server.Expect.POST("/wordlists").
				WithHeader("Authorization", "Bearer "+token).
				WithJSON(map[string]interface{}{
					"name":         fmt.Sprintf("Premium List %d", i),
					"description":  "Test description",
					"languageCode": "en",
				}).
				Expect().
				Status(http.StatusCreated)
		}

		// Exceed threshold to trigger burst detection (10 wordlists per minute)
		for i := 5; i < 12; i++ {
			resp := server.Expect.POST("/wordlists").
				WithHeader("Authorization", "Bearer "+token).
				WithJSON(map[string]interface{}{
					"name":         fmt.Sprintf("Burst List %d", i),
					"description":  "Test description",
					"languageCode": "en",
				}).
				Expect()

			if i < 10 {
				resp.Status(http.StatusCreated)
			} else if i == 10 {
				// First violation - warning
				resp.Status(http.StatusTooManyRequests)
				respBody := resp.JSON().Object()
				respBody.Value("error_code").String().IsEqual("BURST_WARNING")
				respBody.Value("retry_after").Number().IsEqual(60)
				respBody.ContainsKey("error")
				break
			}
		}

		// Verify premium users get same protection as free users
		assert.True(t, true, "Premium users are subject to burst detection")
	})

	t.Run("Error report endpoint burst detection", func(t *testing.T) {
		redisClient.FlushAll(ctx)
		server := createTestServerWithBurstDetection(t)
		defer server.Cleanup()

		token := server.WithTestUser(t)

		// Create a wordlist and word for error reporting
		wordlistResp := server.Expect.POST("/wordlists").
			WithHeader("Authorization", "Bearer "+token).
			WithJSON(map[string]interface{}{
				"name":         "Error Test List",
				"description":  "For error reporting",
				"languageCode": "en",
			}).
			Expect().
			Status(http.StatusCreated)

		wordlistID := wordlistResp.JSON().Object().Value("id").Number().Raw()

		wordResp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
			WithHeader("Authorization", "Bearer "+token).
			WithJSON(map[string]interface{}{
				"name":  "testword",
				"notes": "test notes",
			}).
			Expect().
			Status(http.StatusCreated)

		wordID := wordResp.JSON().Object().Value("id").Number().Raw()

		// Test error report burst detection (threshold: 20)
		for i := 0; i < 22; i++ {
			resp := server.Expect.POST("/errorReports").
				WithHeader("Authorization", "Bearer "+token).
				WithJSON(map[string]interface{}{
					"wordId":      wordID,
					"errorType":   "_unrelated_example",
					"description": fmt.Sprintf("Error report %d", i),
				}).
				Expect()

			if i < 20 {
				resp.Status(http.StatusOK)
			} else if i == 20 {
				// First violation
				resp.Status(http.StatusTooManyRequests)
				respBody := resp.JSON().Object()
				respBody.Value("error_code").String().IsEqual("BURST_WARNING")
				break
			}
		}
	})

	t.Run("Cross-endpoint blocking", func(t *testing.T) {
		redisClient.FlushAll(ctx)
		server := createTestServerWithBurstDetection(t)
		defer server.Cleanup()

		token := server.WithTestUser(t)
		detector := service.NewBurstDetector(redisClient)
		userID := getUserIDFromToken(t, server, token)

		// Manually block user
		err := detector.BlockUserFor24Hours(ctx, userID)
		require.NoError(t, err)

		// User should be blocked on all protected endpoints
		endpoints := []struct {
			method string
			path   string
			body   map[string]interface{}
		}{
			{
				"POST", "/wordlists",
				map[string]interface{}{
					"name":         "Blocked Test",
					"description":  "Should fail",
					"languageCode": "en",
				},
			},
			{
				"POST", "/errorReports",
				map[string]interface{}{
					"wordId":      1,
					"errorType":   "_unrelated_example",
					"description": "Should fail",
				},
			},
		}

		for _, endpoint := range endpoints {
			server.Expect.Request(endpoint.method, endpoint.path).
				WithHeader("Authorization", "Bearer "+token).
				WithJSON(endpoint.body).
				Expect().
				Status(http.StatusForbidden).
				JSON().Object().
				Value("error_code").String().IsEqual("ACCOUNT_SUSPENDED_BURST")
		}
	})

	t.Run("Admin unblock functionality", func(t *testing.T) {
		redisClient.FlushAll(ctx)
		server := createTestServerWithBurstDetection(t)
		defer server.Cleanup()

		token := server.WithTestUser(t)
		detector := service.NewBurstDetector(redisClient)
		userID := getUserIDFromToken(t, server, token)

		// Block user
		err := detector.BlockUserFor24Hours(ctx, userID)
		require.NoError(t, err)

		// Verify user is blocked
		server.Expect.POST("/wordlists").
			WithHeader("Authorization", "Bearer "+token).
			WithJSON(map[string]interface{}{
				"name":         "Should be blocked",
				"description":  "Test",
				"languageCode": "en",
			}).
			Expect().
			Status(http.StatusForbidden)

		// Test admin unblock endpoint
		server.Expect.POST("/static/admin/burst/unblock/{userId}", userID).
			WithHeader("Authorization", os.Getenv("STATIC_AUTHENTICATION")).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			Value("message").String().Contains("unblocked successfully")

		// Verify user is no longer blocked
		server.Expect.POST("/wordlists").
			WithHeader("Authorization", "Bearer "+token).
			WithJSON(map[string]interface{}{
				"name":         "Should work now",
				"description":  "Test",
				"languageCode": "en",
			}).
			Expect().
			Status(http.StatusCreated)
	})

	t.Run("Admin unblock all users", func(t *testing.T) {
		redisClient.FlushAll(ctx)
		server := createTestServerWithBurstDetection(t)
		defer server.Cleanup()

		// Create and block multiple users
		tokens := make([]string, 3)
		userIDs := make([]int64, 3)
		detector := service.NewBurstDetector(redisClient)

		for i := 0; i < 3; i++ {
			tokens[i] = server.WithTestUser(t)
			userIDs[i] = getUserIDFromToken(t, server, tokens[i])
			err := detector.BlockUserFor24Hours(ctx, userIDs[i])
			require.NoError(t, err)
		}

		// Verify all users are blocked
		for i, token := range tokens {
			server.Expect.POST("/wordlists").
				WithHeader("Authorization", "Bearer "+token).
				WithJSON(map[string]interface{}{
					"name":         fmt.Sprintf("Blocked %d", i),
					"description":  "Should fail",
					"languageCode": "en",
				}).
				Expect().
				Status(http.StatusForbidden)
		}

		// Test mass unblock
		server.Expect.POST("/static/admin/burst/unblock-all").
			WithHeader("Authorization", os.Getenv("STATIC_AUTHENTICATION")).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			Value("unblocked_count").Number().IsEqual(3)

		// Verify all users are unblocked
		for i, token := range tokens {
			server.Expect.POST("/wordlists").
				WithHeader("Authorization", "Bearer "+token).
				WithJSON(map[string]interface{}{
					"name":         fmt.Sprintf("Unblocked %d", i),
					"description":  "Should work",
					"languageCode": "en",
				}).
				Expect().
				Status(http.StatusCreated)
		}
	})

	t.Run("Admin blocked users list", func(t *testing.T) {
		redisClient.FlushAll(ctx)
		server := createTestServerWithBurstDetection(t)
		defer server.Cleanup()

		detector := service.NewBurstDetector(redisClient)

		// Initially no blocked users
		server.Expect.GET("/static/admin/burst/blocked-users").
			WithHeader("Authorization", os.Getenv("STATIC_AUTHENTICATION")).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			Value("total_count").Number().IsEqual(0)

		// Block some users
		userIDs := make([]int64, 2)
		for i := 0; i < 2; i++ {
			token := server.WithTestUser(t)
			userIDs[i] = getUserIDFromToken(t, server, token)
			err := detector.BlockUserFor24Hours(ctx, userIDs[i])
			require.NoError(t, err)
		}

		// Check blocked users list
		resp := server.Expect.GET("/static/admin/burst/blocked-users").
			WithHeader("Authorization", os.Getenv("STATIC_AUTHENTICATION")).
			Expect().
			Status(http.StatusOK)

		respBody := resp.JSON().Object()
		respBody.Value("total_count").Number().IsEqual(2)
		blockedUsers := respBody.Value("blocked_users").Array()
		blockedUsers.Length().IsEqual(2)

		// Verify user IDs are in the list
		for _, userID := range userIDs {
			found := false
			for _, user := range blockedUsers.Iter() {
				if user.Object().Value("user_id").Number().Raw() == float64(userID) {
					found = true
					// Verify blocked_until is present and valid
					user.Object().ContainsKey("blocked_until")
					break
				}
			}
			assert.True(t, found, "User ID %d should be in blocked users list", userID)
		}
	})
}

// Helper function to extract user ID from JWT token via API call
func getUserIDFromToken(_ *testing.T, server *setup.TestServer, token string) int64 {
	resp := server.Expect.GET("/users").
		WithHeader("Authorization", "Bearer "+token).
		Expect().
		Status(http.StatusOK)

	userID := resp.JSON().Object().Value("id").Number().Raw()
	return int64(userID)
}
