package integration

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"decorebator.com/internal/app"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/mocks"
	"decorebator.com/tests/integration/setup"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ======================
// HELPER FUNCTIONS FOR REVENUECAT MOCKING
// ======================

// Helper function to create test server with mock RevenueCat service
func newTestServerWithMockRevenueCat(t *testing.T, mockAPIClient service.RevenueCatAPIClient) *setup.TestServer {
	return setup.NewTestServer(t, func(builder *app.ContextBuilder) *app.ContextBuilder {
		// Use the WithRevenueCatServiceFunc method to inject mock API client
		return builder.WithRevenueCatServiceFunc(func(db *pgxpool.Pool) service.RevenueCatService {
			return service.NewRevenueCatService(db, mockAPIClient)
		})
	})
}

// Helper function to create premium customer info for RevenueCat mock
func generatePremiumCustomerInfo() *service.CustomerInfo {
	futureDate := time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339)
	pastDate := time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339)

	return &service.CustomerInfo{
		Subscriber: service.Subscriber{
			Entitlements: map[string]service.Entitlement{
				"Premium": {
					ProductIdentifier: "com.decorebator.premium.monthly",
					ExpiresDate:       &futureDate,
					PurchaseDate:      pastDate,
				},
			},
		},
	}
}

// Helper function to create free customer info for RevenueCat mock
func generateFreeCustomerInfo() *service.CustomerInfo {
	return &service.CustomerInfo{
		Subscriber: service.Subscriber{
			Entitlements: map[string]service.Entitlement{
				// Empty entitlements = free user
			},
		},
	}
}

// ======================
// STEP 1: Basic Infrastructure
// ======================

// TestCreateWordlist_FreeUser_UnderLimit_Success tests that a free user can create a wordlist when under the limit
func TestCreateWordlist_FreeUser_UnderLimit_Success(t *testing.T) {
	// Arrange
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	authToken := server.WithTestUser(t)

	// Act: Create wordlist (user has 0 wordlists, limit is 1)
	response := server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithJSON(map[string]interface{}{
			"name":         "Test Wordlist",
			"description":  "Test Description",
			"languageCode": "en",
		}).
		Expect()

	// Assert: Should succeed
	response.Status(http.StatusCreated)
	response.JSON().Object().Value("name").String().IsEqual("Test Wordlist")
	response.JSON().Object().Value("description").String().IsEqual("Test Description")
	response.JSON().Object().Value("languageCode").String().IsEqual("en")
}

// ======================
// STEP 2: Limit Blocking Test
// ======================

// TestCreateWordlist_FreeUser_AtLimit_Blocked tests that a free user is blocked when at the wordlist limit
func TestCreateWordlist_FreeUser_AtLimit_Blocked(t *testing.T) {
	// Arrange
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	authToken := server.WithTestUser(t)

	// Setup user at limit: create 1 wordlist (free limit is 1)
	server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithJSON(map[string]interface{}{
			"name":         "First Wordlist",
			"description":  "At the limit",
			"languageCode": "en",
		}).
		Expect().Status(http.StatusCreated)

	// Act: Try to create a second wordlist (should be blocked)
	response := server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithJSON(map[string]interface{}{
			"name":         "Blocked Wordlist",
			"description":  "Should be blocked",
			"languageCode": "en",
		}).
		Expect()

	// Assert: Should be blocked with payment required
	response.Status(http.StatusPaymentRequired)
	responseObj := response.JSON().Object()
	responseObj.Value("error").String().Contains("free plan limit reached")
	responseObj.Value("code").String().IsEqual("SUBSCRIPTION_LIMIT_REACHED")
}

// ======================
// STEP 3: Premium User Test
// ======================

// TestCreateWordlist_PremiumUser_NoLimits_Success tests that premium users can bypass wordlist limits
func TestCreateWordlist_PremiumUser_NoLimits_Success(t *testing.T) {
	// Arrange
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	authToken := server.WithPremiumUser(t) // Creates user with premium subscription

	// Setup: Create wordlist that would put free user at limit
	server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithJSON(map[string]interface{}{
			"name":         "First Premium Wordlist",
			"description":  "Premium user first wordlist",
			"languageCode": "en",
		}).
		Expect().Status(http.StatusCreated)

	// Act: Premium user should be able to create more wordlists beyond free limit
	response := server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithJSON(map[string]interface{}{
			"name":         "Second Premium Wordlist",
			"description":  "Should succeed for premium user",
			"languageCode": "en",
		}).
		Expect()

	// Assert: Should succeed because user has premium subscription
	response.Status(http.StatusCreated)
	response.JSON().Object().Value("name").String().IsEqual("Second Premium Wordlist")
	response.JSON().Object().Value("description").String().IsEqual("Should succeed for premium user")
}

