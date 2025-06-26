BEGIN;

-- Add processing status tracking to words table
ALTER TABLE words ADD COLUMN processing_status VARCHAR(20) DEFAULT 'pending';
ALTER TABLE words ADD COLUMN processing_error TEXT;
ALTER TABLE words ADD COLUMN processing_started_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE words ADD COLUMN processing_completed_at TIMESTAMP WITH TIME ZONE;

-- Add check constraint for valid status values
ALTER TABLE words ADD CONSTRAINT check_processing_status 
    CHECK (processing_status IN ('pending', 'processing', 'completed', 'failed'));

-- Create index for efficient status queries
CREATE INDEX idx_words_processing_status ON words(wordlist_id, processing_status);
CREATE INDEX idx_words_status_user ON words(user_id, processing_status) WHERE processing_status != 'completed';

-- Update existing words to have 'completed' status (they were processed before this feature)
UPDATE words 
SET processing_status = 'completed', 
    processing_completed_at = CURRENT_TIMESTAMP
WHERE processing_status = 'pending';

COMMIT;