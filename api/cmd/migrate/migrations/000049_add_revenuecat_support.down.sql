-- Revert trigger to original version
CREATE OR REPLACE FUNCTION update_user_subscription()
RETURNS TRIGGER AS $$
BEGIN
    -- Update user's subscription info based on the latest subscription
    UPDATE users
    SET 
        subscription_plan = NEW.plan,
        subscription_status = NEW.status,
        subscription_ends_at = NEW.current_period_end
    WHERE id = NEW.user_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop indexes
DROP INDEX IF EXISTS idx_revenuecat_events_event_type;
DROP INDEX IF EXISTS idx_revenuecat_events_app_user_id;
DROP INDEX IF EXISTS idx_revenuecat_events_user_id;
DROP INDEX IF EXISTS idx_subscriptions_revenuecat_subscription_id;
DROP INDEX IF EXISTS idx_subscriptions_provider;
DROP INDEX IF EXISTS idx_users_revenuecat_customer_id;

-- Drop RevenueCat events table
DROP TABLE IF EXISTS revenuecat_events;

-- Remove columns from subscriptions table
ALTER TABLE subscriptions
DROP COLUMN IF EXISTS platform,
DROP COLUMN IF EXISTS app_store_product_id,
DROP COLUMN IF EXISTS revenuecat_subscription_id,
DROP COLUMN IF EXISTS provider;

-- Remove columns from users table
ALTER TABLE users
DROP COLUMN IF EXISTS platform,
DROP COLUMN IF EXISTS revenuecat_customer_id;

-- Drop types
DROP TYPE IF EXISTS platform_type;
DROP TYPE IF EXISTS subscription_provider;