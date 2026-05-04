-- +goose Up
-- The very first version of 202605030001_push_subscriptions.sql created
-- push_subscriptions_endpoint_idx as a UNIQUE index on (endpoint) — the file
-- was later edited in place to be a plain CREATE INDEX, but Goose tracks
-- migrations by version number, so test/prod databases that ran the original
-- still carry the unique constraint. That single-column unique index blocks
-- the per-(endpoint, role, order_number) model: a returning customer placing
-- a second order can't insert a new row, and the violation isn't caught by
-- ON CONFLICT (endpoint, role, order_number, restaurant_id) since that
-- target matches push_subscriptions_unique_idx (added in 202605040001),
-- not push_subscriptions_endpoint_idx.
--
-- Replace it with the non-unique lookup index it was always meant to be —
-- needed for DeleteByEndpoint's per-endpoint cleanup on 410 Gone.
DROP INDEX IF EXISTS push_subscriptions_endpoint_idx;

CREATE INDEX push_subscriptions_endpoint_idx ON push_subscriptions (endpoint);

-- +goose Down
DROP INDEX IF EXISTS push_subscriptions_endpoint_idx;

CREATE UNIQUE INDEX push_subscriptions_endpoint_idx ON push_subscriptions (endpoint);
