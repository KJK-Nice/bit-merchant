-- +goose Up
-- Replace the expression-based unique index with a NULLS NOT DISTINCT one so
-- that ON CONFLICT inference can match against plain columns. The old index
-- used COALESCE(...) expressions, but Postgres expression-index inference is
-- finicky about canonical form (literal types, opclass) and the inference
-- failed at runtime with SQLSTATE 42P10. Postgres 15+ supports NULLS NOT
-- DISTINCT, which makes NULL values compare equal in unique indexes — exactly
-- what we want here.
DROP INDEX IF EXISTS push_subscriptions_unique_idx;

CREATE UNIQUE INDEX push_subscriptions_unique_idx
    ON push_subscriptions (endpoint, role, order_number, restaurant_id)
    NULLS NOT DISTINCT;

-- +goose Down
DROP INDEX IF EXISTS push_subscriptions_unique_idx;

CREATE UNIQUE INDEX push_subscriptions_unique_idx
    ON push_subscriptions (endpoint, role, COALESCE(order_number, ''), COALESCE(restaurant_id, ''));
