-- Revert the addition of content snapshot column to error_reports table

-- Drop index first
DROP INDEX IF EXISTS idx_error_reports_content_snapshot;

-- Remove the content snapshot column
ALTER TABLE error_reports 
DROP COLUMN IF EXISTS content_snapshot;