-- Add provider support to subscriptions
CREATE TYPE subscription_provider AS ENUM ('stripe', 'revenuecat');
CREATE TYPE platform_type AS ENUM ('ios', 'android', 'web');

-- Add RevenueCat and platform fields to users table
ALTER TABLE users
ADD COLUMN revenuecat_customer_id VARCHAR(255) UNIQUE DEFAULT NULL,
ADD COLUMN platform platform_type DEFAULT NULL;

-- Add provider and platform fields to subscriptions table
ALTER TABLE subscriptions
ADD COLUMN provider subscription_provider NOT NULL DEFAULT 'stripe',
ADD COLUMN revenuecat_subscription_id VARCHAR(255) DEFAULT NULL,
ADD COLUMN app_store_product_id VARCHAR(255) DEFAULT NULL,
ADD COLUMN platform platform_type DEFAULT NULL;

-- Create RevenueCat events table for webhook audit trail
CREATE TABLE revenuecat_events (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    event_id VARCHAR(255) UNIQUE NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    app_user_id VARCHAR(255) NOT NULL,
    product_id VARCHAR(255),
    entitlement_id VARCHAR(255),
    event_data JSONB NOT NULL,
    processed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for RevenueCat fields
CREATE INDEX idx_users_revenuecat_customer_id ON users(revenuecat_customer_id);
CREATE INDEX idx_subscriptions_provider ON subscriptions(provider);
CREATE INDEX idx_subscriptions_revenuecat_subscription_id ON subscriptions(revenuecat_subscription_id);
CREATE INDEX idx_revenuecat_events_user_id ON revenuecat_events(user_id);
CREATE INDEX idx_revenuecat_events_app_user_id ON revenuecat_events(app_user_id);
CREATE INDEX idx_revenuecat_events_event_type ON revenuecat_events(event_type);

-- Update the trigger to handle both providers
CREATE OR REPLACE FUNCTION update_user_subscription()
RETURNS TRIGGER AS $$
BEGIN
    -- Update user's subscription info based on the latest subscription
    UPDATE users
    SET 
        subscription_plan = NEW.plan,
        subscription_status = NEW.status,
        subscription_ends_at = NEW.current_period_end,
        platform = COALESCE(NEW.platform, platform)
    WHERE id = NEW.user_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;