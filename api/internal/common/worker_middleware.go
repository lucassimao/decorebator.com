package common

import (
	"context"
	"strconv"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// WorkerContextMiddleware enriches the context with worker metadata for all River workers.
// It automatically adds job information to Sentry scope and tracks execution duration.
type WorkerContextMiddleware struct {
	river.MiddlewareDefaults
}

// NewWorkerContextMiddleware creates a new WorkerContextMiddleware instance
func NewWorkerContextMiddleware() *WorkerContextMiddleware {
	return &WorkerContextMiddleware{}
}

// Work wraps the worker execution to add context enrichment and duration tracking
func (m *WorkerContextMiddleware) Work(ctx context.Context, job *rivertype.JobRow, doInner func(ctx context.Context) error) error {
	// Enrich context with worker metadata
	ctx = EnrichWorkerContext(ctx, job.Kind, job.ID, job.Attempt)

	// Add additional job metadata as tags
	if job.Queue != "" && job.Queue != "default" {
		AddWorkerTag(ctx, "queue", job.Queue)
	}
	AddWorkerTag(ctx, "state", string(job.State))
	AddWorkerTag(ctx, "priority", strconv.Itoa(job.Priority))

	// Record start time for duration tracking
	startTime := time.Now()

	// Call the actual worker
	err := doInner(ctx)

	// Record execution duration
	duration := time.Since(startTime)
	RecordWorkerDuration(ctx, duration)

	// Add error information if the job failed
	if err != nil {
		AddWorkerContext(ctx, "error", map[string]interface{}{
			"error_message": err.Error(),
			"failed_after":  duration.String(),
		})
	}

	// Log completion status
	if err != nil {
		Logger.ErrorContext(ctx, "Worker job failed",
			"worker", job.Kind,
			"job_id", job.ID,
			"duration", duration,
			"error", err,
		)
	} else {
		Logger.InfoContext(ctx, "Worker job completed successfully",
			"worker", job.Kind,
			"job_id", job.ID,
			"duration", duration,
		)
	}

	return err
}

// Ensure our middleware implements the WorkerMiddleware interface
var _ rivertype.WorkerMiddleware = (*WorkerContextMiddleware)(nil)
