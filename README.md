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

### CLI (`cmd/`)

The service is started through a cobra CLI configured with viper
 - `go run . serve` - starts the HTTP server (implementation pending)
 - flags: `--host` (default `127.0.0.1`), `--port` (default `8080`)
 - env overrides via viper: `GAME_SERVER_HOST`, `GAME_SERVER_PORT`

### Config (`pkg/config`)

Env access goes through `config.GetEnv`, backed by a map populated **once** by
`config.LoadEnv()`. Call `LoadEnv()` at startup (and in `TestMain`) before anything
reads env — `GetEnv` returns empty for anything loaded afterward.

## Game Logic