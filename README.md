# Tic-Tac-Toe Assessment

This file provides guidance to users and agents when using the project.
The game focuses on a single game server and multiple clients connection to it
 - distribution or multiple game servers are left for future work - but taken in account while designing the architecture.

AI is used in all steps of this project, my stamp on every step

I have focused on robust usage of tokens, validation and sensible security levels
 - token are used to joining lobbies, 
 - player actions - moves, game creation, game state retrieval
 - game - validate access to current game state and moves
Admin role is for token role example: 
 - this is why only admins see history of games 
 - and create lobbies

## Project

A Tic-Tac-Toe game server = Go module `ticTacSolved/task`, Go 1.25.
 - this is deliberately named `ticTacSolved` to avoid naming companies also emphasizes the spirit of the project

The game domain is intentionally **standard-library only**;
the CLI layer uses `cobra` and `viper`
tools used: `sqlc`

## Commands

Common tasks are driven by `just` (see `justfile`); 
 - Docker is used primarily to provide a pinned toolchain for codegen.

- `just build` - build image user by `just compile` (sqlc)
- `just compile` - generates sqlc files 
- `just test` - uses on machine go to run tests
- `just run` - runs the `ttt` server locally (`go run ./cli/server/main serve`)
- `just compile-check` - runs `just compile` and aborts if generated code changed
- `just build-prod` - runs `compile-check`, then builds the `ttt-build-image`
  compile image and the `ttt-server-exam-image` scratch runtime image
- `just run-prod` - runs the scratch image with port 8080 published

## Architecture

Two top-level trees:

- `game/` — the game domain: `state_machine` (rules), `auth` (tokens), `data` (persistence, sqlc-generated).
- `pkg/` — cross-cutting infrastructure reused by the domain: `errs`, `config`, `util`.

### Auth (`game/auth`)

Tokens are HS256 JWTs signed with `JWT_SECRET` and carry their own expiration:
 - `Issue` stamps an `exp` claim (unix seconds) into the token, so `Validate`
   checks expiration from the token itself, no store roundtrip needed
 - a `player_id` claim is required on every issued token

Each player has exactly one active token at a time:
 - issuing a new token replaces the previous one for that player
 - active tokens live in a package-level in-memory cache (`tokenCache`),
   keyed by player id; a ticker inside the cache struct evicts entries
   after a set time (`TokenCacheTTL`)
 - on a cache miss for a still-unexpired token, the store is checked;
   a found token is valid and is cached again
 - the `tokens` table keys on `player_id` (unique) and `SaveToken` upserts,
   so the single-active-token rule also holds at the DB level

### CLI (`cli/`)

The server binary is the `ttt` command, main package at `cli/server/main`
 - `go run ./cli/server/main serve` - starts the HTTP server (or `just run`)
 - flags: `--host` (default `127.0.0.1`), `--port` (default `8080`)
 - env overrides via viper: `GAME_SERVER_HOST`, `GAME_SERVER_PORT`
 - `JWT_SECRET` must be set for token issuing (`just run` defaults it for dev)

The client binary is the `ttt-client` command, main package at `cli/client/main`
 - `go run ./cli/client/main action list` - one shot actions against the server
 - `go run ./cli/client/main --type cli` - interactive play mode
 - flags: `--server --user --password --type (cli|file) --token
   --token-ttl --session-ttl`; every flag is also an env var
   (`TTT_` prefix) and can live in a `.env` file
 - session and refresh tokens are stored via `pkg/session`
   (file store by default, `~/.ttt/session.json`)

### Shared interfaces (`cli/*/internal`)

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

### HTTP API (`pkg/api`)

Paths, header names and request/response DTOs live in `pkg/api` and are
shared by the server handlers and the client.

 - `POST /api/login` - no auth, returns player id, session and refresh tokens
 - `POST /api/refresh` - no auth, refresh token in the body, returns a session
 - `GET /api/games` - bearer session, lists games waiting for players
 - `POST /api/games` - bearer session, creates a game, returns a game token
 - `GET /api/games/{id}` - bearer session, current game state
 - `POST /api/games/{id}/join` - bearer session, join code for private games
 - `POST /api/games/{id}/move` - bearer session plus `X-Game-Token` header

Errors come back as `{"code": "...", "message": "..."}`; the status is
derived from the `pkg/errs` code (`INVALID_INPUT` -> 400,
`INVALID_TOKEN` -> 401, `*_NOT_FOUND` -> 404, `CELL_OCCUPIED`,
`INVALID_TRANSITION`, `GAME_FINISHED`, `OUT_OF_TURN` -> 409,
anything else -> 500).

## Testing

Run everything with `just test` or `go test ./...`.

 - tests are table-driven with `t.Run`, plain `testing` package only
 - errors are asserted by code via `errs.HasCode`, never by message
 - HTTP is exercised with `httptest`; collaborators behind the
   interfaces above are replaced by in-package fakes
 - data packages and generated models (`game/data`, `game/data/gen`)
   are intentionally not unit tested
 - the full plan and the coverage map live in `test.ai.md`

### Production image (`_env/prod`)

Multi stage build, see `_env/prod/Dockerfile`
 - stage `ttt-build-image` compiles a static, self contained binary
   (`CGO_ENABLED=0`, no runtime dependencies)
 - stage `ttt-server-exam-image` is `FROM scratch` with only the binary
 - the image binds `0.0.0.0:8080` and exposes port 8080 so other
   containers can reach the server over http on a shared docker network:
   `docker run --network game-net --name ttt-server ttt-server-exam-image`
 - `just build-prod` first runs `just compile-check`, which regenerates
   sqlc code and aborts the build if the generated files changed

### Config (`pkg/config`)

Env access goes through `config.GetEnv`, backed by a map populated **once** by
`config.LoadEnv()`. Call `LoadEnv()` at startup (and in `TestMain`) before anything
reads env — `GetEnv` returns empty for anything loaded afterward.

## Game Logic