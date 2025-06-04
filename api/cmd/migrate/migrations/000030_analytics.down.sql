BEGIN;

-- 1. Drop triggers and their backing functions

-- Word Mastery: remove trigger and function
DROP TRIGGER IF EXISTS update_word_mastery_timestamp_trigger ON word_mastery;
DROP FUNCTION IF EXISTS update_word_mastery_timestamp();

-- Learning Progress: remove trigger and function
DROP TRIGGER IF EXISTS check_learning_progress_trigger ON learning_progress;
DROP FUNCTION IF EXISTS check_learning_progress_consistency();

-- 2. Drop materialized views (this also removes their indexes)
DROP MATERIALIZED VIEW IF EXISTS mv_quiz_type_performance CASCADE;
DROP MATERIALIZED VIEW IF EXISTS mv_word_mastery_current CASCADE;

-- 3. Drop the standalone analytic functions
DROP FUNCTION IF EXISTS calculate_current_streak(BIGINT);
DROP FUNCTION IF EXISTS get_wordlist_mastery_stats(BIGINT, BIGINT);

-- 4. Drop indexes on leitner_system_tracking
DROP INDEX IF EXISTS leitner_tracking_due_idx;
DROP INDEX IF EXISTS leitner_tracking_word_idx;

-- 5. Drop the analytics tables (in reverse order of creation)
DROP TABLE IF EXISTS box_distribution_snapshot CASCADE;
DROP TABLE IF EXISTS quiz_type_analytics CASCADE;
DROP TABLE IF EXISTS learning_progress CASCADE;
DROP TABLE IF EXISTS word_mastery CASCADE;
DROP TABLE IF EXISTS quiz_performance CASCADE;

-- 6. Remove the added columns from leitner_system_history
ALTER TABLE leitner_system_history
  DROP COLUMN IF EXISTS success,
  DROP COLUMN IF EXISTS quiz_type,
  DROP COLUMN IF EXISTS response_time_ms;

COMMIT;
