# Account deletion operations

Account deletion removes the user and owned learning records in one PostgreSQL
transaction. The same transaction inserts an `account_cleanup` River job for
the collision-safe `users/<user-id>-` profile-object prefix. The job starts
after five minutes, deletes current objects and all visible object versions,
and retries up to 25 times. Every failed attempt is reported through the River
error handler; the final attempt is identifiable by `attempt == max_attempts`.
If a profile upload succeeds but its database update fails, the API first
deletes that exact object synchronously. A transient failure schedules the same
worker queue with only the exact object name; it never uses the account-wide
prefix while the account remains active.

## Read-only health check

Run this query before release and during deletion-incident triage:

```sql
SELECT id, state, attempt, max_attempts, scheduled_at, attempted_at, errors
FROM river_job
WHERE kind = 'account_cleanup'
  AND state IN ('retryable', 'discarded')
ORDER BY scheduled_at;
```

A discarded row is an unresolved erasure obligation. Correct the object-store
problem, retry that River job through the existing administrative retry path,
and verify its exact object or prefix target no longer lists any current objects
or versions. Do not copy job arguments into tickets or alerts.

## Required owner sign-off before store submission

- Verify MinIO bucket versioning/lifecycle configuration and run one deletion
  against a disposable test account containing multiple image versions.
- Record the actual backup retention period and restore-time deletion procedure.
- Record PostHog and Sentry retention settings. Historical vendor events are
  governed by those settings; account deletion does not synchronously erase them.
- Configure an alert for discarded `account_cleanup` jobs or a final-attempt
  Sentry event, and test its routing.
