CREATE TABLE players (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE games (
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL,
    is_public BOOLEAN NOT NULL,
    board TEXT NOT NULL,
    status TEXT NOT NULL,
    player_x TEXT NOT NULL,
    player_o TEXT NOT NULL
);

CREATE TABLE stats (
    player_id TEXT PRIMARY KEY,
    wins INTEGER NOT NULL DEFAULT 0,
    losses INTEGER NOT NULL DEFAULT 0,
    draws INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE tokens (
    token TEXT PRIMARY KEY,
    player_id TEXT NOT NULL,
    expires_at INTEGER NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_tokens_player_active ON tokens (player_id, active);
