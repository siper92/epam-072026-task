#!/usr/bin/env bash
set -euo pipefail
source "/app/tests/lib.sh"

login lobby-p1 secret-1 >/dev/null
login lobby-p2 secret-2 >/dev/null

created="$(ttt lobby-p1 action create)"
game_id="$(json_get '.id | tostring' "$created")"
assert_json '.game_token | length > 0' "$created" "create should return a game token"

waiting="$(ttt lobby-p2 action list)"
assert_json "[.games[].id | tostring] | index(\"$game_id\") != null" "$waiting" \
    "new game should show in the lobby"

joined="$(ttt lobby-p2 action join "$game_id")"
assert_json '.game_token | length > 0' "$joined" "join should return a game token"

after="$(ttt lobby-p1 action list)"
assert_json "[.games[].id | tostring] | index(\"$game_id\") == null" "$after" \
    "joined game should leave the lobby"
