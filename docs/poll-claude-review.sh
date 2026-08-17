#!/usr/bin/env bash

set -euo pipefail

usage() {
  echo "usage: $0 <prompt-file> <result-file> [repository]" >&2
  exit 2
}

fail() {
  echo "Claude review poll failed: $*" >&2
  exit 1
}

[[ $# -ge 2 && $# -le 3 ]] || usage

prompt_file=$1
result_file=$2
review_repository=${3:-$(pwd)}
review_attempt_seconds=${CLAUDE_REVIEW_ATTEMPT_SECONDS:-300}
codex_attempt_seconds=${CODEX_REVIEW_ATTEMPT_SECONDS:-1800}
codex_review_model=${CODEX_REVIEW_MODEL:-gpt-5.6-sol}

[[ -s "$prompt_file" ]] || fail "prompt file is missing or empty: ${prompt_file}"
[[ -d "$review_repository" ]] || fail "repository directory does not exist: ${review_repository}"
[[ -d "$(dirname "$result_file")" ]] || fail "result directory does not exist: $(dirname "$result_file")"
[[ "$review_attempt_seconds" =~ ^[1-9][0-9]*$ ]] || fail "CLAUDE_REVIEW_ATTEMPT_SECONDS must be a positive integer"
[[ "$codex_attempt_seconds" =~ ^[1-9][0-9]*$ ]] || fail "CODEX_REVIEW_ATTEMPT_SECONDS must be a positive integer"
command -v flock >/dev/null || fail "flock is required"
command -v realpath >/dev/null || fail "realpath is required"
command -v timeout >/dev/null || fail "timeout is required"

prompt_file=$(realpath "$prompt_file")
result_file=$(realpath -m "$result_file")
review_repository=$(realpath "$review_repository")
[[ "$prompt_file" != "$result_file" ]] || fail "prompt and result files must differ"

attempt_log="${result_file}.attempts.log"
lock_file="${result_file}.lock"

has_review_verdict() {
  local candidate=$1
  local final_line

  final_line=$(sed '/^[[:space:]]*$/d' "$candidate" | tail -n 1)
  [[ "$final_line" =~ ^##\ (APPROVED|CHANGES\ REQUIRED)[[:space:]]*$ ]]
}

is_substantive_review() {
  local candidate=$1
  local nonblank_lines

  [[ -s "$candidate" ]] || return 1
  nonblank_lines=$(awk 'NF { count++ } END { print count + 0 }' "$candidate")
  [[ "$nonblank_lines" -ge 2 ]] && [[ $(wc -c <"$candidate") -ge 200 ]] && has_review_verdict "$candidate"
}

if is_substantive_review "$result_file"; then
  echo "CLAUDE_REVIEW_READY existing_result=${result_file}"
  exit 0
fi

exec 9>"$lock_file"
flock -n 9 || fail "another poller already owns ${result_file}"

review_prompt=$(<"$prompt_file")
cycle=0
active_candidate=
active_trace=
active_pid=

cleanup() {
  if [[ -n "$active_pid" ]] && kill -0 "$active_pid" 2>/dev/null; then
    kill "$active_pid" 2>/dev/null || true
    wait "$active_pid" 2>/dev/null || true
  fi
  if [[ -n "$active_candidate" ]]; then
    rm -f -- "$active_candidate"
  fi
  if [[ -n "$active_trace" ]]; then
    rm -f -- "$active_trace"
  fi
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

printf 'poll_start time=%s repository=%s\n' "$(date --iso-8601=seconds)" "$review_repository" >>"$attempt_log"

cd "$review_repository"

while true; do
  cycle=$((cycle + 1))

  for model in fable opus sonnet; do
    active_candidate=$(mktemp "${result_file}.${model}.candidate.XXXXXX")
    printf 'CLAUDE_POLL cycle=%s model=%s time=%s\n' \
      "$cycle" "$model" "$(date --iso-8601=seconds)"

    if ! command -v claude >/dev/null; then
      printf 'cycle=%s model=%s exit=127 reason=claude-cli-unavailable time=%s\n' \
        "$cycle" "$model" "$(date --iso-8601=seconds)" >>"$attempt_log"
      rm -f -- "$active_candidate"
      active_candidate=
      continue
    fi

    timeout "${review_attempt_seconds}s" claude \
      -p \
      --model "$model" \
      --permission-mode plan \
      --no-session-persistence \
      --effort high \
      "$review_prompt" >"$active_candidate" 2>&1 &
    active_pid=$!

    heartbeat_seconds=0
    while kill -0 "$active_pid" 2>/dev/null; do
      sleep 1
      heartbeat_seconds=$((heartbeat_seconds + 1))
      if [[ "$heartbeat_seconds" -ge 55 ]] && kill -0 "$active_pid" 2>/dev/null; then
        printf 'CLAUDE_POLL_RUNNING cycle=%s model=%s time=%s\n' \
          "$cycle" "$model" "$(date --iso-8601=seconds)"
        heartbeat_seconds=0
      fi
    done

    if wait "$active_pid"; then
      attempt_status=0
    else
      attempt_status=$?
    fi
    active_pid=

    if [[ "$attempt_status" -eq 0 ]] && is_substantive_review "$active_candidate"; then
      mv -- "$active_candidate" "$result_file"
      active_candidate=
      printf 'CLAUDE_REVIEW_READY model=%s cycle=%s result=%s\n' \
        "$model" "$cycle" "$result_file"
      exit 0
    fi

    if [[ "$attempt_status" -eq 0 ]]; then
      attempt_status=75
    fi
    printf 'cycle=%s model=%s exit=%s time=%s\n' \
      "$cycle" "$model" "$attempt_status" "$(date --iso-8601=seconds)" >>"$attempt_log"
    sed -n '1,8p' "$active_candidate" >>"$attempt_log"
    rm -f -- "$active_candidate"
    active_candidate=
  done

  active_candidate=$(mktemp "${result_file}.codex-xhigh.candidate.XXXXXX")
  active_trace=$(mktemp "${result_file}.codex-xhigh.trace.XXXXXX")
  if ! command -v codex >/dev/null; then
    printf 'cycle=%s model=codex-xhigh exit=127 reason=codex-cli-unavailable time=%s\n' \
      "$cycle" "$(date --iso-8601=seconds)" >>"$attempt_log"
    rm -f -- "$active_candidate" "$active_trace"
    active_candidate=
    active_trace=
  else
    printf 'REVIEW_POLL cycle=%s model=codex-xhigh time=%s\n' \
      "$cycle" "$(date --iso-8601=seconds)"

  timeout "${codex_attempt_seconds}s" codex exec \
    --ephemeral \
    --ignore-user-config \
    --model "$codex_review_model" \
    -C "$review_repository" \
    -s read-only \
    -c 'model_reasoning_effort="xhigh"' \
    -o "$active_candidate" \
    - <"$prompt_file" >"$active_trace" 2>&1 &
  active_pid=$!

  heartbeat_seconds=0
  while kill -0 "$active_pid" 2>/dev/null; do
    sleep 1
    heartbeat_seconds=$((heartbeat_seconds + 1))
    if [[ "$heartbeat_seconds" -ge 55 ]] && kill -0 "$active_pid" 2>/dev/null; then
      printf 'REVIEW_POLL_RUNNING cycle=%s model=codex-xhigh time=%s\n' \
        "$cycle" "$(date --iso-8601=seconds)"
      heartbeat_seconds=0
    fi
  done

  if wait "$active_pid"; then
    attempt_status=0
  else
    attempt_status=$?
  fi
  active_pid=

  if [[ "$attempt_status" -eq 0 ]] && is_substantive_review "$active_candidate"; then
    mv -- "$active_candidate" "$result_file"
    active_candidate=
    rm -f -- "$active_trace"
    active_trace=
    printf 'REVIEW_READY model=codex-xhigh cycle=%s result=%s\n' \
      "$cycle" "$result_file"
    exit 0
  fi

  if [[ "$attempt_status" -eq 0 ]]; then
    attempt_status=75
  fi
  printf 'cycle=%s model=codex-xhigh exit=%s time=%s\n' \
    "$cycle" "$attempt_status" "$(date --iso-8601=seconds)" >>"$attempt_log"
  if [[ -s "$active_candidate" ]]; then
    sed -n '1,8p' "$active_candidate" >>"$attempt_log"
  else
    sed -n '1,8p' "$active_trace" >>"$attempt_log"
  fi
  rm -f -- "$active_candidate" "$active_trace"
  active_candidate=
  active_trace=
  fi

  if [[ "$cycle" -lt 5 ]]; then
    wait_steps=1
  else
    wait_steps=5
  fi

  printf 'CLAUDE_RETRY_WAIT_START cycle=%s steps=%s time=%s\n' \
    "$cycle" "$wait_steps" "$(date --iso-8601=seconds)"
  wait_step=0
  while [[ "$wait_step" -lt "$wait_steps" ]]; do
    sleep 55
    wait_step=$((wait_step + 1))
    printf 'CLAUDE_RETRY_WAIT cycle=%s step=%s/%s time=%s\n' \
      "$cycle" "$wait_step" "$wait_steps" "$(date --iso-8601=seconds)"
  done
done
