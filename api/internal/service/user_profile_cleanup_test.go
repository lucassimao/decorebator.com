package service

import (
	"context"
	"errors"
	"testing"

	"decorebator.com/internal/common"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type profileCleanupJobRecorder struct {
	JobService
	args          ProfileUploadReconciliationArgs
	finalizedArgs ProfileUploadReconciliationArgs
	err           error
}

func (r *profileCleanupJobRecorder) ScheduleProfileUploadReconciliationJob(_ context.Context, args ProfileUploadReconciliationArgs) (int64, error) {
	r.args = args
	return 123, r.err
}

func (r *profileCleanupJobRecorder) FinalizeProfileUploadReconciliationJob(_ context.Context, jobID int64, args ProfileUploadReconciliationArgs, _ ...pgx.Tx) error {
	if jobID != 123 {
		return errors.New("unexpected cleanup intent ID")
	}
	r.finalizedArgs = args
	return r.err
}

func TestCompensateProfileUploadSchedulesOnlyFailedObject(t *testing.T) {
	jobs := &profileCleanupJobRecorder{}
	service := &UserService{
		jobService: jobs,
		profileObjectReferenced: func(context.Context, int64, string) (bool, error) {
			return false, nil
		},
		deleteProfileObject: func(context.Context, string, string, string) error {
			return errors.New("temporary object-store failure")
		},
	}
	receipt := common.UploadReceipt{Bucket: "decorebator", ObjectName: "users/42-failed.jpg", URL: "https://example.test/failed.jpg", VersionID: "v2"}

	err := service.CompensateProfileUpload(context.Background(), 42, receipt)
	require.NoError(t, err)
	if jobs.args.UserID != 42 || jobs.args.ObjectName != receipt.ObjectName || jobs.args.ObjectURL != receipt.URL || jobs.args.ObjectVersionID != receipt.VersionID {
		t.Fatalf("scheduled cleanup = %#v", jobs.args)
	}
}

func TestCompensateProfileUploadStopsAfterDirectVersionDelete(t *testing.T) {
	jobs := &profileCleanupJobRecorder{}
	service := &UserService{
		jobService: jobs,
		profileObjectReferenced: func(context.Context, int64, string) (bool, error) {
			return false, nil
		},
		deleteProfileObject: func(_ context.Context, bucket, objectName, versionID string) error {
			if bucket != "decorebator" || objectName != "users/42-failed.jpg" || versionID != "v2" {
				t.Fatalf("delete target = %q, %q, %q", bucket, objectName, versionID)
			}
			return nil
		},
	}
	receipt := common.UploadReceipt{Bucket: "decorebator", ObjectName: "users/42-failed.jpg", URL: "https://example.test/failed.jpg", VersionID: "v2"}

	require.NoError(t, service.CompensateProfileUpload(context.Background(), 42, receipt))
	if jobs.args.ObjectName != "" {
		t.Fatalf("unexpected retry job for %q", jobs.args.ObjectName)
	}
}

func TestCompensateProfileUploadKeepsReferencedObject(t *testing.T) {
	service := &UserService{
		profileObjectReferenced: func(_ context.Context, userID int64, objectURL string) (bool, error) {
			if userID != 42 || objectURL != "https://example.test/committed.jpg" {
				t.Fatalf("reference check = %d, %q", userID, objectURL)
			}
			return true, nil
		},
		deleteProfileObject: func(context.Context, string, string, string) error {
			t.Fatal("committed object must not be deleted")
			return nil
		},
	}
	receipt := common.UploadReceipt{Bucket: "decorebator", ObjectName: "users/42-committed.jpg", URL: "https://example.test/committed.jpg", VersionID: "v3"}
	require.NoError(t, service.CompensateProfileUpload(context.Background(), 42, receipt))
}

func TestScheduleUncertainProfileUploadCleanupUsesVersionedAllVersionsContract(t *testing.T) {
	jobs := &profileCleanupJobRecorder{}
	service := &UserService{jobService: jobs}
	jobID, err := service.ScheduleUncertainProfileUploadCleanup(
		context.Background(), 42, "decorebator", "users/42-0123456789abcdef0123456789abcdef.jpg",
		"https://example.test/users/42-planned.jpg",
	)
	require.NoError(t, err)
	assert.Equal(t, int64(123), jobID)
	assert.Equal(t, ProfileUploadReconciliationArgs{
		UserID: 42, Bucket: "decorebator", ObjectName: "users/42-0123456789abcdef0123456789abcdef.jpg",
		ObjectURL: "https://example.test/users/42-planned.jpg", DeleteAllVersions: true,
	}, jobs.args)
}
