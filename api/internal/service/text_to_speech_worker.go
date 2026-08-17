package service

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
	WordID      int64        `json:"wordId"`
	UserID      *int64       `json:"userId"`
	ErrorReport *ErrorReport `json:"errorReport"`
}

func (TextToSpeechArgs) Kind() string { return "TextToSpeech" }

func (w *TextToSpeechWorker) Timeout(*river.Job[TextToSpeechArgs]) time.Duration {
	return 2 * time.Minute
}

type TextToSpeechWorker struct {
	river.WorkerDefaults[TextToSpeechArgs]
	wordService           *WordService
	definitionService     *DefinitionService
	leitnerSystemStrategy *LeitnerSystemStrategy
	userService           *UserService
}

func NewTextToSpeechWorker(wordService *WordService, definitionService *DefinitionService, leitnerSystemStrategy *LeitnerSystemStrategy, userService *UserService) *TextToSpeechWorker {
	return &TextToSpeechWorker{
		wordService:           wordService,
		definitionService:     definitionService,
		leitnerSystemStrategy: leitnerSystemStrategy,
		userService:           userService,
	}
}

func (w *TextToSpeechWorker) Work(ctx context.Context, job *river.Job[TextToSpeechArgs]) error {
	logger := common.Logger.With("worker", "texttospeech", "WordID", job.Args.WordID, "UserID", job.Args.UserID)
	skip, pendingErr := w.shouldSkipReportedAudio(ctx, job.Args.ErrorReport)
	if pendingErr != nil || skip {
		return pendingErr
	}

	// Validate user eligibility before processing (skip for admin/system jobs)
	if job.Args.UserID != nil {
		if err := w.userService.ValidateUserEligibilityForWorkers(ctx, *job.Args.UserID); err != nil {
			logger.Warn("User not eligible for text-to-speech",
				"userId", *job.Args.UserID, "wordId", job.Args.WordID, "error", err)
			// Cancel job permanently - user needs to upgrade
			return river.JobCancel(err)
		}
	}

	word, err := w.wordService.GetWordByID(ctx, job.Args.WordID)

	var notFoundErr common.NotFoundError
	if err != nil && errors.As(err, &notFoundErr) {
		return river.JobCancel(err)
	}

	if err != nil || word == nil {
		logger.Error("failed to get word", "error", err)
		return err
	}

	// Get wordlist language for language-specific audio generation
	languageCode, _, err := w.wordService.GetWordlistLanguageAndPronunciation(ctx, job.Args.WordID)
	if err != nil {
		logger.Error("failed to get wordlist language", "error", err)
		return err
	}

	response, err := openai.GenerateAudio(ctx, word.Name, languageCode)
	if err != nil {
		logger.Error("failed to generate audio", "error", err)
		return err
	}

	if response.Error != nil {
		logger.Error("failed to generate audio", "body", response.Error)

		switch response.Error.Code {
		case "rate_limit_exceeded":
			// snoozing between 1 and 2min
			//nolint:gosec // G404 - using weak random for rate limiting jitter, not cryptographic security
			return river.JobSnooze(time.Minute + (time.Duration(rand.Intn(60)) * time.Second))
		case "billing_hard_limit_reached":
			// [TODO] notification here elsewhere
			logger.Warn(response.Error.Message)
		}

		return fmt.Errorf("OpenAI error: %s", response.Error.Message)
	}

	word.AudioURL, err = common.Upload(ctx, response.Data, common.MinIOBucketName(),
		fmt.Sprintf("audio/audio-%d-%s.mp3", word.ID, strings.ReplaceAll(word.Name, " ", "-")), "audio/mpeg")

	if err != nil {
		logger.Error("failed to upload audio", "error", err)
		return err
	}

	logger.Debug("audio generated", "wordId", word.ID, "url", word.AudioURL, "word", word.Name)

	return w.persistRegeneratedWordAudio(ctx, word, job.Args.ErrorReport)
}

func (w *TextToSpeechWorker) shouldSkipReportedAudio(ctx context.Context, report *ErrorReport) (bool, error) {
	if report == nil {
		return false, nil
	}
	if w.leitnerSystemStrategy == nil {
		return false, river.JobCancel(errors.New("no leitner strategy configured for reported audio regeneration"))
	}
	pending, err := w.leitnerSystemStrategy.IsErrorPending(ctx, *report)
	return !pending, err
}

func (w *TextToSpeechWorker) persistRegeneratedWordAudio(ctx context.Context, word *Word, report *ErrorReport) error {
	if report == nil {
		return w.wordService.UpdateWord(ctx, word, nil)
	}
	tx, err := w.leitnerSystemStrategy.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer common.RollbackTx(ctx, tx, "word audio regeneration persistence")
	if err := w.wordService.UpdateWord(ctx, word, &tx); err != nil {
		return err
	}
	if err := w.leitnerSystemStrategy.MarkErrorResolvedTx(ctx, *report, *report, tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit reported audio regeneration: %w", err)
	}
	return nil
}
