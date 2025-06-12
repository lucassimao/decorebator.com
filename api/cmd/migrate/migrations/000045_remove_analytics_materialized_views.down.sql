-- Restore materialized views if we need to rollback

-- Recreate materialized view for current word mastery levels
CREATE MATERIALIZED VIEW mv_word_mastery_current AS
SELECT 
    wm.user_id,
    w.wordlist_id,
    wm.word_id,
    w.name AS word,
    wm.mastery_level,
    wm.total_attempts,
    wm.correct_attempts,
    CASE 
        WHEN wm.total_attempts > 0 
        THEN ROUND(wm.correct_attempts::numeric / wm.total_attempts::numeric, 2)
        ELSE 0
    END AS accuracy_rate,
    wm.streak_count,
    wm.last_seen_at,
    MAX(lst.box_id) AS highest_box_reached
FROM word_mastery wm
JOIN words w ON wm.word_id = w.id
LEFT JOIN leitner_system_tracking lst 
    ON lst.word_id = wm.word_id AND lst.user_id = wm.user_id
GROUP BY wm.user_id, w.wordlist_id, wm.word_id, w.name, wm.mastery_level, 
         wm.total_attempts, wm.correct_attempts, wm.streak_count, wm.last_seen_at;

-- Create unique index for concurrent refresh
CREATE UNIQUE INDEX mv_word_mastery_current_idx 
ON mv_word_mastery_current (user_id, word_id);

-- Create index for better query performance
CREATE INDEX mv_word_mastery_wordlist_idx 
ON mv_word_mastery_current (user_id, wordlist_id, mastery_level DESC);

-- Recreate materialized view for wordlist-level quiz performance
CREATE MATERIALIZED VIEW mv_quiz_type_performance_by_wordlist AS
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

-- Recreate the refresh function
CREATE OR REPLACE FUNCTION refresh_all_materialized_views() RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_word_mastery_current;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_quiz_type_performance_by_wordlist;
END;
$$ LANGUAGE plpgsql;

-- Restore comments
COMMENT ON MATERIALIZED VIEW mv_word_mastery_current IS 'Cached view of current word mastery with calculated metrics';
COMMENT ON TABLE word_mastery IS 'Tracks overall mastery level for each word per user';
COMMENT ON TABLE quiz_performance IS 'Tracks individual quiz attempts with performance metrics';