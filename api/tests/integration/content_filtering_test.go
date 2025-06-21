package integration

import (
	"net/http"
	"testing"

	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
)

// setupMockModeration injects a mock moderation service for testing
func setupMockModeration(_ *testing.T) func() {
	// Create and inject mock moderation service
	mockModeration := service.NewMockModerationService()
	service.SetModerationService(mockModeration)
	service.SetWordlistModerationService(mockModeration)

	// Return cleanup function to restore original services
	cleanup := func() {
		// Restore original moderation services
		service.SetModerationService(service.NewOpenAIModerationService())
		service.SetWordlistModerationService(service.NewOpenAIModerationService())
	}

	return cleanup
}

// Helper function to create a test user
func createTestUser(_ *testing.T, server *setup.TestServer) map[string]interface{} {
	testUser := map[string]interface{}{
		"email":     "testuser@example.com",
		"password":  "password123",
		"firstName": "Test",
		"lastName":  "User",
	}

	resp := server.Expect.POST("/users").
		WithJSON(testUser).
		Expect().
		Status(http.StatusCreated)

	return map[string]interface{}{
		"email":    testUser["email"],
		"password": testUser["password"],
		"token":    resp.Header("Authorization").NotEmpty().Raw(),
	}
}

// Helper function to create a premium test user
func createPremiumTestUser(t *testing.T, server *setup.TestServer) map[string]interface{} {
	token := server.WithPremiumUser(t)

	return map[string]interface{}{
		"token": token,
	}
}

func TestContentFilteringWordlist_Create_ShouldRejectInappropriateName(t *testing.T) {
	// Arrange
	cleanup := setupMockModeration(t)
	defer cleanup()

	server := setup.NewTestServer(t)
	defer server.Cleanup()

	user := createTestUser(t, server)
	token := user["token"].(string)

	testCases := []struct {
		name         string
		wordlistName string
		expectedCode int
		shouldReject bool
		description  string
	}{
		{
			name:         "Profanity in name",
			wordlistName: "damn words",
			expectedCode: http.StatusBadRequest,
			shouldReject: true,
			description:  "Should reject wordlist name containing profanity",
		},
		{
			name:         "Adult content indicator",
			wordlistName: "xxx vocabulary",
			expectedCode: http.StatusBadRequest,
			shouldReject: true,
			description:  "Should reject wordlist name with adult content indicators",
		},
		{
			name:         "Violence indicator",
			wordlistName: "kill list",
			expectedCode: http.StatusBadRequest,
			shouldReject: true,
			description:  "Should reject wordlist name with violence indicators",
		},
		{
			name:         "URL in name",
			wordlistName: "visit https://spam.com",
			expectedCode: http.StatusBadRequest,
			shouldReject: true,
			description:  "Should reject wordlist name containing URLs",
		},
		{
			name:         "Email in name",
			wordlistName: "contact spam@test.com",
			expectedCode: http.StatusBadRequest,
			shouldReject: true,
			description:  "Should reject wordlist name containing email addresses",
		},
		{
			name:         "Phone number in name",
			wordlistName: "call 555-123-4567",
			expectedCode: http.StatusBadRequest,
			shouldReject: true,
			description:  "Should reject wordlist name containing phone numbers",
		},
		{
			name:         "Excessive repetition",
			wordlistName: "aaaaaaaaaaa spam spam spam",
			expectedCode: http.StatusBadRequest,
			shouldReject: true,
			description:  "Should reject wordlist name with excessive repetition",
		},
		{
			name:         "All caps spam",
			wordlistName: "SPAMMY WORDLIST NAME",
			expectedCode: http.StatusBadRequest,
			shouldReject: true,
			description:  "Should reject all caps wordlist names",
		},
		{
			name:         "Spanish profanity",
			wordlistName: "lista estúpida",
			expectedCode: http.StatusBadRequest,
			shouldReject: true,
			description:  "Should reject Spanish profanity in wordlist names",
		},
		{
			name:         "French profanity",
			wordlistName: "liste stupide",
			expectedCode: http.StatusBadRequest,
			shouldReject: true,
			description:  "Should reject French profanity in wordlist names",
		},
		{
			name:         "Valid educational name",
			wordlistName: "Spanish Vocabulary",
			expectedCode: http.StatusCreated,
			shouldReject: false,
			description:  "Should accept valid educational wordlist name",
		},
		{
			name:         "Valid name with description",
			wordlistName: "French Basics",
			expectedCode: http.StatusCreated,
			shouldReject: false,
			description:  "Should accept valid wordlist name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			resp := server.Expect.POST("/wordlists").
				WithHeader("Authorization", "Bearer "+token).
				WithJSON(map[string]interface{}{
					"name":         tc.wordlistName,
					"description":  "Test description",
					"languageCode": "en",
				}).
				Expect()

			// Assert
			resp.Status(tc.expectedCode)

			if tc.shouldReject {
				respBody := resp.JSON().Object()
				respBody.ContainsKey("error")
				err := respBody.Value("error").String().Raw()
				assert.Contains(t, err, "not appropriate", tc.description)
			} else {
				respBody := resp.JSON().Object()
				respBody.ContainsKey("id")
				respBody.Value("name").String().IsEqual(tc.wordlistName)
			}
		})
	}
}

