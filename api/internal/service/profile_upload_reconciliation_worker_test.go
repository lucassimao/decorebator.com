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

func TestProfileUploadReconciliationKindCannotBeConsumedAsLegacyCleanup(t *testing.T) {
	t.Parallel()
	assert.NotEqual(t, (AccountCleanupArgs{}).Kind(), (ProfileUploadReconciliationArgs{}).Kind())
}

func TestProfileUploadReconciliationKeepsCommittedReceipt(t *testing.T) {
	require.NoError(t, common.ConfigureMinIO(common.MinIOConfig{
		Endpoint: "localhost:9000", AccessKey: "test-user", SecretKey: "test-secret", Bucket: "decorebator",
	}))
	objectName := "users/42-0123456789abcdef0123456789abcdef.jpg"
	expectedURL, err := common.MinIOObjectURL(common.MinIOBucketName(), objectName)
	require.NoError(t, err)
	worker := &ProfileUploadReconciliationWorker{
		objectIsReferenced: func(_ context.Context, userID int64, objectURL string) (bool, error) {
			assert.Equal(t, int64(42), userID)
			assert.Equal(t, expectedURL, objectURL)
			return true, nil
		},
		deleteVersion: func(context.Context, string, string, string) error {
			t.Fatal("committed object must not be deleted")
			return nil
		},
	}
	err = worker.Work(context.Background(), &river.Job[ProfileUploadReconciliationArgs]{Args: ProfileUploadReconciliationArgs{
		UserID: 42, Bucket: common.MinIOBucketName(), ObjectName: objectName,
		ObjectURL: expectedURL, ObjectVersionID: "v3",
	}})
	require.NoError(t, err)
}

func TestProfileUploadReconciliationDeletesExactUnreferencedVersion(t *testing.T) {
	require.NoError(t, common.ConfigureMinIO(common.MinIOConfig{
		Endpoint: "localhost:9000", AccessKey: "test-user", SecretKey: "test-secret", Bucket: "decorebator",
	}))
	objectName := "users/42-0123456789abcdef0123456789abcdef.png"
	objectURL, err := common.MinIOObjectURL(common.MinIOBucketName(), objectName)
	require.NoError(t, err)
	worker := &ProfileUploadReconciliationWorker{
		objectIsReferenced: func(context.Context, int64, string) (bool, error) { return false, nil },
		deleteVersion: func(_ context.Context, bucket, objectName, versionID string) error {
			assert.Equal(t, common.MinIOBucketName(), bucket)
			assert.Equal(t, "users/42-0123456789abcdef0123456789abcdef.png", objectName)
			assert.Equal(t, "v2", versionID)
			return nil
		},
	}
	err = worker.Work(context.Background(), &river.Job[ProfileUploadReconciliationArgs]{Args: ProfileUploadReconciliationArgs{
		UserID: 42, Bucket: common.MinIOBucketName(), ObjectName: objectName,
		ObjectURL: objectURL, ObjectVersionID: "v2",
	}})
	require.NoError(t, err)
}

func TestProfileUploadReconciliationKeepsReferencedUnversionedObject(t *testing.T) {
	require.NoError(t, common.ConfigureMinIO(common.MinIOConfig{
		Endpoint: "localhost:9000", AccessKey: "test-user", SecretKey: "test-secret", Bucket: "decorebator",
	}))
	objectName := "users/42-0123456789abcdef0123456789abcdef.jpg"
	objectURL, err := common.MinIOObjectURL(common.MinIOBucketName(), objectName)
	require.NoError(t, err)
	worker := &ProfileUploadReconciliationWorker{
		objectIsReferenced: func(context.Context, int64, string) (bool, error) { return true, nil },
		deleteAllVersions: func(context.Context, string, string) error {
			t.Fatal("referenced unversioned object must not be deleted")
			return nil
		},
	}
	require.NoError(t, worker.Work(context.Background(), &river.Job[ProfileUploadReconciliationArgs]{Args: ProfileUploadReconciliationArgs{
		UserID: 42, Bucket: common.MinIOBucketName(), ObjectName: objectName, ObjectURL: objectURL,
	}}))
}

