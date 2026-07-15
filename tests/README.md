# E2E tests

Player-vs-player tests that drive the real `ttt` server binary and the
`ttt-client` binary over HTTP, all inside the test container.

## How it runs

- `just test` builds the test image (`_env/test/Dockerfile`) and runs it
  with the repo mounted at `/app` and a named volume for the go cache
- the image only provides the environment (go toolchain, bash, jq, curl);
- inside the container `tests/run_tests.sh`:
  - loads the test config: a `.env` mounted at the repo root wins,
    otherwise `_env/test/.env` from the repo mount is used
  - runs the unit tests (`go test ./...`)
  - builds the server and client binaries into a temp dir
  - starts the server as a separate background process
  - runs every script in `tests/scenarios/` and reports PASS/FAIL
- `just test-unit` still runs only the unit tests on the host machine

## Layout

- `run_tests.sh` - orchestrator: unit tests, build, server lifecycle,
  scenario loop
- `lib.sh` - shared helpers: per-player client wrapper, json asserts,
  error-code asserts, server start/stop
- `scenarios/01_auth.sh` - login tokens, session reuse, invalid token
- `scenarios/02_lobby.sh` - create, lobby listing, join, lobby removal
- `scenarios/03_pvp_game.sh` - full game to a win, state checks,
  move after finish
- `scenarios/04_private_game.sh` - join code flow, hidden from lobby
- `scenarios/05_invalid_moves.sh` - out of turn, occupied cell,
  out of bounds, bad game token
- `scenarios/06_token_expiry.sh` - short-ttl session expiry, explicit
  and automatic refresh, expired refresh token, replaced session token
- `scenarios/07_private_token_validation.sh` - join code required and
  checked, no third player, session token vs game token, opponent and
  cross-game token rejection

## Test config

`_env/test/.env` holds the container test settings and is picked up
through the repo mount (or can be mounted directly as `/app/.env`):

- `JWT_SECRET`, `GAME_SERVER_HOST`, `GAME_SERVER_PORT` - server env
- `TEST_SHORT_TTL_SECONDS` - requested token ttl for expiry tests,
  kept at 1 second so the expiry scenarios stay fast
- `TEST_EXPIRY_WAIT_SECONDS` - how long scenarios wait for a short
  token to expire

## Player isolation

Each logical player gets its own `HOME` under the temp work dir, so the
client's file session store (`~/.ttt/session.json`) keeps sessions apart
without any client changes.

## Contract

The scenarios code against the one-shot client actions and JSON output
described in `changes.ai.md` (step 2 work). Until those changes land the
scenarios are expected to fail; the harness itself is complete.
