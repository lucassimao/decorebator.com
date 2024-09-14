BEGIN;

CREATE TABLE IF NOT EXISTS error_reports(
   id serial PRIMARY KEY,
   is_resolved bool,
   error_type varchar NOT NULL, 
   quiz jsonb NOT NULL,
   user_id int NOT NULL,
   created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE error_reports ADD FOREIGN KEY (user_id) REFERENCES users(id);

COMMIT;
