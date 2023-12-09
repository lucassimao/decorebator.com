BEGIN;

CREATE TABLE IF NOT EXISTS words(
   id serial PRIMARY KEY,
   name VARCHAR NOT NULL,
   created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
   updated_at timestamp with time zone,
   user_id int NOT NULL,
   wordlist_id int NOT NULL
);

ALTER TABLE words ADD FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE words ADD FOREIGN KEY (wordlist_id) REFERENCES wordlists(id);

COMMIT;
