CREATE UNIQUE INDEX idx_error_reports_user_definition_word 
ON error_reports(
    user_id, 
    COALESCE(definition_id, -1), 
    COALESCE(word_id, -1)
) 
WHERE status = 'pending';