package api

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/openai"
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

	if err != nil && errors.Is(err, common.NotFoundError{}) {
		return river.JobCancel(errors.New("word not found"))
	}

	if err != nil || word == nil {
		logger.Error("failed to get word", "error", err)
		return err
	}

	response, err := openai.GenerateAudio(word.Name)
	if err != nil {
		logger.Error("failed to generate audio", "error", err)
		return err
	}

	if response.Error != nil {
		logger.Error("failed to generate audio", "body", response.Error)

		switch response.Error.Code {
		case "rate_limit_exceeded":
			// snoozing between 1 and 2min
			return river.JobSnooze(time.Minute + (time.Duration(rand.Intn(60)) * time.Second))
		case "billing_hard_limit_reached":
			// [TODO] notification here elsewhere
			logger.Warn(response.Error.Message)
		}

		return fmt.Errorf("OpenAI error: %s", response.Error.Message)
	}

	word.AudioURL, err = common.Upload(response.Data, "audio",
		fmt.Sprintf("audio-%d-%s.mp3", word.ID, strings.ReplaceAll(word.Name, " ", "-")), "audio/mpeg")

	if err != nil {
		logger.Error("failed to upload audio", "error", err)
		return err
	}

	logger.Debug("audio generated", "wordId", word.ID, "url", word.AudioURL, "word", word.Name)

	return UpdateWord(word)
}
