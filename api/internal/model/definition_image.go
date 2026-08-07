package model

import "github.com/jackc/pgx/v5/pgtype"

type API string

const (
	OPENAI API = "openai"
)

type DefinitionImage struct {
	ID          int64
	API         API
	URL         string
	Description string
	Model       string
	Prompt      string
	IsVisible   bool

	CreatedAt    pgtype.Timestamp
	DefinitionID int64
}

type CreateDefinitionImageDTO struct {
	API          API
	URL          string
	Description  string
	Model        string
	Prompt       string
	DefinitionID int64
}
