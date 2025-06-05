-- Drop definitions that have empty or null part_of_speech
-- This will cascade delete from word_definitions and definition_images due to foreign key constraints

DELETE FROM definitions 
WHERE part_of_speech IS NULL 
   OR part_of_speech = '' 
   OR trim(part_of_speech) = '';

-- Log the cleanup
DO $$
DECLARE
    remaining_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO remaining_count FROM definitions;
    RAISE NOTICE 'Cleanup completed. Remaining definitions: %', remaining_count;
END $$;