func TestProfileUploadReconciliationDeletesOnlyExactUnreferencedUnversionedKey(t *testing.T) {
	require.NoError(t, common.ConfigureMinIO(common.MinIOConfig{
		Endpoint: "localhost:9000", AccessKey: "test-user", SecretKey: "test-secret", Bucket: "decorebator",
	}))
	objectName := "users/42-0123456789abcdef0123456789abcdef.png"
	objectURL, err := common.MinIOObjectURL(common.MinIOBucketName(), objectName)
	require.NoError(t, err)
	worker := &ProfileUploadReconciliationWorker{
		objectIsReferenced: func(context.Context, int64, string) (bool, error) { return false, nil },
		deleteVersion: func(context.Context, string, string, string) error {
			t.Fatal("an empty version ID must use exact-key all-version deletion")
			return nil
		},
		deleteAllVersions: func(_ context.Context, bucket, gotObjectName string) error {
			assert.Equal(t, common.MinIOBucketName(), bucket)
			assert.Equal(t, objectName, gotObjectName)
			return nil
		},
	}
	require.NoError(t, worker.Work(context.Background(), &river.Job[ProfileUploadReconciliationArgs]{Args: ProfileUploadReconciliationArgs{
		UserID: 42, Bucket: common.MinIOBucketName(), ObjectName: objectName, ObjectURL: objectURL,
	}}))
}

func TestProfileUploadReconciliationRejectsExactVersionURLKeyMismatch(t *testing.T) {
	require.NoError(t, common.ConfigureMinIO(common.MinIOConfig{
		Endpoint: "localhost:9000", AccessKey: "test-user", SecretKey: "test-secret", Bucket: "decorebator",
	}))
	otherURL, err := common.MinIOObjectURL(
		common.MinIOBucketName(),
		"users/42-fedcba9876543210fedcba9876543210.jpg",
	)
	require.NoError(t, err)
	worker := &ProfileUploadReconciliationWorker{
		objectIsReferenced: func(context.Context, int64, string) (bool, error) {
			t.Fatal("mismatched exact-version URL must fail before reference lookup")
			return false, nil
		},
		deleteVersion: func(context.Context, string, string, string) error {
			t.Fatal("mismatched exact-version URL must never delete")
			return nil
		},
	}
	err = worker.Work(context.Background(), &river.Job[ProfileUploadReconciliationArgs]{Args: ProfileUploadReconciliationArgs{
		UserID: 42, Bucket: common.MinIOBucketName(),
		ObjectName: "users/42-0123456789abcdef0123456789abcdef.jpg",
		ObjectURL:  otherURL, ObjectVersionID: "v2",
	}})
	var cancel *river.JobCancelError
	require.True(t, errors.As(err, &cancel))
}

func TestProfileUploadReconciliationDeletesAllVersionsOnlyForValidatedRandomKey(t *testing.T) {
	require.NoError(t, common.ConfigureMinIO(common.MinIOConfig{
		Endpoint: "localhost:9000", AccessKey: "test-user", SecretKey: "test-secret", Bucket: "decorebator",
	}))
	objectName := "users/42-0123456789abcdef0123456789abcdef.jpg"
	objectURL, err := common.MinIOObjectURL("decorebator", objectName)
	require.NoError(t, err)
	deleted := false
	worker := &ProfileUploadReconciliationWorker{
		objectIsReferenced: func(context.Context, int64, string) (bool, error) { return false, nil },
		deleteAllVersions: func(_ context.Context, bucket, objectName string) error {
			deleted = true
			assert.Equal(t, common.MinIOBucketName(), bucket)
			assert.Equal(t, "users/42-0123456789abcdef0123456789abcdef.jpg", objectName)
			return nil
		},
	}
	err = worker.Work(context.Background(), &river.Job[ProfileUploadReconciliationArgs]{Args: ProfileUploadReconciliationArgs{
		UserID: 42, Bucket: common.MinIOBucketName(), ObjectName: objectName, ObjectURL: objectURL, DeleteAllVersions: true,
	}})
	require.NoError(t, err)
	assert.True(t, deleted)

	err = worker.Work(context.Background(), &river.Job[ProfileUploadReconciliationArgs]{Args: ProfileUploadReconciliationArgs{
		UserID: 42, Bucket: common.MinIOBucketName(), ObjectName: "users/42-../../other.jpg", ObjectURL: objectURL, DeleteAllVersions: true,
	}})
	var cancel *river.JobCancelError
	require.True(t, errors.As(err, &cancel))
}

