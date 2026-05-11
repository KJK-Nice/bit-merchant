-- +goose Up
ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS modifiers JSONB NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS special_instructions TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE order_items
    DROP COLUMN IF EXISTS modifiers,
    DROP COLUMN IF EXISTS special_instructions;
