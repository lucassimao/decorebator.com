-- Drop the virtual column and its index
DROP INDEX IF EXISTS idx_definitions_is_verb_type;
ALTER TABLE definitions DROP COLUMN IF EXISTS is_verb_type;