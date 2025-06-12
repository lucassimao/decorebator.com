package setup

import (
	"fmt"
	"math/rand" // nosec G404 - math/rand is acceptable for test fixtures, not cryptographic use
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
)

// GenerateTestUser generates a random test user
func GenerateTestUser() map[string]interface{} {
	fake := gofakeit.New(0)

	return map[string]interface{}{
		"email":     fake.Email(),
		"password":  "password123",
		"firstName": fake.FirstName(),
		"lastName":  fake.LastName(),
	}
}

// GenerateTestWordlist generates a random test wordlist
func GenerateTestWordlist() map[string]interface{} {
	fake := gofakeit.New(0)

	languages := []string{"en", "es", "fr", "de", "it", "pt", "ja"}

	return map[string]interface{}{
		"name":        fake.Sentence(3),
		"language":    languages[rand.Intn(len(languages))], //nolint:gosec // G404 - test fixtures only
		"description": fake.Sentence(10),
	}
}

// GenerateTestWord generates a random test word
func GenerateTestWord() map[string]interface{} {
	fake := gofakeit.New(0)

	return map[string]interface{}{
		"name":  fake.Word(),
		"notes": fake.Sentence(5),
	}
}

// GenerateTestDefinition generates a random test definition
func GenerateTestDefinition() map[string]interface{} {
	fake := gofakeit.New(0)

	partsOfSpeech := []string{"noun", "verb", "adjective", "adverb", "preposition"}

	return map[string]interface{}{
		"meaning":      fake.Sentence(8),
		"partOfSpeech": partsOfSpeech[rand.Intn(len(partsOfSpeech))], //nolint:gosec // G404 - test fixtures only
		"examples":     []string{fake.Sentence(6), fake.Sentence(7)},
		"source":       "test",
	}
}

// GenerateJWTToken generates a test JWT token
func GenerateJWTToken(userID int64) string {
	// This is a simplified version for testing
	// In real implementation, use proper JWT signing
	return fmt.Sprintf("test-token-%d-%s", userID, uuid.New().String()[:8])
}

// TestUserCredentials represents test user credentials
type TestUserCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TestUser represents a test user
type TestUser struct {
	ID               int64  `json:"id"`
	Email            string `json:"email"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	SubscriptionPlan string `json:"subscriptionPlan"`
	Token            string `json:"token,omitempty"`
}

// TestWordlist represents a test wordlist
type TestWordlist struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Language    string `json:"language"`
	Description string `json:"description,omitempty"`
	UserID      int64  `json:"userId"`
	WordCount   int    `json:"wordCount"`
}

// TestWord represents a test word
type TestWord struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Notes      string `json:"notes,omitempty"`
	WordlistID int64  `json:"wordlistId"`
	AudioURL   string `json:"audioUrl,omitempty"`
}

// TestDefinition represents a test definition
type TestDefinition struct {
	ID           int64    `json:"id"`
	WordID       int64    `json:"wordId"`
	Meaning      string   `json:"meaning"`
	PartOfSpeech string   `json:"partOfSpeech"`
	Examples     []string `json:"examples"`
	Source       string   `json:"source"`
}

// CreateTestUserSet creates a set of test users with different subscription levels
func CreateTestUserSet() []TestUser {
	return []TestUser{
		{
			Email:            "free@test.com",
			FirstName:        "Free",
			LastName:         "User",
			SubscriptionPlan: "free",
		},
		{
			Email:            "monthly@test.com",
			FirstName:        "Monthly",
			LastName:         "User",
			SubscriptionPlan: "monthly",
		},
		{
			Email:            "annual@test.com",
			FirstName:        "Annual",
			LastName:         "User",
			SubscriptionPlan: "annual",
		},
	}
}

// CreateTestWordlistSet creates a set of test wordlists
func CreateTestWordlistSet(userID int64) []TestWordlist {
	return []TestWordlist{
		{
			Name:        "Travel Essentials",
			Language:    "en",
			Description: "Essential words for traveling",
			UserID:      userID,
		},
		{
			Name:        "Business Vocabulary",
			Language:    "en",
			Description: "Professional business terms",
			UserID:      userID,
		},
		{
			Name:        "Vocabulario Básico",
			Language:    "es",
			Description: "Palabras básicas en español",
			UserID:      userID,
		},
	}
}

// CreateTestWordSet creates a set of test words
func CreateTestWordSet(wordlistID int64) []TestWord {
	return []TestWord{
		{
			Name:       "hello",
			Notes:      "Common greeting",
			WordlistID: wordlistID,
		},
		{
			Name:       "goodbye",
			Notes:      "Parting phrase",
			WordlistID: wordlistID,
		},
		{
			Name:       "please",
			Notes:      "Polite request word",
			WordlistID: wordlistID,
		},
		{
			Name:       "thank you",
			Notes:      "Expression of gratitude",
			WordlistID: wordlistID,
		},
		{
			Name:       "excuse me",
			Notes:      "Polite interruption phrase",
			WordlistID: wordlistID,
		},
	}
}

// MockStripeEvent generates a mock Stripe webhook event
func MockStripeEvent(eventType string, data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"id":      fmt.Sprintf("evt_%s", uuid.New().String()[:16]),
		"object":  "event",
		"type":    eventType,
		"created": time.Now().Unix(),
		"data": map[string]interface{}{
			"object": data,
		},
		"livemode":         false,
		"pending_webhooks": 1,
		"api_version":      "2020-08-27",
	}
}

// MockOpenAIResponse generates a mock OpenAI API response
func MockOpenAIResponse(responseType string) map[string]interface{} {
	switch responseType {
	case "chat_completion":
		return map[string]interface{}{
			"id":      fmt.Sprintf("chatcmpl-%s", uuid.New().String()[:16]),
			"object":  "chat.completion",
			"created": time.Now().Unix(),
			"model":   "gpt-4",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": `{"meaning": "A greeting used when meeting someone", "partOfSpeech": "interjection", "examples": ["Hello, how are you?", "She said hello to her neighbor"]}`,
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     50,
				"completion_tokens": 25,
				"total_tokens":      75,
			},
		}
	case "image_generation":
		return map[string]interface{}{
			"created": time.Now().Unix(),
			"data": []map[string]interface{}{
				{
					"url": fmt.Sprintf("https://test-images.example.com/%s.png", uuid.New().String()),
				},
			},
		}
	case "tts":
		return map[string]interface{}{
			"audio_url": fmt.Sprintf("https://test-audio.example.com/%s.mp3", uuid.New().String()),
		}
	default:
		return map[string]interface{}{
			"error": "Unknown response type",
		}
	}
}
