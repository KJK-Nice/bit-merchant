-- +goose Up
-- Per-restaurant order-number counter. Replaces the previous app-side
-- rand.Intn(10000) generator, which collided under concurrent load (birthday
-- paradox at ~118 orders) and capped namespaces at 10k orders per restaurant.
--
-- Atomic increment via INSERT … ON CONFLICT DO UPDATE … RETURNING. The UPDATE
-- takes the row lock implicitly, so the read-modify-write is race-free without
-- app-side retries.
CREATE TABLE restaurant_order_counters (
    restaurant_id TEXT     PRIMARY KEY REFERENCES restaurants(id) ON DELETE CASCADE,
    last_number   INTEGER  NOT NULL DEFAULT 0
);

-- Backfill from existing orders so live restaurants don't restart at 1 and
-- immediately collide with historical numbers. The regex extracts the numeric
-- portion of any existing order_number; NULLIF/COALESCE handle restaurants
-- whose orders contain only non-numeric characters or have no orders at all.
INSERT INTO restaurant_order_counters (restaurant_id, last_number)
SELECT
    restaurant_id,
    COALESCE(MAX(NULLIF(regexp_replace(order_number, '\D', '', 'g'), '')::int), 0)
FROM orders
GROUP BY restaurant_id
ON CONFLICT (restaurant_id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS restaurant_order_counters;
