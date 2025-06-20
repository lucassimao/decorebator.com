-- Remove pronunciation system support from wordlists
DROP INDEX IF EXISTS idx_wordlists_pronunciation_system;

ALTER TABLE wordlists 
DROP COLUMN IF EXISTS pronunciation_system;

DROP TYPE IF EXISTS pronunciation_system;