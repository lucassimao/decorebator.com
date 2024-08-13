package model

import "github.com/jackc/pgx/pgtype"

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

type Definition struct {
	ID                int64
	Token             string
	Language          string
	Meaning           string       `json:"meaning"`
	PartOfSpeech      string       `json:"part_of_speech"`
	Examples          []string     `json:"examples"`
	Inflections       []Inflection `json:"inflections"`
	Source            DefinitionSource
	SourceId          string
	Sounds            []Sound
	PhoneticNotations []PhoneticNotation

	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
}
