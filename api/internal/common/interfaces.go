package common

import "decorebator.com/internal/model"

type ContentGenerationService interface {
	FetchDefinition(wordId int64) (jobId int64, err error)
	TextToSpeech(wordId int64) (jobId int64, err error)
	GenerateImage(definitionId int64, customPrompt string) (jobId int64, err error)
}

type SpacedRepetitionStrategy interface {
	CreateQuiz(wordlistID, userID int64) (*model.Quiz, error)
	SaveQuizResult(id int64, success bool) error
	IncludeDefinitions(wordId, userId int64, definitionIds []int64) error
}
