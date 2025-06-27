BEGIN;

-- Add consecutive failures tracking to leitner_system_tracking table
-- This enables progressive skip penalties for Box 1 words that fail repeatedly
-- instead of the fixed 10-minute delay that contradicts "immediate" review concept
ALTER TABLE leitner_system_tracking 
ADD COLUMN consecutive_failures INTEGER DEFAULT 0;

-- Add constraint to ensure failure count is non-negative
ALTER TABLE leitner_system_tracking 
ADD CONSTRAINT check_consecutive_failures_non_negative 
    CHECK (consecutive_failures >= 0);

-- Create index for efficient queries when selecting words based on failure count
-- This supports the progressive penalty calculation in word selection
CREATE INDEX idx_leitner_consecutive_failures 
ON leitner_system_tracking(user_id, definition_id, consecutive_failures);

-- Initialize existing records to 0 failures for backward compatibility
-- All existing tracking records are considered to have no current failure streak
UPDATE leitner_system_tracking 
SET consecutive_failures = 0 
WHERE consecutive_failures IS NULL;

-- Make the column NOT NULL after setting default values
ALTER TABLE leitner_system_tracking 
ALTER COLUMN consecutive_failures SET NOT NULL;

COMMIT;