-- +goose Up
CREATE TABLE matchmaking_queue (
    user_id UUID PRIMARY KEY,
    status TEXT NOT NULL,
    game_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE matchmaking_queue;