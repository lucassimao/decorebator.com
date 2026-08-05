BEGIN;

CREATE TABLE provider_event_inbox (
    id BIGSERIAL PRIMARY KEY,
    store TEXT NOT NULL CHECK (store IN ('apple', 'google')),
    environment TEXT CHECK (environment IN ('production', 'sandbox')),
    idempotency_key VARCHAR(128) NOT NULL UNIQUE,
    semantic_key VARCHAR(128),
    provider_event_type TEXT NOT NULL,
    provider_occurred_at TIMESTAMPTZ NOT NULL,
    state TEXT NOT NULL CHECK (state IN ('processing', 'completed', 'retryable', 'rejected')),
    outcome TEXT CHECK (outcome IN ('applied', 'unchanged', 'pending', 'retry', 'rejected')),
    result_code TEXT CHECK (result_code IN (
        'entitlement_applied',
        'purchase_pending',
        'entitlement_revoked',
        'retryable_provider_error',
        'invalid_purchase',
        'account_mismatch',
        'unknown_product',
        'stale_provider_event'
    )),
    attempt_count INTEGER NOT NULL DEFAULT 1 CHECK (attempt_count > 0),
    lease_token UUID,
    lease_expires_at TIMESTAMPTZ,
    next_attempt_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (char_length(idempotency_key) > 0),
    CHECK (semantic_key IS NULL OR char_length(semantic_key) > 0),
    CHECK (store <> 'apple' OR environment IS NOT NULL),
    CHECK (
        (state = 'processing' AND outcome IS NULL AND result_code IS NULL AND lease_token IS NOT NULL AND lease_expires_at IS NOT NULL AND next_attempt_at IS NULL)
        OR (state = 'retryable' AND outcome = 'retry' AND result_code = 'retryable_provider_error' AND lease_token IS NULL AND lease_expires_at IS NULL AND next_attempt_at IS NOT NULL)
        OR (state IN ('completed', 'rejected') AND outcome IS NOT NULL AND result_code IS NOT NULL AND lease_token IS NULL AND lease_expires_at IS NULL AND next_attempt_at IS NULL)
    )
);

CREATE UNIQUE INDEX provider_event_inbox_semantic_key_unique
ON provider_event_inbox (semantic_key)
WHERE semantic_key IS NOT NULL;

COMMIT;
