-- Create materialized view for wordlist-level quiz performance
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_quiz_type_performance_by_wordlist AS
SELECT 
    qp.user_id,
    qp.wordlist_id,
    qp.quiz_type,
    COUNT(*) as total_attempts,
    SUM(CASE WHEN qp.is_correct THEN 1 ELSE 0 END) as correct_attempts,
    CASE 
        WHEN COUNT(*) > 0 
        THEN ROUND(SUM(CASE WHEN qp.is_correct THEN 1 ELSE 0 END)::numeric / COUNT(*)::numeric * 100, 1)
        ELSE 0
    END AS success_rate,
    AVG(qp.response_time_ms)::INT as average_response_time_ms,
    MAX(qp.created_at) as last_updated
FROM quiz_performance qp
GROUP BY qp.user_id, qp.wordlist_id, qp.quiz_type;

-- Create indexes for better query performance
CREATE INDEX idx_quiz_perf_wordlist_user_id ON mv_quiz_type_performance_by_wordlist(user_id);
CREATE INDEX idx_quiz_perf_wordlist_wordlist_id ON mv_quiz_type_performance_by_wordlist(wordlist_id);
CREATE INDEX idx_quiz_perf_wordlist_composite ON mv_quiz_type_performance_by_wordlist(user_id, wordlist_id);

-- Add to the refresh function to update this view along with others
CREATE OR REPLACE FUNCTION refresh_all_materialized_views() RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_word_mastery_current;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_quiz_type_performance;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_quiz_type_performance_by_wordlist;
END;
$$ LANGUAGE plpgsql;