package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/openai"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

func TestMeaningAudioWorkerCompletesOutsideDefinitionTransaction(t *testing.T) {
	generated := false
	uploaded := false
	updated := false
	worker := newMeaningAudioWorkerWithDeps(nil, meaningAudioWorkerDeps{
		getDefinition: func(context.Context, int64) (*model.Definition, error) {
			return &model.Definition{ID: 41, Meaning: "a useful meaning"}, nil
		},
		getLanguage: func(context.Context, int64) (string, error) { return "pt-BR", nil },
		generateAudio: func(_ context.Context, text, language string) (*openai.GenerateAudioResponse, error) {
			generated = true
			if text != "a useful meaning" || language != "pt-BR" {
				t.Fatalf("generator got text=%q language=%q", text, language)
			}
			return &openai.GenerateAudioResponse{Data: []byte("audio")}, nil
		},
		uploadAudio: func(_ context.Context, data []byte, objectName string) (string, error) {
			uploaded = true
			if string(data) != "audio" || objectName != "audio/meaning-41.mp3" {
				t.Fatalf("upload got data=%q object=%q", data, objectName)
			}
			return "https://objects.test/audio/meaning-41.mp3", nil
		},
		setURLIfEmpty: func(_ context.Context, definitionID int64, url string) (bool, error) {
			updated = true
			if definitionID != 41 || url != "https://objects.test/audio/meaning-41.mp3" {
				t.Fatalf("update got definition=%d url=%q", definitionID, url)
			}
			return true, nil
		},
	})

	err := worker.Work(context.Background(), &river.Job[MeaningAudioArgs]{Args: MeaningAudioArgs{DefinitionID: 41, WordID: 9}})
	if err != nil {
		t.Fatal(err)
	}
	if !generated || !uploaded || !updated {
		t.Fatalf("pipeline calls generated=%t uploaded=%t updated=%t", generated, uploaded, updated)
	}
}

func TestMeaningAudioWorkerSkipsCompletedDefinition(t *testing.T) {
	worker := newMeaningAudioWorkerWithDeps(nil, meaningAudioWorkerDeps{
		getDefinition: func(context.Context, int64) (*model.Definition, error) {
			return &model.Definition{ID: 41, Meaning: "meaning", MeaningAudioURL: "already-set"}, nil
		},
		generateAudio: func(context.Context, string, string) (*openai.GenerateAudioResponse, error) {
			t.Fatal("completed definition regenerated paid audio")
			return nil, nil
		},
	})
	if err := worker.Work(context.Background(), &river.Job[MeaningAudioArgs]{Args: MeaningAudioArgs{DefinitionID: 41}}); err != nil {
		t.Fatal(err)
	}
}

func TestMeaningAudioWorkerTreatsLostCompareAndSetAsCompleted(t *testing.T) {
	updated := false
	worker := newMeaningAudioWorkerWithDeps(nil, meaningAudioWorkerDeps{
		getDefinition: func(context.Context, int64) (*model.Definition, error) {
			return &model.Definition{ID: 41, Meaning: "meaning"}, nil
		},
		getLanguage: func(context.Context, int64) (string, error) { return "en", nil },
		generateAudio: func(context.Context, string, string) (*openai.GenerateAudioResponse, error) {
			return &openai.GenerateAudioResponse{Data: []byte("audio")}, nil
		},
		uploadAudio: func(context.Context, []byte, string) (string, error) { return "candidate", nil },
		setURLIfEmpty: func(context.Context, int64, string) (bool, error) {
			updated = true
			return false, nil
		},
	})
	if err := worker.Work(context.Background(), testMeaningAudioJob()); err != nil {
		t.Fatalf("lost compare-and-set should be an idempotent success: %v", err)
	}
	if !updated {
		t.Fatal("compare-and-set was not attempted")
	}
}

func TestMeaningAudioWorkerCancelsMissingDefinition(t *testing.T) {
	worker := newMeaningAudioWorkerWithDeps(nil, meaningAudioWorkerDeps{
		getDefinition: func(context.Context, int64) (*model.Definition, error) { return nil, nil },
	})
	var cancel *river.JobCancelError
	if err := worker.Work(context.Background(), testMeaningAudioJob()); !errors.As(err, &cancel) {
		t.Fatalf("error = %T %v, want JobCancelError", err, err)
	}
}

