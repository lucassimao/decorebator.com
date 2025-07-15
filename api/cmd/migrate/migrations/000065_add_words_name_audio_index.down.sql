-- Remove optimized index for word audio URL lookups by name
DROP INDEX IF EXISTS idx_words_name_audio_id;