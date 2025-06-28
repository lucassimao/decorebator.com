-- Fix example_hash column size from VARCHAR(32) to VARCHAR(64)
-- SHA256 hash produces 32 bytes which becomes 64 characters when hex-encoded
-- This resolves the "value too long for type character varying(32)" error

ALTER TABLE example_usage 
ALTER COLUMN example_hash TYPE VARCHAR(64);