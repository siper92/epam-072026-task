#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib.sh"

out="$(login auth-alice secret-a)"
assert_json 'has("player_id")' "$out" "login should return a player id"
assert_json '.session_token | length > 0' "$out" "login should return a session token"
assert_json '.refresh_token | length > 0' "$out" "login should return a refresh token"

ttt auth-alice action list >/dev/null

again="$(login auth-alice secret-a)"
assert_json '.session_token | length > 0' "$again" "second login should issue a session token"

expect_err_code INVALID_TOKEN ttt auth-mallory --token bogus-token action list
