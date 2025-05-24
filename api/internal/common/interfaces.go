package common

import (
	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5"
)

type SpacedRepetitionStrategy interface {
	CreateQuiz(wordlistID, userID int64) (*model.Quiz, error)
	SaveQuizResult(id int64, success bool, tx *pgx.Tx) error
	IncludeDefinitions(wordId, userId int64, definitionIds []int64, tx pgx.Tx) error
}