// ======================
// STEP 4: RevenueCat Optimistic Subscription Test
// ======================

// TestCreateWordlist_OptimisticFlag_RevenueCatPremium_Allowed tests optimistic subscription verification with RevenueCat
func TestCreateWordlist_OptimisticFlag_RevenueCatPremium_Allowed(t *testing.T) {
	// Arrange: Mock RevenueCat to return premium subscription
	mockAPIClient := &mocks.MockRevenueCatAPIClient{
		GetCustomerInfoFunc: func(_ context.Context, _ string) (*service.CustomerInfo, error) {
			return generatePremiumCustomerInfo(), nil
		},
	}

	server := newTestServerWithMockRevenueCat(t, mockAPIClient)
	defer server.Cleanup()

	authToken := server.WithTestUser(t) // Creates free user locally

	// Setup user at limit: create 1 wordlist (should be blocked without optimistic flag)
	server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithJSON(map[string]interface{}{
			"name":         "First Wordlist",
			"description":  "At the limit",
			"languageCode": "en",
		}).
		Expect().Status(http.StatusCreated)

	// Act: Attempt to create wordlist with optimistic flag
	response := server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithQuery("hasOptimisticSubscription", "true").
		WithJSON(map[string]interface{}{
			"name":         "Optimistic Wordlist",
			"description":  "Should succeed via RevenueCat",
			"languageCode": "en",
		}).
		Expect()

	// Assert: Should succeed due to RevenueCat premium verification
	response.Status(http.StatusCreated)
	response.JSON().Object().Value("name").String().IsEqual("Optimistic Wordlist")
	response.JSON().Object().Value("description").String().IsEqual("Should succeed via RevenueCat")

	// Verify RevenueCat was called
	if len(mockAPIClient.GetCustomerInfoCalls) != 1 {
		t.Errorf("Expected RevenueCat to be called once, but was called %d times", len(mockAPIClient.GetCustomerInfoCalls))
	}
}

// ======================
// STEP 5: RevenueCat Failure Scenarios
// ======================

// TestCreateWordlist_OptimisticFlag_RevenueCatFree_Blocked tests when RevenueCat also says user is free
func TestCreateWordlist_OptimisticFlag_RevenueCatFree_Blocked(t *testing.T) {
	// Arrange: Mock RevenueCat to return free subscription
	mockAPIClient := &mocks.MockRevenueCatAPIClient{
		GetCustomerInfoFunc: func(_ context.Context, _ string) (*service.CustomerInfo, error) {
			return generateFreeCustomerInfo(), nil
		},
	}

	server := newTestServerWithMockRevenueCat(t, mockAPIClient)
	defer server.Cleanup()

	authToken := server.WithTestUser(t)

	// Setup user at limit: create 1 wordlist
	server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithJSON(map[string]interface{}{
			"name":         "First Wordlist",
			"description":  "At the limit",
			"languageCode": "en",
		}).
		Expect().Status(http.StatusCreated)

	// Act: Attempt to create wordlist with optimistic flag
	response := server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithQuery("hasOptimisticSubscription", "true").
		WithJSON(map[string]interface{}{
			"name":         "Should Be Blocked",
			"description":  "RevenueCat also says free",
			"languageCode": "en",
		}).
		Expect()

	// Assert: Should be blocked since RevenueCat also says free
	response.Status(http.StatusPaymentRequired)
	responseObj := response.JSON().Object()
	responseObj.Value("error").String().Contains("free plan limit reached")
	responseObj.Value("code").String().IsEqual("SUBSCRIPTION_LIMIT_REACHED")

	// Verify RevenueCat was called
	if len(mockAPIClient.GetCustomerInfoCalls) != 1 {
		t.Errorf("Expected RevenueCat to be called once, but was called %d times", len(mockAPIClient.GetCustomerInfoCalls))
	}
}

// ======================
// STEP 6: RevenueCat API Error Scenarios
// ======================

