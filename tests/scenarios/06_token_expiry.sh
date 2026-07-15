#!/usr/bin/env bash
set -euo pipefail
source "/app/tests/lib.sh"

short_ttl="${TEST_SHORT_TTL_SECONDS:-1}"
expiry_wait="${TEST_EXPIRY_WAIT_SECONDS:-2}"

out="$(login exp-alice secret-a --session-ttl "$short_ttl")"
stale_session="$(json_get '.session_token' "$out")"

sleep "$expiry_wait"

expect_err_code INVALID_TOKEN ttt exp-stale --token "$stale_session" action list

refreshed="$(ttt exp-alice action refresh)"
assert_json '.session_token | length > 0' "$refreshed" \
    "refresh should issue a session token"
new_session="$(json_get '.session_token' "$refreshed")"
if [ "$new_session" = "$stale_session" ]; then
    fail "refresh should issue a different session token"
fi

listed="$(ttt exp-alice action list)"
assert_json 'has("games")' "$listed" "list should work after refresh"

login exp-bob secret-b --session-ttl "$short_ttl" >/dev/null
sleep "$expiry_wait"
auto="$(ttt exp-bob action list)"
assert_json 'has("games")' "$auto" \
    "client should auto refresh an expired session"

login exp-carol secret-c --session-ttl "$short_ttl" --token-ttl "$short_ttl" >/dev/null
sleep "$expiry_wait"
expect_err_code INVALID_TOKEN ttt exp-carol action refresh
relogged="$(ttt exp-carol --user exp-carol --password secret-c action list)"
assert_json 'has("games")' "$relogged" \
    "client should re-login when the refresh token expired"

first="$(login exp-dave secret-d)"
first_session="$(json_get '.session_token' "$first")"
login exp-dave secret-d --session-ttl 600 >/dev/null
expect_err_code INVALID_TOKEN ttt exp-stale --token "$first_session" action list
