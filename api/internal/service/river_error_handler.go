package service

import (
	"context"
	"fmt"

	"decorebator.com/internal/common"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// RiverErrorHandler implements River's ErrorHandler interface to provide
// unified error reporting to Sentry for all background job failures.
// It preserves River's default error handling behavior while adding observability.
type RiverErrorHandler struct{}

// NewRiverErrorHandler creates a new RiverErrorHandler instance
func NewRiverErrorHandler() *RiverErrorHandler {
	return &RiverErrorHandler{}
}

// HandleError is called by River when a job encounters an error.
// It reports the error to Sentry and returns nil to maintain River's
// default retry/cancellation behavior.
func (h *RiverErrorHandler) HandleError(ctx context.Context, job *rivertype.JobRow, err error) *river.ErrorHandlerResult {
	// Create error message with job context
	errorMsg := fmt.Sprintf("River job failed: %s (ID: %d, Attempt: %d/%d)",
		job.Kind, job.ID, job.Attempt, job.MaxAttempts)

	// Report error to Sentry with structured context
	common.CaptureError(ctx, err, errorMsg,
		"job_id", job.ID,
		"job_kind", job.Kind,
		"job_queue", job.Queue,
		"attempt", job.Attempt,
		"max_attempts", job.MaxAttempts,
		"job_state", string(job.State),
		"error", err,
	)

	// Return nil to preserve River's default error handling behavior
	// This allows River to handle retries and cancellation as configured
	return nil
}

// HandlePanic is called by River when a job panics.
// It reports the panic to Sentry and returns nil to maintain River's
// default panic handling behavior.
func (h *RiverErrorHandler) HandlePanic(ctx context.Context, job *rivertype.JobRow, panicVal any, trace string) *river.ErrorHandlerResult {
	// Create panic message with job context
	panicMsg := fmt.Sprintf("River job panicked: %s (ID: %d, Attempt: %d/%d)",
		job.Kind, job.ID, job.Attempt, job.MaxAttempts)

	// Create error from panic value
	panicErr := fmt.Errorf("panic: %v", panicVal)

	// Report panic to Sentry with structured context including stack trace
	common.CaptureError(ctx, panicErr, panicMsg,
		"job_id", job.ID,
		"job_kind", job.Kind,
		"job_queue", job.Queue,
		"attempt", job.Attempt,
		"max_attempts", job.MaxAttempts,
		"job_state", string(job.State),
		"panic_value", panicVal,
		"stack_trace", trace,
		"error", panicErr,
	)

	// Return nil to preserve River's default panic handling behavior
	// This allows River to handle the panic as configured
	return nil
}
