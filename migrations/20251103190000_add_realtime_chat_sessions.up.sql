CREATE TABLE IF NOT EXISTS realtime_chat_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wordlist_id BIGINT REFERENCES wordlists(id) ON DELETE SET NULL,
    session_uuid UUID NOT NULL,
    conversation_id TEXT,
    language_code TEXT,
    model TEXT,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ NOT NULL,
    duration_ms INTEGER NOT NULL,
    total_audio_tokens_in INTEGER NOT NULL DEFAULT 0,
    total_audio_tokens_out INTEGER NOT NULL DEFAULT 0,
    total_text_tokens_in INTEGER NOT NULL DEFAULT 0,
    total_text_tokens_out INTEGER NOT NULL DEFAULT 0,
    total_turns INTEGER NOT NULL DEFAULT 0,
    turns JSONB NOT NULL DEFAULT '[]'::jsonb,
    rate_limits JSONB NOT NULL DEFAULT '[]'::jsonb,
    error_count INTEGER NOT NULL DEFAULT 0,
    client_version TEXT,
    device_info JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_realtime_chat_sessions_user_created
    ON realtime_chat_sessions (user_id, created_at DESC);

