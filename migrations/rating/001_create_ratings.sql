-- +goose Up
CREATE TABLE ratings (
    user_id UUID PRIMARY KEY,
    rating INT NOT NULL DEFAULT 1000 CHECK (rating >= 0),
    games_played INT NOT NULL DEFAULT 0 CHECK (games_played >= 0),
    wins INT NOT NULL DEFAULT 0 CHECK (wins >= 0),
    losses INT NOT NULL DEFAULT 0 CHECK (losses >= 0),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE ratings;