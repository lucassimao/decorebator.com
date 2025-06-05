BEGIN;


CREATE TABLE IF NOT EXISTS leitner_system_history(
   id serial PRIMARY KEY,
   at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
   box_id int not null,
   leitner_system_tracking_id int NOT NULL,
   success boolean,
   quiz_type varchar(50),
   response_time_ms int

);

ALTER TABLE leitner_system_history ADD FOREIGN KEY (leitner_system_tracking_id) 
REFERENCES leitner_system_tracking(id)
ON DELETE CASCADE;

COMMIT;