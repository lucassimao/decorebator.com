BEGIN;

-- in leitner_system_tracking table
DROP INDEX IF EXISTS idx_unique_definition_id_word_id;

ALTER TABLE leitner_system_tracking DROP COLUMN word_id;

CREATE UNIQUE INDEX idx_unique_definition_id_user_id
    ON leitner_system_tracking (definition_id, user_id);

COMMIT;