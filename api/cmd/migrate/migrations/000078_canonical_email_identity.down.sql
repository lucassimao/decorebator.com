BEGIN;

SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';

CREATE UNIQUE INDEX users_email_unique_lower ON users (LOWER(email));
DROP INDEX users_email_unique_canonical;

COMMIT;
