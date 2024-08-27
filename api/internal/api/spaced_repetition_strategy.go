package api

import "decorebator.com/internal/model"

type SpacedRepetitionStrategy interface {
	CreateQuiz(wordlistID, userID int64) (*model.Quiz, error)
	SaveQuizResult(id int64, success bool) error
	IncludeDefinitions(userID int64, definitions []model.Definition) error
}
