BEGIN;

CREATE TABLE IF NOT EXISTS definitions(
   id serial PRIMARY KEY,
   token VARCHAR NOT NULL,
   language VARCHAR NOT NULL,
   part_of_speech VARCHAR NOT NULL,
   meaning TEXT NOT NULL,
   examples TEXT[],
   inflections jsonb,
   created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
   updated_at timestamp with time zone
);

CREATE TABLE IF NOT EXISTS word_definitions(
   id serial PRIMARY KEY,
   word_id int NOT NULL,
   definition_id int NOT NULL,
   created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE word_definitions ADD FOREIGN KEY (word_id) REFERENCES words(id);
ALTER TABLE word_definitions ADD FOREIGN KEY (definition_id) REFERENCES definitions(id);

COMMIT;
