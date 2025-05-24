BEGIN;

CREATE TABLE IF NOT EXISTS definition_images(
   id serial PRIMARY KEY,
   api VARCHAR NOT NULL,
   url TEXT NOT NULL,
   description TEXT NOT NULL,
   model VARCHAR NOT NULL,
   prompt TEXT NOT NULL,
   is_visible BOOLEAN NOT NULL,
   created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
   definition_id int NOT NULL
);

ALTER TABLE definition_images ADD FOREIGN KEY (definition_id) REFERENCES definitions(id);

INSERT INTO definition_images(api,description, model, prompt,is_visible,definition_id, url)
    SELECT 'openai',examples[1], 'dall-e-3', examples[1], true,id, image_url 
    from definitions 
    where image_url is not null;

alter table definitions drop column image_url;

COMMIT;
