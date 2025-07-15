-- Add optimized index for word audio URL lookups by name
-- This partial index only includes words with valid audio URLs and is pre-sorted by id DESC
CREATE INDEX IF NOT EXISTS idx_words_name_audio_id ON words(name, id DESC) 
WHERE audio_url IS NOT NULL AND LENGTH(audio_url) > 0;