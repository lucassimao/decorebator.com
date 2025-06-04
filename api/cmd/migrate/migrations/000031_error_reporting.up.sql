-- Migration: Add error reporting support to Leitner system

-- Add temporary skip column to leitner_system_tracking
ALTER TABLE leitner_system_tracking 
ADD COLUMN temporarily_skipped_until TIMESTAMP WITH TIME ZONE;

DROP TABLE error_reports;

-- Create definition error reports table
CREATE TABLE error_reports (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    definition_id BIGINT NOT NULL REFERENCES definitions(id) ON DELETE CASCADE,
    word_id BIGINT NOT NULL REFERENCES words(id) ON DELETE CASCADE,
    error_type VARCHAR(50) NOT NULL,
    description TEXT,
    reported_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'resolved', 'dismissed')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create unique index to prevent duplicate error reports per user/definition
CREATE UNIQUE INDEX idx_error_reports_user_definition 
ON error_reports(user_id, definition_id) 
WHERE status = 'pending';

-- Create indexes for efficient querying
CREATE INDEX idx_error_reports_status ON error_reports(status);
CREATE INDEX idx_error_reports_reported_at ON error_reports(reported_at);
CREATE INDEX idx_error_reports_definition_id ON error_reports(definition_id);

-- Create index for temporarily skipped definitions
CREATE INDEX idx_leitner_system_tracking_skipped_until 
ON leitner_system_tracking(temporarily_skipped_until) 
WHERE temporarily_skipped_until IS NOT NULL;