// TestCreateWordlist_OptimisticFlag_RevenueCatError_Blocked tests fallback when RevenueCat API fails
func TestCreateWordlist_OptimisticFlag_RevenueCatError_Blocked(t *testing.T) {
	// Arrange: Mock RevenueCat to return error
	mockAPIClient := &mocks.MockRevenueCatAPIClient{
		GetCustomerInfoFunc: func(_ context.Context, _ string) (*service.CustomerInfo, error) {
			return nil, errors.New("RevenueCat API timeout")
		},
	}

	server := newTestServerWithMockRevenueCat(t, mockAPIClient)
	defer server.Cleanup()

	authToken := server.WithTestUser(t)

	// Setup user at limit: create 1 wordlist
	server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithJSON(map[string]interface{}{
			"name":         "First Wordlist",
			"description":  "At the limit",
			"languageCode": "en",
		}).
		Expect().Status(http.StatusCreated)

	// Act: Attempt to create wordlist with optimistic flag when RevenueCat API fails
	response := server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithQuery("hasOptimisticSubscription", "true").
		WithJSON(map[string]interface{}{
			"name":         "Should Be Blocked",
			"description":  "API error fallback",
			"languageCode": "en",
		}).
		Expect()

	// Assert: Should be blocked due to API error fallback (conservative approach)
	response.Status(http.StatusPaymentRequired)
	responseObj := response.JSON().Object()
	responseObj.Value("error").String().Contains("free plan limit reached")
	responseObj.Value("code").String().IsEqual("SUBSCRIPTION_LIMIT_REACHED")

	// Verify RevenueCat was called despite error
	if len(mockAPIClient.GetCustomerInfoCalls) != 1 {
		t.Errorf("Expected RevenueCat to be called once, but was called %d times", len(mockAPIClient.GetCustomerInfoCalls))
	}
}

// ======================
// STEP 7: Word Limit Enforcement
// ======================

// TestAddWord_FreeUser_AtLimit_Blocked tests that free users are blocked when adding words beyond the limit
func TestAddWord_FreeUser_AtLimit_Blocked(t *testing.T) {
	// Arrange
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	authToken := server.WithTestUser(t)

	// Create a wordlist first
	wordlistResponse := server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithJSON(map[string]interface{}{
			"name":         "Test Wordlist",
			"description":  "For word limit testing",
			"languageCode": "en",
		}).
		Expect().Status(http.StatusCreated)

	wordlistID := int(wordlistResponse.JSON().Object().Value("id").Number().Raw())

	// Setup: Add 10 words (free limit is 10 words per wordlist)
	for i := 1; i <= 10; i++ {
		server.Expect.POST(fmt.Sprintf("/wordlists/%d/words", wordlistID)).
			WithHeader("Authorization", authToken).
			WithJSON(map[string]interface{}{
				"name":  fmt.Sprintf("word%d", i),
				"notes": fmt.Sprintf("Test word %d", i),
			}).
			Expect().Status(http.StatusCreated)
	}

	// Act: Try to add 11th word (should be blocked)
	response := server.Expect.POST(fmt.Sprintf("/wordlists/%d/words", wordlistID)).
		WithHeader("Authorization", authToken).
		WithJSON(map[string]interface{}{
			"name":  "word11",
			"notes": "Should be blocked",
		}).
		Expect()

	// Assert: Should be blocked with payment required
	response.Status(http.StatusPaymentRequired)
	responseObj := response.JSON().Object()
	responseObj.Value("error").String().Contains("free plan limit reached")
	responseObj.Value("code").String().IsEqual("SUBSCRIPTION_LIMIT_REACHED")
}

// ======================
// STEP 8: Word Limit with Optimistic Subscription
// ======================

