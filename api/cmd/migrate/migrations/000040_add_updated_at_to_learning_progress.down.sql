-- Rollback migration for adding updated_at to learning_progress

-- Remove columns
ALTER TABLE learning_progress DROP COLUMN IF EXISTS updated_at;
ALTER TABLE box_distribution_snapshot DROP COLUMN IF EXISTS updated_at;