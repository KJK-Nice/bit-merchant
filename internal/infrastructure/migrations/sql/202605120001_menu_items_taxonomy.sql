-- +goose Up
ALTER TABLE menu_items
    ADD COLUMN IF NOT EXISTS spice_level                TEXT,
    ADD COLUMN IF NOT EXISTS sku                        TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS schedule                   TEXT NOT NULL DEFAULT 'ALL_DAY',
    ADD COLUMN IF NOT EXISTS is_vegan                   BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS is_dairy_free              BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS is_halal                   BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS is_nut_free                BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS allergens                  JSONB NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS badges                     JSONB NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS allow_special_instructions BOOLEAN NOT NULL DEFAULT TRUE;

-- Backfill spice_level from the deprecated is_spicy boolean.
UPDATE menu_items
   SET spice_level = 'MEDIUM'
 WHERE spice_level IS NULL AND is_spicy = TRUE;

-- +goose Down
ALTER TABLE menu_items
    DROP COLUMN IF EXISTS spice_level,
    DROP COLUMN IF EXISTS sku,
    DROP COLUMN IF EXISTS schedule,
    DROP COLUMN IF EXISTS is_vegan,
    DROP COLUMN IF EXISTS is_dairy_free,
    DROP COLUMN IF EXISTS is_halal,
    DROP COLUMN IF EXISTS is_nut_free,
    DROP COLUMN IF EXISTS allergens,
    DROP COLUMN IF EXISTS badges,
    DROP COLUMN IF EXISTS allow_special_instructions;
