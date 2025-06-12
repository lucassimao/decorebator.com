-- Composite index for overview queries
CREATE INDEX IF NOT EXISTS idx_word_mastery_overview
ON word_mastery(user_id, word_id)
INCLUDE (mastery_level, streak_count);

-- Learning progress index
CREATE INDEX IF NOT EXISTS idx_learning_progress_lookup
ON learning_progress(user_id, wordlist_id, date DESC);

-- Words lookup index for analytics
CREATE INDEX IF NOT EXISTS idx_words_analytics
ON words(user_id, wordlist_id)
WHERE learned = FALSE;

-- Composite index for progress summary
CREATE INDEX IF NOT EXISTS idx_word_mastery_progress
ON word_mastery(user_id, word_id)
INCLUDE (mastery_level, updated_at);

-- Wordlist progress join optimization
CREATE INDEX IF NOT EXISTS idx_words_wordlist_progress
ON words(wordlist_id)
WHERE learned = FALSE;

-- Streak calculation optimization
CREATE INDEX IF NOT EXISTS idx_learning_progress_streak
ON learning_progress(user_id, wordlist_id, date DESC)
WHERE total_quiz_attempts > 0;