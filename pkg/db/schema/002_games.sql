CREATE TABLE games (
    id TEXT PRIMARY KEY,
    player_x TEXT NOT NULL REFERENCES users (id),
    player_o TEXT NOT NULL REFERENCES users (id),
    grid TEXT NOT NULL,
    status TEXT NOT NULL,
    winner_id TEXT NOT NULL DEFAULT '',
    move_count INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_games_player_x ON games (player_x);

CREATE INDEX idx_games_player_o ON games (player_o);
