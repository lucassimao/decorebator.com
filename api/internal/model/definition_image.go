package model

import "github.com/jackc/pgx/pgtype"

type Api string

const (
	OPENAI Api = "openai"
)

type DefinitionImage struct {
	ID          int64
	Api         Api
	URL         string
	Description string
	Model       string
	Prompt      string
	IsVisible   bool

	CreatedAt    pgtype.Timestamp
	DefinitionId int64
}

type CreateDefinitionImageDTO struct {
	Api          Api
	URL          string
	Description  string
	Model        string
	Prompt       string
	DefinitionId int64
}