func TestProfileUploadReconciliationRejectsMissingOrMismatchedPlannedURL(t *testing.T) {
	require.NoError(t, common.ConfigureMinIO(common.MinIOConfig{
		Endpoint: "localhost:9000", AccessKey: "test-user", SecretKey: "test-secret", Bucket: "decorebator",
	}))
	worker := &ProfileUploadReconciliationWorker{
		objectIsReferenced: func(context.Context, int64, string) (bool, error) {
			t.Fatal("invalid URL must be rejected before reference access")
			return false, nil
		},
		deleteAllVersions: func(context.Context, string, string) error {
			t.Fatal("invalid URL must not reach storage")
			return nil
		},
	}
	for _, objectURL := range []string{
		"",
		"http://localhost:9000/decorebator/users/42-other.jpg",
		"https://user:password@example.test/decorebator/users/42-0123456789abcdef0123456789abcdef.jpg",
	} {
		err := worker.Work(context.Background(), &river.Job[ProfileUploadReconciliationArgs]{Args: ProfileUploadReconciliationArgs{
			UserID: 42, Bucket: "decorebator", ObjectName: "users/42-0123456789abcdef0123456789abcdef.jpg",
			ObjectURL: objectURL, DeleteAllVersions: true,
		}})
		var cancel *river.JobCancelError
		require.True(t, errors.As(err, &cancel))
	}
}

func TestProfileUploadReconciliationKeepsEnqueueTimeURLAcrossStorageRotation(t *testing.T) {
	t.Cleanup(func() {
		require.NoError(t, common.ConfigureMinIO(common.MinIOConfig{
			Endpoint: "localhost:9000", AccessKey: "test-user", SecretKey: "test-secret", Bucket: "decorebator",
		}))
	})
	objectName := "users/42-0123456789abcdef0123456789abcdef.jpg"
	require.NoError(t, common.ConfigureMinIO(common.MinIOConfig{
		Endpoint: "old.internal:9000", AccessKey: "test-user", SecretKey: "test-secret",
		Bucket: "former-media", PublicBaseURL: "https://old-media.example.test",
	}))
	objectURL, err := common.MinIOObjectURL("former-media", objectName)
	require.NoError(t, err)
	require.NoError(t, common.ConfigureMinIO(common.MinIOConfig{
		Endpoint: "new.internal:9000", AccessKey: "test-user", SecretKey: "test-secret",
		Bucket: "current-media", LegacyBuckets: []string{"former-media"}, PublicBaseURL: "https://new-media.example.test",
	}))
	worker := &ProfileUploadReconciliationWorker{
		objectIsReferenced: func(_ context.Context, userID int64, gotURL string) (bool, error) {
			assert.Equal(t, int64(42), userID)
			assert.Equal(t, objectURL, gotURL)
			return true, nil
		},
		deleteAllVersions: func(context.Context, string, string) error {
			t.Fatal("referenced enqueue-time URL must preserve the object")
			return nil
		},
	}
	require.NoError(t, worker.Work(context.Background(), &river.Job[ProfileUploadReconciliationArgs]{Args: ProfileUploadReconciliationArgs{
		UserID: 42, Bucket: "former-media", ObjectName: objectName, ObjectURL: objectURL, DeleteAllVersions: true,
	}}))
}
