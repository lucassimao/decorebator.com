-- Rollback content filtering features

-- Drop views
DROP VIEW IF EXISTS content_moderation_summary;

-- Drop functions
DROP FUNCTION IF EXISTS get_user_content_stats(INTEGER);
DROP FUNCTION IF EXISTS log_content_status_change();

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_log_words_content_status_change ON words;
DROP TRIGGER IF EXISTS trigger_log_wordlists_content_status_change ON wordlists;

-- Drop content moderation log table
DROP TABLE IF EXISTS content_moderation_log;

-- Drop indexes
DROP INDEX IF EXISTS idx_words_content_status;
DROP INDEX IF EXISTS idx_words_content_status_user;
DROP INDEX IF EXISTS idx_wordlists_content_status;
DROP INDEX IF EXISTS idx_wordlists_content_status_user;

-- Remove constraints
ALTER TABLE words DROP CONSTRAINT IF EXISTS check_words_flagged_reason;
ALTER TABLE wordlists DROP CONSTRAINT IF EXISTS check_wordlists_flagged_reason;

-- Remove content filtering columns from words table
ALTER TABLE words 
DROP COLUMN IF EXISTS content_status,
DROP COLUMN IF EXISTS flagged_reason,
DROP COLUMN IF EXISTS content_reviewed_at;

-- Remove content filtering columns from wordlists table
ALTER TABLE wordlists
DROP COLUMN IF EXISTS content_status,
DROP COLUMN IF EXISTS flagged_reason,
DROP COLUMN IF EXISTS content_reviewed_at;

-- Drop content status enum
DROP TYPE IF EXISTS content_status;