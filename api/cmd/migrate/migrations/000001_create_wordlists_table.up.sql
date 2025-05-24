CREATE TABLE IF NOT EXISTS wordlists(
   id serial PRIMARY KEY,
   name VARCHAR NOT NULL,
   description text NOT NULL,
   created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
   updated_at timestamp with time zone,
   user_id int
);