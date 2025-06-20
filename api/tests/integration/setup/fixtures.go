package setup

import (
	"fmt"
	"math/rand" // nosec G404 - math/rand is acceptable for test fixtures, not cryptographic use
	"time"

	httphandlers "decorebator.com/internal/http"
	"decorebator.com/internal/model"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// GenerateTestUser generates a random test user using production model
func GenerateTestUser() *model.User {
	fake := gofakeit.New(uint64(time.Now().UnixNano())) // Use time-based seed for unique data

	// Generate a realistic bcrypt hash for testing
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	return &model.User{
		FirstName:        fake.FirstName(),
		LastName:         fake.LastName(),
		Email:            fake.Email(),
		PasswordHash:     string(passwordHash),
		SubscriptionPlan: model.PlanFree,
		// pgtype fields will be set by database operations
	}
}

// GenerateSignupInput generates signup input using production SignupInput struct
func GenerateSignupInput() httphandlers.SignupInput {
	fake := gofakeit.New(uint64(time.Now().UnixNano())) // Use time-based seed for unique data

	return httphandlers.SignupInput{
		Email:     fake.Email(),
		Password:  "password123",
		FirstName: fake.FirstName(),
		LastName:  fake.LastName(),
	}
}

// GenerateTestWordlist generates a random test wordlist using production model
func GenerateTestWordlist(userID int64) *model.Wordlist {
	fake := gofakeit.New(uint64(time.Now().UnixNano())) // Use time-based seed for unique data

	languages := []string{"en", "es", "fr", "de", "it", "pt", "ja"}

	return &model.Wordlist{
		Name:         fake.Sentence(3),
		Description:  fake.Sentence(10),
		UserID:       userID,
		LanguageCode: languages[rand.Intn(len(languages))], //nolint:gosec // G404 - test fixtures only
		// pgtype fields will be set by database operations
	}
}

// GenerateTestWord generates a random test word using production model
func GenerateTestWord(wordlistID, userID int64) *model.Word {
	fake := gofakeit.New(uint64(time.Now().UnixNano())) // Use time-based seed for unique data

	return &model.Word{
		Name:       fake.Word(),
		Notes:      fake.Sentence(5),
		WordlistID: wordlistID,
		UserID:     userID,
		Learned:    false,
		// pgtype fields will be set by database operations
	}
}

// GenerateTestDefinition generates a random test definition using production model  
func GenerateTestDefinition() *model.Definition {
	fake := gofakeit.New(uint64(time.Now().UnixNano())) // Use time-based seed for unique data

	partsOfSpeech := []string{"noun", "verb", "adjective", "adverb", "preposition"}

	return &model.Definition{
		Token:        fake.Word(),
		Language:     "en",
		Meaning:      fake.Sentence(8),
		PartOfSpeech: partsOfSpeech[rand.Intn(len(partsOfSpeech))], //nolint:gosec // G404 - test fixtures only
		Examples:     []string{fake.Sentence(6), fake.Sentence(7)},
		Source:       model.ChatGPT,
		// pgtype fields will be set by database operations
	}
}

// GenerateJWTToken generates a test JWT token
func GenerateJWTToken(userID int64) string {
	// This is a simplified version for testing
	// In real implementation, use proper JWT signing
	return fmt.Sprintf("test-token-%d-%s", userID, uuid.New().String()[:8])
}

// CreateTestUserSet creates a set of test users with different subscription levels
func CreateTestUserSet() []*model.User {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	return []*model.User{
		{
			Email:            "free@test.com",
			FirstName:        "Free",
			LastName:         "User",
			PasswordHash:     string(passwordHash),
			SubscriptionPlan: model.PlanFree,
		},
		{
			Email:            "monthly@test.com",
			FirstName:        "Monthly",
			LastName:         "User",
			PasswordHash:     string(passwordHash),
			SubscriptionPlan: model.PlanMonthly,
		},
		{
			Email:            "annual@test.com",
			FirstName:        "Annual",
			LastName:         "User",
			PasswordHash:     string(passwordHash),
			SubscriptionPlan: model.PlanAnnual,
		},
	}
}

// CreateTestWordlistSet creates a set of test wordlists
func CreateTestWordlistSet(userID int64) []*model.Wordlist {
	return []*model.Wordlist{
		{
			Name:         "Travel Essentials",
			LanguageCode: "en",
			Description:  "Essential words for traveling",
			UserID:       userID,
		},
		{
			Name:         "Business Vocabulary",
			LanguageCode: "en",
			Description:  "Professional business terms",
			UserID:       userID,
		},
		{
			Name:         "Vocabulario Básico",
			LanguageCode: "es",
			Description:  "Palabras básicas en español",
			UserID:       userID,
		},
	}
}

// CreateTestWordSet creates a set of test words
func CreateTestWordSet(wordlistID, userID int64) []*model.Word {
	return []*model.Word{
		{
			Name:       "hello",
			Notes:      "Common greeting",
			WordlistID: wordlistID,
			UserID:     userID,
			Learned:    false,
		},
		{
			Name:       "goodbye",
			Notes:      "Parting phrase",
			WordlistID: wordlistID,
			UserID:     userID,
			Learned:    false,
		},
		{
			Name:       "please",
			Notes:      "Polite request word",
			WordlistID: wordlistID,
			UserID:     userID,
			Learned:    false,
		},
		{
			Name:       "thank you",
			Notes:      "Expression of gratitude",
			WordlistID: wordlistID,
			UserID:     userID,
			Learned:    false,
		},
		{
			Name:       "excuse me",
			Notes:      "Polite interruption phrase",
			WordlistID: wordlistID,
			UserID:     userID,
			Learned:    false,
		},
	}
}
