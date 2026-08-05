# Read-only reports

Run `./docs/queries/run-activation-cohort-report.sh` from the repository root to produce the database baseline for 7-, 30-, 90-, and 365-day signup cohorts. The script reads the controlled store-approval registry and performs a `SELECT` only.

The database can measure wordlist creation, word creation, and quiz answers. It cannot prove `quiz_session_started`, `quiz_session_completed`, app opens, failed client flows, or PostHog event deduplication. Those post-instrumentation metrics must be joined from the production analytics export only after the event contract is live; do not compare `database_quiz_answerers` with `quiz_session_completed` as if they were the same measure.