func TestContentFilteringWordlist_Create_ShouldRejectInappropriateDescription(t *testing.T) {
	// Arrange
	cleanup := setupMockModeration(t)
	defer cleanup()

	server := setup.NewTestServer(t)
	defer server.Cleanup()

	user := createTestUser(t, server)
	token := user["token"].(string)

	testCases := []struct {
		name        string
		description string
		shouldPass  bool
		testDesc    string
	}{
		{
			name:        "Profanity in description",
			description: "This is a damn good vocabulary list with stupid words",
			shouldPass:  false,
			testDesc:    "Should reject description containing profanity",
		},
		{
			name:        "Adult content in description",
			description: "Adult vocabulary with xxx content for mature learners",
			shouldPass:  false,
			testDesc:    "Should reject description with adult content indicators",
		},
		{
			name:        "Violence in description",
			description: "Learn words about violence and murder for advanced students",
			shouldPass:  false,
			testDesc:    "Should reject description with violence content",
		},
		{
			name:        "URL in description",
			description: "Visit https://spam.com for more vocabulary resources",
			shouldPass:  false,
			testDesc:    "Should reject description containing URLs",
		},
		{
			name:        "Email in description",
			description: "Contact me at spam@test.com for vocabulary help",
			shouldPass:  false,
			testDesc:    "Should reject description containing email addresses",
		},
		{
			name:        "Valid educational description",
			description: "A comprehensive collection of Spanish vocabulary for beginners",
			shouldPass:  true,
			testDesc:    "Should accept valid educational description",
		},
		{
			name:        "Empty description",
			description: "",
			shouldPass:  true,
			testDesc:    "Should accept empty description",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			resp := server.Expect.POST("/wordlists").
				WithHeader("Authorization", "Bearer "+token).
				WithJSON(map[string]interface{}{
					"name":         "Test Wordlist",
					"description":  tc.description,
					"languageCode": "en",
				}).
				Expect()

			// Assert
			if tc.shouldPass {
				resp.Status(http.StatusCreated)
				respBody := resp.JSON().Object()
				respBody.ContainsKey("id")
			} else {
				resp.Status(http.StatusBadRequest)
				respBody := resp.JSON().Object()
				respBody.ContainsKey("error")
				err := respBody.Value("error").String().Raw()
				assert.Contains(t, err, "not appropriate", tc.testDesc)
			}
		})
	}
}

