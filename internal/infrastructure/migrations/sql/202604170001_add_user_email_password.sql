-- +goose Up
ALTER TABLE auth_users ADD COLUMN email TEXT;
ALTER TABLE auth_users ADD COLUMN password_hash TEXT;
CREATE UNIQUE INDEX idx_auth_users_email ON auth_users (LOWER(email)) WHERE email IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_auth_users_email;
ALTER TABLE auth_users DROP COLUMN IF EXISTS password_hash;
ALTER TABLE auth_users DROP COLUMN IF EXISTS email;
