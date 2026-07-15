#!/usr/bin/env bash

log() {
    echo "[test] $*"
}

fail() {
    echo "[test] FAIL: $*" >&2
    exit 1
}

ttt() {
    local player="$1"
    shift
    mkdir -p "$TTT_WORK/home-$player"
    HOME="$TTT_WORK/home-$player" "$TTT_CLIENT_BIN" \
        --server "$TTT_SERVER_URL" \
        "$@"
}

login() {
    local player="$1"
    local password="$2"
    ttt "$player" --user "$player" --password "$password" action login
}

json_get() {
    local filter="$1"
    local json="$2"
    jq -er "$filter" <<<"$json"
}

assert_eq() {
    local want="$1"
    local got="$2"
    local msg="$3"
    if [ "$want" != "$got" ]; then
        fail "$msg: want [$want] got [$got]"
    fi
}

assert_json() {
    local filter="$1"
    local json="$2"
    local msg="$3"
    if ! jq -e "$filter" <<<"$json" >/dev/null; then
        fail "$msg: filter [$filter] on [$json]"
    fi
}

expect_err() {
    local out
    if out="$("$@" 2>&1)"; then
        fail "expected command to fail, got: $out"
    fi
}

expect_err_code() {
    local want="$1"
    shift
    local out
    local got
    if out="$("$@" 2>&1)"; then
        fail "expected error [$want], command succeeded: $out"
    fi
    got="$(jq -r '.code' <<<"$out" 2>/dev/null || echo "")"
    assert_eq "$want" "$got" "unexpected error code, output [$out]"
}

start_server() {
    (
        cd "$TTT_WORK"
        exec env \
            GAME_SERVER_HOST="$TTT_SERVER_HOST" \
            GAME_SERVER_PORT="$TTT_SERVER_PORT" \
            "$TTT_SERVER_BIN" serve
    ) >"$TTT_WORK/server.log" 2>&1 &
    TTT_SERVER_PID=$!
    wait_for_server
}

wait_for_server() {
    local i
    for i in $(seq 1 50); do
        if curl -s -o /dev/null "$TTT_SERVER_URL/api/games"; then
            log "server is up on $TTT_SERVER_URL"
            return 0
        fi
        sleep 0.2
    done
    echo "[test] server did not start" >&2
    cat "$TTT_WORK/server.log" >&2 || true
    return 1
}

stop_server() {
    if [ -n "${TTT_SERVER_PID:-}" ]; then
        kill "$TTT_SERVER_PID" 2>/dev/null || true
        wait "$TTT_SERVER_PID" 2>/dev/null || true
    fi
}
