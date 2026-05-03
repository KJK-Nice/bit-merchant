-- +goose Up
CREATE TABLE push_subscriptions (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    role          TEXT        NOT NULL CHECK (role IN ('customer', 'kitchen')),
    order_number  TEXT,
    restaurant_id TEXT        REFERENCES restaurants(id) ON DELETE CASCADE,
    endpoint      TEXT        NOT NULL,
    auth_key      TEXT        NOT NULL,
    p256dh_key    TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX push_subscriptions_endpoint_idx ON push_subscriptions(endpoint);

-- +goose Down
DROP TABLE IF EXISTS push_subscriptions;
