-- Remove index on user_id for subscriptions table
DROP INDEX IF EXISTS idx_subscriptions_user_id;