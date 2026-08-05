BEGIN;

ALTER TABLE provider_event_inbox
DROP CONSTRAINT provider_event_inbox_result_code_check;

ALTER TABLE provider_event_inbox
ADD CONSTRAINT provider_event_inbox_result_code_check CHECK (result_code IN (
    'entitlement_applied', 'purchase_pending', 'entitlement_revoked',
    'retryable_provider_error', 'invalid_purchase', 'account_mismatch',
    'unknown_product', 'stale_provider_event', 'provider_test_event',
    'revocation_retained'
));

CREATE TABLE store_account_identities (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    store TEXT NOT NULL CHECK (store IN ('apple', 'google')),
    lookup_digest CHAR(64) NOT NULL,
    encrypted_identifier BYTEA NOT NULL,
    encryption_key_version SMALLINT NOT NULL CHECK (encryption_key_version > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, store),
    UNIQUE (store, lookup_digest),
    UNIQUE (id, user_id, store),
    CHECK (octet_length(encrypted_identifier) > 28)
);

CREATE TABLE store_entitlements (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    store TEXT NOT NULL CHECK (store IN ('apple', 'google')),
    environment TEXT NOT NULL CHECK (environment IN ('production', 'sandbox')),
    product_id TEXT NOT NULL CHECK (char_length(product_id) > 0),
    entitlement TEXT NOT NULL CHECK (char_length(entitlement) > 0),
    status TEXT NOT NULL CHECK (status IN (
        'pending', 'active', 'grace_period', 'on_hold', 'paused', 'expired', 'revoked'
    )),
    period_start TIMESTAMPTZ,
    period_end TIMESTAMPTZ,
    auto_renew_enabled BOOLEAN NOT NULL,
    canceled_at TIMESTAMPTZ,
    cancellation_reason TEXT,
    revoked_at TIMESTAMPTZ,
    revocation_reason TEXT,
    last_verified_at TIMESTAMPTZ NOT NULL,
    last_snapshot_observed_at TIMESTAMPTZ,
    last_provider_snapshot_signed_at TIMESTAMPTZ,
    last_event_occurred_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, store, environment, entitlement),
    UNIQUE (id, user_id, store, environment),
    CHECK (period_start IS NULL OR period_end IS NULL OR period_end >= period_start),
    CHECK (status NOT IN ('active', 'grace_period') OR period_end IS NOT NULL),
    CHECK ((status = 'revoked') = (revoked_at IS NOT NULL)),
    CHECK (revocation_reason IS NULL OR revoked_at IS NOT NULL),
    CHECK (last_snapshot_observed_at IS NOT NULL OR last_event_occurred_at IS NOT NULL)
);

CREATE TABLE store_purchase_bindings (
    id BIGSERIAL PRIMARY KEY,
    entitlement_id BIGINT NOT NULL,
    account_identity_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    store TEXT NOT NULL CHECK (store IN ('apple', 'google')),
    environment TEXT NOT NULL CHECK (environment IN ('production', 'sandbox')),
    provider_record_digest CHAR(64) NOT NULL,
    encrypted_provider_record BYTEA NOT NULL,
    encryption_key_version SMALLINT NOT NULL CHECK (encryption_key_version > 0),
    linked_record_digest CHAR(64),
    encrypted_linked_record BYTEA,
    linked_encryption_key_version SMALLINT CHECK (linked_encryption_key_version > 0),
    provider_version TEXT,
    acknowledged BOOLEAN,
    post_revocation_candidate BOOLEAN NOT NULL DEFAULT FALSE,
    is_current BOOLEAN NOT NULL DEFAULT TRUE,
    last_verified_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (entitlement_id, user_id, store, environment)
        REFERENCES store_entitlements(id, user_id, store, environment) ON DELETE CASCADE,
    FOREIGN KEY (account_identity_id, user_id, store)
        REFERENCES store_account_identities(id, user_id, store) ON DELETE CASCADE,
    CONSTRAINT store_purchase_bindings_provider_record_unique
        UNIQUE (store, environment, provider_record_digest),
    CHECK (octet_length(encrypted_provider_record) > 28),
    CHECK (
        (linked_record_digest IS NULL AND encrypted_linked_record IS NULL AND linked_encryption_key_version IS NULL)
        OR (store = 'google' AND linked_record_digest IS NOT NULL AND encrypted_linked_record IS NOT NULL
            AND linked_encryption_key_version IS NOT NULL AND octet_length(encrypted_linked_record) > 28)
    ),
    CHECK ((store = 'apple' AND acknowledged IS NULL) OR (store = 'google' AND acknowledged IS NOT NULL))
);

CREATE UNIQUE INDEX store_purchase_bindings_one_current_per_entitlement
ON store_purchase_bindings (entitlement_id)
WHERE is_current;

CREATE INDEX store_purchase_bindings_linked_record_lookup
ON store_purchase_bindings (store, environment, linked_record_digest)
WHERE linked_record_digest IS NOT NULL;

CREATE INDEX store_purchase_bindings_entitlement_fk
ON store_purchase_bindings (entitlement_id);

CREATE INDEX store_purchase_bindings_account_fk
ON store_purchase_bindings (account_identity_id);

CREATE INDEX store_entitlements_access_lookup
ON store_entitlements (user_id, entitlement, environment, status);

COMMIT;
