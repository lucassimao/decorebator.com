-- Rollback rate limiting and cooldown tracking for error reports

-- Drop error report limits table
DROP TABLE IF EXISTS error_report_limits;

-- Remove last_regenerated_at columns
ALTER TABLE words 
DROP COLUMN IF EXISTS audio_last_regenerated_at;

ALTER TABLE definitions 
DROP COLUMN IF EXISTS last_regenerated_at;

-- Drop cooldown tracking table
DROP TABLE IF EXISTS error_report_cooldowns;

-- Remove tracking columns from error_reports
ALTER TABLE error_reports 
DROP COLUMN IF EXISTS regeneration_count,
DROP COLUMN IF EXISTS ip_address;