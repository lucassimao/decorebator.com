-- Create subscription plans enum
CREATE TYPE subscription_plan AS ENUM ('free', 'monthly', 'annual');

-- Create subscription status enum
CREATE TYPE subscription_status AS ENUM ('active', 'cancelled', 'past_due', 'trialing', 'unpaid');

-- Add subscription fields to users table
ALTER TABLE users
ADD COLUMN subscription_plan subscription_plan NOT NULL DEFAULT 'free',
ADD COLUMN subscription_status subscription_status DEFAULT NULL,
ADD COLUMN stripe_customer_id VARCHAR(255) UNIQUE DEFAULT NULL,
ADD COLUMN subscription_ends_at TIMESTAMP DEFAULT NULL;

-- Create subscriptions table to track subscription history
CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stripe_subscription_id VARCHAR(255) UNIQUE NOT NULL,
    stripe_customer_id VARCHAR(255) NOT NULL,
    plan subscription_plan NOT NULL,
    status subscription_status NOT NULL,
    current_period_start TIMESTAMP NOT NULL,
    current_period_end TIMESTAMP NOT NULL,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
    cancelled_at TIMESTAMP DEFAULT NULL,
    trial_end TIMESTAMP DEFAULT NULL,
    amount_cents INTEGER NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create subscription events table for audit trail
CREATE TABLE subscription_events (
    id SERIAL PRIMARY KEY,
    subscription_id INTEGER NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    stripe_event_id VARCHAR(255) UNIQUE NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB NOT NULL,
    processed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_stripe_subscription_id ON subscriptions(stripe_subscription_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscription_events_subscription_id ON subscription_events(subscription_id);
CREATE INDEX idx_users_stripe_customer_id ON users(stripe_customer_id);
CREATE INDEX idx_users_subscription_plan ON users(subscription_plan);

-- Add trigger to update users subscription info when subscription changes
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

CREATE TRIGGER update_user_subscription_trigger
AFTER INSERT OR UPDATE ON subscriptions
FOR EACH ROW
EXECUTE FUNCTION update_user_subscription();