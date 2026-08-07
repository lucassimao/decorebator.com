BEGIN;

SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';

-- Serialize against every AUTH-3 signup/token writer before cleanup. Writers
-- take a shared lock on this singleton in their transaction, so either they
-- commit before this update and are cleaned below, or they observe FALSE and
-- refuse to create new AUTH-3 state while old binaries are restored.
UPDATE auth_hardening_rollout_state
SET writes_enabled = FALSE, updated_at = NOW()
WHERE singleton = TRUE;

-- Roll back feature state without removing compatibility schema. A later
-- contract migration may drop these tables only after all AUTH-3 binaries and
-- workers have been retired. V2 reset envelopes are deliberately non-hex, so
-- restored pre-AUTH-3 binaries reject links issued by AUTH-3 instead of
-- bypassing durable single-use consumption.
DELETE FROM users
WHERE id IN (SELECT user_id FROM pending_email_verifications);

UPDATE password_reset_tokens
SET consumed_at = COALESCE(consumed_at, NOW());

COMMIT;
