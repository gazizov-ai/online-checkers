-- +goose Up

CREATE INDEX IF NOT EXISTS idx_matchmaking_queue_waiting_created_at
ON matchmaking_queue (created_at, user_id)
WHERE status = 'waiting';

CREATE INDEX IF NOT EXISTS idx_matchmaking_queue_matching_updated_at
ON matchmaking_queue (updated_at, user_id)
WHERE status = 'matching';

-- +goose Down

DROP INDEX IF EXISTS idx_matchmaking_queue_matching_updated_at;
DROP INDEX IF EXISTS idx_matchmaking_queue_waiting_created_at;