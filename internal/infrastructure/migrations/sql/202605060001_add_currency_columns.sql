-- +goose Up
ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS base_currency TEXT NOT NULL DEFAULT 'USD';

ALTER TABLE menu_items ADD COLUMN IF NOT EXISTS currency TEXT NOT NULL DEFAULT 'USD';
ALTER TABLE menu_items ADD COLUMN IF NOT EXISTS price_minor BIGINT NOT NULL DEFAULT 0;

ALTER TABLE orders ADD COLUMN IF NOT EXISTS currency TEXT NOT NULL DEFAULT 'USD';

ALTER TABLE order_items ADD COLUMN IF NOT EXISTS currency TEXT NOT NULL DEFAULT 'USD';
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS unit_price_minor BIGINT NOT NULL DEFAULT 0;
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS subtotal_minor BIGINT NOT NULL DEFAULT 0;

ALTER TABLE payments ADD COLUMN IF NOT EXISTS currency TEXT NOT NULL DEFAULT 'USD';
ALTER TABLE payments ADD COLUMN IF NOT EXISTS amount_minor BIGINT NOT NULL DEFAULT 0;

UPDATE menu_items
SET price_minor = ROUND(price * 100)::BIGINT
WHERE price_minor = 0 AND price > 0;

UPDATE order_items
SET unit_price_minor = ROUND(unit_price * 100)::BIGINT,
    subtotal_minor   = ROUND(subtotal   * 100)::BIGINT
WHERE unit_price_minor = 0 AND unit_price > 0;

UPDATE payments
SET amount_minor = ROUND(amount * 100)::BIGINT
WHERE amount_minor = 0 AND amount > 0;

CREATE INDEX IF NOT EXISTS idx_restaurants_base_currency ON restaurants(base_currency);

-- +goose Down
DROP INDEX IF EXISTS idx_restaurants_base_currency;

ALTER TABLE payments DROP COLUMN IF EXISTS amount_minor;
ALTER TABLE payments DROP COLUMN IF EXISTS currency;

ALTER TABLE order_items DROP COLUMN IF EXISTS subtotal_minor;
ALTER TABLE order_items DROP COLUMN IF EXISTS unit_price_minor;
ALTER TABLE order_items DROP COLUMN IF EXISTS currency;

ALTER TABLE orders DROP COLUMN IF EXISTS currency;

ALTER TABLE menu_items DROP COLUMN IF EXISTS price_minor;
ALTER TABLE menu_items DROP COLUMN IF EXISTS currency;

ALTER TABLE restaurants DROP COLUMN IF EXISTS base_currency;
