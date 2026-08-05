package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/openai"
	"github.com/riverqueue/river"
)

type MeaningAudioArgs struct {
	DefinitionID int64  `json:"definitionId" river:"unique"`
	WordID       int64  `json:"wordId"`
	UserID       *int64 `json:"userId"`
}

func (MeaningAudioArgs) Kind() string { return "DefinitionMeaningAudio" }

type meaningAudioWorkerDeps struct {
	getDefinition func(context.Context, int64) (*model.Definition, error)
	getLanguage   func(context.Context, int64) (string, error)
	generateAudio AudioGenerator
	uploadAudio   func(context.Context, []byte, string) (string, error)
	setURLIfEmpty func(context.Context, int64, string) (bool, error)
}

type MeaningAudioWorker struct {
	river.WorkerDefaults[MeaningAudioArgs]
	userService *UserService
	deps        meaningAudioWorkerDeps
}

func NewMeaningAudioWorker(definitionService *DefinitionService, wordService *WordService, userService *UserService) *MeaningAudioWorker {
	return newMeaningAudioWorkerWithDeps(userService, meaningAudioWorkerDeps{
		getDefinition: definitionService.GetDefinitionByID,
		getLanguage: func(ctx context.Context, wordID int64) (string, error) {
			language, _, err := wordService.GetWordlistLanguageAndPronunciation(ctx, wordID)
			return language, err
		},
		generateAudio: openai.GenerateAudio,
		uploadAudio: func(ctx context.Context, data []byte, objectName string) (string, error) {
			return common.Upload(ctx, data, "decorebator", objectName, "audio/mpeg")
		},
		setURLIfEmpty: definitionService.SetMeaningAudioURLIfEmpty,
	})
}

func newMeaningAudioWorkerWithDeps(userService *UserService, deps meaningAudioWorkerDeps) *MeaningAudioWorker {
	return &MeaningAudioWorker{userService: userService, deps: deps}
}

func (w *MeaningAudioWorker) Timeout(*river.Job[MeaningAudioArgs]) time.Duration {
	return 2 * time.Minute
}

func (w *MeaningAudioWorker) Work(ctx context.Context, job *river.Job[MeaningAudioArgs]) error {
	startedAt := time.Now()
	logger := common.Logger.With(
		"worker", "meaning_audio",
		"definitionId", job.Args.DefinitionID,
		"wordId", job.Args.WordID,
	)
	if job.Args.UserID != nil {
		if w.userService == nil {
			return river.JobCancel(errors.New("meaning audio worker has no user service"))
		}
		if err := w.userService.ValidateUserEligibilityForWorkers(ctx, *job.Args.UserID); err != nil {
			return meaningAudioEligibilityError(err)
		}
	}

	definition, err := w.deps.getDefinition(ctx, job.Args.DefinitionID)
	if err != nil {
		if errors.Is(err, common.NotFoundError{}) {
			return river.JobCancel(err)
		}
		return err
	}
	if definition == nil {
		return river.JobCancel(common.NotFoundError{ID: job.Args.DefinitionID, Entity: "definition"})
	}
	if definition.MeaningAudioURL != "" {
		logger.Debug("meaning audio already exists", "duration", time.Since(startedAt))
		return nil
	}
	if definition.Meaning == "" {
		return river.JobCancel(fmt.Errorf("definition %d has no meaning", job.Args.DefinitionID))
	}

	language, err := w.deps.getLanguage(ctx, job.Args.WordID)
	if err != nil {
		return err
	}
	response, err := w.deps.generateAudio(ctx, definition.Meaning, language)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return err
	}
	if response == nil {
		return errors.New("meaning audio generator returned no response")
	}
	if response.Error != nil {
		providerErr := fmt.Errorf("meaning audio provider error %q: %s", response.Error.Code, response.Error.Message)
		switch response.Error.Code {
		case openAIErrorRateLimitExceeded:
			//nolint:gosec // G404 - jitter does not require cryptographic randomness.
			return river.JobSnooze(time.Minute + time.Duration(rand.Intn(60))*time.Second)
		case openAIErrorBillingLimit:
			return river.JobCancel(providerErr)
		default:
			return providerErr
		}
	}
	if len(response.Data) == 0 {
		return errors.New("meaning audio generator returned empty audio")
	}

	objectName := fmt.Sprintf("audio/meaning-%d.mp3", job.Args.DefinitionID)
	audioURL, err := w.deps.uploadAudio(ctx, response.Data, objectName)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return err
	}
	updated, err := w.deps.setURLIfEmpty(ctx, job.Args.DefinitionID, audioURL)
	if err != nil {
		return err
	}
	logger.Info("meaning audio job completed", "urlUpdated", updated, "duration", time.Since(startedAt))
	return nil
}

func meaningAudioEligibilityError(err error) error {
	var businessErr common.BusinessError
	if errors.As(err, &businessErr) {
		return river.JobCancel(err)
	}
	return err
}
