ALTER TABLE leitner_system_history DROP COLUMN success;

DROP INDEX idx_leitner_tracking_due;

-- Add review schedule table
DROP TABLE leitner_review_schedule;