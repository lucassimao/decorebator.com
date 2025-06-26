-- Rename cancelled_at to canceled_at to use consistent American English spelling
ALTER TABLE subscriptions RENAME COLUMN cancelled_at TO canceled_at;