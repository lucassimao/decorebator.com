#!/usr/bin/env bash
set -euo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
repo_root=$(cd "$script_dir/../.." && pwd)
api_dir="$repo_root/api"

: "${EXPO_PUBLIC_TEST_USER_EMAIL:?EXPO_PUBLIC_TEST_USER_EMAIL is required}"
: "${EXPO_PUBLIC_TEST_USER_PASSWORD:?EXPO_PUBLIC_TEST_USER_PASSWORD is required}"

if ! command -v htpasswd >/dev/null 2>&1; then
  echo "htpasswd is required to generate the local test password hash" >&2
  exit 1
fi

password_hash=$(printf '%s\n' "$EXPO_PUBLIC_TEST_USER_PASSWORD" | htpasswd -niBC 10 "" | tr -d ':\n')

SEED_EMAIL="$EXPO_PUBLIC_TEST_USER_EMAIL" \
SEED_PASSWORD_HASH="$password_hash" \
  docker compose -f "$api_dir/docker-compose.yml" exec -T \
  -e SEED_EMAIL -e SEED_PASSWORD_HASH postgres sh -s <<'CONTAINER_SCRIPT'
set -eu

psql -v ON_ERROR_STOP=1 -q \
  -U "$POSTGRES_USER" \
  -d "$POSTGRES_DB" \
  -v seed_email="$SEED_EMAIL" \
  -v seed_password_hash="$SEED_PASSWORD_HASH" <<'SQL'
BEGIN;

INSERT INTO users (
  first_name,
  last_name,
  email,
  password_hash,
  preferred_language
)
SELECT
  'Maestro',
  'Tester',
  :'seed_email',
  :'seed_password_hash',
  'en'
WHERE NOT EXISTS (
  SELECT 1 FROM users WHERE lower(email) = lower(:'seed_email')
);

UPDATE users
SET password_hash = :'seed_password_hash'
WHERE lower(email) = lower(:'seed_email');

WITH fixture_user AS (
  SELECT id
  FROM users
  WHERE lower(email) = lower(:'seed_email')
), inserted_deck AS (
  INSERT INTO wordlists (
    name,
    description,
    user_id,
    language_code,
    pronunciation_system,
    words_count
  )
  SELECT
    'Maestro Deck',
    'Local end-to-end fixture',
    fixture_user.id,
    'en',
    'ipa',
    1
  FROM fixture_user
  WHERE NOT EXISTS (
    SELECT 1
    FROM wordlists
    WHERE user_id = fixture_user.id
      AND name = 'Maestro Deck'
  )
  RETURNING id, user_id
), fixture_deck AS (
  SELECT id, user_id FROM inserted_deck
  UNION ALL
  SELECT wordlists.id, wordlists.user_id
  FROM wordlists
  JOIN fixture_user ON fixture_user.id = wordlists.user_id
  WHERE wordlists.name = 'Maestro Deck'
  LIMIT 1
), inserted_word AS (
  INSERT INTO words (
    name,
    user_id,
    wordlist_id,
    processing_status,
    processing_completed_at
  )
  SELECT
    'serendipity',
    fixture_deck.user_id,
    fixture_deck.id,
    'completed',
    NOW()
  FROM fixture_deck
  WHERE NOT EXISTS (
    SELECT 1
    FROM words
    WHERE wordlist_id = fixture_deck.id
      AND name = 'serendipity'
  )
  RETURNING id
), fixture_word AS (
  SELECT id FROM inserted_word
  UNION ALL
  SELECT words.id
  FROM words
  JOIN fixture_deck ON fixture_deck.id = words.wordlist_id
  WHERE words.name = 'serendipity'
  LIMIT 1
), inserted_definition AS (
  INSERT INTO definitions (
    token,
    language,
    part_of_speech,
    part_of_speech_normalized,
    meaning,
    examples,
    source
  )
  SELECT
    'serendipity',
    'en',
    'noun',
    'noun',
    'A fortunate discovery made by chance',
    ARRAY['A useful idea appeared through serendipity.'],
    'local-maestro-fixture'
  WHERE NOT EXISTS (
    SELECT 1
    FROM definitions
    WHERE token = 'serendipity'
      AND source = 'local-maestro-fixture'
  )
  RETURNING id
), fixture_definition AS (
  SELECT id FROM inserted_definition
  UNION ALL
  SELECT id
  FROM definitions
  WHERE token = 'serendipity'
    AND source = 'local-maestro-fixture'
  LIMIT 1
)
INSERT INTO word_definitions (word_id, definition_id)
SELECT fixture_word.id, fixture_definition.id
FROM fixture_word, fixture_definition
WHERE NOT EXISTS (
  SELECT 1
  FROM word_definitions
  WHERE word_id = fixture_word.id
    AND definition_id = fixture_definition.id
);

UPDATE wordlists
SET words_count = (
  SELECT COUNT(*)
  FROM words
  WHERE words.wordlist_id = wordlists.id
)
WHERE user_id = (
  SELECT id FROM users WHERE lower(email) = lower(:'seed_email')
)
  AND name = 'Maestro Deck';

COMMIT;
SQL
CONTAINER_SCRIPT

echo "Local Maestro fixture is ready."
