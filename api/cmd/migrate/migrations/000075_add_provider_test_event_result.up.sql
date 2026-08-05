BEGIN;

ALTER TABLE provider_event_inbox
DROP CONSTRAINT provider_event_inbox_result_code_check;

ALTER TABLE provider_event_inbox
ADD CONSTRAINT provider_event_inbox_result_code_check CHECK (result_code IN (
    'entitlement_applied',
    'purchase_pending',
    'entitlement_revoked',
    'retryable_provider_error',
    'invalid_purchase',
    'account_mismatch',
    'unknown_product',
    'stale_provider_event',
    'provider_test_event'
));

COMMIT;
