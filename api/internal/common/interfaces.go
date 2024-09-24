package common

type ContentGenerationService interface {
	FetchDefinition(wordId int64) (jobId int64, err error)
	TextToSpeech(wordId int64) (jobId int64, err error)
	GenerateImage(definitionId int64, customPrompt string) (jobId int64, err error)
}