// TestAddWord_OptimisticFlag_RevenueCatPremium_Allowed tests optimistic word limit bypass with RevenueCat premium
func TestAddWord_OptimisticFlag_RevenueCatPremium_Allowed(t *testing.T) {
	// Arrange: Mock RevenueCat to return premium subscription
	mockAPIClient := &mocks.MockRevenueCatAPIClient{
		GetCustomerInfoFunc: func(_ context.Context, _ string) (*service.CustomerInfo, error) {
			return generatePremiumCustomerInfo(), nil
		},
	}

	server := newTestServerWithMockRevenueCat(t, mockAPIClient)
	defer server.Cleanup()

	authToken := server.WithTestUser(t) // Creates free user locally

	// Create a wordlist first
	wordlistResponse := server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithJSON(map[string]interface{}{
			"name":         "Premium Test Wordlist",
			"description":  "For premium word limit testing",
			"languageCode": "en",
		}).
		Expect().Status(http.StatusCreated)

	wordlistID := int(wordlistResponse.JSON().Object().Value("id").Number().Raw())

	// Setup: Add 10 words to reach free limit
	for i := 1; i <= 10; i++ {
		server.Expect.POST(fmt.Sprintf("/wordlists/%d/words", wordlistID)).
			WithHeader("Authorization", authToken).
			WithJSON(map[string]interface{}{
				"name":  fmt.Sprintf("word%d", i),
				"notes": fmt.Sprintf("Test word %d", i),
			}).
			Expect().Status(http.StatusCreated)
	}

	// Act: Add 11th word with optimistic flag (should succeed via RevenueCat verification)
	response := server.Expect.POST(fmt.Sprintf("/wordlists/%d/words", wordlistID)).
		WithHeader("Authorization", authToken).
		WithQuery("hasOptimisticSubscription", "true").
		WithJSON(map[string]interface{}{
			"name":  "premium-word",
			"notes": "Should succeed via RevenueCat",
		}).
		Expect()

	// Assert: Should succeed due to RevenueCat premium verification
	response.Status(http.StatusCreated)
	response.JSON().Object().Value("name").String().IsEqual("premium-word")
	response.JSON().Object().Value("notes").String().IsEqual("Should succeed via RevenueCat")

	// Verify RevenueCat was called
	if len(mockAPIClient.GetCustomerInfoCalls) != 1 {
		t.Errorf("Expected RevenueCat to be called once, but was called %d times", len(mockAPIClient.GetCustomerInfoCalls))
	}
}

// ======================
// STEP 9: Edge Cases and Cross-Wordlist Limits
// ======================

// TestCreateWordlist_FreeUser_WithExistingWordlist_StillBlocked tests that wordlist limits are enforced globally
func TestCreateWordlist_FreeUser_WithExistingWordlist_StillBlocked(t *testing.T) {
	// Arrange
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	authToken := server.WithTestUser(t)

	// Create first wordlist (uses up the 1 wordlist limit for free users)
	firstWordlistResponse := server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithJSON(map[string]interface{}{
			"name":         "First Wordlist",
			"description":  "Taking up the free limit",
			"languageCode": "en",
		}).
		Expect().Status(http.StatusCreated)

	firstWordlistID := int(firstWordlistResponse.JSON().Object().Value("id").Number().Raw())

	// Add some words to the first wordlist (within limits)
	for i := 1; i <= 5; i++ {
		server.Expect.POST(fmt.Sprintf("/wordlists/%d/words", firstWordlistID)).
			WithHeader("Authorization", authToken).
			WithJSON(map[string]interface{}{
				"name":  fmt.Sprintf("word%d", i),
				"notes": fmt.Sprintf("Word %d in first wordlist", i),
			}).
			Expect().Status(http.StatusCreated)
	}

	// Act: Try to create a second wordlist (should be blocked)
	response := server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithJSON(map[string]interface{}{
			"name":         "Second Wordlist",
			"description":  "Should be blocked",
			"languageCode": "es",
		}).
		Expect()

	// Assert: Should be blocked even though first wordlist isn't at word capacity
	response.Status(http.StatusPaymentRequired)
	responseObj := response.JSON().Object()
	responseObj.Value("error").String().Contains("free plan limit reached")
	responseObj.Value("code").String().IsEqual("SUBSCRIPTION_LIMIT_REACHED")

	// Verify we can still add words to existing wordlist (within word limits)
	server.Expect.POST(fmt.Sprintf("/wordlists/%d/words", firstWordlistID)).
		WithHeader("Authorization", authToken).
		WithJSON(map[string]interface{}{
			"name":  "additional-word",
			"notes": "Should still work in existing wordlist",
		}).
		Expect().Status(http.StatusCreated)
}

// ======================
// STEP 10: Comprehensive End-to-End Flow
// ======================

