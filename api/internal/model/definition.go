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

	CreatedAt pgtype.Timestamp `json:"createdAt"`
	UpdatedAt pgtype.Timestamp `json:"updatedAt"`
}
