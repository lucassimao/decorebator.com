BEGIN;

SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';

CREATE TABLE auth_session_families (
    id UUID PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL,
    idle_expires_at TIMESTAMPTZ NOT NULL,
    absolute_expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    revocation_reason TEXT,
    CHECK (idle_expires_at > created_at),
    CHECK (absolute_expires_at > created_at),
    CHECK (idle_expires_at <= absolute_expires_at),
    CHECK ((revoked_at IS NULL) = (revocation_reason IS NULL))
);

CREATE INDEX auth_session_families_user_id_idx
ON auth_session_families (user_id);

CREATE TABLE auth_refresh_tokens (
    id UUID PRIMARY KEY,
    family_id UUID NOT NULL REFERENCES auth_session_families(id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    created_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    replacement_id UUID REFERENCES auth_refresh_tokens(id),
    CHECK (replacement_id IS NULL OR consumed_at IS NOT NULL)
);

CREATE UNIQUE INDEX auth_refresh_tokens_one_active_per_family
ON auth_refresh_tokens (family_id)
WHERE consumed_at IS NULL;

CREATE INDEX auth_refresh_tokens_family_id_idx
ON auth_refresh_tokens (family_id);

COMMIT;
