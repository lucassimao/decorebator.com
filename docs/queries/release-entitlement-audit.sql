\set ON_ERROR_STOP on

BEGIN READ ONLY;

WITH registry AS (
    SELECT value::BIGINT AS user_id
    FROM jsonb_array_elements_text(:'registry_json'::JSONB)
),
legacy_user_plans AS (
    SELECT user_row.id AS user_id
    FROM users user_row
    WHERE user_row.subscription_plan IN ('monthly', 'annual')
),
legacy_effective_rows AS (
    SELECT DISTINCT subscription.user_id
    FROM subscriptions subscription
    WHERE subscription.status = 'active'
       OR (
           subscription.status = 'past_due'
           AND subscription.current_period_end + INTERVAL '3 days' > NOW()
       )
),
legacy_union AS (
    SELECT user_id FROM legacy_user_plans
    UNION
    SELECT user_id FROM legacy_effective_rows
)
SELECT
    COUNT(*) AS legacy_effective_accounts,
    COUNT(*) FILTER (WHERE registry.user_id IS NOT NULL) AS approved_test_accounts,
    COUNT(*) FILTER (WHERE registry.user_id IS NULL) AS unclassified_accounts
FROM legacy_union
LEFT JOIN registry USING (user_id);

SELECT to_regclass('public.store_entitlements') IS NOT NULL AS canonical_tables_present \gset

\if :canonical_tables_present
WITH registry AS (
    SELECT value::BIGINT AS user_id
    FROM jsonb_array_elements_text(:'registry_json'::JSONB)
),
canonical_effective AS (
    SELECT DISTINCT entitlement.user_id
    FROM store_entitlements entitlement
    WHERE entitlement.environment = 'production'
      AND entitlement.entitlement = 'premium'
      AND entitlement.status IN ('active', 'grace_period')
      AND entitlement.period_end > NOW()
      AND (
          entitlement.store = 'apple'
          OR EXISTS (
              SELECT 1
              FROM store_purchase_bindings binding
              WHERE binding.entitlement_id = entitlement.id
                AND binding.is_current
                AND binding.acknowledged IS TRUE
          )
      )
),
canonical_sandbox AS (
    SELECT DISTINCT user_id
    FROM store_entitlements
    WHERE environment = 'sandbox'
)
SELECT
    (SELECT COUNT(*) FROM canonical_effective) AS production_store_effective_accounts,
    (
        SELECT COUNT(*)
        FROM canonical_effective effective
        JOIN registry USING (user_id)
    ) AS approved_production_store_test_accounts,
    (
        SELECT COUNT(*)
        FROM canonical_effective effective
        LEFT JOIN registry USING (user_id)
        WHERE registry.user_id IS NULL
    ) AS unclassified_production_store_accounts,
    (SELECT COUNT(*) FROM canonical_sandbox) AS sandbox_store_accounts;
\else
SELECT
    0::BIGINT AS production_store_effective_accounts,
    0::BIGINT AS approved_production_store_test_accounts,
    0::BIGINT AS unclassified_production_store_accounts,
    0::BIGINT AS sandbox_store_accounts,
    'canonical store tables are not installed in this database' AS note;
\endif

ROLLBACK;
