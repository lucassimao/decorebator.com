-- Allow definition_id to be NULL in error_reports table
-- This is needed for word-level errors like sound not playing

-- First, drop the existing foreign key constraint
ALTER TABLE error_reports 
DROP CONSTRAINT error_reports_definition_id_fkey;

-- Modify the column to allow NULL
ALTER TABLE error_reports 
ALTER COLUMN definition_id DROP NOT NULL;

-- Re-add the foreign key constraint but allowing NULL values
ALTER TABLE error_reports 
ADD CONSTRAINT error_reports_definition_id_fkey 
FOREIGN KEY (definition_id) REFERENCES definitions(id) ON DELETE CASCADE;