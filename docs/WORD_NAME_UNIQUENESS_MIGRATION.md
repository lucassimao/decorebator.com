# Word-name uniqueness migration

Migration `000077_add_word_name_uniqueness` prevents concurrent creates, renames,
and moves from producing names that are equal under the application's canonical
`LOWER(BTRIM(name))` policy within one wordlist. Production application remains
owner-gated under `OPS-MIG-1`.

Before applying it, run this read-only report:

```sql
SELECT
  wordlist_id,
  LOWER(BTRIM(name)) AS canonical_name,
  COUNT(*) AS duplicate_count,
  ARRAY_AGG(id ORDER BY id) AS word_ids,
  ARRAY_AGG(name ORDER BY id) AS stored_names
FROM words
GROUP BY wordlist_id, LOWER(BTRIM(name))
HAVING COUNT(*) > 1
ORDER BY wordlist_id, canonical_name;

SELECT id, wordlist_id, name, LOWER(BTRIM(name)) AS canonical_name
FROM words
WHERE name <> LOWER(BTRIM(name))
ORDER BY wordlist_id, id;
```

The duplicate report must be empty. The second report inventories legacy rows
that the application may normalize on a later explicit update; those rows do
not block the migration unless their canonical names collide. If duplicate rows
are returned, do not delete or merge them automatically: retain the lowest ID
only after an owner has reviewed the associated definitions, Leitner tracking,
quiz history, audio, notes, and learned state and approved a deterministic
merge. The migration fails before creating the index when the duplicate
precondition is not met.

## Locking and failure recovery

The checked-in migration uses ordinary `CREATE UNIQUE INDEX` inside a
transaction. PostgreSQL permits reads while that index is built but blocks
writes to `words`; `SET LOCAL lock_timeout = '5s'` prevents an indefinite wait
to acquire the required table lock. The owner should schedule the production
application for a measured low-write window after recording row count and build
duration on a representative copy.

If that write-block window is not acceptable, `OPS-MIG-1` must prepare and
validate a separate non-transactional `CREATE UNIQUE INDEX CONCURRENTLY`
procedure. It cannot be substituted inside this transaction. A failed
concurrent build can leave an invalid index, which must be identified through
`pg_index.indisvalid` and dropped before retrying.

The preflight should be run before invoking golang-migrate. If migration 77
still fails, golang-migrate records version 77 as dirty even though this
transaction rolls back. Confirm that no valid or invalid index named
`words_wordlist_id_name_unique` remains, resolve the reported data or lock
condition, run `migrate force 76`, and then retry the normal migration. Never
force the version without first proving the migration transaction rolled back.

After applying locally, verify the constraint and duplicate report:

```sql
SELECT indexrelid::regclass AS index_name, indisvalid, indisready
FROM pg_index
WHERE indexrelid = 'words_wordlist_id_name_unique'::regclass;

SELECT wordlist_id, LOWER(BTRIM(name)) AS canonical_name, COUNT(*)
FROM words
GROUP BY wordlist_id, LOWER(BTRIM(name))
HAVING COUNT(*) > 1;
```

Rollback drops only the new index; it does not modify word data.
