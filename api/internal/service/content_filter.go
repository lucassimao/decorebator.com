package service

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ContentSeverity represents the severity level of content violations
type ContentSeverity string

const (
	SeverityLow    ContentSeverity = "low"
	SeverityMedium ContentSeverity = "medium"
	SeverityHigh   ContentSeverity = "high"
)

// ContentFilterResult represents the result of content filtering
type ContentFilterResult struct {
	IsAppropriate bool
	Severity      ContentSeverity
	Reason        string
	FlaggedWords  []string
}

// ContentFilterService handles content moderation for user-generated content
type ContentFilterService struct {
	profanityLists map[string][]string
	patterns       map[string]*regexp.Regexp
}

// NewContentFilterService creates a new content filter service
func NewContentFilterService() *ContentFilterService {
	service := &ContentFilterService{
		profanityLists: make(map[string][]string),
		patterns:       make(map[string]*regexp.Regexp),
	}

	service.initializeProfanityLists()
	service.initializePatterns()

	return service
}

// initializeProfanityLists sets up profanity word lists for different languages
func (cfs *ContentFilterService) initializeProfanityLists() {
	// English profanity and inappropriate content
	cfs.profanityLists["en"] = []string{
		// Common profanity (keeping it educational-appropriate)
		"damn", "hell", "crap", "stupid", "idiot", "moron", "dumb",
		// Adult content indicators
		"sex", "porn", "xxx", "adult", "naked", "nude",
		// Violence indicators
		"kill", "murder", "violence", "weapon", "gun", "knife",
		// Hate speech indicators
		"hate", "racist", "nazi", "terrorist",
		// Spam indicators
		"free money", "click here", "buy now", "limited time",
	}

	// Spanish
	cfs.profanityLists["es"] = []string{
		"tonto", "estúpido", "idiota", "maldito",
		"sexo", "adulto", "desnudo",
		"matar", "violencia", "arma",
		"odio", "racista",
	}

	// French
	cfs.profanityLists["fr"] = []string{
		"stupide", "idiot", "maudit", "bête",
		"sexe", "adulte", "nu",
		"tuer", "violence", "arme",
		"haine", "raciste",
	}

	// German
	cfs.profanityLists["de"] = []string{
		"dumm", "blöd", "idiot", "verdammt",
		"sex", "erwachsene", "nackt",
		"töten", "gewalt", "waffe",
		"hass", "rassist",
	}

	// Italian
	cfs.profanityLists["it"] = []string{
		"stupido", "idiota", "maledetto", "scemo",
		"sesso", "adulto", "nudo",
		"uccidere", "violenza", "arma",
		"odio", "razzista",
	}

	// Portuguese
	cfs.profanityLists["pt"] = []string{
		"estúpido", "idiota", "maldito", "burro",
		"sexo", "adulto", "nu",
		"matar", "violência", "arma",
		"ódio", "racista",
	}

	// Japanese (romanized)
	cfs.profanityLists["ja"] = []string{
		"baka", "aho", "kuso",
		"ecchi", "hentai", "sekkusu",
		"korosu", "boryoku",
		"kirai", "sabetsu",
	}
}

