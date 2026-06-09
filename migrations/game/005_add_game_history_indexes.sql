-- +goose Up

CREATE INDEX IF NOT EXISTS idx_games_white_player_history
ON games (
    white_player_id,
    (CASE WHEN status = 'active' THEN 0 ELSE 1 END),
    (CASE WHEN status = 'active' THEN created_at ELSE finished_at END) DESC,
    id DESC
);

CREATE INDEX IF NOT EXISTS idx_games_black_player_history
ON games (
    black_player_id,
    (CASE WHEN status = 'active' THEN 0 ELSE 1 END),
    (CASE WHEN status = 'active' THEN created_at ELSE finished_at END) DESC,
    id DESC
);

-- +goose Down

DROP INDEX IF EXISTS idx_games_black_player_history;
DROP INDEX IF EXISTS idx_games_white_player_history;