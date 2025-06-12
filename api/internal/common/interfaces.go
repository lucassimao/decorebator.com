package common

import (
	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5"
)

type QuizResult struct {
	UserID                  int64
	WordlistID              int64
	WordID                  int64
	DefinitionID            int64
	LeitnerSystemTrackingID int64
	QuizType                model.QuizType
	BoxID                   int64
	IsCorrect               bool
	ResponseTimeMs          int
}

type SpacedRepetitionStrategy interface {
	CreateQuiz(wordlistID, userID int64) (*model.Quiz, error)
	SaveQuizResult(result QuizResult, isPremium bool, tx *pgx.Tx) error
	IncludeDefinitions(wordID, userID int64, definitionIDs []int64, tx pgx.Tx) error
}
