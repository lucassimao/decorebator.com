-- Add rate limiting and cooldown tracking for error reports

-- Add tracking columns to error_reports
ALTER TABLE error_reports 
ADD COLUMN ip_address INET,
ADD COLUMN regeneration_count INT DEFAULT 0;

-- Create cooldown tracking table
CREATE TABLE error_report_cooldowns (
  user_id        BIGINT     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  word_id        BIGINT,
  definition_id  BIGINT,
  error_type     VARCHAR(50) NOT NULL,
  cooldown_until TIMESTAMPTZ NOT NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- enforce one row per (user, wordId-or--1, definitionId-or--1, error_type)
CREATE UNIQUE INDEX uq_error_report_cooldowns
  ON error_report_cooldowns (
    user_id,
    COALESCE(word_id, -1),
    COALESCE(definition_id, -1),
    error_type
  );


-- Create indexes for efficient querying
CREATE INDEX idx_error_report_cooldowns_user_id ON error_report_cooldowns(user_id);
CREATE INDEX idx_error_report_cooldowns_cooldown_until ON error_report_cooldowns(cooldown_until);

-- Add last_regenerated_at to track recent regenerations
ALTER TABLE definitions 
ADD COLUMN last_regenerated_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE words 
ADD COLUMN audio_last_regenerated_at TIMESTAMP WITH TIME ZONE;

-- Create table for tracking daily/hourly limits (will use Redis primarily, this is for backup/analytics)
CREATE TABLE error_report_limits (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    hourly_count INT DEFAULT 0,
    daily_count INT DEFAULT 0,
    last_hour_reset TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_day_reset TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_error_report_limits_last_hour_reset ON error_report_limits(last_hour_reset);
CREATE INDEX idx_error_report_limits_last_day_reset ON error_report_limits(last_day_reset);