-- +goose Up
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS server_called_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS bill_requested_at TIMESTAMPTZ NULL;

-- +goose Down
ALTER TABLE orders
    DROP COLUMN IF EXISTS bill_requested_at,
    DROP COLUMN IF EXISTS server_called_at;
