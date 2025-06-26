BEGIN;

-- Drop indexes
DROP INDEX IF EXISTS idx_words_processing_status;
DROP INDEX IF EXISTS idx_words_status_user;

-- Drop constraint
ALTER TABLE words DROP CONSTRAINT IF EXISTS check_processing_status;

-- Drop columns
ALTER TABLE words DROP COLUMN IF EXISTS processing_status;
ALTER TABLE words DROP COLUMN IF EXISTS processing_error;
ALTER TABLE words DROP COLUMN IF EXISTS processing_started_at;
ALTER TABLE words DROP COLUMN IF EXISTS processing_completed_at;

COMMIT;