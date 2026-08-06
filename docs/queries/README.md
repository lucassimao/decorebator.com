# Read-only reports

Run `./docs/queries/run-activation-cohort-report.sh` from the repository root to produce the database baseline for 7-, 30-, 90-, and 365-day signup cohorts. The script reads the controlled store-approval registry and performs a `SELECT` only.

Run `./docs/queries/run-release-entitlement-audit.sh` immediately before an IAP release or legacy-provider cutover. It opens an explicit read-only transaction and reports only aggregate counts for legacy access, canonical production store access, sandbox rows, approved store-test accounts, and unclassified accounts. A release requires every unclassified count to be zero; the output must not be reused from an earlier release window.

The database can measure wordlist creation, word creation, and quiz answers. It cannot prove `quiz_session_started`, `quiz_session_completed`, app opens, failed client flows, or PostHog event deduplication. Those post-instrumentation metrics must be joined from the production analytics export only after the event contract is live; do not compare `database_quiz_answerers` with `quiz_session_completed` as if they were the same measure.
