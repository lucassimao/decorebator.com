BEGIN;

CREATE TABLE IF NOT EXISTS users(
   id serial PRIMARY KEY,
   first_name VARCHAR NOT NULL,
   email VARCHAR NOT NULL,
   last_name VARCHAR NOT NULL,
   password_hash text NOT NULL,
   created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
   updated_at timestamp with time zone
);

ALTER TABLE wordlists ADD FOREIGN KEY (user_id) REFERENCES users(id);

COMMIT;