// TestSubscriptionLimits_CompleteFlow_AllScenarios tests the complete subscription limit enforcement flow
func TestSubscriptionLimits_CompleteFlow_AllScenarios(t *testing.T) {
	// Test multiple scenarios with RevenueCat integration
	callCount := 0
	mockAPIClient := &mocks.MockRevenueCatAPIClient{
		GetCustomerInfoFunc: func(_ context.Context, _ string) (*service.CustomerInfo, error) {
			callCount++
			// First call returns premium, second call returns free (simulating subscription change)
			if callCount == 1 {
				return generatePremiumCustomerInfo(), nil
			}
			return generateFreeCustomerInfo(), nil
		},
	}

	server := newTestServerWithMockRevenueCat(t, mockAPIClient)
	defer server.Cleanup()

	authToken := server.WithTestUser(t)

	// Phase 1: User starts as free, creates 1 wordlist
	wordlistResponse := server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithJSON(map[string]interface{}{
			"name":         "Test Wordlist",
			"description":  "Complete flow test",
			"languageCode": "en",
		}).
		Expect().Status(http.StatusCreated)

	wordlistID := int(wordlistResponse.JSON().Object().Value("id").Number().Raw())

	// Phase 2: Try to create second wordlist without optimistic flag (should be blocked)
	server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithJSON(map[string]interface{}{
			"name":         "Should Be Blocked",
			"description":  "No optimistic flag",
			"languageCode": "es",
		}).
		Expect().Status(http.StatusPaymentRequired)

	// Phase 3: Try with optimistic flag - RevenueCat returns premium (should succeed)
	secondWordlistResponse := server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithQuery("hasOptimisticSubscription", "true").
		WithJSON(map[string]interface{}{
			"name":         "Premium Success",
			"description":  "Via RevenueCat premium",
			"languageCode": "fr",
		}).
		Expect().Status(http.StatusCreated)

	secondWordlistID := int(secondWordlistResponse.JSON().Object().Value("id").Number().Raw())

	// Phase 4: Add words to both wordlists
	for i := 1; i <= 5; i++ {
		// Add to first wordlist
		server.Expect.POST(fmt.Sprintf("/wordlists/%d/words", wordlistID)).
			WithHeader("Authorization", authToken).
			WithJSON(map[string]interface{}{
				"name":  fmt.Sprintf("word1-%d", i),
				"notes": fmt.Sprintf("Word %d in first wordlist", i),
			}).
			Expect().Status(http.StatusCreated)

		// Add to second wordlist
		server.Expect.POST(fmt.Sprintf("/wordlists/%d/words", secondWordlistID)).
			WithHeader("Authorization", authToken).
			WithJSON(map[string]interface{}{
				"name":  fmt.Sprintf("word2-%d", i),
				"notes": fmt.Sprintf("Word %d in second wordlist", i),
			}).
			Expect().Status(http.StatusCreated)
	}

	// Phase 5: Try third wordlist with optimistic flag - RevenueCat now returns free (should be blocked)
	server.Expect.POST("/wordlists").
		WithHeader("Authorization", authToken).
		WithQuery("hasOptimisticSubscription", "true").
		WithJSON(map[string]interface{}{
			"name":         "Should Be Blocked Now",
			"description":  "RevenueCat now says free",
			"languageCode": "de",
		}).
		Expect().Status(http.StatusPaymentRequired)

	// Phase 6: Verify word limits still work - fill first wordlist to capacity
	for i := 6; i <= 10; i++ {
		server.Expect.POST(fmt.Sprintf("/wordlists/%d/words", wordlistID)).
			WithHeader("Authorization", authToken).
			WithJSON(map[string]interface{}{
				"name":  fmt.Sprintf("word1-%d", i),
				"notes": fmt.Sprintf("Word %d fills capacity", i),
			}).
			Expect().Status(http.StatusCreated)
	}

	// Phase 7: Try to add 11th word to first wordlist (should be blocked by word limit)
	server.Expect.POST(fmt.Sprintf("/wordlists/%d/words", wordlistID)).
		WithHeader("Authorization", authToken).
		WithJSON(map[string]interface{}{
			"name":  "overflow-word",
			"notes": "Should be blocked by word limit",
		}).
		Expect().Status(http.StatusPaymentRequired)

	// Verify RevenueCat was called twice (once for each optimistic request)
	if len(mockAPIClient.GetCustomerInfoCalls) != 2 {
		t.Errorf("Expected RevenueCat to be called twice, but was called %d times", len(mockAPIClient.GetCustomerInfoCalls))
	}
}
