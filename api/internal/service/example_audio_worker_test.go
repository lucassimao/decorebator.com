package service

import (
	"errors"
	"testing"

	"decorebator.com/internal/common"
	"github.com/riverqueue/river"
)

func TestExampleAudioDefinitionLookupRetriesTransientErrorsAndCancelsNotFound(t *testing.T) {
	transient := errors.New("database temporarily unavailable")
	if got := exampleAudioDefinitionLookupError(transient); !errors.Is(got, transient) {
		t.Fatalf("transient lookup error = %v, want retryable original", got)
	}

	notFound := common.NotFoundError{ID: 42, Entity: "definition"}
	var cancelErr *river.JobCancelError
	if got := exampleAudioDefinitionLookupError(notFound); !errors.As(got, &cancelErr) {
		t.Fatalf("not-found lookup error = %T %v, want JobCancelError", got, got)
	}
}
