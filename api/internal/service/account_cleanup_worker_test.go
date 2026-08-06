package service

import (
	"context"
	"errors"
	"testing"

	"github.com/riverqueue/river"
)

func TestAccountCleanupWorkerDeletesExactPrefix(t *testing.T) {
	var bucket, prefix string
	worker := &AccountCleanupWorker{deletePrefix: func(_ context.Context, gotBucket, gotPrefix string) error {
		bucket, prefix = gotBucket, gotPrefix
		return nil
	}}

	err := worker.Work(context.Background(), &river.Job[AccountCleanupArgs]{Args: AccountCleanupArgs{
		ProfileObjectPrefix: "users/42-",
	}})
	if err != nil {
		t.Fatalf("Work() error = %v", err)
	}
	if bucket != "decorebator" || prefix != "users/42-" {
		t.Fatalf("deletePrefix() = %q, %q", bucket, prefix)
	}
}

func TestAccountCleanupWorkerReturnsStorageErrorForRiverRetry(t *testing.T) {
	want := errors.New("storage unavailable")
	worker := &AccountCleanupWorker{deletePrefix: func(context.Context, string, string) error { return want }}
	err := worker.Work(context.Background(), &river.Job[AccountCleanupArgs]{Args: AccountCleanupArgs{
		ProfileObjectPrefix: "users/42-",
	}})
	if !errors.Is(err, want) {
		t.Fatalf("Work() error = %v, want wrapped storage error", err)
	}
}

func TestAccountCleanupWorkerDeletesOnlyFailedObject(t *testing.T) {
	var bucket, objectName string
	worker := &AccountCleanupWorker{
		deletePrefix: func(context.Context, string, string) error {
			t.Fatal("prefix deletion must not run for a failed-upload cleanup")
			return nil
		},
		deleteObject: func(_ context.Context, gotBucket, gotObject string) error {
			bucket, objectName = gotBucket, gotObject
			return nil
		},
	}

	err := worker.Work(context.Background(), &river.Job[AccountCleanupArgs]{Args: AccountCleanupArgs{
		ProfileObjectName: "users/42-failed.jpg",
	}})
	if err != nil {
		t.Fatalf("Work() error = %v", err)
	}
	if bucket != "decorebator" || objectName != "users/42-failed.jpg" {
		t.Fatalf("deleteObject() = %q, %q", bucket, objectName)
	}
}

func TestAccountCleanupWorkerRejectsAmbiguousTarget(t *testing.T) {
	worker := &AccountCleanupWorker{}
	err := worker.Work(context.Background(), &river.Job[AccountCleanupArgs]{Args: AccountCleanupArgs{
		ProfileObjectPrefix: "users/42-",
		ProfileObjectName:   "users/42-failed.jpg",
	}})
	var cancel *river.JobCancelError
	if !errors.As(err, &cancel) {
		t.Fatalf("Work() error = %v, want JobCancelError", err)
	}
}
