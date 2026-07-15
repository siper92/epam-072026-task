#!/usr/bin/env bash
set -euo pipefail
source "/app/tests/lib.sh"

p1_login="$(login ptok-p1 secret-1)"
p1_session="$(json_get '.session_token' "$p1_login")"
login ptok-p2 secret-2 >/dev/null
login ptok-p3 secret-3 >/dev/null

created="$(ttt ptok-p1 action create --private)"
game_id="$(json_get '.id | tostring' "$created")"
x_token="$(json_get '.game_token' "$created")"
join_code="$(json_get '.join_code' "$created")"

expect_err_code INVALID_INPUT ttt ptok-p2 action join "$game_id"
expect_err_code INVALID_INPUT ttt ptok-p2 action join "$game_id" --code wrong-code

joined="$(ttt ptok-p2 action join "$game_id" --code "$join_code")"
o_token="$(json_get '.game_token' "$joined")"

expect_err_code INVALID_TRANSITION ttt ptok-p3 action join "$game_id" --code "$join_code"

expect_err_code INVALID_TOKEN ttt ptok-p1 action move "$game_id" a1 --game-token "$p1_session"
expect_err_code INVALID_TOKEN ttt ptok-p1 action move "$game_id" a1 --game-token "$o_token"

other="$(ttt ptok-p3 action create --private)"
other_token="$(json_get '.game_token' "$other")"
expect_err_code INVALID_TOKEN ttt ptok-p3 action move "$game_id" a1 --game-token "$other_token"

state="$(ttt ptok-p1 action move "$game_id" a1 --game-token "$x_token")"
assert_json '.board[0][0] == "x"' "$state" \
    "creator should move with the matching game token"
