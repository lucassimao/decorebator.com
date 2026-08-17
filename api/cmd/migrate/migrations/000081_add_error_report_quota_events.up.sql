-- Keep every committed error-report submission for rolling quota accounting.
-- error_reports represents the current pending remediation state and can be
-- updated in place, so it cannot accurately count repeated submissions.
CREATE TABLE error_report_quota_events (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reported_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_error_report_quota_events_user_reported_at
    ON error_report_quota_events (user_id, reported_at);

-- Preserve the currently active rolling-window usage when this migration is
-- deployed. Historical repeated upserts before this table existed were not
-- durably distinguishable, so each existing report correctly contributes one
-- committed submission from its last reported_at value.
INSERT INTO error_report_quota_events (user_id, reported_at)
SELECT user_id, reported_at
FROM error_reports
WHERE user_id IS NOT NULL
  AND reported_at > NOW() - INTERVAL '24 hours';
