# EPAM Tic-Tac-Toe Assessment

This file provides guidance to users and agents when using the project.
The game focuses on a single game server and multiple clients connection to it
 - distribution or multiple game servers are left for future work - but taken in account while designing the architecture.

I have focused on robust usage of tokens, validation and sensible security levels
 - token are used to joining lobbies, 
 - player actions - moves, game creation, game state retrieval
 - game - validate access to current game state and moves
Admin role is for token role example: 
 - this is why only admins see history of games 
 - and create lobbies

## Project

A Tic-Tac-Toe game server (EPAM assessment task). Go module `epam/task`, Go 1.25.
The implementation is intentionally **standard-library only** — the sole external
tool is `sqlc`, used at build time for code generation, not as a runtime dependency.

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

// token usage - game token specifics

### Config (`pkg/config`)

Env access goes through `config.GetEnv`, backed by a map populated **once** by
`config.LoadEnv()`. Call `LoadEnv()` at startup (and in `TestMain`) before anything
reads env — `GetEnv` returns empty for anything loaded afterward.

## Game Logic