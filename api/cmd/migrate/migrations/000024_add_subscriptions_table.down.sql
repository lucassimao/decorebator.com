-- Drop trigger and function
DROP TRIGGER IF EXISTS update_user_subscription_trigger ON subscriptions;
DROP FUNCTION IF EXISTS update_user_subscription();

-- Drop indexes
DROP INDEX IF EXISTS idx_users_subscription_plan;
DROP INDEX IF EXISTS idx_users_stripe_customer_id;
DROP INDEX IF EXISTS idx_subscription_events_subscription_id;
DROP INDEX IF EXISTS idx_subscriptions_status;
DROP INDEX IF EXISTS idx_subscriptions_stripe_subscription_id;
DROP INDEX IF EXISTS idx_subscriptions_user_id;

-- Drop tables
DROP TABLE IF EXISTS subscription_events;
DROP TABLE IF EXISTS subscriptions;

-- Remove columns from users table
ALTER TABLE users
DROP COLUMN IF EXISTS subscription_ends_at,
DROP COLUMN IF EXISTS stripe_customer_id,
DROP COLUMN IF EXISTS subscription_status,
DROP COLUMN IF EXISTS subscription_plan;

-- Drop enums
DROP TYPE IF EXISTS subscription_status;
DROP TYPE IF EXISTS subscription_plan;