func TestContentFilteringWord_Create_ShouldRejectInappropriateWords(t *testing.T) {
	// Arrange
	cleanup := setupMockModeration(t)
	defer cleanup()

	server := setup.NewTestServer(t)
	defer server.Cleanup()

	user := createTestUser(t, server)
	token := user["token"].(string)

	// Create a test wordlist first
	wordlistResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", "Bearer "+token).
		WithJSON(map[string]interface{}{
			"name":         "Test Wordlist",
			"description":  "For testing",
			"languageCode": "en",
		}).
		Expect().Status(http.StatusCreated)

	wordlistID := wordlistResp.JSON().Object().Value("id").Number().Raw()

	testCases := []struct {
		name       string
		word       string
		notes      string
		shouldPass bool
		testDesc   string
	}{
		{
			name:       "Profanity word",
			word:       "damn",
			notes:      "",
			shouldPass: false,
			testDesc:   "Should reject profane words",
		},
		{
			name:       "Adult content word",
			word:       "xxx",
			notes:      "",
			shouldPass: false,
			testDesc:   "Should reject adult content words",
		},
		{
			name:       "Violence word",
			word:       "kill",
			notes:      "",
			shouldPass: false,
			testDesc:   "Should reject violence-related words",
		},
		{
			name:       "Hate speech word",
			word:       "racist",
			notes:      "",
			shouldPass: false,
			testDesc:   "Should reject hate speech words",
		},
		{
			name:       "URL as word",
			word:       "https://spam.com",
			notes:      "",
			shouldPass: false,
			testDesc:   "Should reject URLs as words",
		},
		{
			name:       "Email as word",
			word:       "spam@test.com",
			notes:      "",
			shouldPass: false,
			testDesc:   "Should reject email addresses as words",
		},
		{
			name:       "Phone number as word",
			word:       "555-123-4567",
			notes:      "",
			shouldPass: false,
			testDesc:   "Should reject phone numbers as words",
		},
		{
			name:       "Long number sequence",
			word:       "123456789",
			notes:      "",
			shouldPass: false,
			testDesc:   "Should reject long number sequences",
		},
		{
			name:       "Word too long",
			word:       "supercalifragilisticexpialidocious",
			notes:      "",
			shouldPass: false,
			testDesc:   "Should reject words that are too long",
		},
		{
			name:       "Valid word with inappropriate notes",
			word:       "hello",
			notes:      "This word is damn useful for idiots",
			shouldPass: false,
			testDesc:   "Should reject words with inappropriate notes",
		},
		{
			name:       "Spanish profanity",
			word:       "estúpido",
			notes:      "",
			shouldPass: false,
			testDesc:   "Should reject Spanish profanity",
		},
		{
			name:       "French profanity",
			word:       "stupide",
			notes:      "",
			shouldPass: false,
			testDesc:   "Should reject French profanity",
		},
		{
			name:       "Valid educational word",
			word:       "hello",
			notes:      "A common greeting",
			shouldPass: true,
			testDesc:   "Should accept valid educational words",
		},
		{
			name:       "Valid word without notes",
			word:       "world",
			notes:      "",
			shouldPass: true,
			testDesc:   "Should accept valid words without notes",
		},
		{
			name:       "Valid accented word",
			word:       "café",
			notes:      "French word for coffee",
			shouldPass: true,
			testDesc:   "Should accept valid words with accents",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			resp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
				WithHeader("Authorization", "Bearer "+token).
				WithJSON(map[string]interface{}{
					"name":  tc.word,
					"notes": tc.notes,
				}).
				Expect()

			// Assert
			if tc.shouldPass {
				resp.Status(http.StatusCreated)
				respBody := resp.JSON().Object()
				respBody.ContainsKey("id")
				respBody.Value("name").String().IsEqual(tc.word)
			} else {
				resp.Status(http.StatusBadRequest)
				respBody := resp.JSON().Object()
				respBody.ContainsKey("error")
				err := respBody.Value("error").String().Raw()
				assert.Contains(t, err, "not appropriate", tc.testDesc)
			}
		})
	}
}

func TestContentFilteringMultiLanguage_FreeUser_ShouldDetectInappropriateContent(t *testing.T) {
	// Arrange
	cleanup := setupMockModeration(t)
	defer cleanup()

	server := setup.NewTestServer(t)
	defer server.Cleanup()

	user := createTestUser(t, server)
	token := user["token"].(string)

	// Create one wordlist to use for all language tests (free plan limit)
	wordlistResp := server.Expect.POST("/wordlists").
		WithHeader("Authorization", "Bearer "+token).
		WithJSON(map[string]interface{}{
			"name":         "Multi-Language Test",
			"description":  "Testing multi-language filtering",
			"languageCode": "en",
		}).
		Expect().Status(http.StatusCreated)

	wordlistID := wordlistResp.JSON().Object().Value("id").Number().Raw()

	// Test inappropriate words from different languages - should be rejected
	inappropriateWords := []struct {
		word     string
		language string
	}{
		{"estúpido", "Spanish"},
		{"stupide", "French"},
		{"dumm", "German"},
		{"idiota", "Italian"},
		{"maldito", "Portuguese"},
	}

	for _, test := range inappropriateWords {
		t.Run("reject_"+test.word+"_"+test.language, func(t *testing.T) {
			resp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
				WithHeader("Authorization", "Bearer "+token).
				WithJSON(map[string]interface{}{
					"name":  test.word,
					"notes": "",
				}).
				Expect()

			resp.Status(http.StatusBadRequest)
			respBody := resp.JSON().Object()
			respBody.ContainsKey("error")
			err := respBody.Value("error").String().Raw()
			assert.Contains(t, err, "not appropriate", "Should reject "+test.word+" ("+test.language+")")
		})
	}

	// Test appropriate words from different languages - should be accepted (within free limit)
	appropriateWords := []struct {
		word     string
		language string
	}{
		{"hola", "Spanish"},
		{"bonjour", "French"},
		{"hallo", "German"},
		{"ciao", "Italian"},
		{"olá", "Portuguese"},
	}

	for i, test := range appropriateWords {
		if i >= 5 { // Free plan has 10 word limit, save some for other tests
			break
		}
		t.Run("accept_"+test.word+"_"+test.language, func(_ *testing.T) {
			resp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
				WithHeader("Authorization", "Bearer "+token).
				WithJSON(map[string]interface{}{
					"name":  test.word,
					"notes": "",
				}).
				Expect()

			resp.Status(http.StatusCreated)
			respBody := resp.JSON().Object()
			respBody.ContainsKey("id")
			respBody.Value("name").String().IsEqual(test.word)
		})
	}
}

