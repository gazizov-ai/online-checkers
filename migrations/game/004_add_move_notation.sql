-- +goose Up
ALTER TABLE moves
ADD COLUMN turn_number INT NOT NULL,
ADD COLUMN sequence_number INT NOT NULL,
ADD COLUMN is_capture BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN notation TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE moves
DROP COLUMN notation,
DROP COLUMN is_capture,
DROP COLUMN sequence_number,
DROP COLUMN turn_number;