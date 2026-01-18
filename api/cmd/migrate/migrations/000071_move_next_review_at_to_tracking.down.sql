BEGIN;

CREATE TABLE IF NOT EXISTS leitner_review_schedule (
    id BIGSERIAL PRIMARY KEY,
    leitner_system_tracking_id BIGINT REFERENCES leitner_system_tracking(id),
    next_review_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

DROP INDEX IF EXISTS idx_leitner_system_tracking_next_review_at;

ALTER TABLE leitner_system_tracking
DROP COLUMN IF EXISTS next_review_at;

COMMIT;
