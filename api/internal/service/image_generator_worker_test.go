package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"decorebator.com/internal/common"
	"github.com/riverqueue/river"
)

func TestBuildImagePromptHasNoFormattingLeakage(t *testing.T) {
	for _, language := range []string{"en", "es", "fr", "de", "it", "pt", "ja"} {
		t.Run(language, func(t *testing.T) {
			prompt, err := buildImagePrompt("A sample sentence", "sample", "an example", language)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(prompt, "%!") {
				t.Fatalf("prompt contains fmt leakage: %q", prompt)
			}
			for _, value := range []string{"A sample sentence", "sample", "an example"} {
				if !strings.Contains(prompt, value) {
					t.Fatalf("prompt does not contain %q", value)
				}
			}
		})
	}
	prompt, err := buildImagePrompt("sentence", "token", "meaning", "xx")
	if err == nil || prompt != "" {
		t.Fatalf("unsupported language returned prompt=%q err=%v", prompt, err)
	}
}

func TestImageGeneratorErrorReportWithoutStrategyCancelsInsteadOfPanicking(t *testing.T) {
	worker := &ImageGeneratorWorker{}
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("resolveErrorReport panicked: %v", recovered)
		}
	}()
	err := worker.resolveErrorReport(context.Background(), 42, ErrorReport{})
	if err == nil {
		t.Fatal("expected missing-strategy cancellation")
	}
	var cancelErr *river.JobCancelError
	if !errors.As(err, &cancelErr) {
		t.Fatalf("expected JobCancelError to prevent paid retries, got %T: %v", err, err)
	}
}

func TestImageDefinitionLookupRetriesTransientErrorsAndCancelsNotFound(t *testing.T) {
	transient := errors.New("database temporarily unavailable")
	if got := imageDefinitionLookupError(transient); !errors.Is(got, transient) {
		t.Fatalf("transient lookup error = %v, want retryable original", got)
	}

	notFound := common.NotFoundError{ID: 42, Entity: "definition"}
	var cancelErr *river.JobCancelError
	if got := imageDefinitionLookupError(notFound); !errors.As(got, &cancelErr) {
		t.Fatalf("not-found lookup error = %T %v, want JobCancelError", got, got)
	}
}
