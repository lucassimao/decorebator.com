-- Add success tracking to history
ALTER TABLE leitner_system_history 
ADD COLUMN success BOOLEAN;

-- Add index for performance
CREATE INDEX idx_leitner_tracking_due 
ON leitner_system_tracking(user_id, box_id, updated_at);

-- Add review schedule table
CREATE TABLE leitner_review_schedule (
    id BIGSERIAL PRIMARY KEY,
    leitner_system_tracking_id BIGINT REFERENCES leitner_system_tracking(id),
    next_review_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);