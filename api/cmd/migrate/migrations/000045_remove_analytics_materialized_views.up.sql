-- Remove materialized views as we now use Redis caching for analytics
-- This migration should be run after the new caching system has been stable in production

-- Drop the materialized views
DROP MATERIALIZED VIEW IF EXISTS mv_quiz_type_performance_by_wordlist;
DROP MATERIALIZED VIEW IF EXISTS mv_word_mastery_current;

-- Drop the refresh function as it's no longer needed
DROP FUNCTION IF EXISTS refresh_all_materialized_views();

-- Add comment to document the change
COMMENT ON TABLE word_mastery IS 'Tracks overall mastery level for each word per user. Previously used materialized views, now queries are cached in Redis';
COMMENT ON TABLE quiz_performance IS 'Tracks individual quiz attempts with performance metrics. Analytics are now served from Redis cache instead of materialized views';