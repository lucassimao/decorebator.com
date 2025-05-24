BEGIN;

ALTER TABLE definitions ADD COLUMN image_url TEXT;

UPDATE definitions
    SET image_url = url
    FROM definition_images
    WHERE definitions.id = definition_images.definition_id;


DROP TABLE definition_images;

COMMIT;
