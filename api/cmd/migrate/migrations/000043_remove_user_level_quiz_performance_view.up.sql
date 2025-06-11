-- Drop the user-level quiz performance materialized view as we now only use wordlist-level
DROP MATERIALIZED VIEW IF EXISTS mv_quiz_type_performance;

-- Update the refresh function to only refresh the remaining views
CREATE OR REPLACE FUNCTION refresh_all_materialized_views() RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_word_mastery_current;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_quiz_type_performance_by_wordlist;
END;
$$ LANGUAGE plpgsql;