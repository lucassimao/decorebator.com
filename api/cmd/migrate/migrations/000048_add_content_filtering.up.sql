-- Add content filtering support to the database

-- Create content status enum
CREATE TYPE content_status AS ENUM ('approved', 'pending', 'rejected');

-- Add content filtering fields to words table
ALTER TABLE words 
ADD COLUMN content_status content_status DEFAULT 'approved' NOT NULL,
ADD COLUMN flagged_reason TEXT DEFAULT NULL,
ADD COLUMN content_reviewed_at TIMESTAMP DEFAULT NULL;

-- Add content filtering fields to wordlists table  
ALTER TABLE wordlists
ADD COLUMN content_status content_status DEFAULT 'approved' NOT NULL,
ADD COLUMN flagged_reason TEXT DEFAULT NULL,
ADD COLUMN content_reviewed_at TIMESTAMP DEFAULT NULL;

-- Create indexes for efficient content filtering queries
CREATE INDEX idx_words_content_status ON words(content_status);
CREATE INDEX idx_words_content_status_user ON words(user_id, content_status);
CREATE INDEX idx_wordlists_content_status ON wordlists(content_status);
CREATE INDEX idx_wordlists_content_status_user ON wordlists(user_id, content_status);

-- Create table for tracking content moderation activities
CREATE TABLE content_moderation_log (
    id SERIAL PRIMARY KEY,
    entity_type VARCHAR(20) NOT NULL CHECK (entity_type IN ('word', 'wordlist')),
    entity_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    old_status content_status,
    new_status content_status NOT NULL,
    reason TEXT,
    flagged_words TEXT[], -- Array of specific words that triggered the filter
    severity VARCHAR(10) CHECK (severity IN ('low', 'medium', 'high')),
    automated BOOLEAN DEFAULT TRUE, -- TRUE for automated filtering, FALSE for manual review
    reviewed_by INTEGER REFERENCES users(id) ON DELETE SET NULL, -- Admin who reviewed (for manual reviews)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for moderation log
CREATE INDEX idx_content_moderation_log_entity ON content_moderation_log(entity_type, entity_id);
CREATE INDEX idx_content_moderation_log_user ON content_moderation_log(user_id);
CREATE INDEX idx_content_moderation_log_status ON content_moderation_log(new_status);
CREATE INDEX idx_content_moderation_log_automated ON content_moderation_log(automated);
CREATE INDEX idx_content_moderation_log_created_at ON content_moderation_log(created_at);

-- Create function to automatically log content status changes
CREATE OR REPLACE FUNCTION log_content_status_change()
RETURNS TRIGGER AS $$
BEGIN
    -- Only log if content_status actually changed
    IF OLD.content_status IS DISTINCT FROM NEW.content_status THEN
        INSERT INTO content_moderation_log (
            entity_type,
            entity_id,
            user_id,
            old_status,
            new_status,
            reason,
            automated,
            created_at
        ) VALUES (
            TG_TABLE_NAME::TEXT,
            NEW.id,
            NEW.user_id,
            OLD.content_status,
            NEW.content_status,
            NEW.flagged_reason,
            TRUE, -- Assume automated unless manually set
            CURRENT_TIMESTAMP
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers to automatically log content status changes
CREATE TRIGGER trigger_log_words_content_status_change
    AFTER UPDATE ON words
    FOR EACH ROW
    EXECUTE FUNCTION log_content_status_change();

CREATE TRIGGER trigger_log_wordlists_content_status_change
    AFTER UPDATE ON wordlists
    FOR EACH ROW
    EXECUTE FUNCTION log_content_status_change();

-- Update existing words and wordlists to approved status (they were created before filtering)
UPDATE words SET content_status = 'approved', content_reviewed_at = CURRENT_TIMESTAMP WHERE content_status IS NULL;
UPDATE wordlists SET content_status = 'approved', content_reviewed_at = CURRENT_TIMESTAMP WHERE content_status IS NULL;

-- Add constraint to ensure flagged_reason is provided when status is rejected
ALTER TABLE words ADD CONSTRAINT check_words_flagged_reason 
    CHECK (content_status != 'rejected' OR flagged_reason IS NOT NULL);

ALTER TABLE wordlists ADD CONSTRAINT check_wordlists_flagged_reason 
    CHECK (content_status != 'rejected' OR flagged_reason IS NOT NULL);

-- Create view for content moderation dashboard (future use)
CREATE VIEW content_moderation_summary AS
SELECT 
    entity_type,
    new_status as status,
    COUNT(*) as count,
    COUNT(CASE WHEN automated = FALSE THEN 1 END) as manual_reviews,
    COUNT(CASE WHEN severity = 'high' THEN 1 END) as high_severity,
    COUNT(CASE WHEN severity = 'medium' THEN 1 END) as medium_severity,
    COUNT(CASE WHEN severity = 'low' THEN 1 END) as low_severity,
    MAX(created_at) as last_activity
FROM content_moderation_log
WHERE created_at >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY entity_type, new_status;

-- Create function to get user's content filtering stats
CREATE OR REPLACE FUNCTION get_user_content_stats(user_id_param INTEGER)
RETURNS TABLE (
    total_words INTEGER,
    approved_words INTEGER,
    pending_words INTEGER,
    rejected_words INTEGER,
    total_wordlists INTEGER,
    approved_wordlists INTEGER,
    pending_wordlists INTEGER,
    rejected_wordlists INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        (SELECT COUNT(*)::INTEGER FROM words WHERE user_id = user_id_param) as total_words,
        (SELECT COUNT(*)::INTEGER FROM words WHERE user_id = user_id_param AND content_status = 'approved') as approved_words,
        (SELECT COUNT(*)::INTEGER FROM words WHERE user_id = user_id_param AND content_status = 'pending') as pending_words,
        (SELECT COUNT(*)::INTEGER FROM words WHERE user_id = user_id_param AND content_status = 'rejected') as rejected_words,
        (SELECT COUNT(*)::INTEGER FROM wordlists WHERE user_id = user_id_param) as total_wordlists,
        (SELECT COUNT(*)::INTEGER FROM wordlists WHERE user_id = user_id_param AND content_status = 'approved') as approved_wordlists,
        (SELECT COUNT(*)::INTEGER FROM wordlists WHERE user_id = user_id_param AND content_status = 'pending') as pending_wordlists,
        (SELECT COUNT(*)::INTEGER FROM wordlists WHERE user_id = user_id_param AND content_status = 'rejected') as rejected_wordlists;
END;
$$ LANGUAGE plpgsql;