// hasCharacterRepetition checks for excessive character repetition (5+ consecutive chars)
func (cfs *ContentFilterService) hasCharacterRepetition(text string) bool {
	if len(text) < 5 {
		return false
	}

	count := 1
	for i := 1; i < len(text); i++ {
		if text[i] == text[i-1] {
			count++
			if count >= 5 {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}

// hasWordRepetition checks for the same word repeated 3+ times consecutively
func (cfs *ContentFilterService) hasWordRepetition(text string) bool {
	words := strings.Fields(strings.ToLower(text))
	if len(words) < 3 {
		return false
	}

	for i := 0; i <= len(words)-3; i++ {
		if words[i] == words[i+1] && words[i+1] == words[i+2] {
			return true
		}
	}
	return false
}

// containsWholeWord checks if text contains the word as a whole word (not substring)
func (cfs *ContentFilterService) containsWholeWord(text, word string) bool {
	// Convert both to lowercase for case-insensitive matching
	lowerText := strings.ToLower(text)
	lowerWord := strings.ToLower(word)

	// Split text into words
	words := strings.FieldsFunc(lowerText, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	// Check if any word matches exactly
	for _, w := range words {
		if w == lowerWord {
			return true
		}
	}

	return false
}

// initializePatterns sets up regex patterns for content detection
func (cfs *ContentFilterService) initializePatterns() {
	// URL detection pattern
	cfs.patterns["url"] = regexp.MustCompile(`https?://[^\s]+|www\.[^\s]+|\b[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`)

	// Email detection pattern
	cfs.patterns["email"] = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)

	// Phone number pattern
	cfs.patterns["phone"] = regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b|\+\d{1,3}[-.\s]?\d{1,14}`)

	// Note: Word repetition is checked separately in hasWordRepetition function
	// No regex pattern needed due to RE2 limitations with backreferences

	// All caps pattern (potential spam/aggression)
	cfs.patterns["allcaps"] = regexp.MustCompile(`^[A-Z\s]{10,}$`)

	// Number spam pattern
	cfs.patterns["numberspam"] = regexp.MustCompile(`\d{5,}`)
}

// ValidateWord checks if a word is appropriate for the educational platform
func (cfs *ContentFilterService) ValidateWord(word string) ContentFilterResult {
	if word == "" {
		return ContentFilterResult{
			IsAppropriate: false,
			Severity:      SeverityMedium,
			Reason:        "Word cannot be empty",
			FlaggedWords:  []string{},
		}
	}

	// Normalize the word for checking
	normalizedWord := strings.ToLower(strings.TrimSpace(word))

	// Check word length (prevent spam)
	if len(normalizedWord) > 100 {
		return ContentFilterResult{
			IsAppropriate: false,
			Severity:      SeverityMedium,
			Reason:        "Word is too long (maximum 100 characters)",
			FlaggedWords:  []string{word},
		}
	}

	// Check for non-educational content across all languages for comprehensive filtering
	for lang := range cfs.profanityLists {
		if result := cfs.checkProfanity(normalizedWord, lang); !result.IsAppropriate {
			return result
		}
	}

	// Check for patterns (URLs, emails, etc.)
	if result := cfs.checkPatterns(normalizedWord); !result.IsAppropriate {
		return result
	}

	// Check for non-alphabetic content (except allowed characters)
	if result := cfs.checkCharacters(normalizedWord); !result.IsAppropriate {
		return result
	}

	return ContentFilterResult{
		IsAppropriate: true,
		Severity:      "",
		Reason:        "",
		FlaggedWords:  []string{},
	}
}

// ValidateWordlistName checks if a wordlist name is appropriate
func (cfs *ContentFilterService) ValidateWordlistName(name string) ContentFilterResult {
	if name == "" {
		return ContentFilterResult{
			IsAppropriate: false,
			Severity:      SeverityMedium,
			Reason:        "Wordlist name cannot be empty",
			FlaggedWords:  []string{},
		}
	}

	// Normalize the name
	normalizedName := strings.ToLower(strings.TrimSpace(name))

	// Check length
	if len(normalizedName) > 50 {
		return ContentFilterResult{
			IsAppropriate: false,
			Severity:      SeverityMedium,
			Reason:        "Wordlist name is too long (maximum 50 characters)",
			FlaggedWords:  []string{name},
		}
	}

	// Check for profanity in multiple languages
	for lang := range cfs.profanityLists {
		if result := cfs.checkProfanity(normalizedName, lang); !result.IsAppropriate {
			return result
		}
	}

	// Check for patterns
	if result := cfs.checkPatterns(normalizedName); !result.IsAppropriate {
		return result
	}

	return ContentFilterResult{
		IsAppropriate: true,
		Severity:      "",
		Reason:        "",
		FlaggedWords:  []string{},
	}
}

// ValidateDescription checks if a description is appropriate
func (cfs *ContentFilterService) ValidateDescription(description string) ContentFilterResult {
	if description == "" {
		// Empty descriptions are allowed
		return ContentFilterResult{
			IsAppropriate: true,
			Severity:      "",
			Reason:        "",
			FlaggedWords:  []string{},
		}
	}

	// Normalize the description
	normalizedDesc := strings.ToLower(strings.TrimSpace(description))

	// Check length
	if len(normalizedDesc) > 200 {
		return ContentFilterResult{
			IsAppropriate: false,
			Severity:      SeverityMedium,
			Reason:        "Description is too long (maximum 200 characters)",
			FlaggedWords:  []string{},
		}
	}

	// Check for profanity in multiple languages
	words := strings.Fields(normalizedDesc)
	flaggedWords := []string{}

	for _, word := range words {
		for lang := range cfs.profanityLists {
			if result := cfs.checkProfanity(word, lang); !result.IsAppropriate {
				flaggedWords = append(flaggedWords, word)
			}
		}
	}

	if len(flaggedWords) > 0 {
		return ContentFilterResult{
			IsAppropriate: false,
			Severity:      SeverityHigh,
			Reason:        "Description contains inappropriate content",
			FlaggedWords:  flaggedWords,
		}
	}

	// Check for patterns
	if result := cfs.checkPatterns(normalizedDesc); !result.IsAppropriate {
		return result
	}

	return ContentFilterResult{
		IsAppropriate: true,
		Severity:      "",
		Reason:        "",
		FlaggedWords:  []string{},
	}
}

// checkProfanity checks for profanity and inappropriate content
func (cfs *ContentFilterService) checkProfanity(text, language string) ContentFilterResult {
	// Default to English if language not supported
	if _, exists := cfs.profanityLists[language]; !exists {
		language = "en"
	}

	words := cfs.profanityLists[language]
	flaggedWords := []string{}

	for _, word := range words {
		// Use word boundary matching to avoid false positives like "hell" in "hello"
		if cfs.containsWholeWord(text, word) {
			flaggedWords = append(flaggedWords, word)
		}
	}

	if len(flaggedWords) > 0 {
		severity := SeverityMedium
		reason := "Contains inappropriate language"

		// Determine severity based on content type
		for _, word := range flaggedWords {
			if cfs.isHighSeverityWord(word) {
				severity = SeverityHigh
				reason = "Contains inappropriate content not suitable for educational use"
				break
			}
		}

		return ContentFilterResult{
			IsAppropriate: false,
			Severity:      severity,
			Reason:        reason,
			FlaggedWords:  flaggedWords,
		}
	}

	return ContentFilterResult{
		IsAppropriate: true,
		Severity:      "",
		Reason:        "",
		FlaggedWords:  []string{},
	}
}

// checkPatterns checks for problematic patterns in text
func (cfs *ContentFilterService) checkPatterns(text string) ContentFilterResult {
	// Check for URLs
	if cfs.patterns["url"].MatchString(text) {
		return ContentFilterResult{
			IsAppropriate: false,
			Severity:      SeverityHigh,
			Reason:        "URLs are not allowed in vocabulary content",
			FlaggedWords:  []string{"URL detected"},
		}
	}

	// Check for emails
	if cfs.patterns["email"].MatchString(text) {
		return ContentFilterResult{
			IsAppropriate: false,
			Severity:      SeverityHigh,
			Reason:        "Email addresses are not allowed in vocabulary content",
			FlaggedWords:  []string{"Email detected"},
		}
	}

	// Check for phone numbers
	if cfs.patterns["phone"].MatchString(text) {
		return ContentFilterResult{
			IsAppropriate: false,
			Severity:      SeverityHigh,
			Reason:        "Phone numbers are not allowed in vocabulary content",
			FlaggedWords:  []string{"Phone number detected"},
		}
	}

	// Check for excessive character repetition
	if cfs.hasCharacterRepetition(text) {
		return ContentFilterResult{
			IsAppropriate: false,
			Severity:      SeverityMedium,
			Reason:        "Excessive character repetition detected (potential spam)",
			FlaggedWords:  []string{"Repetitive content"},
		}
	}

	// Check for word repetition
	if cfs.hasWordRepetition(text) {
		return ContentFilterResult{
			IsAppropriate: false,
			Severity:      SeverityMedium,
			Reason:        "Excessive word repetition detected (potential spam)",
			FlaggedWords:  []string{"Repetitive content"},
		}
	}

	// Check for all caps (potential spam/aggression)
	if cfs.patterns["allcaps"].MatchString(text) {
		return ContentFilterResult{
			IsAppropriate: false,
			Severity:      SeverityLow,
			Reason:        "Excessive use of capital letters",
			FlaggedWords:  []string{"All caps text"},
		}
	}

	// Check for number spam
	if cfs.patterns["numberspam"].MatchString(text) {
		return ContentFilterResult{
			IsAppropriate: false,
			Severity:      SeverityMedium,
			Reason:        "Long number sequences are not appropriate for vocabulary",
			FlaggedWords:  []string{"Number spam"},
		}
	}

	return ContentFilterResult{
		IsAppropriate: true,
		Severity:      "",
		Reason:        "",
		FlaggedWords:  []string{},
	}
}

// checkCharacters validates that content contains appropriate characters for vocabulary
func (cfs *ContentFilterService) checkCharacters(text string) ContentFilterResult {
	// Allow letters, numbers, spaces, hyphens, apostrophes, and common punctuation
	allowedChars := func(r rune) bool {
		return unicode.IsLetter(r) ||
			unicode.IsNumber(r) ||
			unicode.IsSpace(r) ||
			r == '-' || r == '\'' || r == '.' || r == ',' ||
			r == '(' || r == ')' || r == '[' || r == ']' ||
			r == '!' || r == '?'
	}

	for _, r := range text {
		if !allowedChars(r) {
			return ContentFilterResult{
				IsAppropriate: false,
				Severity:      SeverityLow,
				Reason:        fmt.Sprintf("Contains inappropriate character: %c", r),
				FlaggedWords:  []string{string(r)},
			}
		}
	}

	return ContentFilterResult{
		IsAppropriate: true,
		Severity:      "",
		Reason:        "",
		FlaggedWords:  []string{},
	}
}

// isHighSeverityWord determines if a word should trigger high severity response
func (cfs *ContentFilterService) isHighSeverityWord(word string) bool {
	highSeverityWords := []string{
		"sex", "porn", "xxx", "adult", "naked", "nude",
		"kill", "murder", "violence", "weapon", "gun", "knife",
		"hate", "racist", "nazi", "terrorist",
		"sexo", "adulto", "desnudo", "matar", "violencia", "arma", "odio", "racista",
		"sexe", "adulte", "nu", "tuer", "violence", "arme", "haine", "raciste",
		"sex", "erwachsene", "nackt", "töten", "gewalt", "waffe", "hass", "rassist",
		"sesso", "adulto", "nudo", "uccidere", "violenza", "arma", "odio", "razzista",
		"sexo", "adulto", "nu", "matar", "violência", "arma", "ódio", "racista",
		"ecchi", "hentai", "sekkusu", "korosu", "boryoku", "sabetsu",
	}

	for _, highWord := range highSeverityWords {
		if strings.Contains(word, highWord) {
			return true
		}
	}

	return false
}

// GetContentGuidelines returns user-friendly content guidelines
func (cfs *ContentFilterService) GetContentGuidelines() []string {
	return []string{
		"Keep content educational and appropriate for all ages",
		"Use vocabulary words that help language learning",
		"Avoid profanity, adult content, or offensive language",
		"Don't include personal information (emails, phone numbers, URLs)",
		"Keep wordlist names and descriptions family-friendly",
		"Focus on legitimate vocabulary and educational content",
		"Respect cultural sensitivities across different languages",
	}
}
