CREATE TABLE IF NOT EXISTS push_receipts (
    id BIGSERIAL PRIMARY KEY,
    expo_ticket_id TEXT NOT NULL UNIQUE,
    expo_token TEXT NOT NULL,
    status TEXT,
    message TEXT,
    details JSONB,
    checked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_push_receipts_unchecked
ON push_receipts(checked_at) WHERE checked_at IS NULL;
