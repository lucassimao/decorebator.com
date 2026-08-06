package service

import (
	"context"
	"errors"
	"testing"
)

type profileCleanupJobRecorder struct {
	JobService
	objectName string
	err        error
}

func (r *profileCleanupJobRecorder) ScheduleProfileObjectCleanupJob(_ context.Context, objectName string) error {
	r.objectName = objectName
	return r.err
}

func TestCompensateProfileUploadSchedulesOnlyFailedObject(t *testing.T) {
	jobs := &profileCleanupJobRecorder{}
	service := &UserService{
		jobService: jobs,
		deleteProfileObject: func(context.Context, string, string) error {
			return errors.New("temporary object-store failure")
		},
	}

	err := service.CompensateProfileUpload(context.Background(), "users/42-failed.jpg")
	if err != nil {
		t.Fatalf("CompensateProfileUpload() error = %v", err)
	}
	if jobs.objectName != "users/42-failed.jpg" {
		t.Fatalf("scheduled object = %q, want exact failed object", jobs.objectName)
	}
}

func TestCompensateProfileUploadStopsAfterDirectDelete(t *testing.T) {
	jobs := &profileCleanupJobRecorder{}
	service := &UserService{
		jobService: jobs,
		deleteProfileObject: func(_ context.Context, bucket, objectName string) error {
			if bucket != "decorebator" || objectName != "users/42-failed.jpg" {
				t.Fatalf("delete target = %q, %q", bucket, objectName)
			}
			return nil
		},
	}

	if err := service.CompensateProfileUpload(context.Background(), "users/42-failed.jpg"); err != nil {
		t.Fatalf("CompensateProfileUpload() error = %v", err)
	}
	if jobs.objectName != "" {
		t.Fatalf("unexpected retry job for %q", jobs.objectName)
	}
}
