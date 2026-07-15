#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib.sh"

login inv-x secret-x >/dev/null
login inv-o secret-o >/dev/null

created="$(ttt inv-x action create)"
game_id="$(json_get '.id | tostring' "$created")"
x_token="$(json_get '.game_token' "$created")"

joined="$(ttt inv-o action join "$game_id")"
o_token="$(json_get '.game_token' "$joined")"

expect_err_code OUT_OF_TURN ttt inv-o action move "$game_id" a1 --game-token "$o_token"

ttt inv-x action move "$game_id" a1 --game-token "$x_token" >/dev/null

expect_err_code CELL_OCCUPIED ttt inv-o action move "$game_id" a1 --game-token "$o_token"
expect_err_code OUT_OF_BOUNDS ttt inv-o action move "$game_id" d4 --game-token "$o_token"
expect_err_code INVALID_TOKEN ttt inv-o action move "$game_id" b1 --game-token bogus-game-token
