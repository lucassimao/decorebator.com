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
	ID                int64                  `json:"id"`
	Token             string                 `json:"token"`
	Language          string                 `json:"language"`
	Meaning           string                 `json:"meaning"`
	PartOfSpeech      string                 `json:"partOfSpeech"`
	Examples          []string               `json:"examples"`
	Inflections       []Inflection           `json:"inflections"`
	Source            DefinitionSource       `json:"source"`
	SourceId          *string                `json:"sourceId"`
	Sounds            []Sound                `json:"sounds"`
	PhoneticNotations []PhoneticNotation     `json:"phoneticNotations"`

	CreatedAt pgtype.Timestamp `json:"createdAt"`
	UpdatedAt pgtype.Timestamp `json:"updatedAt"`
}
