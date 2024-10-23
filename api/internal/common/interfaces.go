package common

import (
	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5"
)

type ContentGenerationService interface {
	FetchDefinition(wordId int64, tx pgx.Tx) (jobId int64, err error)
	TextToSpeech(wordId int64, tx *pgx.Tx) (jobId int64, err error)
	GenerateImage(definitionId int64, customPrompt string, tx *pgx.Tx) (jobId int64, err error)
}

type SpacedRepetitionStrategy interface {
	CreateQuiz(wordlistID, userID int64) (*model.Quiz, error)
	SaveQuizResult(id int64, success bool, tx *pgx.Tx) error
	IncludeDefinitions(wordId, userId int64, definitionIds []int64, tx pgx.Tx) error
}
