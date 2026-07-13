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

CREATE TABLE tokens (
    token TEXT PRIMARY KEY,
    expires_at INTEGER NOT NULL
);
