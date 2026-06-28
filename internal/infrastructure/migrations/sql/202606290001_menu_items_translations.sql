-- +goose Up
ALTER TABLE menu_items
    ADD COLUMN IF NOT EXISTS translations JSONB NOT NULL DEFAULT '{}';

-- +goose Down
ALTER TABLE menu_items
    DROP COLUMN IF EXISTS translations;
