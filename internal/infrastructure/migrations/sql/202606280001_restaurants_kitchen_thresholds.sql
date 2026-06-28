-- +goose Up
ALTER TABLE restaurants
    ADD COLUMN IF NOT EXISTS kitchen_warning_minutes INT NOT NULL DEFAULT 8,
    ADD COLUMN IF NOT EXISTS kitchen_overdue_minutes INT NOT NULL DEFAULT 12;

-- +goose Down
ALTER TABLE restaurants
    DROP COLUMN IF EXISTS kitchen_overdue_minutes,
    DROP COLUMN IF EXISTS kitchen_warning_minutes;
