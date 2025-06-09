-- Rollback: Make definition_id NOT NULL again
-- WARNING: This will fail if there are any NULL definition_id values

-- First, drop the foreign key constraint
ALTER TABLE error_reports 
DROP CONSTRAINT error_reports_definition_id_fkey;

-- Make the column NOT NULL again
ALTER TABLE error_reports 
ALTER COLUMN definition_id SET NOT NULL;

-- Re-add the foreign key constraint with NOT NULL
ALTER TABLE error_reports 
ADD CONSTRAINT error_reports_definition_id_fkey 
FOREIGN KEY (definition_id) REFERENCES definitions(id) ON DELETE CASCADE;