# Tic-Tac-Toe Assessment

This file provides guidance to users and agents when using the project.
The game focuses on a single game server and multiple clients connecting to it
 - distribution or multiple game servers are left for future work - but taken
   in account while designing the architecture.

AI is used in all steps of this project, my stamp on every step

I have focused on robust usage of tokens, validation and sensible security levels
 - tokens are used for joining lobbies
 - player actions - moves, game creation, game state retrieval
 - game token - validates access to moves in the current game

The game is not perfect I have tried to save time and such

## Project

A Tic-Tac-Toe game server = Go module `ticTacSolved/task`, Go 1.25.
 - this is deliberately named `ticTacSolved` to avoid naming companies,
   also emphasizes the spirit of the project

The game domain is intentionally **standard-library only**;
the HTTP layer uses `gin`, the CLI layer uses `cobra` and `viper`;
tools used: `sqlc`

Two top-level trees plus infrastructure:

- `game/` - the game domain: `state_machine` (rules), `auth` (tokens),
  `service` (lobby, play, queue, leaderboard), `data` (persistence,
  sqlc-generated)
- `cli/` - the `ttt` server binary (`cli/server`) and the `ttt-client`
  binary (`cli/client`)
- `pkg/` - cross-cutting infrastructure reused by the domain:
  `errs`, `config`, `util`, `api`, `session`
- `tests/` - end to end scenarios that drive the real binaries over HTTP
- `_env/` - Dockerfiles for the dev toolchain, the test runner and the
  production image

## How to run

Requirements: Go 1.25+; `just` and Docker are optional
(only needed for codegen, the containerized test run and the prod image).

`JWT_SECRET` must be set for the server to issue tokens
(`just run` defaults it for dev).

### Start the server

```
JWT_SECRET=dev-secret go run ./cli/server serve
```

 - flags: `--host` (default `127.0.0.1`), `--port` (default `8080`),
   `--storage` (default `memory`), `--db` (default `./_local/db.sqlite3`)
 - env overrides via viper: `GAME_SERVER_HOST`, `GAME_SERVER_PORT`,
   `GAME_DB_STORAGE`, `GAME_DB_PATH`, `GAME_DB_DRIVER`
 - defaults are documented in `cli/server/.env.default`

### Play a game - two terminals

Terminal 1 (player one):

```
go run ./cli/client --type cli --user alice --password one
ttt> queue
```

Terminal 2 (player two):

```
go run ./cli/client --type cli --user bob --password two
ttt> queue
```

The queue pairs the two waiting players into one game automatically.
Then alternate moves until a win or a draw; the final state is shown
to both players:

```
ttt> move a1        - move by cell name (a1..c3)
ttt> move 0 2       - or by row and column (0..2)
ttt> show           - render the current board
ttt> watch          - stream live updates until the game finishes
```

Other interactive commands: `list`, `create [private]`, `join <id> [code]`,
`leaders`, `help`, `quit`.

Instead of the queue, players can also pair manually:
one runs `create` (or `create private` and shares the join code),
the other runs `join <id>` from `list` (plus the code for private games).

### One shot client actions

Every interactive command also exists as a standalone action
(used by the e2e scenarios, supports `--output json`):

```
go run ./cli/client action login
go run ./cli/client action list
go run ./cli/client action create
go run ./cli/client action join <id>
go run ./cli/client action queue
go run ./cli/client action leaderboard
go run ./cli/client action show <id>
go run ./cli/client action state <id>
go run ./cli/client action move <row> <col>
go run ./cli/client action refresh
```

Client flags: `--server --user --password --type (cli|file) --token
--token-ttl --session-ttl --session-file --output (human|json)
--game-token`; every flag is also an env var (`TTT_` prefix) and can
live in a `.env` file - see `cli/client/.env.default`.
Session and refresh tokens are stored via `pkg/session`
(file store, `~/.ttt/session.json` by default).

## Commands (justfile)

Common tasks are driven by `just` (see `justfile`);
Docker is used primarily to provide a pinned toolchain for codegen.

- `just build` - builds the dev image used by `just compile` (sqlc)
- `just compile` - generates sqlc files inside the dev image
- `just compile-check` - runs `just compile` and aborts if generated code changed
- `just build-test` - builds the e2e test runner image
- `just test` - runs unit tests plus the e2e scenarios inside the test image
- `just test-unit` - runs `go test ./...` with the on machine go
- `just run` - runs the `ttt` server locally with a dev `JWT_SECRET`
- `just build-prod` - runs `compile-check`, then builds the `ttt-build-image`
  compile image and the `ttt-server-image` scratch runtime image
- `just run-prod` - runs the scratch image with port 8080 published

## Storage

The server supports two storage backends behind the same `data.Store`
interface, selected at startup:

- `memory` (default) - in-process store, state is lost on restart
- `sqlite` - sql store, survives restarts; the schema
  (`game/data/schema/schema.sql`) is applied automatically on first start

Choosing the storage and the database location:

```
go run ./cli/server serve --storage sqlite
go run ./cli/server serve --storage sqlite --db ./_local/other.sqlite3
GAME_DB_STORAGE=sqlite GAME_DB_PATH=/var/data/ttt.sqlite3 go run ./cli/server serve
```

 - `--storage` / `GAME_DB_STORAGE` - `memory` or `sqlite`,
   anything else fails startup with `INVALID_INPUT`
 - `--db` / `GAME_DB_PATH` - sqlite database file path,
   defaults to `./_local/db.sqlite3`; the parent directory is created
   automatically
 - `GAME_DB_DRIVER` (env only) - sql driver name, defaults to `sqlite`

