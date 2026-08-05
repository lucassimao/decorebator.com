BEGIN;

DROP TABLE IF EXISTS store_purchase_bindings;
DROP TABLE IF EXISTS store_entitlements;
DROP TABLE IF EXISTS store_account_identities;

UPDATE provider_event_inbox
SET result_code = 'stale_provider_event', updated_at = NOW()
WHERE result_code = 'revocation_retained';

ALTER TABLE provider_event_inbox
DROP CONSTRAINT provider_event_inbox_result_code_check;

ALTER TABLE provider_event_inbox
ADD CONSTRAINT provider_event_inbox_result_code_check CHECK (result_code IN (
    'entitlement_applied', 'purchase_pending', 'entitlement_revoked',
    'retryable_provider_error', 'invalid_purchase', 'account_mismatch',
    'unknown_product', 'stale_provider_event', 'provider_test_event'
));

COMMIT;
