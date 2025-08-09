package common

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/getsentry/sentry-go"
)

// WorkerContextKey is the key for worker metadata in context
type WorkerContextKey struct{}

// WorkerMetadata contains metadata about the worker execution environment
type WorkerMetadata struct {
	WorkerKind string    // The type of worker (e.g., "DefinitionFetcher")
	JobID      int64     // The River job ID
	JobAttempt int       // Current attempt number
	StartTime  time.Time // When the job started
}

// EnrichWorkerContext adds worker metadata to the context and configures Sentry scope
func EnrichWorkerContext(ctx context.Context, workerKind string, jobID int64, jobAttempt int) context.Context {
	metadata := &WorkerMetadata{
		WorkerKind: workerKind,
		JobID:      jobID,
		JobAttempt: jobAttempt,
		StartTime:  time.Now(),
	}

	// Add metadata to context
	ctx = context.WithValue(ctx, WorkerContextKey{}, metadata)

	// Configure Sentry scope with worker metadata
	if hub := sentry.GetHubFromContext(ctx); hub == nil {
		// Create a new hub if none exists
		hub = sentry.CurrentHub().Clone()
		ctx = sentry.SetHubOnContext(ctx, hub)
	}

	hub := sentry.GetHubFromContext(ctx)
	hub.ConfigureScope(func(scope *sentry.Scope) {
		// Set worker-specific tags
		scope.SetTag("worker.kind", workerKind)
		scope.SetTag("worker.job_id", fmt.Sprintf("%d", jobID))
		scope.SetTag("worker.attempt", fmt.Sprintf("%d", jobAttempt))
		scope.SetTag("environment", os.Getenv("ENV"))

		// Set worker context
		scope.SetContext("worker", map[string]interface{}{
			"kind":        workerKind,
			"job_id":      jobID,
			"job_attempt": jobAttempt,
			"started_at":  metadata.StartTime.Format(time.RFC3339),
		})

		// Set runtime context
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		scope.SetContext("runtime", map[string]interface{}{
			"go_version":      runtime.Version(),
			"num_cpu":         runtime.NumCPU(),
			"num_goroutine":   runtime.NumGoroutine(),
			"memory_alloc_mb": memStats.Alloc / 1024 / 1024,
			"memory_sys_mb":   memStats.Sys / 1024 / 1024,
			"gc_runs":         memStats.NumGC,
		})
	})

	return ctx
}

// GetWorkerMetadata retrieves worker metadata from context
func GetWorkerMetadata(ctx context.Context) *WorkerMetadata {
	if metadata, ok := ctx.Value(WorkerContextKey{}).(*WorkerMetadata); ok {
		return metadata
	}
	return nil
}

// RecordWorkerDuration records how long a worker job took to complete
func RecordWorkerDuration(ctx context.Context, duration time.Duration) {
	if metadata := GetWorkerMetadata(ctx); metadata != nil {
		if hub := sentry.GetHubFromContext(ctx); hub != nil {
			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetContext("performance", map[string]interface{}{
					"duration_ms":       duration.Milliseconds(),
					"duration":          duration.String(),
					"completed_at":      time.Now().Format(time.RFC3339),
					"total_duration_ms": time.Since(metadata.StartTime).Milliseconds(),
				})
			})
		}
	}
}

// AddWorkerTag adds a custom tag to the Sentry scope for the current worker
func AddWorkerTag(ctx context.Context, key, value string) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("worker.custom."+key, value)
		})
	}
}

// AddWorkerContext adds custom context data to the Sentry scope for the current worker
func AddWorkerContext(ctx context.Context, contextKey string, data map[string]interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetContext("worker."+contextKey, data)
		})
	}
}
