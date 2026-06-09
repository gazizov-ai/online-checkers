-- +goose Up

CREATE INDEX IF NOT EXISTS idx_ratings_leaderboard
ON ratings (
    rating DESC,
    wins DESC,
    games_played ASC,
    updated_at DESC
)
INCLUDE (user_id, losses);

-- +goose Down

DROP INDEX IF EXISTS idx_ratings_leaderboard;