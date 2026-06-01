-- +goose Up
CREATE TABLE moves (
    id UUID PRIMARY KEY,
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    player_id UUID NOT NULL,
    move_number INT NOT NULL,
    from_row INT NOT NULL,
    from_col INT NOT NULL,
    to_row INT NOT NULL,
    to_col INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT uniq_game_move UNIQUE (game_id, move_number)
);

-- +goose Down
DROP TABLE moves;