Note: the module intentionally keeps the domain dependency-free, so no
sqlite driver is linked into the build yet; running with
`--storage sqlite` requires adding a driver (for example
`modernc.org/sqlite`, cgo free) via `go get` and a blank import in
`cli/server/main.go`.

## Auth design

Tokens are HS256 JWTs signed with `JWT_SECRET` and carry their own expiration:
 - `Issue` stamps an `exp` claim (unix seconds) into the token, so `Validate`
   checks expiration from the token itself, no store roundtrip needed
 - a `player_id` claim is required on every issued token

JWT was chosen over opaque session ids because validation is stateless
(signature plus `exp`), which keeps the hot request path off the store;
the tradeoff - early revocation - is handled by the
single-active-token rule below.

Register and login are one flow: `POST /api/login` derives a stable
player id from `sha256(user:password)` and creates the player on first
login. This is a deliberate simplification for the assessment - there is
no separate credential store, so a different password simply becomes a
different player identity.

Login returns three lifetimes of credentials:
 - session token - short lived (default 15m, capped at 1h),
   sent as `Authorization: Bearer ...` on every authenticated call
 - refresh token - longer lived (default 24h, capped at 7d),
   exchanged at `POST /api/refresh` for a new session without re-login
 - game token - issued when creating, joining or queueing into a game
   (30m), sent as `X-Game-Token` on moves so only the two participants
   of that game can move in it

Each player has exactly one active session token at a time:
 - issuing a new token replaces the previous one for that player
 - active tokens live in a package-level in-memory cache (`tokenCache`),
   keyed by player id; a ticker inside the cache struct evicts entries
   after a set time (`TokenCacheTTL`)
 - on a cache miss for a still-unexpired token, the store is checked;
   a found token is valid and is cached again
 - the `tokens` table keys on `player_id` (unique) and `SaveToken` upserts,
   so the single-active-token rule also holds at the DB level

Unauthenticated or invalid requests are rejected with `401` and
`{"code": "INVALID_TOKEN", "message": "..."}`.

## Shared interfaces (`cli/*/internal`)

The CLI layers expose small interfaces so parts can be shared and
faked in tests without touching the network or a real token service.

Client (`cli/client/internal/interfaces.go`):
 - `SessionAuth` - `Session`, `Login`, `Refresh` (token lifecycle)
 - `Lobby` - `WaitingGames`, `CreateGame`, `JoinGame`
 - `GamePlay` - `GetGame`, `Move`
 - `GameClient` - composition of the three, implemented by `*Client`;
   the cobra commands and the interactive loop depend only on it

Server (`cli/server/internal`):
 - `Authenticator` - `Login`, `Refresh` (used by the login handlers)
 - `SessionValidator` - `ValidateSession` (all `RequireSession` needs)
 - `Tokens` - composition of the two, implemented by `tokenService`
 - `Runner` - `Addr`, `Run`, implemented by `*Server`

The naming mirrors `game/service` (`Lobby`, `GamePlay`, `GameService`)
so the client interface reads as the remote twin of the domain service.

## Testing

Unit tests - run with `just test-unit` or `go test ./...`:

 - tests are table-driven with `t.Run`, plain `testing` package only
 - errors are asserted by code via `errs.HasCode`, never by message
 - HTTP is exercised with `httptest`; collaborators behind the
   interfaces above are replaced by in-package fakes
 - data packages and generated models (`game/data`, `game/data/gen`)
   are intentionally not unit tested

End to end tests - run with `just test` (requires Docker):

 - builds the test image (`_env/test/Dockerfile`) and runs
   `tests/run_tests.sh` inside it: unit tests first, then it builds the
   real server and client binaries, starts the server as a background
   process and runs every script in `tests/scenarios/`
 - scenarios cover auth, lobby, a full pvp game to a win, private games
   with join codes, and invalid moves - see `tests/README.md`
 - each logical player gets its own `HOME`, so the file session store
   keeps the players apart without any client changes

## Docker

Three images, all under `_env/`:

 - `_env/dev/Dockerfile` (`just build`) - pinned toolchain for codegen,
   ships `sqlc`; `just compile` mounts the repo and regenerates
   `game/data/gen`
 - `_env/test/Dockerfile` (`just build-test`) - go toolchain plus
   `bash`, `curl`, `jq` for the e2e scenario runner
 - `_env/prod/Dockerfile` (`just build-prod`) - multi stage production
   build

Production image:
 - stage `ttt-build-image` compiles a static, self contained binary
   (`CGO_ENABLED=0`, no runtime dependencies)
 - stage `ttt-server-image` is `FROM scratch` with only the binary
 - the image binds `0.0.0.0:8080` and exposes port 8080 so other
   containers can reach the server over http on a shared docker network:
   `docker run --network game-net --name ttt-server ttt-server-image`
 - `just build-prod` first runs `just compile-check`, which regenerates
   sqlc code and aborts the build if the generated files changed
 - run it with `just run-prod` (sets `JWT_SECRET` and publishes 8080)

### Config (`pkg/config`)

Env access goes through `config.GetEnv`, backed by a map populated **once** by
`config.LoadEnv()`. Call `LoadEnv()` at startup (and in `TestMain`) before anything
reads env - `GetEnv` returns empty for anything loaded afterward.
