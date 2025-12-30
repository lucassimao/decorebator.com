-- Revert push notification support
DROP INDEX IF EXISTS idx_push_tokens_user_active;
DROP TABLE IF EXISTS push_tokens;

ALTER TABLE users
DROP COLUMN IF EXISTS last_practice_at,
DROP COLUMN IF EXISTS notifications_enabled;
