-- +goose Up
CREATE TABLE IF NOT EXISTS session_restaurant_visits (
    session_id TEXT NOT NULL,
    restaurant_id TEXT NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    first_visited_at TIMESTAMPTZ NOT NULL,
    last_visited_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (session_id, restaurant_id)
);

CREATE INDEX IF NOT EXISTS idx_session_restaurant_visits_session
    ON session_restaurant_visits (session_id);

-- +goose Down
DROP INDEX IF EXISTS idx_session_restaurant_visits_session;
DROP TABLE IF EXISTS session_restaurant_visits;
