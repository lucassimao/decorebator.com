#!/usr/bin/env bash
set -euo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
repo_root=$(cd "$script_dir/../.." && pwd)
registry_file="$repo_root/docs/fixtures/store-approval-accounts.json"
query_file="$script_dir/release-entitlement-audit.sql"

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required to load the controlled store-approval registry" >&2
  exit 1
fi

if ! command -v psql >/dev/null 2>&1; then
  echo "psql is required to run the read-only entitlement audit" >&2
  exit 1
fi

registry_json=$(jq -c '[.accounts[] | select(.classification == "owner-confirmed-store-approval") | .userId]' "$registry_file")
db_url="${DATABASE_URL:-}"

if [[ -z "$db_url" && -f "$repo_root/api/.env.prod" ]]; then
  db_url=$(sed -n 's/^DATABASE_URL=//p' "$repo_root/api/.env.prod")
fi

if [[ -z "$db_url" ]]; then
  echo "DATABASE_URL is required; no value was found in the environment or api/.env.prod" >&2
  exit 1
fi

psql "$db_url" \
  -X \
  -v registry_json="$registry_json" \
  -f "$query_file"
