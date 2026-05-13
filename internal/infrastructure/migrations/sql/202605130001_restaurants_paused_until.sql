-- +goose Up
ALTER TABLE restaurants
    ADD COLUMN IF NOT EXISTS paused_until TIMESTAMPTZ NULL;

-- +goose Down
ALTER TABLE restaurants
    DROP COLUMN IF EXISTS paused_until;
