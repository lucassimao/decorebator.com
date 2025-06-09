-- Add missing updated_at column to learning_progress table
-- This column is required by the analytics service which sets it manually in SQL

ALTER TABLE learning_progress 
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;

-- Backfill existing rows only if they don't have updated_at set
UPDATE learning_progress 
SET updated_at = CURRENT_TIMESTAMP 
WHERE updated_at IS NULL;

-- Also add updated_at to box_distribution_snapshot table for consistency
ALTER TABLE box_distribution_snapshot 
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;

-- Backfill existing box distribution rows only if they don't have updated_at set
UPDATE box_distribution_snapshot 
SET updated_at = CURRENT_TIMESTAMP 
WHERE updated_at IS NULL;