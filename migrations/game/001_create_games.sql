-- +goose Up 
CREATE TABLE games (
    id UUID PRIMARY KEY,
    white_player_id UUID NOT NULL,
    black_player_id UUID NOT NULL,
    status TEXT NOT NULL,
    winner_id UUID,
    board_state JSONB NOT NULL,
    current_turn TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    finished_at TIMESTAMP
);

-- +goose Down
DROP TABLE games;