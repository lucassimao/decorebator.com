-- 1. Add the column (default 0 so new lists start at 0)
ALTER TABLE wordlists
  ADD COLUMN words_count integer NOT NULL DEFAULT 0;

-- 2. Backfill existing lists
UPDATE wordlists wl
SET words_count = COALESCE(sub.count, 0)
FROM (
  SELECT wordlist_id, COUNT(*) AS count
  FROM words
  GROUP BY wordlist_id
) AS sub
WHERE wl.id = sub.wordlist_id;