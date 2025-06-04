-- Analytics Tables for Leitner System (PostgreSQL v17)

-- 1. Quiz Performance Tracking
CREATE TABLE quiz_performance (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wordlist_id BIGINT NOT NULL REFERENCES wordlists(id) ON DELETE CASCADE,
    word_id BIGINT NOT NULL REFERENCES words(id) ON DELETE CASCADE,
    definition_id BIGINT NOT NULL REFERENCES definitions(id) ON DELETE CASCADE,
    leitner_system_tracking_id BIGINT NOT NULL REFERENCES leitner_system_tracking(id) ON DELETE CASCADE,
    quiz_type VARCHAR(50) NOT NULL,
    box_id INT NOT NULL CHECK (box_id BETWEEN 1 AND 7),
    is_correct BOOLEAN NOT NULL,
    response_time_ms INT CHECK (response_time_ms >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
) ;

-- Indexes for quiz_performance
CREATE INDEX quiz_perf_user_date_idx ON quiz_performance (user_id, created_at DESC);
CREATE INDEX quiz_perf_word_idx ON quiz_performance (word_id, created_at DESC);
CREATE INDEX quiz_perf_type_idx ON quiz_performance (quiz_type, is_correct);

-- 2. Word Mastery Tracking
CREATE TABLE word_mastery (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    word_id BIGINT NOT NULL REFERENCES words(id) ON DELETE CASCADE,
    mastery_level NUMERIC(3,2) NOT NULL DEFAULT 0.00 CHECK (mastery_level BETWEEN 0 AND 1),
    total_attempts INT NOT NULL DEFAULT 0 CHECK (total_attempts >= 0),
    correct_attempts INT NOT NULL DEFAULT 0 CHECK (correct_attempts >= 0),
    streak_count INT NOT NULL DEFAULT 0 CHECK (streak_count >= 0),
    max_streak INT NOT NULL DEFAULT 0 CHECK (max_streak >= 0),
    last_seen_at TIMESTAMPTZ,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_user_word UNIQUE (user_id, word_id),
    CONSTRAINT correct_attempts_check CHECK (correct_attempts <= total_attempts)
);

-- Indexes for word_mastery
CREATE INDEX word_mastery_level_idx ON word_mastery (user_id, mastery_level DESC);
CREATE INDEX word_mastery_updated_idx ON word_mastery (updated_at DESC);

-- 3. Learning Progress Summary
CREATE TABLE learning_progress (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wordlist_id BIGINT NOT NULL REFERENCES wordlists(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    words_studied INT NOT NULL DEFAULT 0 CHECK (words_studied >= 0),
    new_words_added INT NOT NULL DEFAULT 0 CHECK (new_words_added >= 0),
    words_mastered INT NOT NULL DEFAULT 0 CHECK (words_mastered >= 0),
    total_quiz_attempts INT NOT NULL DEFAULT 0 CHECK (total_quiz_attempts >= 0),
    correct_attempts INT NOT NULL DEFAULT 0 CHECK (correct_attempts >= 0),
    average_response_time_ms INT CHECK (average_response_time_ms >= 0),
    study_time_seconds INT NOT NULL DEFAULT 0 CHECK (study_time_seconds >= 0),
    CONSTRAINT unique_user_wordlist_date UNIQUE (user_id, wordlist_id, date),
    CONSTRAINT correct_attempts_check CHECK (correct_attempts <= total_quiz_attempts)
) ;

-- Indexes for learning_progress
CREATE INDEX learning_progress_date_idx ON learning_progress (user_id, date DESC);
CREATE INDEX learning_progress_wordlist_idx ON learning_progress (wordlist_id, date DESC);

-- 4. Quiz Type Performance Analytics
CREATE TABLE quiz_type_analytics (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quiz_type VARCHAR(50) NOT NULL,
    total_attempts INT NOT NULL DEFAULT 0 CHECK (total_attempts >= 0),
    correct_attempts INT NOT NULL DEFAULT 0 CHECK (correct_attempts >= 0),
    average_response_time_ms INT CHECK (average_response_time_ms >= 0),
    last_updated TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_user_quiz_type UNIQUE (user_id, quiz_type),
    CONSTRAINT correct_attempts_check CHECK (correct_attempts <= total_attempts)
);

-- Index for quiz_type_analytics
CREATE INDEX quiz_type_analytics_user_idx ON quiz_type_analytics (user_id);

-- 5. Box Distribution Snapshot
CREATE TABLE box_distribution_snapshot (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wordlist_id BIGINT NOT NULL REFERENCES wordlists(id) ON DELETE CASCADE,
    snapshot_date DATE NOT NULL DEFAULT CURRENT_DATE,
    box_1_count INT NOT NULL DEFAULT 0 CHECK (box_1_count >= 0),
    box_2_count INT NOT NULL DEFAULT 0 CHECK (box_2_count >= 0),
    box_3_count INT NOT NULL DEFAULT 0 CHECK (box_3_count >= 0),
    box_4_count INT NOT NULL DEFAULT 0 CHECK (box_4_count >= 0),
    box_5_count INT NOT NULL DEFAULT 0 CHECK (box_5_count >= 0),
    box_6_count INT NOT NULL DEFAULT 0 CHECK (box_6_count >= 0),
    box_7_count INT NOT NULL DEFAULT 0 CHECK (box_7_count >= 0),
    CONSTRAINT unique_snapshot UNIQUE (user_id, wordlist_id, snapshot_date)
);

-- Index for box_distribution_snapshot
CREATE INDEX box_distribution_date_idx ON box_distribution_snapshot (user_id, wordlist_id, snapshot_date DESC);

-- Update leitner_system_history with analytics fields
ALTER TABLE leitner_system_history 
ADD COLUMN IF NOT EXISTS success BOOLEAN,
ADD COLUMN IF NOT EXISTS quiz_type VARCHAR(50),
ADD COLUMN IF NOT EXISTS response_time_ms INT CHECK (response_time_ms >= 0);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS leitner_tracking_due_idx 
ON leitner_system_tracking(user_id, box_id, updated_at)
WHERE updated_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS leitner_tracking_word_idx 
ON leitner_system_tracking(word_id, user_id);

-- Materialized view for current word mastery levels
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

-- Materialized view for quiz type performance
CREATE MATERIALIZED VIEW mv_quiz_type_performance AS
SELECT 
    qta.user_id,
    qta.quiz_type,
    qta.total_attempts,
    qta.correct_attempts,
    CASE 
        WHEN qta.total_attempts > 0 
        THEN ROUND(qta.correct_attempts::numeric / qta.total_attempts::numeric * 100, 1)
        ELSE 0
    END AS success_rate,
    qta.average_response_time_ms,
    qta.last_updated
FROM quiz_type_analytics qta
WHERE qta.total_attempts > 0;

-- Create unique index for concurrent refresh
CREATE UNIQUE INDEX mv_quiz_type_performance_idx 
ON mv_quiz_type_performance (user_id, quiz_type);

-- Function to update word mastery updated_at timestamp
CREATE OR REPLACE FUNCTION update_word_mastery_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for updating word_mastery timestamp
CREATE TRIGGER update_word_mastery_timestamp_trigger
BEFORE UPDATE ON word_mastery
FOR EACH ROW
EXECUTE FUNCTION update_word_mastery_timestamp();

-- Function to maintain learning progress consistency
CREATE OR REPLACE FUNCTION check_learning_progress_consistency()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.correct_attempts > NEW.total_quiz_attempts THEN
        RAISE EXCEPTION 'correct_attempts cannot exceed total_quiz_attempts';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for learning progress consistency
CREATE TRIGGER check_learning_progress_trigger
BEFORE INSERT OR UPDATE ON learning_progress
FOR EACH ROW
EXECUTE FUNCTION check_learning_progress_consistency();

-- Scheduled maintenance for materialized views
-- Note: These should be scheduled via pg_cron or external scheduler
-- REFRESH MATERIALIZED VIEW CONCURRENTLY mv_word_mastery_current;
-- REFRESH MATERIALIZED VIEW CONCURRENTLY mv_quiz_type_performance;

-- Function to calculate current streak
CREATE OR REPLACE FUNCTION calculate_current_streak(p_user_id BIGINT)
RETURNS INT AS $$
DECLARE
    v_streak INT := 0;
BEGIN
    WITH daily_activity AS (
        SELECT date, SUM(total_quiz_attempts) AS attempts
        FROM learning_progress
        WHERE user_id = p_user_id
        GROUP BY date
        ORDER BY date DESC
    ),
    streak_calc AS (
        SELECT 
            date,
            date - (ROW_NUMBER() OVER (ORDER BY date DESC))::int AS streak_group
        FROM daily_activity
        WHERE attempts > 0
    )
    SELECT COUNT(*)::int INTO v_streak
    FROM streak_calc
    WHERE streak_group = (
        SELECT streak_group 
        FROM streak_calc 
        WHERE date = CURRENT_DATE
        LIMIT 1
    );
    
    RETURN COALESCE(v_streak, 0);
END;
$$ LANGUAGE plpgsql STABLE;

-- Function to get word mastery statistics
CREATE OR REPLACE FUNCTION get_wordlist_mastery_stats(p_user_id BIGINT, p_wordlist_id BIGINT)
RETURNS TABLE (
    total_words INT,
    words_mastered INT,
    average_mastery NUMERIC,
    words_in_progress INT,
    words_not_started INT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(DISTINCT w.id)::INT AS total_words,
        COUNT(DISTINCT CASE WHEN wm.mastery_level >= 0.8 THEN w.id END)::INT AS words_mastered,
        COALESCE(AVG(wm.mastery_level), 0)::NUMERIC AS average_mastery,
        COUNT(DISTINCT CASE WHEN wm.mastery_level > 0 AND wm.mastery_level < 0.8 THEN w.id END)::INT AS words_in_progress,
        COUNT(DISTINCT CASE WHEN wm.word_id IS NULL THEN w.id END)::INT AS words_not_started
    FROM words w
    LEFT JOIN word_mastery wm ON w.id = wm.word_id AND wm.user_id = p_user_id
    WHERE w.wordlist_id = p_wordlist_id;
END;
$$ LANGUAGE plpgsql STABLE;

-- Add comments for documentation
COMMENT ON TABLE quiz_performance IS 'Tracks individual quiz attempts with performance metrics';
COMMENT ON TABLE word_mastery IS 'Tracks overall mastery level for each word per user';
COMMENT ON TABLE learning_progress IS 'Daily aggregated learning statistics';
COMMENT ON TABLE quiz_type_analytics IS 'Performance metrics grouped by quiz type';
COMMENT ON TABLE box_distribution_snapshot IS 'Daily snapshot of word distribution across Leitner boxes';
COMMENT ON MATERIALIZED VIEW mv_word_mastery_current IS 'Cached view of current word mastery with calculated metrics';
COMMENT ON FUNCTION calculate_current_streak IS 'Calculates the current daily streak for a user';
COMMENT ON FUNCTION get_wordlist_mastery_stats IS 'Returns aggregate mastery statistics for a wordlist';