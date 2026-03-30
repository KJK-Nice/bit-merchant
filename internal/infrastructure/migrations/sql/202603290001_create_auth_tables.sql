-- +goose Up
CREATE TABLE IF NOT EXISTS auth_users (
    id TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    credentials_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS auth_memberships (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    restaurant_id TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS auth_invitations (
    id TEXT PRIMARY KEY,
    restaurant_id TEXT NOT NULL,
    role TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL,
    used_by_user_id TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS auth_sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NULL,
    restaurant_id TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_auth_memberships_user ON auth_memberships (user_id);
CREATE INDEX IF NOT EXISTS idx_auth_memberships_restaurant ON auth_memberships (restaurant_id);
CREATE INDEX IF NOT EXISTS idx_auth_invitations_restaurant ON auth_invitations (restaurant_id);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_user ON auth_sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires ON auth_sessions (expires_at);

-- +goose Down
DROP TABLE IF EXISTS auth_sessions;
DROP TABLE IF EXISTS auth_invitations;
DROP TABLE IF EXISTS auth_memberships;
DROP TABLE IF EXISTS auth_users;
