-- Drop indexes first
DROP INDEX IF EXISTS idx_quiz_perf_wordlist_user_id;
DROP INDEX IF EXISTS idx_quiz_perf_wordlist_wordlist_id;
DROP INDEX IF EXISTS idx_quiz_perf_wordlist_composite;

-- Drop the materialized view
DROP MATERIALIZED VIEW IF EXISTS mv_quiz_type_performance_by_wordlist;

-- Restore the original refresh function
CREATE OR REPLACE FUNCTION refresh_all_materialized_views() RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_word_mastery_current;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_quiz_type_performance;
END;
$$ LANGUAGE plpgsql;