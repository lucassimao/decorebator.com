# Account deletion operations

Account deletion removes the user and owned learning records in one PostgreSQL
transaction. The same transaction inserts an `account_cleanup` River job for
the captured current/legacy bucket set and collision-safe `users/<user-id>-`
profile-object prefix. The job starts
after five minutes, deletes current objects and all visible object versions,
and retries up to 25 times. Every failed attempt is reported through the River
error handler; the final attempt is identifiable by `attempt == max_attempts`.
If a profile upload succeeds but its database update fails, the API first
deletes that exact bucket/key/version synchronously. Every uncertain PUT or
ambiguous persistence outcome schedules `profile_upload_reconciliation_v1` in
the same queue with the captured bucket and exact random object key; it never
uses the account-wide prefix while the account remains active. Deploy the new
worker before API producers. Before changing `MINIO_BUCKET`, add the former
bucket to `MINIO_LEGACY_BUCKETS` and retain it until both cleanup job kinds have
zero `state <> 'completed'` rows referencing that bucket.
Pre-upgrade prefix jobs without captured bucket provenance intentionally run
against the complete configured current/legacy set, so that set must be
provisioned before deploying the worker and retained until those jobs drain.
Pre-upgrade name-only jobs matching the former
`users/<user-id>-<unix-seconds>.(jpg|jpeg|png)` contract are also drained: the
worker checks the current user's exact URL suffix first, preserves referenced
objects, and otherwise deletes every version of only that exact key from the
complete configured bucket set. Other name-only shapes remain quarantined.

## Read-only health check

Run this query before release and during deletion-incident triage:

```sql
SELECT id, state, attempt, max_attempts, scheduled_at, attempted_at, errors
FROM river_job
WHERE kind IN ('account_cleanup', 'profile_upload_reconciliation_v1')
  AND state <> 'completed'
ORDER BY scheduled_at;
```

Every scheduled, available, pending, running, retryable, discarded, or
cancelled row is an outstanding obligation; cancellation is not successful
cleanup. Correct the object-store
problem, ensure the captured bucket remains explicitly allowlisted, retry that
River job through the existing administrative retry path, and verify its exact
bucket/key/version or bucket/prefix target no longer lists any current objects
or versions. Uncertain-PUT jobs intentionally delete every version of only the
validated random exact key; persisted receipts delete only their captured
version after the database reference check. Do not copy job arguments into
tickets or alerts.

Use this bucket-scoped cutover count. Every row for the legacy bucket and the
`<missing-provenance>` row must reach zero before removing that bucket from the
allowlist; a missing-provenance job can still target every configured bucket:

```sql
WITH cleanup_jobs AS (
  SELECT id, state, args
  FROM river_job
  WHERE kind IN ('account_cleanup', 'profile_upload_reconciliation_v1')
    AND state <> 'completed'
), cleanup_buckets AS (
  SELECT id, state, jsonb_array_elements_text(args->'profile_buckets') AS bucket
  FROM cleanup_jobs
  WHERE jsonb_typeof(args->'profile_buckets') = 'array'
  UNION ALL
  SELECT id, state, COALESCE(args->>'bucket', args->>'profile_bucket') AS bucket
  FROM cleanup_jobs
  WHERE COALESCE(args->>'bucket', args->>'profile_bucket') IS NOT NULL
), outstanding AS (
  SELECT bucket, state, COUNT(*) AS outstanding_jobs
  FROM cleanup_buckets
  GROUP BY bucket, state
  UNION ALL
  SELECT '<missing-provenance>' AS bucket, state, COUNT(*) AS outstanding_jobs
  FROM cleanup_jobs
  WHERE NOT COALESCE((
    jsonb_typeof(args->'profile_buckets') = 'array'
    AND jsonb_array_length(args->'profile_buckets') > 0
  ), false)
    AND COALESCE(args->>'bucket', args->>'profile_bucket') IS NULL
  GROUP BY state
)
SELECT bucket, state, outstanding_jobs
FROM outstanding
ORDER BY bucket, state;
```

## Legacy name-only deployment gate

Before deploying this worker, this query must return zero. It identifies
legacy name-only rows that cannot be safely converted by the automated,
reference-aware compatibility path:

```sql
SELECT id, state, args->>'profile_object_name' AS object_name
FROM river_job
WHERE kind = 'account_cleanup'
  AND state <> 'completed'
  AND COALESCE(args->>'profile_object_name', '') <> ''
  AND (args->>'profile_object_name') !~
      '^users/[1-9][0-9]*-[1-9][0-9]*\.(jpg|jpeg|png)$'
ORDER BY id;
```

If any row exists, stop the deployment and discharge it through a separately
reviewed one-off migration. For each exact row, enumerate that exact key in
every configured bucket, construct its candidate public URLs without copying
credentials, and check `users.profile_picture_url` for exact equality with
every candidate. Preserve the object when any user references it. Otherwise,
delete every version of only that exact key and independently verify absence.
After either proof, remove only the guarded River row by exact `id`, `kind`,
current `state`, and unchanged `args`, recording the affected-row count. Never
retry a quarantined row unchanged, broaden the key to a prefix, or remove a
legacy bucket while this query or the general health check is nonzero.

## Required owner sign-off before store submission

- Verify MinIO bucket versioning/lifecycle configuration and run one deletion
  against a disposable test account containing multiple image versions.
- Record the actual backup retention period and restore-time deletion procedure.
- Record PostHog and Sentry retention settings. Historical vendor events are
  governed by those settings; account deletion does not synchronously erase them.
- Configure an alert for discarded `account_cleanup` and
  `profile_upload_reconciliation_v1` jobs or either final-attempt Sentry event,
  and test its routing.
