-- Add 'trialing' back to subscription_status enum
ALTER TYPE subscription_status RENAME TO subscription_status_old;
CREATE TYPE subscription_status AS ENUM ('active', 'cancelled', 'past_due', 'trialing', 'unpaid');

-- Update both tables that use the enum
ALTER TABLE subscriptions ALTER COLUMN status TYPE subscription_status USING status::text::subscription_status;
ALTER TABLE users ALTER COLUMN subscription_status TYPE subscription_status USING subscription_status::text::subscription_status;

-- Drop the old type
DROP TYPE subscription_status_old;