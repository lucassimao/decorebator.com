-- Revert the changes made in the up migration

-- Drop the unique index and restore the unique constraint
DROP INDEX IF EXISTS subscriptions_stripe_subscription_id_key;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_stripe_subscription_id_key UNIQUE (stripe_subscription_id);

-- Drop the unique constraint for RevenueCat subscription IDs
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_revenuecat_subscription_id_key;

-- Drop the check constraint
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS check_provider_fields;

-- Make Stripe fields NOT NULL again (this will fail if there are NULL values)
-- You may need to delete or update rows with NULL values before running this
ALTER TABLE subscriptions
ALTER COLUMN stripe_subscription_id SET NOT NULL,
ALTER COLUMN stripe_customer_id SET NOT NULL;