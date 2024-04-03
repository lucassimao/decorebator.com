BEGIN;

CREATE TABLE IF NOT EXISTS wiktionary(
   id serial PRIMARY KEY,
   word VARCHAR NOT NULL,
   data jsonb,
   created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
   updated_at timestamp with time zone
);

COMMIT;
