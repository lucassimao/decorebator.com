package service

import "strings"

// MockModerationService is a test implementation that simulates OpenAI moderation responses
type MockModerationService struct {
	// flaggedContent maps content to rejection reasons for testing
	flaggedContent map[string]string
}

var _ ModerationService = (*MockModerationService)(nil)

// NewMockModerationService creates a new mock moderation service for testing
func NewMockModerationService() *MockModerationService {
	return &MockModerationService{
		flaggedContent: map[string]string{
			// Profanity content
			"damn":       "Content contains harassment or threatening behavior",
			"damn words": "Content contains harassment or threatening behavior",
			"This is a damn good vocabulary list with stupid words": "Content contains harassment or threatening behavior",
			"This word is damn useful for idiots":                   "Content contains harassment or threatening behavior",

			// Adult/sexual content
			"xxx":            "Content contains sexual or adult material not appropriate for educational use",
			"xxx vocabulary": "Content contains sexual or adult material not appropriate for educational use",
			"Adult vocabulary with xxx content for mature learners": "Content contains sexual or adult material not appropriate for educational use",

			// Violence content
			"kill":      "Content contains violent or graphic material",
			"kill list": "Content contains violent or graphic material",
			"Learn words about violence and murder for advanced students": "Content contains violent or graphic material",

			// Hate speech
			"racist": "Content contains hate speech or threatening language",

			// Spam/harassment patterns (URLs, emails, phone numbers)
			"https://spam.com":       "Content contains harassment or threatening behavior",
			"visit https://spam.com": "Content contains harassment or threatening behavior",
			"Visit https://spam.com for more vocabulary resources": "Content contains harassment or threatening behavior",
			"spam@test.com":         "Content contains harassment or threatening behavior",
			"contact spam@test.com": "Content contains harassment or threatening behavior",
			"Contact me at spam@test.com for vocabulary help": "Content contains harassment or threatening behavior",
			"555-123-4567":      "Content contains harassment or threatening behavior",
			"call 555-123-4567": "Content contains harassment or threatening behavior",

			// Repetitive/spam content
			"aaaaaaaaaaa spam spam spam": "Content contains harassment or threatening behavior",
			"SPAMMY WORDLIST NAME":       "Content contains harassment or threatening behavior",

			// Multi-language profanity
			"estúpido":       "Content contains harassment or threatening behavior",
			"lista estúpida": "Content contains harassment or threatening behavior",
			"idiota":         "Content contains harassment or threatening behavior",
			"maldito":        "Content contains harassment or threatening behavior",
			"stupide":        "Content contains harassment or threatening behavior",
			"liste stupide":  "Content contains harassment or threatening behavior",
			"idiot":          "Content contains harassment or threatening behavior",
			"maudit":         "Content contains harassment or threatening behavior",
			"dumm":           "Content contains harassment or threatening behavior",
			"blöd":           "Content contains harassment or threatening behavior",
			"verdammt":       "Content contains harassment or threatening behavior",
			"stupido":        "Content contains harassment or threatening behavior",
			"maledetto":      "Content contains harassment or threatening behavior",

			// Long sequences/invalid content
			"123456789":    "Content contains harassment or threatening behavior",
			"verylongword": "Content contains harassment or threatening behavior",
			"spam.com":     "Content contains harassment or threatening behavior",
		},
	}
}

// Validate checks if content is appropriate using predefined test rules
func (m *MockModerationService) Validate(text string) ContentFilterResult {
	// Basic validation first
	if text == "" {
		return ContentFilterResult{
			IsAppropriate: false,
			Reason:        "Content cannot be empty",
			FlaggedWords:  []string{},
		}
	}

	if len(text) > 2000 {
		return ContentFilterResult{
			IsAppropriate: false,
			Reason:        "Content is too long (maximum 2000 characters)",
			FlaggedWords:  []string{},
		}
	}

	// Check against flagged content
	if reason, flagged := m.flaggedContent[text]; flagged {
		return ContentFilterResult{
			IsAppropriate: false,
			Reason:        reason,
			FlaggedWords:  []string{}, // OpenAI doesn't provide specific flagged words
		}
	}

	// Check for case-insensitive matches for some patterns
	lowerText := strings.ToLower(text)
	for content, reason := range m.flaggedContent {
		if strings.ToLower(content) == lowerText {
			return ContentFilterResult{
				IsAppropriate: false,
				Reason:        reason,
				FlaggedWords:  []string{},
			}
		}
	}

	// Content is appropriate
	return ContentFilterResult{
		IsAppropriate: true,
		Reason:        "",
		FlaggedWords:  []string{},
	}
}

// AddFlaggedContent allows tests to add custom flagged content
func (m *MockModerationService) AddFlaggedContent(content, reason string) {
	m.flaggedContent[content] = reason
}

// ClearFlaggedContent clears all flagged content (useful for test isolation)
func (m *MockModerationService) ClearFlaggedContent() {
	m.flaggedContent = make(map[string]string)
}
