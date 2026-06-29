-- +goose Up
CREATE TABLE IF NOT EXISTS auth_password_reset_tokens (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user ON auth_password_reset_tokens (user_id);

-- +goose Down
DROP TABLE IF EXISTS auth_password_reset_tokens;
