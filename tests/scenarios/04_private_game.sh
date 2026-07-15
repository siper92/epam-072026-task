#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib.sh"

login priv-p1 secret-1 >/dev/null
login priv-p2 secret-2 >/dev/null

created="$(ttt priv-p1 action create --private)"
game_id="$(json_get '.id | tostring' "$created")"
join_code="$(json_get '.join_code' "$created")"

waiting="$(ttt priv-p2 action list)"
assert_json "[.games[].id | tostring] | index(\"$game_id\") == null" "$waiting" \
    "private game should not show in the lobby"

expect_err ttt priv-p2 action join "$game_id" --code wrong-code

joined="$(ttt priv-p2 action join "$game_id" --code "$join_code")"
assert_json '.game_token | length > 0' "$joined" "join with code should return a game token"

state="$(ttt priv-p1 action state "$game_id")"
assert_json '.status == "in_progress"' "$state" "private game should start after join"
