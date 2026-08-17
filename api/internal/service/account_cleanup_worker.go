package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"decorebator.com/internal/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

const AccountCleanupQueue = "account_cleanup"

var profileObjectPrefixPattern = regexp.MustCompile(`^users/[1-9][0-9]*-$`)
var legacyProfileObjectPattern = regexp.MustCompile(`^users/([1-9][0-9]*)-([1-9][0-9]*)\.(jpg|jpeg|png)$`)

type AccountCleanupArgs struct {
	ProfileBucket       string   `json:"profile_bucket,omitempty"`
	ProfileBuckets      []string `json:"profile_buckets,omitempty"`
	ProfileObjectPrefix string   `json:"profile_object_prefix"`
	ProfileObjectName   string   `json:"profile_object_name"`
}

func (AccountCleanupArgs) Kind() string { return "account_cleanup" }

type profileObjectPrefixDeleter func(context.Context, string, string) error
type legacyProfileObjectReferenceChecker func(context.Context, int64, string) (bool, error)

type AccountCleanupWorker struct {
	river.WorkerDefaults[AccountCleanupArgs]
	deletePrefix           profileObjectPrefixDeleter
	deleteAllVersions      profileObjectPrefixDeleter
	legacyObjectReferenced legacyProfileObjectReferenceChecker
}

func NewAccountCleanupWorker(db *pgxpool.Pool) *AccountCleanupWorker {
	return &AccountCleanupWorker{
		deletePrefix:      common.MinIODeletePrefix,
		deleteAllVersions: common.MinIODeleteObjectAllVersions,
		legacyObjectReferenced: func(ctx context.Context, userID int64, objectName string) (bool, error) {
			var profileURL string
			err := db.QueryRow(ctx, `SELECT COALESCE(profile_picture_url, '') FROM users WHERE id=$1`, userID).Scan(&profileURL)
			if errors.Is(err, pgx.ErrNoRows) {
				return false, nil
			}
			if err != nil {
				return false, err
			}
			return profileURL != "" && common.ValidPublicObjectURLForKey(profileURL, objectName), nil
		},
	}
}

func (w *AccountCleanupWorker) Work(ctx context.Context, job *river.Job[AccountCleanupArgs]) error {
	hasPrefix := job.Args.ProfileObjectPrefix != ""
	hasObject := job.Args.ProfileObjectName != ""
	if hasPrefix == hasObject {
		return river.JobCancel(fmt.Errorf("exactly one profile object cleanup target is required"))
	}
	if hasObject {
		return w.cleanupLegacyObject(ctx, job.Args.ProfileObjectName)
	}
	if !profileObjectPrefixPattern.MatchString(job.Args.ProfileObjectPrefix) {
		return river.JobCancel(fmt.Errorf("invalid account profile cleanup prefix"))
	}
	buckets := job.Args.ProfileBuckets
	if len(buckets) == 0 && job.Args.ProfileBucket != "" {
		buckets = []string{job.Args.ProfileBucket}
	}
	if len(buckets) == 0 {
		// Pre-upgrade prefix jobs had no bucket provenance. During cutover the
		// configured current/legacy allowlist is the complete safe cleanup set.
		buckets = common.MinIOCleanupBuckets()
	}
	for _, bucket := range buckets {
		if !common.MinIOBucketAllowed(bucket) {
			return river.JobCancel(fmt.Errorf("account cleanup bucket is not allowlisted"))
		}
	}
	for _, bucket := range buckets {
		if err := w.deletePrefix(ctx, bucket, job.Args.ProfileObjectPrefix); err != nil {
			return fmt.Errorf("delete account profile objects from %q: %w", bucket, err)
		}
	}
	return nil
}

func (w *AccountCleanupWorker) cleanupLegacyObject(ctx context.Context, objectName string) error {
	matches := legacyProfileObjectPattern.FindStringSubmatch(objectName)
	if len(matches) != 4 || w.legacyObjectReferenced == nil || w.deleteAllVersions == nil {
		return river.JobCancel(fmt.Errorf("legacy name-only profile cleanup quarantined: %q", objectName))
	}
	userID, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil || userID <= 0 {
		return river.JobCancel(fmt.Errorf("invalid legacy profile object owner"))
	}
	buckets := common.MinIOCleanupBuckets()
	for _, bucket := range buckets {
		if !common.MinIOBucketAllowed(bucket) {
			return river.JobCancel(fmt.Errorf("legacy profile cleanup bucket is not allowlisted"))
		}
	}
	referenced, err := w.legacyObjectReferenced(ctx, userID, objectName)
	if err != nil {
		return fmt.Errorf("check legacy profile object reference: %w", err)
	}
	if referenced {
		return nil
	}
	for _, bucket := range buckets {
		if err := w.deleteAllVersions(ctx, bucket, objectName); err != nil {
			return fmt.Errorf("delete legacy profile object from %q: %w", bucket, err)
		}
	}
	return nil
}
