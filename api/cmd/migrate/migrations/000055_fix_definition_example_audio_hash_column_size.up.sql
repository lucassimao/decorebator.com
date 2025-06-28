-- Fix definition_example_audio.example_hash column size from VARCHAR(32) to VARCHAR(64)
-- SHA256 hash produces 32 bytes which becomes 64 characters when hex-encoded
-- This resolves potential "value too long for type character varying(32)" errors in example audio generation

ALTER TABLE definition_example_audio 
ALTER COLUMN example_hash TYPE VARCHAR(64);