-- +goose Up
-- Decouple push subscriptions from individual orders. The previous schema
-- keyed each row on (endpoint, role, order_number, restaurant_id), so a
-- customer who enabled notifications on order A had to re-subscribe on
-- order B's status page to be notified — defeating the "enable once per
-- device" UX expectation.
--
-- The new model:
--   * push_subscriptions: one row per (endpoint, role); holds the
--     encryption material that identifies the device.
--   * push_subscription_scopes: many-to-many between subscriptions and the
--     things they want pings for (orders, restaurants).
--
-- Existing test subscriptions are dropped — affected users re-subscribe on
-- their next page visit (the per-page subscribe POST is idempotent).
DROP TABLE IF EXISTS push_subscriptions CASCADE;

CREATE TABLE push_subscriptions (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    role       TEXT        NOT NULL CHECK (role IN ('customer', 'kitchen')),
    endpoint   TEXT        NOT NULL,
    auth_key   TEXT        NOT NULL,
    p256dh_key TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (endpoint, role)
);

CREATE TABLE push_subscription_scopes (
    subscription_id UUID        NOT NULL REFERENCES push_subscriptions(id) ON DELETE CASCADE,
    scope_type      TEXT        NOT NULL CHECK (scope_type IN ('order', 'restaurant')),
    scope_id        TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (subscription_id, scope_type, scope_id)
);

-- Reverse lookup index: fan-out from a scope (e.g. "order 0215") to every
-- device that wants pings for it. Used by the notifier on every event.
CREATE INDEX push_subscription_scopes_lookup_idx
    ON push_subscription_scopes (scope_type, scope_id);

-- +goose Down
DROP TABLE IF EXISTS push_subscription_scopes;
DROP TABLE IF EXISTS push_subscriptions;

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
CREATE UNIQUE INDEX push_subscriptions_unique_idx
    ON push_subscriptions (endpoint, role, order_number, restaurant_id) NULLS NOT DISTINCT;
CREATE INDEX push_subscriptions_endpoint_idx ON push_subscriptions (endpoint);
