package model

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// PronunciationSystem represents the pronunciation system for a wordlist
type PronunciationSystem string

const (
	PronunciationSystemIPA      PronunciationSystem = "ipa"      // International Phonetic Alphabet (European languages)
	PronunciationSystemRomaji   PronunciationSystem = "romaji"   // Romaji for Japanese
	PronunciationSystemHiragana PronunciationSystem = "hiragana" // Hiragana/Katakana for Japanese
	PronunciationSystemPinyin   PronunciationSystem = "pinyin"   // Pinyin for Chinese (Mandarin)
	PronunciationSystemHangul   PronunciationSystem = "hangul"   // Hangul for Korean
)

// GetDefaultPronunciationSystem returns the default pronunciation system for a language
func GetDefaultPronunciationSystem(languageCode string) PronunciationSystem {
	switch languageCode {
	case "ja":
		return PronunciationSystemRomaji
	case "zh", "zh-cn", "zh-tw":
		return PronunciationSystemPinyin
	case "ko", "kr":
		return PronunciationSystemHangul
	default:
		return PronunciationSystemIPA
	}
}

// GetSupportedPronunciationSystems returns the supported pronunciation systems for a language
func GetSupportedPronunciationSystems(languageCode string) []PronunciationSystem {
	switch languageCode {
	case "ja":
		return []PronunciationSystem{PronunciationSystemRomaji, PronunciationSystemHiragana}
	case "zh", "zh-cn", "zh-tw":
		return []PronunciationSystem{PronunciationSystemPinyin}
	case "ko", "kr":
		return []PronunciationSystem{PronunciationSystemHangul}
	default:
		return []PronunciationSystem{PronunciationSystemIPA}
	}
}

// IsPronunciationSystemSupported reports whether a pronunciation system is
// valid for the requested language according to the create/read contract.
func IsPronunciationSystemSupported(languageCode string, pronunciationSystem PronunciationSystem) bool {
	for _, supported := range GetSupportedPronunciationSystems(languageCode) {
		if supported == pronunciationSystem {
			return true
		}
	}
	return false
}

// CanChangePronunciationSystem returns true if the language allows changing pronunciation system
func CanChangePronunciationSystem(languageCode string) bool {
	switch languageCode {
	case "ja":
		return true // Japanese allows switching between Romaji and Hiragana/Katakana
	default:
		return false // All other languages are locked to their default system
	}
}

type Wordlist struct {
	ID                  int64               `json:"id"`
	Name                string              `json:"name"`
	Description         string              `json:"description"`
	CreatedAt           pgtype.Timestamptz  `json:"createdAt"`
	UpdatedAt           pgtype.Timestamptz  `json:"updatedAt"`
	UserID              int64               `json:"userId"`
	LanguageCode        string              `json:"languageCode"`
	PronunciationSystem PronunciationSystem `json:"pronunciationSystem"`

	// Computed dinamically based on words table
	WordsCount        *int `json:"wordsCount"`
	WordsLearnedCount *int `json:"wordsLearnedCount"`
}

func (w Wordlist) MarshalJSON() ([]byte, error) {
	var (
		createdAtValue any
		updatedAtValue any
	)
	if w.CreatedAt.Valid {
		createdAtValue = w.CreatedAt.Time.UTC().Format(time.RFC3339)
	}
	if w.UpdatedAt.Valid {
		updatedAtValue = w.UpdatedAt.Time.UTC().Format(time.RFC3339)
	}

	m := map[string]any{
		"id":                  w.ID,
		"name":                w.Name,
		"description":         w.Description,
		"languageCode":        w.LanguageCode,
		"languageName":        getLanguageName(w.LanguageCode),
		"pronunciationSystem": w.PronunciationSystem,
		"createdAt":           createdAtValue, // será string ou nil → "null"
		"updatedAt":           updatedAtValue, // será string ou nil → "null"
		"userId":              w.UserID,
	}

	if w.WordsCount != nil {
		m["wordsCount"] = *w.WordsCount
	}
	if w.WordsLearnedCount != nil {
		m["wordsLearnedCount"] = *w.WordsLearnedCount
	}

	return json.Marshal(m)
}

// getLanguageName returns the full language name for a given language code
func getLanguageName(languageCode string) string {
	languageMap := map[string]string{
		"en": "English",
		"es": "Spanish",
		"fr": "French",
		"de": "German",
		"it": "Italian",
		"pt": "Portuguese",
		"ja": "Japanese",
	}
	if name, exists := languageMap[languageCode]; exists {
		return name
	}
	return languageCode // fallback to code if not found
}

func (w *Wordlist) UnmarshalJSON(data []byte) error {
	type Alias Wordlist // Create an alias to avoid recursion

	aux := &struct {
		CreatedAt *string `json:"createdAt"`
		UpdatedAt *string `json:"updatedAt"`
		*Alias
	}{
		Alias: (*Alias)(w),
	}

	// Unmarshal into aux struct
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Handle CreatedAt
	if aux.CreatedAt != nil {
		createdAtTime, err := time.Parse(time.RFC3339, *aux.CreatedAt)
		if err != nil {
			return err
		}
		w.CreatedAt = pgtype.Timestamptz{Time: createdAtTime, Valid: true}
	} else {
		w.CreatedAt = pgtype.Timestamptz{Valid: false}
	}

	// Handle UpdatedAt
	if aux.UpdatedAt != nil {
		updatedAtTime, err := time.Parse(time.RFC3339, *aux.UpdatedAt)
		if err != nil {
			return err
		}
		w.UpdatedAt = pgtype.Timestamptz{Time: updatedAtTime, Valid: true}
	} else {
		w.UpdatedAt = pgtype.Timestamptz{Valid: false}
	}

	return nil
}
