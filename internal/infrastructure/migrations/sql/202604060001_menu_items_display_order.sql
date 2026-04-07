-- +goose Up
ALTER TABLE menu_items ADD COLUMN IF NOT EXISTS display_order INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE menu_items DROP COLUMN IF EXISTS display_order;
