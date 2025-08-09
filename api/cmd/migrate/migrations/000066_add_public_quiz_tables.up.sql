-- Create enum for quiz difficulty
CREATE TYPE quiz_difficulty AS ENUM ('easy', 'medium', 'hard');

-- Create public quizzes table
CREATE TABLE public_quizzes (
    id BIGSERIAL PRIMARY KEY,
    slug TEXT UNIQUE NOT NULL, -- Immutable slug for public access
    wordlist_id BIGINT NOT NULL REFERENCES wordlists(id) ON DELETE CASCADE,
    creator_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Public quiz metadata
    title TEXT NOT NULL,
    description TEXT,
    difficulty quiz_difficulty NOT NULL DEFAULT 'medium',
    time_limit_minutes INT NOT NULL CHECK (time_limit_minutes BETWEEN 1 AND 15),
    
    -- Social features
    preview_image_url TEXT,
    play_count INT DEFAULT 0,
    share_count INT DEFAULT 0,
    average_score DECIMAL(5,2),
    
    -- Status and timestamps
    is_active BOOLEAN DEFAULT true,
    published_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create quiz attempts table for anonymous and authenticated players
CREATE TABLE quiz_attempts (
    id BIGSERIAL PRIMARY KEY,
    public_quiz_id BIGINT NOT NULL REFERENCES public_quizzes(id) ON DELETE CASCADE,
    player_id BIGINT REFERENCES users(id) ON DELETE SET NULL, -- NULL for anonymous
    guest_name TEXT, -- Display name for anonymous players
    
    -- Quiz results
    score_percentage DECIMAL(5,2) NOT NULL,
    accuracy_percentage DECIMAL(5,2) NOT NULL,
    completion_time_seconds INT NOT NULL,
    
    -- Timestamps
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_public_quizzes_slug ON public_quizzes(slug);
CREATE INDEX idx_public_quizzes_creator_id ON public_quizzes(creator_id);
CREATE INDEX idx_public_quizzes_wordlist_id ON public_quizzes(wordlist_id);
CREATE INDEX idx_public_quizzes_is_active ON public_quizzes(is_active);

CREATE INDEX idx_quiz_attempts_public_quiz_id ON quiz_attempts(public_quiz_id);
CREATE INDEX idx_quiz_attempts_player_id ON quiz_attempts(player_id);
CREATE INDEX idx_quiz_attempts_completed_at ON quiz_attempts(completed_at);
CREATE INDEX idx_quiz_attempts_score_completion ON quiz_attempts(public_quiz_id, score_percentage DESC, completion_time_seconds ASC);
