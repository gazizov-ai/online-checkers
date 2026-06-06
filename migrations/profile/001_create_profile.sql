-- +goose Up
CREATE TABLE profiles (
    user_id UUID PRIMARY KEY,
    username TEXT NOT NULL,
    display_name TEXT,
    country_code TEXT CHECK (
        country_code IS NULL
        OR country_code ~ '^[A-Z]{2}$'
    ),
    avatar_url TEXT,
    bio TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE profile_processed_events (
    event_id UUID PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS profile_processed_events;
DROP TABLE IF EXISTS profiles;