-- Revert: remove guest_email column from quiz_attempts
DROP INDEX IF EXISTS unique_quiz_attempt_guest_email;

ALTER TABLE quiz_attempts
DROP COLUMN IF EXISTS guest_email;


