-- Add virtual column to compute isVerbType based on part_of_speech_normalized
ALTER TABLE definitions 
ADD COLUMN is_verb_type BOOLEAN GENERATED ALWAYS AS (
    part_of_speech_normalized = 'verb' OR part_of_speech_normalized = 'phrasal verb'
) STORED;

-- Create index for query performance
CREATE INDEX idx_definitions_is_verb_type ON definitions(is_verb_type);