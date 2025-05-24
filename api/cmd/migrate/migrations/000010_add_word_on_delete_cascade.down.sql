-- Drop the existing foreign key constraint
ALTER TABLE words DROP CONSTRAINT words_wordlist_id_fkey;


-- Recreate the foreign key constraints without the cascade on delete
ALTER TABLE words
ADD CONSTRAINT words_wordlist_id_fkey
FOREIGN KEY (wordlist_id) REFERENCES wordlists(id);