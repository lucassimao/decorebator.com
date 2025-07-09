-- Revert email constraint modernization

-- Step 1: Recreate original case-sensitive unique index
CREATE UNIQUE INDEX users_email_unique ON users (email);

-- Step 2: Drop the case-insensitive unique index
DROP INDEX IF EXISTS users_email_unique_lower;