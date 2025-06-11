-- Recreate the user-level quiz performance materialized view
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_quiz_type_performance AS
SELECT 
    user_id,
    quiz_type,
    COUNT(*) as total_attempts,
    ROUND(SUM(CASE WHEN success_rate > 0 THEN 1 ELSE 0 END)::numeric / NULLIF(COUNT(*), 0) * 100, 1) as success_rate,
    ROUND(AVG(average_response_time_ms)) as average_response_time_ms,
    MAX(last_updated) as last_updated
FROM quiz_type_analytics
GROUP BY user_id, quiz_type;

-- Create index for better query performance
CREATE INDEX IF NOT EXISTS idx_mv_quiz_type_performance_user_id ON mv_quiz_type_performance(user_id);

-- Restore the refresh function to include all views
CREATE OR REPLACE FUNCTION refresh_all_materialized_views() RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_word_mastery_current;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_quiz_type_performance;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_quiz_type_performance_by_wordlist;
END;
$$ LANGUAGE plpgsql;