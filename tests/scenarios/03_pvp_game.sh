#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib.sh"

login pvp-x secret-x >/dev/null
login pvp-o secret-o >/dev/null

created="$(ttt pvp-x action create)"
game_id="$(json_get '.id | tostring' "$created")"
x_token="$(json_get '.game_token' "$created")"

joined="$(ttt pvp-o action join "$game_id")"
o_token="$(json_get '.game_token' "$joined")"

state="$(ttt pvp-x action state "$game_id")"
assert_json '.status == "in_progress"' "$state" "game should start after join"
assert_json '.next == "x"' "$state" "creator should move first"

state="$(ttt pvp-x action move "$game_id" a1 --game-token "$x_token")"
assert_json '.board[0][0] == "x"' "$state" "a1 should hold x"
assert_json '.next == "o"' "$state" "o should be next after x moves"

ttt pvp-o action move "$game_id" b1 --game-token "$o_token" >/dev/null
ttt pvp-x action move "$game_id" a2 --game-token "$x_token" >/dev/null
ttt pvp-o action move "$game_id" b2 --game-token "$o_token" >/dev/null
final="$(ttt pvp-x action move "$game_id" a3 --game-token "$x_token")"

assert_json '.status == "finished"' "$final" "top row should finish the game"
assert_json '.winner == "x"' "$final" "x should win with the top row"

seen="$(ttt pvp-o action state "$game_id")"
assert_json '.status == "finished"' "$seen" "both players should see the finished game"

expect_err_code GAME_FINISHED ttt pvp-o action move "$game_id" c1 --game-token "$o_token"
