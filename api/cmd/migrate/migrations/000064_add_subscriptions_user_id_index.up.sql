-- Add index on user_id for subscriptions table to optimize subscription lookups
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);