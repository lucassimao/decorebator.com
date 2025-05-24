BEGIN;

-- in leitner_system_tracking table
DROP INDEX IF EXISTS idx_unique_definition_id_user_id;


ALTER TABLE leitner_system_tracking ADD COLUMN word_id int;

ALTER TABLE leitner_system_tracking
    ADD CONSTRAINT leitner_system_tracking_word_id_fk
    FOREIGN KEY (word_id) REFERENCES words(id)
    ON DELETE CASCADE;


UPDATE leitner_system_tracking lst 
    SET word_id=(SELECT w.id FROM word_definitions wd JOIN words w ON w.id=wd.word_id WHERE definition_id=lst.definition_id AND w.user_id=lst.user_id );


ALTER TABLE leitner_system_tracking ALTER COLUMN word_id SET NOT NULL;

CREATE UNIQUE INDEX idx_unique_definition_id_word_id
    ON leitner_system_tracking (definition_id, word_id);

COMMIT;