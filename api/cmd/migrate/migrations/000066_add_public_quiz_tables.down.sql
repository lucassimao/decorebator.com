-- Drop tables (in reverse order due to foreign key constraints)
DROP TABLE IF EXISTS quiz_attempts;
DROP TABLE IF EXISTS public_quizzes;

-- Drop enum
DROP TYPE IF EXISTS quiz_difficulty;
