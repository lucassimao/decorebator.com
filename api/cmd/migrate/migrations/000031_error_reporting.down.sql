-- Down migration
-- File: migrations/xxx_add_error_reporting.down.sql

-- Drop indexes
DROP INDEX IF EXISTS idx_leitner_system_tracking_skipped_until;
DROP INDEX IF EXISTS idx_error_reports_definition_id;
DROP INDEX IF EXISTS idx_error_reports_reported_at;
DROP INDEX IF EXISTS idx_error_reports_status;
DROP INDEX IF EXISTS idx_error_reports_user_definition;

-- Drop table
DROP TABLE IF EXISTS error_reports;
-- Rollback to previous version
CREATE TABLE IF NOT EXISTS error_reports(
   id serial PRIMARY KEY,
   is_resolved bool,
   error_type varchar NOT NULL, 
   quiz jsonb NOT NULL,
   user_id int NOT NULL,
   created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE error_reports ADD FOREIGN KEY (user_id) REFERENCES users(id);

-- Remove column
ALTER TABLE leitner_system_tracking 
DROP COLUMN IF EXISTS temporarily_skipped_until;