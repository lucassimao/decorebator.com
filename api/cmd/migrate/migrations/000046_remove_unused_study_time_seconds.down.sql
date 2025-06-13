-- Restore study_time_seconds column to learning_progress table
-- Note: All values will be 0 since this field was never properly maintained

ALTER TABLE learning_progress ADD COLUMN study_time_seconds INT NOT NULL DEFAULT 0 CHECK (study_time_seconds >= 0);