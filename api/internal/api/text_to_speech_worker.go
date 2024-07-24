package api

import (
	"context"

	"decorebator.com/internal/common"
	"github.com/riverqueue/river"
)

type TextToSpeechArgs struct {
	WordId int64 `json:"wordId"`
}

func (TextToSpeechArgs) Kind() string { return "TextToSpeech" }

type TextToSpeechWorker struct {
	river.WorkerDefaults[TextToSpeechArgs]
}

func (w *TextToSpeechWorker) Work(ctx context.Context, job *river.Job[TextToSpeechArgs]) error {
	logger := common.Logger.With("worker", "texttospeech", "WordId", job.Args.WordId)

	word, err := GetWordById(job.Args.WordId)

	if err != nil || word == nil {
		logger.Error("failed to get word", "error", err)
		return err
	}

	return nil
}
