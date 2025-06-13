-- Remove unused study_time_seconds column from learning_progress table
-- This field was never properly maintained and always remained 0

ALTER TABLE learning_progress DROP COLUMN study_time_seconds;