-- +goose Up
ALTER TABLE menu_items
    ADD COLUMN IF NOT EXISTS option_groups JSONB NOT NULL DEFAULT '[]';

-- +goose Down
ALTER TABLE menu_items
    DROP COLUMN IF EXISTS option_groups;
