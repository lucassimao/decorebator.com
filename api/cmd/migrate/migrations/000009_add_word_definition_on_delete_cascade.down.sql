-- Drop the existing foreign key constraint
ALTER TABLE word_definitions DROP CONSTRAINT word_definitions_definition_id_fkey;
ALTER TABLE word_definitions DROP CONSTRAINT word_definitions_word_id_fkey;


-- Recreate the foreign key constraints without the cascade on delete
ALTER TABLE word_definitions
ADD CONSTRAINT word_definitions_definition_id_fkey
FOREIGN KEY (definition_id) REFERENCES definitions(id);

ALTER TABLE word_definitions
ADD CONSTRAINT word_definitions_word_id_fkey
FOREIGN KEY (word_id) REFERENCES words(id);