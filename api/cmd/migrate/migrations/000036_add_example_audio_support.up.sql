-- Create table for storing multiple example audio files per definition
CREATE TABLE IF NOT EXISTS definition_example_audio (
    id SERIAL PRIMARY KEY,
    definition_id BIGINT NOT NULL,
    example_text TEXT NOT NULL,
    example_hash VARCHAR(32) NOT NULL,
    audio_url TEXT NOT NULL,
    inflection_type VARCHAR(50), -- NULL for main examples, or tense/form for inflection examples
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(definition_id, example_hash),
    FOREIGN KEY (definition_id) REFERENCES definitions(id) ON DELETE CASCADE
);

-- Indexes for efficient querying
CREATE INDEX idx_definition_example_audio_definition_id ON definition_example_audio(definition_id);
CREATE INDEX idx_definition_example_audio_hash ON definition_example_audio(definition_id, example_hash);
CREATE INDEX idx_definition_example_audio_inflection ON definition_example_audio(definition_id, inflection_type);

-- Create table for tracking example audio usage fairness
CREATE TABLE IF NOT EXISTS example_audio_usage (
    id SERIAL PRIMARY KEY,
    definition_id BIGINT NOT NULL,
    example_audio_id BIGINT NOT NULL,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    usage_count INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(definition_id, example_audio_id),
    FOREIGN KEY (definition_id) REFERENCES definitions(id) ON DELETE CASCADE,
    FOREIGN KEY (example_audio_id) REFERENCES definition_example_audio(id) ON DELETE CASCADE
);

CREATE INDEX idx_example_audio_usage_definition_id ON example_audio_usage(definition_id);
CREATE INDEX idx_example_audio_usage_last_used ON example_audio_usage(last_used_at);
CREATE INDEX idx_example_audio_usage_count ON example_audio_usage(usage_count);