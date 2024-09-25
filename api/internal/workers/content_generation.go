package workers

import (
	"context"
	"errors"

	"decorebator.com/internal/common"
	"github.com/riverqueue/river"
)

const IMAGE_GENERATOR_QUEUE = "image_generator"
const TEXT_TO_SPEECH_QUEUE = "text_to_speech"
const DEFINITION_FETCHER_QUEUE = "definition_fetcher"

type ContentGenerationServiceImpl struct{}

func (*ContentGenerationServiceImpl) GenerateImage(definitionId int64, customPrompt string) (int64, error) {
	opts := river.InsertOpts{
		Queue: IMAGE_GENERATOR_QUEUE,
	}

	return triggerWorker(&opts, ImageGeneratorArgs{
		DefinitionId: definitionId,
		CustomPrompt: customPrompt,
	})
}

func (*ContentGenerationServiceImpl) TextToSpeech(wordId int64) (int64, error) {
	opts := river.InsertOpts{
		Queue: TEXT_TO_SPEECH_QUEUE,
	}

	return triggerWorker(&opts, TextToSpeechArgs{
		WordId: wordId,
	})
}

func (*ContentGenerationServiceImpl) FetchDefinition(wordId int64) (int64, error) {
	opts := river.InsertOpts{
		Queue: DEFINITION_FETCHER_QUEUE,
	}

	return triggerWorker(&opts, DefinitionFetcherArgs{
		WordId: wordId,
	})
}

func NewContentGenerationService() common.ContentGenerationService {
	return &ContentGenerationServiceImpl{}
}

func triggerWorker(opts *river.InsertOpts, args river.JobArgs) (int64, error) {
	logger := common.Logger.With("func", "triggerWorker", "Kind", args.Kind())

	riverClient, err := GetRiverClient()
	if err != nil {
		logger.Error("failed to open river connection", "error", err)
		return -1, errors.New("could not open river client")
	}

	result, err := riverClient.Insert(context.Background(), args, opts)

	if err != nil {
		logger.Error("failed to trigger river job", "error", err)
		return -1, errors.New("could not insert job")
	}

	return result.Job.ID, nil
}
