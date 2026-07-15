#!/usr/bin/env bash
set -euo pipefail

TESTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$TESTS_DIR/.." && pwd)"

TTT_ENV_FILE="${TTT_ENV_FILE:-$ROOT_DIR/.env}"
if [ ! -f "$TTT_ENV_FILE" ]; then
    TTT_ENV_FILE="$ROOT_DIR/_env/test/.env"
fi
if [ -f "$TTT_ENV_FILE" ]; then
    echo "[test] loading env file $TTT_ENV_FILE"
    set -a
    source "$TTT_ENV_FILE"
    set +a
fi

export TTT_WORK="$(mktemp -d)"
export TTT_SERVER_BIN="$TTT_WORK/bin/ttt"
export TTT_CLIENT_BIN="$TTT_WORK/bin/ttt-client"
export TTT_SERVER_HOST="${GAME_SERVER_HOST:-127.0.0.1}"
export TTT_SERVER_PORT="${GAME_SERVER_PORT:-18080}"
export TTT_SERVER_URL="http://$TTT_SERVER_HOST:$TTT_SERVER_PORT"
export JWT_SECRET="${JWT_SECRET:-test-secret}"
export TEST_SHORT_TTL_SECONDS="${TEST_SHORT_TTL_SECONDS:-1}"
export TEST_EXPIRY_WAIT_SECONDS="${TEST_EXPIRY_WAIT_SECONDS:-2}"

source "$TESTS_DIR/lib.sh"

log "running unit tests"
(cd "$ROOT_DIR" && go test ./...)

log "building server and client binaries"
mkdir -p "$TTT_WORK/bin"
(cd "$ROOT_DIR" && go build -o "$TTT_SERVER_BIN" ./cli/server)
(cd "$ROOT_DIR" && go build -o "$TTT_CLIENT_BIN" ./cli/client)

start_server
trap stop_server EXIT

failed=0
for scenario in "$TESTS_DIR"/scenarios/*.sh; do
    name="$(basename "$scenario")"
    log "scenario: $name"
    if bash "$scenario"; then
        log "PASS: $name"
    else
        log "FAIL: $name"
        log "last server log lines:"
        tail -n 40 "$TTT_WORK/server.log" >&2 || true
        failed=1
    fi
done

if [ "$failed" -ne 0 ]; then
    log "RESULT: FAIL"
else
    log "RESULT: PASS"
fi
exit "$failed"
