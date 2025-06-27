BEGIN;

-- Remove the index first
DROP INDEX IF EXISTS idx_leitner_consecutive_failures;

-- Remove the constraint
ALTER TABLE leitner_system_tracking 
DROP CONSTRAINT IF EXISTS check_consecutive_failures_non_negative;

-- Remove the consecutive_failures column
ALTER TABLE leitner_system_tracking 
DROP COLUMN IF EXISTS consecutive_failures;

COMMIT;