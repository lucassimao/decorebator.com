package service

import (
	"context"
	"fmt"
	"regexp"

	"decorebator.com/internal/common"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

const profileUploadReconciliationKind = "profile_upload_reconciliation_v1"

var profileObjectNamePattern = regexp.MustCompile(`^users/([1-9][0-9]*)-[0-9a-f]{32}\.(jpg|png)$`)

type ProfileUploadReconciliationArgs struct {
	UserID            int64  `json:"user_id"`
	Bucket            string `json:"bucket"`
	ObjectName        string `json:"object_name"`
	ObjectURL         string `json:"object_url,omitempty"`
	ObjectVersionID   string `json:"object_version_id,omitempty"`
	DeleteAllVersions bool   `json:"delete_all_versions,omitempty"`
}

func (ProfileUploadReconciliationArgs) Kind() string { return profileUploadReconciliationKind }

type ProfileUploadReconciliationWorker struct {
	river.WorkerDefaults[ProfileUploadReconciliationArgs]
	objectIsReferenced func(context.Context, int64, string) (bool, error)
	deleteVersion      func(context.Context, string, string, string) error
	deleteAllVersions  func(context.Context, string, string) error
}

func NewProfileUploadReconciliationWorker(db *pgxpool.Pool) *ProfileUploadReconciliationWorker {
	return &ProfileUploadReconciliationWorker{
		objectIsReferenced: func(ctx context.Context, userID int64, objectURL string) (bool, error) {
			var referenced bool
			err := db.QueryRow(ctx, `SELECT EXISTS(
				SELECT 1 FROM users WHERE id=$1 AND profile_picture_url=$2
			)`, userID, objectURL).Scan(&referenced)
			return referenced, err
		},
		deleteVersion:     common.MinIODeleteObjectVersion,
		deleteAllVersions: common.MinIODeleteObjectAllVersions,
	}
}

func (w *ProfileUploadReconciliationWorker) Work(ctx context.Context, job *river.Job[ProfileUploadReconciliationArgs]) error {
	args := job.Args
	if args.UserID <= 0 || !profileObjectNameBelongsToUser(args.ObjectName, args.UserID) {
		return river.JobCancel(fmt.Errorf("invalid profile reconciliation target"))
	}
	if args.Bucket == "" || !common.MinIOBucketAllowed(args.Bucket) {
		return river.JobCancel(fmt.Errorf("profile reconciliation bucket is not allowlisted"))
	}
	if !common.ValidPublicObjectURLForKey(args.ObjectURL, args.ObjectName) {
		return river.JobCancel(fmt.Errorf("profile reconciliation URL does not match its key"))
	}
	if args.DeleteAllVersions {
		if args.ObjectVersionID != "" {
			return river.JobCancel(fmt.Errorf("all-version cleanup cannot carry an object version"))
		}
		referenced, err := w.objectIsReferenced(ctx, args.UserID, args.ObjectURL)
		if err != nil {
			return fmt.Errorf("check planned profile object reference: %w", err)
		}
		if referenced {
			return nil
		}
		if err := w.deleteAllVersions(ctx, args.Bucket, args.ObjectName); err != nil {
			return fmt.Errorf("delete uncertain profile upload: %w", err)
		}
		return nil
	}
	referenced, err := w.objectIsReferenced(ctx, args.UserID, args.ObjectURL)
	if err != nil {
		return fmt.Errorf("check profile object reference: %w", err)
	}
	if referenced {
		return nil
	}
	if args.ObjectVersionID == "" {
		if err := w.deleteAllVersions(ctx, args.Bucket, args.ObjectName); err != nil {
			return fmt.Errorf("delete unversioned profile upload: %w", err)
		}
		return nil
	}
	if err := w.deleteVersion(ctx, args.Bucket, args.ObjectName, args.ObjectVersionID); err != nil {
		return fmt.Errorf("delete failed profile upload: %w", err)
	}
	return nil
}

func profileObjectNameBelongsToUser(objectName string, userID int64) bool {
	matches := profileObjectNamePattern.FindStringSubmatch(objectName)
	return len(matches) == 3 && matches[1] == fmt.Sprint(userID)
}
