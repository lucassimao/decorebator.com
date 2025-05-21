BEGIN;

ALTER TABLE definition_images DROP constraint  definition_images_definition_id_fkey;

ALTER TABLE definition_images
ADD CONSTRAINT definition_images_definition_id_fkey
FOREIGN KEY (definition_id) REFERENCES definitions(id);
COMMIT;
