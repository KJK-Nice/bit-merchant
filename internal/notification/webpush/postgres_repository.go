package webpush

import (
	"database/sql"

	"bitmerchant/internal/common"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Upsert refreshes the encryption material for (endpoint, role) and returns
// the subscription's id via sub.ID. The UNIQUE (endpoint, role) constraint
// in the migration makes ON CONFLICT inference reliable without expression
// parens or NULLS NOT DISTINCT — see the schema rationale in
// 202605040004_push_subscriptions_with_scopes.sql.
func (r *PostgresRepository) Upsert(sub *Subscription) error {
	return r.db.QueryRow(`
		INSERT INTO push_subscriptions (role, endpoint, auth_key, p256dh_key)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (endpoint, role) DO UPDATE SET
		    auth_key   = EXCLUDED.auth_key,
		    p256dh_key = EXCLUDED.p256dh_key
		RETURNING id`,
		sub.Role, sub.Endpoint, sub.AuthKey, sub.P256DHKey,
	).Scan(&sub.ID)
}

// AddScope is idempotent: a second call with the same scope is a silent no-op.
// We don't need ON CONFLICT … DO UPDATE because the row carries no mutable
// payload — the scope's identity is the row.
func (r *PostgresRepository) AddScope(subscriptionID string, scopeType ScopeType, scopeID string) error {
	_, err := r.db.Exec(`
		INSERT INTO push_subscription_scopes (subscription_id, scope_type, scope_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (subscription_id, scope_type, scope_id) DO NOTHING`,
		subscriptionID, string(scopeType), scopeID,
	)
	return err
}

func (r *PostgresRepository) FindByOrderNumber(orderNumber string) ([]*Subscription, error) {
	rows, err := r.db.Query(`
		SELECT s.id, s.role, s.endpoint, s.auth_key, s.p256dh_key
		FROM push_subscriptions s
		JOIN push_subscription_scopes sc ON sc.subscription_id = s.id
		WHERE s.role = 'customer'
		  AND sc.scope_type = 'order'
		  AND sc.scope_id = $1`, orderNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSubscriptions(rows)
}

func (r *PostgresRepository) FindByRestaurantID(restaurantID common.RestaurantID) ([]*Subscription, error) {
	rows, err := r.db.Query(`
		SELECT s.id, s.role, s.endpoint, s.auth_key, s.p256dh_key
		FROM push_subscriptions s
		JOIN push_subscription_scopes sc ON sc.subscription_id = s.id
		WHERE s.role = 'kitchen'
		  AND sc.scope_type = 'restaurant'
		  AND sc.scope_id = $1`, string(restaurantID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSubscriptions(rows)
}

// DeleteByEndpoint deletes every subscription with this endpoint (typically
// just one per role), and CASCADE drops their scope rows.
func (r *PostgresRepository) DeleteByEndpoint(endpoint string) error {
	_, err := r.db.Exec(`DELETE FROM push_subscriptions WHERE endpoint = $1`, endpoint)
	return err
}

func scanSubscriptions(rows *sql.Rows) ([]*Subscription, error) {
	var out []*Subscription
	for rows.Next() {
		var s Subscription
		if err := rows.Scan(&s.ID, &s.Role, &s.Endpoint, &s.AuthKey, &s.P256DHKey); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	return out, rows.Err()
}
