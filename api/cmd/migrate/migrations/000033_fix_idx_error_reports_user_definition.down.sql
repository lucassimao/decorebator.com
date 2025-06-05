-- Fix the unique index to match the conflict resolution:
DROP INDEX idx_error_reports_user_definition;

CREATE UNIQUE INDEX idx_error_reports_user_definition 
ON error_reports(user_id, definition_id) 
WHERE status = 'pending';