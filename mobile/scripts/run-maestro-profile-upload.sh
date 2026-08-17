#!/usr/bin/env bash
set -euo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
mobile_dir=$(cd "$script_dir/.." && pwd)
repo_root=$(cd "$mobile_dir/.." && pwd)
api_dir="$repo_root/api"
maestro_bin="${MAESTRO_BIN:-$HOME/.maestro/bin/maestro}"

: "${EXPO_PUBLIC_TEST_USER_EMAIL:?EXPO_PUBLIC_TEST_USER_EMAIL is required}"

profile_url() {
  docker compose -f "$api_dir/docker-compose.yml" exec -T \
    -e SEED_EMAIL="$EXPO_PUBLIC_TEST_USER_EMAIL" postgres sh -s <<'CONTAINER'
psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -At -v seed_email="$SEED_EMAIL" <<'SQL'
SELECT COALESCE(profile_picture_url, '')
FROM users
WHERE lower(email) = lower(:'seed_email')
LIMIT 1;
SQL
CONTAINER
}

baseline_url=$(profile_url)
"$maestro_bin" test "$mobile_dir/.maestro/profile-image-upload.yaml"

current_url=""
for _ in $(seq 1 30); do
  current_url=$(profile_url)
  if [ -n "$current_url" ] && [ "$current_url" != "$baseline_url" ]; then
    break
  fi
  sleep 1
done

if [ -z "$current_url" ] || [ "$current_url" = "$baseline_url" ]; then
  echo "Profile upload did not persist a new database URL." >&2
  exit 1
fi
if [[ ! "$current_url" =~ ^http://[^/]+/([a-z0-9][a-z0-9.-]*)/(users/[0-9]+-[0-9a-f]{32}\.jpg)$ ]]; then
  echo "Persisted profile URL does not satisfy the local server-owned key contract." >&2
  exit 1
fi

bucket="${BASH_REMATCH[1]}"
object_key="${BASH_REMATCH[2]}"
"$HOME/go/bin/godotenv" -f "$api_dir/.env" docker run --rm \
  --network decorebator_default \
  --entrypoint /bin/sh \
  -e MINIO_ROOT_USER \
  -e MINIO_ROOT_PASSWORD \
  -e TARGET_BUCKET="$bucket" \
  -e TARGET_OBJECT="$object_key" \
  minio/mc -c '
    set -eu
    mc alias set local http://minio:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" >/dev/null
    mc stat "local/$TARGET_BUCKET/$TARGET_OBJECT" >/dev/null
  '

echo "Profile upload persisted a fresh server-owned URL and MinIO object."
