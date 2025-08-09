-- Add guest_email column to quiz_attempts to store optional contact
ALTER TABLE quiz_attempts
ADD COLUMN IF NOT EXISTS guest_email TEXT;

-- Ensure one entry per (quiz_id, guest_email) when email is provided
CREATE UNIQUE INDEX IF NOT EXISTS unique_quiz_attempt_guest_email
ON quiz_attempts(public_quiz_id, guest_email)
WHERE guest_email IS NOT NULL;


