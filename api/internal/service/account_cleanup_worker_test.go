package service

import (
	"context"
	"errors"
	"testing"

	"decorebator.com/internal/common"
	"github.com/riverqueue/river"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountCleanupWorkerUsesCapturedLegacyBucket(t *testing.T) {
	require.NoError(t, common.ConfigureMinIO(common.MinIOConfig{
		Endpoint: "localhost:9000", AccessKey: "test-user", SecretKey: "test-secret",
		Bucket: "current-media", LegacyBuckets: []string{"former-media"},
	}))
	t.Cleanup(func() {
		require.NoError(t, common.ConfigureMinIO(common.MinIOConfig{
			Endpoint: "localhost:9000", AccessKey: "test-user", SecretKey: "test-secret", Bucket: "decorebator",
		}))
	})
	var buckets []string
	worker := &AccountCleanupWorker{deletePrefix: func(_ context.Context, bucket, prefix string) error {
		buckets = append(buckets, bucket)
		assert.Equal(t, "users/42-", prefix)
		return nil
	}}
	require.NoError(t, worker.Work(context.Background(), &river.Job[AccountCleanupArgs]{Args: AccountCleanupArgs{
		ProfileBuckets: []string{"current-media", "former-media"}, ProfileObjectPrefix: "users/42-",
	}}))
	assert.Equal(t, []string{"current-media", "former-media"}, buckets)
}

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

func TestAccountCleanupWorkerQuarantinesLegacyNameOnlyJob(t *testing.T) {
	worker := &AccountCleanupWorker{deletePrefix: func(context.Context, string, string) error {
		t.Fatal("legacy name-only cleanup must not delete storage")
		return nil
	}}
	err := worker.Work(context.Background(), &river.Job[AccountCleanupArgs]{Args: AccountCleanupArgs{
		ProfileObjectName: "users/42-failed.jpg",
	}})
	var cancel *river.JobCancelError
	if !errors.As(err, &cancel) {
		t.Fatalf("Work() error = %v, want JobCancelError", err)
	}
}

func TestAccountCleanupWorkerPreservesReferencedLegacyObject(t *testing.T) {
	worker := &AccountCleanupWorker{
		legacyObjectReferenced: func(_ context.Context, userID int64, objectName string) (bool, error) {
			assert.Equal(t, int64(42), userID)
			assert.Equal(t, "users/42-1700000000.jpg", objectName)
			return true, nil
		},
		deleteAllVersions: func(context.Context, string, string) error {
			t.Fatal("referenced legacy object must not be deleted")
			return nil
		},
	}
	require.NoError(t, worker.Work(context.Background(), &river.Job[AccountCleanupArgs]{Args: AccountCleanupArgs{
		ProfileObjectName: "users/42-1700000000.jpg",
	}}))
}

func TestAccountCleanupWorkerDeletesUnreferencedLegacyObjectFromEveryBucket(t *testing.T) {
	require.NoError(t, common.ConfigureMinIO(common.MinIOConfig{
		Endpoint: "localhost:9000", AccessKey: "test-user", SecretKey: "test-secret",
		Bucket: "current-media", LegacyBuckets: []string{"former-media"},
	}))
	t.Cleanup(func() {
		require.NoError(t, common.ConfigureMinIO(common.MinIOConfig{
			Endpoint: "localhost:9000", AccessKey: "test-user", SecretKey: "test-secret", Bucket: "decorebator",
		}))
	})
	var deletedBuckets []string
	worker := &AccountCleanupWorker{
		legacyObjectReferenced: func(context.Context, int64, string) (bool, error) { return false, nil },
		deleteAllVersions: func(_ context.Context, bucket, objectName string) error {
			assert.Equal(t, "users/42-1700000000.png", objectName)
			deletedBuckets = append(deletedBuckets, bucket)
			return nil
		},
	}
	require.NoError(t, worker.Work(context.Background(), &river.Job[AccountCleanupArgs]{Args: AccountCleanupArgs{
		ProfileObjectName: "users/42-1700000000.png",
	}}))
	assert.Equal(t, []string{"current-media", "former-media"}, deletedBuckets)
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

func TestAccountCleanupWorkerRejectsUnsafePrefixesBeforeAnyDelete(t *testing.T) {
	worker := &AccountCleanupWorker{deletePrefix: func(context.Context, string, string) error {
		t.Fatal("unsafe prefix must not reach storage")
		return nil
	}}
	for _, prefix := range []string{"users/", "users/0-", "users/42", "users/42-/../", "../users/42-"} {
		err := worker.Work(context.Background(), &river.Job[AccountCleanupArgs]{Args: AccountCleanupArgs{
			ProfileBuckets: []string{common.MinIOBucketName()}, ProfileObjectPrefix: prefix,
		}})
		var cancel *river.JobCancelError
		require.True(t, errors.As(err, &cancel), prefix)
	}
}

func TestAccountCleanupWorkerValidatesEveryBucketBeforeDeleting(t *testing.T) {
	worker := &AccountCleanupWorker{deletePrefix: func(context.Context, string, string) error {
		t.Fatal("no bucket may be deleted when any target is invalid")
		return nil
	}}
	err := worker.Work(context.Background(), &river.Job[AccountCleanupArgs]{Args: AccountCleanupArgs{
		ProfileBuckets: []string{common.MinIOBucketName(), "not-allowlisted"}, ProfileObjectPrefix: "users/42-",
	}})
	var cancel *river.JobCancelError
	require.True(t, errors.As(err, &cancel))
}
