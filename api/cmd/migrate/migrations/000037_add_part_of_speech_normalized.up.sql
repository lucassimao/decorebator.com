-- Add normalized part of speech column for language-agnostic logic
ALTER TABLE definitions ADD COLUMN part_of_speech_normalized VARCHAR(50);

-- Create index for query performance
CREATE INDEX idx_definitions_part_of_speech_normalized ON definitions(part_of_speech_normalized);

-- Populate normalized values for existing English data
UPDATE definitions 
SET part_of_speech_normalized = part_of_speech 
WHERE language = 'en';

-- Populate normalized values for existing Spanish data
UPDATE definitions 
SET part_of_speech_normalized = CASE 
    WHEN part_of_speech = 'verbo' THEN 'verb'
    WHEN part_of_speech = 'sustantivo' THEN 'noun'
    WHEN part_of_speech = 'pronombre' THEN 'pronoun'
    WHEN part_of_speech = 'adjetivo' THEN 'adjective'
    WHEN part_of_speech = 'adverbio' THEN 'adverb'
    WHEN part_of_speech LIKE 'preposici%n' THEN 'preposition'
    WHEN part_of_speech LIKE 'conjunci%n' THEN 'conjunction'
    WHEN part_of_speech LIKE 'interjecci%n' THEN 'interjection'
    ELSE part_of_speech
END
WHERE language = 'es';

-- Populate normalized values for existing French data
UPDATE definitions 
SET part_of_speech_normalized = CASE 
    WHEN part_of_speech = 'verbe' THEN 'verb'
    WHEN part_of_speech = 'nom' THEN 'noun'
    WHEN part_of_speech = 'pronom' THEN 'pronoun'
    WHEN part_of_speech = 'adjectif' THEN 'adjective'
    WHEN part_of_speech = 'adverbe' THEN 'adverb'
    WHEN part_of_speech LIKE 'pr%position' THEN 'preposition'
    WHEN part_of_speech = 'conjonction' THEN 'conjunction'
    WHEN part_of_speech = 'interjection' THEN 'interjection'
    ELSE part_of_speech
END
WHERE language = 'fr';

-- Populate normalized values for existing German data
UPDATE definitions 
SET part_of_speech_normalized = CASE 
    WHEN part_of_speech = 'Verb' THEN 'verb'
    WHEN part_of_speech = 'Substantiv' THEN 'noun'
    WHEN part_of_speech = 'Pronomen' THEN 'pronoun'
    WHEN part_of_speech = 'Adjektiv' THEN 'adjective'
    WHEN part_of_speech = 'Adverb' THEN 'adverb'
    WHEN part_of_speech LIKE 'Pr%position' THEN 'preposition'
    WHEN part_of_speech = 'Konjunktion' THEN 'conjunction'
    WHEN part_of_speech = 'Interjektion' THEN 'interjection'
    ELSE part_of_speech
END
WHERE language = 'de';

-- Populate normalized values for existing Italian data
UPDATE definitions 
SET part_of_speech_normalized = CASE 
    WHEN part_of_speech = 'verbo' THEN 'verb'
    WHEN part_of_speech = 'sostantivo' THEN 'noun'
    WHEN part_of_speech = 'pronome' THEN 'pronoun'
    WHEN part_of_speech = 'aggettivo' THEN 'adjective'
    WHEN part_of_speech = 'avverbio' THEN 'adverb'
    WHEN part_of_speech = 'preposizione' THEN 'preposition'
    WHEN part_of_speech = 'congiunzione' THEN 'conjunction'
    WHEN part_of_speech = 'interiezione' THEN 'interjection'
    ELSE part_of_speech
END
WHERE language = 'it';

-- Populate normalized values for existing Portuguese data
UPDATE definitions 
SET part_of_speech_normalized = CASE 
    WHEN part_of_speech = 'verbo' THEN 'verb'
    WHEN part_of_speech = 'substantivo' THEN 'noun'
    WHEN part_of_speech = 'pronome' THEN 'pronoun'
    WHEN part_of_speech = 'adjetivo' THEN 'adjective'
    WHEN part_of_speech LIKE 'adv%rbio' THEN 'adverb'
    WHEN part_of_speech LIKE 'preposi%' THEN 'preposition'
    WHEN part_of_speech LIKE 'conjun%' THEN 'conjunction'
    WHEN part_of_speech LIKE 'interjei%' THEN 'interjection'
    ELSE part_of_speech
END
WHERE language = 'pt';

-- Populate normalized values for existing Japanese data (using LIKE patterns)
UPDATE definitions 
SET part_of_speech_normalized = CASE 
    WHEN part_of_speech LIKE '%詞' AND part_of_speech LIKE '動%' THEN 'verb'
    WHEN part_of_speech LIKE '%詞' AND part_of_speech LIKE '名%' THEN 'noun'
    WHEN part_of_speech LIKE '%詞' AND part_of_speech LIKE '代名%' THEN 'pronoun'
    WHEN part_of_speech LIKE '%詞' AND part_of_speech LIKE '形容%' THEN 'adjective'
    WHEN part_of_speech LIKE '%詞' AND part_of_speech LIKE '副%' THEN 'adverb'
    WHEN part_of_speech LIKE '%詞' AND part_of_speech LIKE '助%' THEN 'preposition'
    WHEN part_of_speech LIKE '%詞' AND part_of_speech LIKE '接続%' THEN 'conjunction'
    WHEN part_of_speech LIKE '%詞' AND part_of_speech LIKE '感動%' THEN 'interjection'
    ELSE part_of_speech
END
WHERE language = 'ja';