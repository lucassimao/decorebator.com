-- Add pronunciation_system column to wordlists table
ALTER TABLE wordlists 
ADD COLUMN pronunciation_system varchar NOT NULL DEFAULT 'ipa';

-- Add index for performance
CREATE INDEX idx_wordlists_pronunciation_system ON wordlists(pronunciation_system);

-- Update existing wordlists based on their language
UPDATE wordlists 
SET pronunciation_system = CASE 
    WHEN language_code = 'ja' THEN 'romaji'
    WHEN language_code IN ('zh', 'zh-cn', 'zh-tw') THEN 'pinyin'
    WHEN language_code IN ('ko', 'kr') THEN 'hangul'
    ELSE 'ipa'
END;