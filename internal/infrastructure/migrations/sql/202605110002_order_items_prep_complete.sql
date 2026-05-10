-- +goose Up
ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS prep_complete BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE order_items
    DROP COLUMN IF EXISTS prep_complete;
