BEGIN;

ALTER TABLE leitner_system_tracking
ADD COLUMN IF NOT EXISTS next_review_at TIMESTAMP WITH TIME ZONE;

UPDATE leitner_system_tracking
SET next_review_at = CASE box_id
	WHEN 1 THEN COALESCE(updated_at, created_at, NOW())
	WHEN 2 THEN COALESCE(updated_at, created_at, NOW()) + INTERVAL '6 hours'
	WHEN 3 THEN COALESCE(updated_at, created_at, NOW()) + INTERVAL '24 hours'
	WHEN 4 THEN COALESCE(updated_at, created_at, NOW()) + INTERVAL '72 hours'
	WHEN 5 THEN COALESCE(updated_at, created_at, NOW()) + INTERVAL '168 hours'
	WHEN 6 THEN COALESCE(updated_at, created_at, NOW()) + INTERVAL '336 hours'
	WHEN 7 THEN COALESCE(updated_at, created_at, NOW()) + INTERVAL '720 hours'
	ELSE COALESCE(updated_at, created_at, NOW())
END
WHERE next_review_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_leitner_system_tracking_next_review_at
ON leitner_system_tracking(next_review_at);

DROP TABLE IF EXISTS leitner_review_schedule;

COMMIT;
