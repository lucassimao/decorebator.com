BEGIN;

CREATE TABLE IF NOT EXISTS leitner_system_tracking(
   id serial PRIMARY KEY,
   box_id int NOT NULL,
   created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
   updated_at timestamp with time zone,
   user_id int NOT NULL,
   definition_id int NOT NULL
);

ALTER TABLE leitner_system_tracking ADD FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE leitner_system_tracking ADD FOREIGN KEY (definition_id) REFERENCES definitions(id) ON DELETE CASCADE;

CREATE INDEX idx_user_id ON leitner_system_tracking (user_id);
CREATE UNIQUE INDEX idx_unique_definition_id_user_id ON leitner_system_tracking (definition_id, user_id);

COMMIT;
