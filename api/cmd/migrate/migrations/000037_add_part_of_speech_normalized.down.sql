-- Drop the index
DROP INDEX IF EXISTS idx_definitions_part_of_speech_normalized;

-- Remove the normalized part of speech column
ALTER TABLE definitions DROP COLUMN IF EXISTS part_of_speech_normalized;