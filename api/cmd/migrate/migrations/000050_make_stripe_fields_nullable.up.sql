-- Make Stripe-specific fields nullable to support both Stripe and RevenueCat subscriptions
ALTER TABLE subscriptions
ALTER COLUMN stripe_subscription_id DROP NOT NULL,
ALTER COLUMN stripe_customer_id DROP NOT NULL;

-- Add a check constraint to ensure provider-specific fields are set correctly
ALTER TABLE subscriptions
ADD CONSTRAINT check_provider_fields CHECK (
    (provider = 'stripe' AND stripe_subscription_id IS NOT NULL AND stripe_customer_id IS NOT NULL) OR
    (provider = 'revenuecat' AND revenuecat_subscription_id IS NOT NULL)
);

-- Add a unique constraint for RevenueCat subscription IDs
ALTER TABLE subscriptions
ADD CONSTRAINT subscriptions_revenuecat_subscription_id_key UNIQUE (revenuecat_subscription_id);

-- Update the unique constraint for stripe_subscription_id to allow NULLs
-- First drop the existing unique constraint
ALTER TABLE subscriptions DROP CONSTRAINT subscriptions_stripe_subscription_id_key;

-- Then add a new unique index that allows NULL values
CREATE UNIQUE INDEX subscriptions_stripe_subscription_id_key ON subscriptions(stripe_subscription_id) WHERE stripe_subscription_id IS NOT NULL;