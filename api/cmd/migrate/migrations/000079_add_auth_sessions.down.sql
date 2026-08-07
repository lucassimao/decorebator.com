BEGIN;

SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';

DROP TABLE auth_refresh_tokens;
DROP TABLE auth_session_families;

COMMIT;
