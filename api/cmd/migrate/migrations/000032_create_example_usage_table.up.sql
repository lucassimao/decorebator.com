CREATE TABLE example_usage (
    definition_id BIGINT NOT NULL,
    example_hash VARCHAR(32) NOT NULL,
    last_used_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (definition_id, example_hash),
    
    FOREIGN KEY (definition_id) REFERENCES definitions(id) ON DELETE CASCADE
);

-- Index for efficient querying by last_used_at
CREATE INDEX idx_example_usage_last_used_at ON example_usage(last_used_at);

-- Index for efficient querying by definition_id and recent usage
CREATE INDEX idx_example_usage_definition_recent ON example_usage(definition_id, last_used_at);