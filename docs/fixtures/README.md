# Controlled fixtures

`store-approval-accounts.json` is the source of truth for excluding the four owner-confirmed store-approval accounts from activation cohorts. Reports must exclude by the listed numeric `userId`; email patterns, subscription status, and historical provider names are not classification rules.

The file contains no email addresses, payment identifiers, credentials, or store secrets. Each entry records the owner-confirmation date and links to the zero-entitlement evidence. Re-verify it read-only before every production cohort report and before subscription cutover; the `reviewAfter` date is a stale-registry warning, not permission to extend the list automatically.

Any addition, removal, or ID change requires owner confirmation, a dated evidence update, and an independent review before merge. The registry is an input to the pending cohort-report implementation; until that report consumes it, no production metric may claim that test-account exclusion is enforced.
