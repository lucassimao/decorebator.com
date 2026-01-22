package model

import (
	"time"

	"github.com/jackc/pgx/pgtype"
)

type Inflection struct {
	Inflection string   `json:"inflection"`
	Tense      string   `json:"tense"`
	Examples   []string `json:"examples"`
}

type Accent string

const (
	US     Accent = "US"
	CANADA Accent = "CA"
	UK     Accent = "UK"
)

type Sound struct {
	Accent Accent `json:"accent"`
	Link   string `json:"link"`
}

type PhoneticNotation struct {
	Ipa    string `json:"ipa"`
	Accent Accent `json:"accent"`
}

type DefinitionSource string

const (
	ChatGPT    DefinitionSource = "ChatGPT"
	Wiktionary DefinitionSource = "wiktionary"
)

type DefinitionExampleAudio struct {
	ID             int64     `json:"id"`
	DefinitionID   int64     `json:"definitionId"`
	ExampleText    string    `json:"exampleText"`
	ExampleHash    string    `json:"exampleHash"`
	AudioURL       string    `json:"audioUrl"`
	InflectionType string    `json:"inflectionType,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

type Definition struct {
	ID                     int64                    `json:"id"`
	Token                  string                   `json:"token"`
	Language               string                   `json:"language"`
	Meaning                string                   `json:"meaning"`
	PartOfSpeech           string                   `json:"partOfSpeech"`
	PartOfSpeechNormalized string                   `json:"partOfSpeechNormalized"`
	IsVerbType             bool                     `json:"isVerbType"` // Computed flag indicating if this is a verb/phrasal verb
	Examples               []string                 `json:"examples"`
	Inflections            []Inflection             `json:"inflections"`
	Source                 DefinitionSource         `json:"source"`
	SourceID               *string                  `json:"sourceId"`
	Sounds                 []Sound                  `json:"sounds"`
	PhoneticNotations      []PhoneticNotation       `json:"phoneticNotations"`
	ExampleAudioFiles      []DefinitionExampleAudio `json:"exampleAudioFiles"`
	MeaningAudioURL        string                   `json:"meaningAudioUrl,omitempty"`

	CreatedAt pgtype.Timestamp `json:"createdAt"`
	UpdatedAt pgtype.Timestamp `json:"updatedAt"`
}

// GetAllExamples returns all available examples for this definition,
// using the appropriate source based on part of speech.
// For verbs: returns examples from inflections
// For non-verbs: returns main examples
func (d *Definition) GetAllExamples() []string {
	var allExamples []string

	if d.IsVerbType {
		// For verbs: collect from inflections
		for _, inflection := range d.Inflections {
			allExamples = append(allExamples, inflection.Examples...)
		}
	} else {
		// For non-verbs: use main examples
		allExamples = append(allExamples, d.Examples...)
	}

	return allExamples
}

// GetLongestExample returns the longest example sentence for this definition,
// using the appropriate source based on part of speech.
func (d *Definition) GetLongestExample() string {
	allExamples := d.GetAllExamples()

	longestExample := ""
	for _, example := range allExamples {
		if len(example) > len(longestExample) {
			longestExample = example
		}
	}

	return longestExample
}

// HasExamples returns true if this definition has any examples available,
// checking the appropriate source based on part of speech.
func (d *Definition) HasExamples() bool {
	if d.IsVerbType {
		for _, inflection := range d.Inflections {
			if len(inflection.Examples) > 0 {
				return true
			}
		}
		return false
	}

	return len(d.Examples) > 0
}