func TestContentFilteringMultiLanguage_PremiumUser_ShouldDetectInappropriateContent(t *testing.T) {
	// Arrange
	cleanup := setupMockModeration(t)
	defer cleanup()

	server := setup.NewTestServer(t)
	defer server.Cleanup()

	user := createPremiumTestUser(t, server)
	token := user["token"].(string)

	multiLanguageTests := []struct {
		language  string
		badWords  []string
		goodWords []string
		testDesc  string
	}{
		{
			language:  "es",
			badWords:  []string{"estúpido", "idiota", "maldito"},
			goodWords: []string{"hola", "mundo", "vocabulario"},
			testDesc:  "Should filter Spanish inappropriate content",
		},
		{
			language:  "fr",
			badWords:  []string{"stupide", "idiot", "maudit"},
			goodWords: []string{"bonjour", "monde", "vocabulaire"},
			testDesc:  "Should filter French inappropriate content",
		},
		{
			language:  "de",
			badWords:  []string{"dumm", "blöd", "verdammt"},
			goodWords: []string{"hallo", "welt", "vokabular"},
			testDesc:  "Should filter German inappropriate content",
		},
		{
			language:  "it",
			badWords:  []string{"stupido", "idiota", "maledetto"},
			goodWords: []string{"ciao", "mondo", "vocabolario"},
			testDesc:  "Should filter Italian inappropriate content",
		},
		{
			language:  "pt",
			badWords:  []string{"estúpido", "idiota", "maldito"},
			goodWords: []string{"olá", "mundo", "vocabulário"},
			testDesc:  "Should filter Portuguese inappropriate content",
		},
	}

	for _, langTest := range multiLanguageTests {
		t.Run(langTest.language, func(t *testing.T) {
			// Create separate wordlist for each language (premium has unlimited wordlists)
			wordlistResp := server.Expect.POST("/wordlists").
				WithHeader("Authorization", "Bearer "+token).
				WithJSON(map[string]interface{}{
					"name":         "Test " + langTest.language,
					"description":  "Testing " + langTest.language + " filtering",
					"languageCode": langTest.language,
				}).
				Expect().Status(http.StatusCreated)

			wordlistID := wordlistResp.JSON().Object().Value("id").Number().Raw()

			// Test bad words should be rejected
			for _, badWord := range langTest.badWords {
				t.Run("reject_"+badWord, func(t *testing.T) {
					resp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
						WithHeader("Authorization", "Bearer "+token).
						WithJSON(map[string]interface{}{
							"name":  badWord,
							"notes": "",
						}).
						Expect()

					resp.Status(http.StatusBadRequest)
					respBody := resp.JSON().Object()
					respBody.ContainsKey("error")
					err := respBody.Value("error").String().Raw()
					assert.Contains(t, err, "not appropriate", "Should reject "+badWord+" in "+langTest.language)
				})
			}

			// Test good words should be accepted
			for _, goodWord := range langTest.goodWords {
				t.Run("accept_"+goodWord, func(_ *testing.T) {
					resp := server.Expect.POST("/wordlists/{wordlistId}/words", wordlistID).
						WithHeader("Authorization", "Bearer "+token).
						WithJSON(map[string]interface{}{
							"name":  goodWord,
							"notes": "",
						}).
						Expect()

					resp.Status(http.StatusCreated)
					respBody := resp.JSON().Object()
					respBody.ContainsKey("id")
					respBody.Value("name").String().IsEqual(goodWord)
				})
			}
		})
	}
}
