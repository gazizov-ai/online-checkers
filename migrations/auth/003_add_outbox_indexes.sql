-- +goose Up

CREATE INDEX IF NOT EXISTS idx_auth_outbox_events_pending
ON auth_outbox_events (created_at, id)
WHERE status = 'pending';

-- +goose Down

DROP INDEX IF EXISTS idx_auth_outbox_events_pending;