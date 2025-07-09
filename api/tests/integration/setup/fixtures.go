package setup

import (
	"crypto/rand" // For secure random numbers
	"fmt"
	"math/big"
	mathrand "math/rand" // nosec G404 - math/rand is acceptable for test fixtures, not cryptographic use
	"time"

	httphandlers "decorebator.com/internal/http"
	"decorebator.com/internal/model"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// GenerateUniqueEventID creates a unique event ID for tests
func GenerateUniqueEventID(prefix string) string {
	// Use nanoseconds for higher precision
	timestamp := time.Now().UnixNano()

	// Add crypto random number for uniqueness
	randomNum, err := rand.Int(rand.Reader, big.NewInt(999999))
	if err != nil {
		// Fallback to timestamp only if crypto fails
		return fmt.Sprintf("%s_%d", prefix, timestamp)
	}

	return fmt.Sprintf("%s_%d_%d", prefix, timestamp, randomNum.Int64())
}

// GenerateTestUser generates a random test user using production model
func GenerateTestUser() *model.User {
	// Use UnixNano for better uniqueness across rapid calls
	now := time.Now().UnixNano()
	var seed uint64
	if now >= 0 {
		seed = uint64(now)
	} else {
		seed = uint64(-now) // Handle negative values safely
	}
	fake := gofakeit.New(seed)

	// Generate a realistic bcrypt hash for testing
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// Add UUID to email to ensure uniqueness
	uuid := uuid.New().String()[:8] // Use first 8 chars of UUID
	email := fmt.Sprintf("test+%s@example.com", uuid)

	return &model.User{
		FirstName:        fake.FirstName(),
		LastName:         fake.LastName(),
		Email:            email,
		PasswordHash:     string(passwordHash),
		SubscriptionPlan: model.PlanFree,
		// pgtype fields will be set by database operations
	}
}

// GenerateSignupInput generates signup input using production SignupInput struct
func GenerateSignupInput() httphandlers.SignupInput {
	// Use UnixNano for better uniqueness across rapid calls
	now := time.Now().UnixNano()
	var seed uint64
	if now >= 0 {
		seed = uint64(now)
	} else {
		seed = uint64(-now) // Handle negative values safely
	}
	fake := gofakeit.New(seed)

	// Add UUID to email to ensure uniqueness
	uuid := uuid.New().String()[:8] // Use first 8 chars of UUID
	email := fmt.Sprintf("test+%s@example.com", uuid)

	return httphandlers.SignupInput{
		Email:     email,
		Password:  "password123",
		FirstName: fake.FirstName(),
		LastName:  fake.LastName(),
	}
}

// GenerateTestWordlist generates a random test wordlist using production model
func GenerateTestWordlist(userID int64) *model.Wordlist {
	// Use Unix timestamp (seconds) with safe conversion to avoid overflow
	now := time.Now().Unix()
	var seed uint64
	if now >= 0 {
		seed = uint64(now)
	} else {
		seed = uint64(-now) // Handle negative values safely
	}
	fake := gofakeit.New(seed)

	languages := []string{"en", "es", "fr", "de", "it", "pt", "ja"}
	languageCode := languages[mathrand.Intn(len(languages))] //nolint:gosec // G404 - test fixtures only

	return &model.Wordlist{
		Name:                fake.Sentence(3),
		Description:         fake.Sentence(10),
		UserID:              userID,
		LanguageCode:        languageCode,
		PronunciationSystem: model.GetDefaultPronunciationSystem(languageCode),
		// pgtype fields will be set by database operations
	}
}

// GenerateTestWord generates a random test word using production model
func GenerateTestWord(wordlistID, userID int64) *model.Word {
	// Use Unix timestamp (seconds) with safe conversion to avoid overflow
	now := time.Now().Unix()
	var seed uint64
	if now >= 0 {
		seed = uint64(now)
	} else {
		seed = uint64(-now) // Handle negative values safely
	}
	fake := gofakeit.New(seed)

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
	// Use Unix timestamp (seconds) with safe conversion to avoid overflow
	now := time.Now().Unix()
	var seed uint64
	if now >= 0 {
		seed = uint64(now)
	} else {
		seed = uint64(-now) // Handle negative values safely
	}
	fake := gofakeit.New(seed)

	partsOfSpeech := []string{"noun", "verb", "adjective", "adverb", "preposition"}

	return &model.Definition{
		Token:        fake.Word(),
		Language:     "en",
		Meaning:      fake.Sentence(8),
		PartOfSpeech: partsOfSpeech[mathrand.Intn(len(partsOfSpeech))], //nolint:gosec // G404 - test fixtures only
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
			Name:                "Travel Essentials",
			LanguageCode:        "en",
			Description:         "Essential words for traveling",
			UserID:              userID,
			PronunciationSystem: model.PronunciationSystemIPA,
		},
		{
			Name:                "Business Vocabulary",
			LanguageCode:        "en",
			Description:         "Professional business terms",
			UserID:              userID,
			PronunciationSystem: model.PronunciationSystemIPA,
		},
		{
			Name:                "Vocabulario Básico",
			LanguageCode:        "es",
			Description:         "Palabras básicas en español",
			UserID:              userID,
			PronunciationSystem: model.PronunciationSystemIPA,
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

// Error Reporting Specific Fixtures

// CreateErrorReportDefinition creates a definition optimized for error reporting tests
func CreateErrorReportDefinition(token, language, partOfSpeech string) *model.Definition {
	// Use Unix timestamp for variety in content
	now := time.Now().Unix()
	var seed uint64
	if now >= 0 {
		seed = uint64(now)
	} else {
		seed = uint64(-now)
	}
	fake := gofakeit.New(seed)

	// Create language-specific content
	var meaning string
	var examples []string

	switch language {
	case "en":
		meaning = fake.Sentence(8)
		examples = []string{
			fmt.Sprintf("Example with [%s] in context", token),
			fmt.Sprintf("Another [%s] usage example", token),
		}
	case "es":
		meaning = "Una definición en español para pruebas"
		examples = []string{
			fmt.Sprintf("Ejemplo con [%s] en contexto", token),
			fmt.Sprintf("Otro ejemplo de [%s] en uso", token),
		}
	case "de":
		meaning = "Eine deutsche Definition für Tests"
		examples = []string{
			fmt.Sprintf("Beispiel mit [%s] im Kontext", token),
			fmt.Sprintf("Ein weiteres [%s] Verwendungsbeispiel", token),
		}
	case "fr":
		meaning = "Une définition française pour les tests"
		examples = []string{
			fmt.Sprintf("Example avec [%s] en contexte", token),
			fmt.Sprintf("Un autre example d'utilisation de [%s]", token),
		}
	case "it":
		meaning = "Una definizione italiana per i test"
		examples = []string{
			fmt.Sprintf("Esempio con [%s] nel contesto", token),
			fmt.Sprintf("Un altro esempio di utilizzo di [%s]", token),
		}
	case "pt":
		meaning = "Uma definição portuguesa para testes"
		examples = []string{
			fmt.Sprintf("Exemplo com [%s] no contexto", token),
			fmt.Sprintf("Outro exemplo de uso de [%s]", token),
		}
	case "ja":
		meaning = "テスト用の日本語定義"
		examples = []string{
			fmt.Sprintf("[%s]を使った例文", token),
			fmt.Sprintf("[%s]の別の使用例", token),
		}
	default:
		meaning = fake.Sentence(8)
		examples = []string{
			fmt.Sprintf("Example with [%s] in context", token),
			fmt.Sprintf("Another [%s] usage example", token),
		}
	}

	return &model.Definition{
		Token:        token,
		Language:     language,
		PartOfSpeech: partOfSpeech,
		Meaning:      meaning,
		Examples:     examples,
		Source:       model.ChatGPT,
	}
}

// CreateMultiLanguageDefinitions creates definitions for common error testing scenarios
func CreateMultiLanguageDefinitions() map[string]*model.Definition {
	return map[string]*model.Definition{
		"en_noun": CreateErrorReportDefinition("water", "en", "noun"),
		"en_verb": CreateErrorReportDefinition("run", "en", "verb"),
		"es_noun": CreateErrorReportDefinition("agua", "es", "sustantivo"),
		"de_noun": CreateErrorReportDefinition("Hund", "de", "Substantiv"),
		"fr_noun": CreateErrorReportDefinition("chat", "fr", "nom"),
		"it_noun": CreateErrorReportDefinition("gatto", "it", "sostantivo"),
		"pt_noun": CreateErrorReportDefinition("casa", "pt", "substantivo"),
		"ja_noun": CreateErrorReportDefinition("猫", "ja", "名詞"),
	}
}

// CreateErrorReportPayload creates standardized error report JSON payload
func CreateErrorReportPayload(wordID, definitionID int64, errorType string, quizDetails map[string]interface{}) map[string]interface{} {
	payload := map[string]interface{}{
		"wordId":    wordID,
		"errorType": errorType,
	}

	// Add definitionId only if it's not zero (some error types don't require it)
	if definitionID != 0 {
		payload["definitionId"] = definitionID
	}

	// Add quiz details if provided
	if quizDetails != nil {
		payload["quizDetails"] = quizDetails
	}

	return payload
}

// CreateQuizDetails creates standardized quiz context for error reports
func CreateQuizDetails(quizType, value, context string, options []string, answerIndex int) map[string]interface{} {
	return map[string]interface{}{
		"quizType":    quizType,
		"value":       value,
		"options":     options,
		"answerIndex": answerIndex,
		"context":     context,
	}
}
