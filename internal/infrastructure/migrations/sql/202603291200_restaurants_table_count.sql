-- +goose Up
ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS table_count INTEGER NOT NULL DEFAULT 1;

-- +goose Down
ALTER TABLE restaurants DROP COLUMN IF EXISTS table_count;
