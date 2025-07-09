-- Remove wordlists user_id performance index
-- This rollback drops the index created for wordlist query optimization

DROP INDEX IF EXISTS idx_wordlists_user_id_created_at;