func TestMeaningAudioWorkerRetryAndPermanentErrors(t *testing.T) {
	t.Run("transport retry", func(t *testing.T) {
		want := errors.New("temporary transport failure")
		worker := testMeaningAudioWorker(func(context.Context, string, string) (*openai.GenerateAudioResponse, error) {
			return nil, want
		})
		if err := worker.Work(context.Background(), testMeaningAudioJob()); !errors.Is(err, want) {
			t.Fatalf("error = %v, want retryable %v", err, want)
		}
	})

	t.Run("rate limit snooze", func(t *testing.T) {
		worker := testMeaningAudioWorker(func(context.Context, string, string) (*openai.GenerateAudioResponse, error) {
			return audioProviderError(openAIErrorRateLimitExceeded), nil
		})
		var snooze *river.JobSnoozeError
		if err := worker.Work(context.Background(), testMeaningAudioJob()); !errors.As(err, &snooze) {
			t.Fatalf("error = %T %v, want JobSnoozeError", err, err)
		}
	})

	t.Run("billing cancellation", func(t *testing.T) {
		worker := testMeaningAudioWorker(func(context.Context, string, string) (*openai.GenerateAudioResponse, error) {
			return audioProviderError(openAIErrorBillingLimit), nil
		})
		var cancel *river.JobCancelError
		if err := worker.Work(context.Background(), testMeaningAudioJob()); !errors.As(err, &cancel) {
			t.Fatalf("error = %T %v, want JobCancelError", err, err)
		}
	})
}

func TestMeaningAudioEligibilityRetriesInfrastructureAndCancelsBusinessDecision(t *testing.T) {
	transient := errors.New("database temporarily unavailable")
	if got := meaningAudioEligibilityError(transient); !errors.Is(got, transient) {
		t.Fatalf("transient eligibility error = %v, want retryable original", got)
	}

	limit := common.BusinessError{Message: "free tier limit reached"}
	var cancel *river.JobCancelError
	if got := meaningAudioEligibilityError(limit); !errors.As(got, &cancel) {
		t.Fatalf("business eligibility error = %T %v, want JobCancelError", got, got)
	}
}

func TestMeaningAudioJobContract(t *testing.T) {
	args := MeaningAudioArgs{DefinitionID: 41, WordID: 9}
	if args.Kind() != "DefinitionMeaningAudio" {
		t.Fatalf("kind = %q", args.Kind())
	}
	if got := (&MeaningAudioWorker{}).Timeout(nil); got != 2*time.Minute {
		t.Fatalf("timeout = %s, want 2m", got)
	}
	opts := meaningAudioInsertOpts()
	if opts.Queue != MeaningAudioQueue || opts.MaxAttempts != 5 || !opts.UniqueOpts.ByArgs || !opts.UniqueOpts.ByQueue {
		t.Fatalf("unexpected insert opts: %+v", opts)
	}
	argsType := reflect.TypeOf(args)
	for _, field := range []struct {
		name string
		want string
	}{
		{name: "DefinitionID", want: "unique"},
		{name: "WordID", want: ""},
		{name: "UserID", want: ""},
	} {
		got, ok := argsType.FieldByName(field.name)
		if !ok || got.Tag.Get("river") != field.want {
			t.Fatalf("%s river tag = %q, want %q", field.name, got.Tag.Get("river"), field.want)
		}
	}
}

func TestMeaningAudioEnqueueFailureRemainsFatalToDefinitionTransaction(t *testing.T) {
	want := errors.New("queue unavailable")
	jobService := &meaningAudioScheduleFailure{err: want}

	err := scheduleRequiredMeaningAudio(context.Background(), jobService, 41, 9, nil, nil)
	if !errors.Is(err, want) {
		t.Fatalf("required meaning-audio enqueue error must propagate and trigger the caller's deferred rollback: %v", err)
	}
	if jobService.calls != 1 {
		t.Fatalf("expected exactly one required enqueue attempt, got %d", jobService.calls)
	}
}

func testMeaningAudioWorker(generator AudioGenerator) *MeaningAudioWorker {
	return newMeaningAudioWorkerWithDeps(nil, meaningAudioWorkerDeps{
		getDefinition: func(context.Context, int64) (*model.Definition, error) {
			return &model.Definition{ID: 41, Meaning: "meaning"}, nil
		},
		getLanguage:   func(context.Context, int64) (string, error) { return "en", nil },
		generateAudio: generator,
		uploadAudio: func(context.Context, []byte, string) (string, error) {
			return "unused", nil
		},
		setURLIfEmpty: func(context.Context, int64, string) (bool, error) { return true, nil },
	})
}

func testMeaningAudioJob() *river.Job[MeaningAudioArgs] {
	return &river.Job[MeaningAudioArgs]{Args: MeaningAudioArgs{DefinitionID: 41, WordID: 9}}
}

type meaningAudioScheduleFailure struct {
	JobService
	err   error
	calls int
}

func (f *meaningAudioScheduleFailure) ScheduleMeaningAudioJob(context.Context, int64, int64, *int64, *pgx.Tx) error {
	f.calls++
	return f.err
}

func audioProviderError(code string) *openai.GenerateAudioResponse {
	response := &openai.GenerateAudioResponse{}
	response.Error = &struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Param   string `json:"param"`
		Type    string `json:"type"`
	}{Code: code, Message: "provider error"}
	return response
}
