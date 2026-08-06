# Canonical email identity migration

Migration `000078_canonical_email_identity` makes the database uniqueness key
match the application identity function: surrounding ASCII whitespace is
ignored, letters are lowercased, plus aliases remain distinct, and non-ASCII
addresses are rejected. The explicit syntax contract accepts an unquoted
dot-separated local part, DNS-style domain labels, and a 2-63 letter final
label; local parts are limited to 64 bytes and the full address to 254 bytes.
PostgreSQL uses explicit ASCII `TRANSLATE`, not
collation-dependent `LOWER`, so its key is byte-equivalent to Go in every
database locale. The migration does not rewrite existing email display values.

Production application remains owner-gated under `OPS-MIG-1`.

## Read-only preflight

Run both reports against the direct production database before applying the
migration. Do not copy email values into tickets or chat; retain only the count
unless an owner-approved private remediation requires row details.

```sql
BEGIN READ ONLY;

SELECT COUNT(*) AS non_ascii_identity_count
FROM users
WHERE OCTET_LENGTH(email) <> CHAR_LENGTH(email);

WITH canonicalized AS (
  SELECT id,
         TRANSLATE(
           BTRIM(email, E' \t\n\r\f\v'),
           'ABCDEFGHIJKLMNOPQRSTUVWXYZ',
           'abcdefghijklmnopqrstuvwxyz'
         ) AS canonical_email
  FROM users
)
SELECT COUNT(*) AS unsupported_ascii_identity_count,
       ARRAY_AGG(id ORDER BY id) AS user_ids
FROM canonicalized
WHERE OCTET_LENGTH(canonical_email) > 254
   OR OCTET_LENGTH(SPLIT_PART(canonical_email, '@', 1)) > 64
   OR canonical_email !~ '^[a-z0-9!#$%&''*+/=?^_`{|}~-]+(\.[a-z0-9!#$%&''*+/=?^_`{|}~-]+)*@([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$';

SELECT TRANSLATE(
         BTRIM(email, E' \t\n\r\f\v'),
         'ABCDEFGHIJKLMNOPQRSTUVWXYZ',
         'abcdefghijklmnopqrstuvwxyz'
       ) AS canonical_email,
       COUNT(*) AS collision_count,
       ARRAY_AGG(id ORDER BY id) AS user_ids
FROM users
GROUP BY TRANSLATE(
  BTRIM(email, E' \t\n\r\f\v'),
  'ABCDEFGHIJKLMNOPQRSTUVWXYZ',
  'abcdefghijklmnopqrstuvwxyz'
)
HAVING COUNT(*) > 1
ORDER BY collision_count DESC, canonical_email;

ROLLBACK;
```

All three reports must be empty/zero. If any is not, stop. The owner must choose
which test account to retain or rename; there are no real paying users to
migrate, but the operation is still an explicit production write and is not
performed by the autonomous lane.

The count-only production baseline run in a forced read-only transaction on
2026-08-06 found 38 users and zero non-ASCII rows, surrounding-whitespace rows,
noncanonical stored rows, unsupported ASCII identities under the exact syntax
contract, or canonical
collision groups. This is evidence, not migration authorization: rerun the
preflight immediately before the owner-approved migration because rows can
change after this observation.

## Apply and verify

Deploy the application normalization contract first. Then rerun the read-only
preflight and apply migration 78 in the same owner-controlled maintenance
window; this prevents the legacy write path from introducing a new unsupported
identity between the final inventory and index creation. The normalized lookup
remains correct before the index exists, although it is not yet index-backed.

The migration uses a five-second lock timeout and thirty-second statement
timeout. Its preflight and index creation run in one transaction; any collision,
timeout, or concurrent conflicting insert rolls the entire migration back and
preserves `users_email_unique_lower`.

After migration, verify without exposing addresses:

```sql
SELECT indexname, indexdef
FROM pg_indexes
WHERE schemaname = 'public'
  AND indexname IN ('users_email_unique_lower', 'users_email_unique_canonical');

SELECT indisvalid, indisready
FROM pg_index
WHERE indexrelid = 'users_email_unique_canonical'::regclass;
```

Expected: only `users_email_unique_canonical` exists and both flags are true.

## Dirty-version recovery

Migration 78 is transactional. If an operator manually forced schema version
78 after a failed attempt, first verify that `users_email_unique_canonical` is
absent and `users_email_unique_lower` is present, then force the migration
version back to 77 and rerun normally. If either index state differs, stop for a
forward-repair review instead of guessing.

The down migration recreates the prior `LOWER(email)` unique index before
dropping the canonical index, under the same bounded timeouts.

## Automated disposable drill

`TestCanonicalEmailIdentityMigrationDrill` creates and removes its own temporary
PostgreSQL database. It migrates to version 77; proves canonical-collision,
non-ASCII, and malformed-ASCII failures preserve the old index; exercises dirty
version recovery; verifies successful migration, index validity/readiness, and
23505 enforcement; then migrates down and confirms the prior index is restored.

Run it only against the disposable Docker test stack:

```bash
docker compose -f docker-compose.test.yml run --rm --no-deps main \
  go test -race -count=1 -run TestCanonicalEmailIdentityMigrationDrill \
  ./tests/integration/...
```
