BEGIN;

SET LOCAL lock_timeout = '5s';

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM words
    GROUP BY wordlist_id, LOWER(BTRIM(name))
    HAVING COUNT(*) > 1
  ) THEN
    RAISE EXCEPTION 'cannot enforce normalized word-name uniqueness while duplicate rows exist'
      USING HINT = 'Run the documented read-only duplicate report and resolve each duplicate group before retrying.';
  END IF;
END
$$;

CREATE UNIQUE INDEX words_wordlist_id_name_unique
  ON words (wordlist_id, LOWER(BTRIM(name)));

COMMIT;
