-- Rollback example_hash column size from VARCHAR(64) to VARCHAR(32)
-- WARNING: This may fail if there are existing hashes longer than 32 characters

ALTER TABLE example_usage 
ALTER COLUMN example_hash TYPE VARCHAR(32);