package common

type WorkerTrigger interface {
	TriggerDefinitionFetcher(wordId int64) (int64, error)
	TriggerTextToSpeech(wordId int64) (int64, error)
	TriggerImageGenerator(definitionId int64, customPrompt string) (int64, error)
}
