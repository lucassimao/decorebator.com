-- Add content snapshot column to error_reports table for complete definition preservation
-- This enables full historical tracking of definitions that get deleted during error processing
-- and precise tracking of which specific content was flagged for AI prompt refinement
-- Note: snapshot type is determined by existing error_type column, no need for redundant column

ALTER TABLE error_reports 
ADD COLUMN content_snapshot JSONB;

-- Add index for performance on new column
CREATE INDEX idx_error_reports_content_snapshot ON error_reports USING GIN(content_snapshot);

-- Add comment explaining the new column
COMMENT ON COLUMN error_reports.content_snapshot IS 'Complete JSONB snapshot of definition/word content before deletion - includes flagged_content_index when specific items are flagged. Snapshot type is determined by error_type column.';