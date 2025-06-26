-- Revert back to cancelled_at
ALTER TABLE subscriptions RENAME COLUMN canceled_at TO cancelled_at;