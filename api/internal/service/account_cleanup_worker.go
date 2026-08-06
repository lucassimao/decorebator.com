package service

import (
	"context"
	"fmt"

	"decorebator.com/internal/common"
	"github.com/riverqueue/river"
)

const AccountCleanupQueue = "account_cleanup"
const profilePictureBucket = "decorebator"

type AccountCleanupArgs struct {
	ProfileObjectPrefix string `json:"profile_object_prefix"`
	ProfileObjectName   string `json:"profile_object_name"`
}

func (AccountCleanupArgs) Kind() string { return "account_cleanup" }

type profileObjectPrefixDeleter func(context.Context, string, string) error

type AccountCleanupWorker struct {
	river.WorkerDefaults[AccountCleanupArgs]
	deletePrefix profileObjectPrefixDeleter
	deleteObject profileObjectPrefixDeleter
}

func NewAccountCleanupWorker() *AccountCleanupWorker {
	return &AccountCleanupWorker{
		deletePrefix: common.MinIODeletePrefix,
		deleteObject: common.MinIODeleteObject,
	}
}

func (w *AccountCleanupWorker) Work(ctx context.Context, job *river.Job[AccountCleanupArgs]) error {
	hasPrefix := job.Args.ProfileObjectPrefix != ""
	hasObject := job.Args.ProfileObjectName != ""
	if hasPrefix == hasObject {
		return river.JobCancel(fmt.Errorf("exactly one profile object cleanup target is required"))
	}
	if hasObject {
		if err := w.deleteObject(ctx, profilePictureBucket, job.Args.ProfileObjectName); err != nil {
			return fmt.Errorf("delete failed profile object: %w", err)
		}
		return nil
	}
	if err := w.deletePrefix(ctx, profilePictureBucket, job.Args.ProfileObjectPrefix); err != nil {
		return fmt.Errorf("delete account profile objects: %w", err)
	}
	return nil
}
