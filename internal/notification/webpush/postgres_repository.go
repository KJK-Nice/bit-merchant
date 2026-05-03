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

func (r *PostgresRepository) Upsert(sub *Subscription) error {
	var restaurantID *string
	if sub.RestaurantID != "" {
		s := string(sub.RestaurantID)
		restaurantID = &s
	}
	_, err := r.db.Exec(`
		INSERT INTO push_subscriptions (role, order_number, restaurant_id, endpoint, auth_key, p256dh_key)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (endpoint) DO UPDATE SET
		    role          = EXCLUDED.role,
		    order_number  = EXCLUDED.order_number,
		    restaurant_id = EXCLUDED.restaurant_id,
		    auth_key      = EXCLUDED.auth_key,
		    p256dh_key    = EXCLUDED.p256dh_key`,
		sub.Role,
		nullString(sub.OrderNumber),
		restaurantID,
		sub.Endpoint,
		sub.AuthKey,
		sub.P256DHKey,
	)
	return err
}

func (r *PostgresRepository) FindByOrderNumber(orderNumber string) ([]*Subscription, error) {
	rows, err := r.db.Query(`
		SELECT id, role, COALESCE(order_number,''), COALESCE(restaurant_id::text,''), endpoint, auth_key, p256dh_key
		FROM push_subscriptions WHERE role = 'customer' AND order_number = $1`, orderNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSubscriptions(rows)
}

func (r *PostgresRepository) FindByRestaurantID(restaurantID common.RestaurantID) ([]*Subscription, error) {
	rows, err := r.db.Query(`
		SELECT id, role, COALESCE(order_number,''), COALESCE(restaurant_id::text,''), endpoint, auth_key, p256dh_key
		FROM push_subscriptions WHERE role = 'kitchen' AND restaurant_id = $1`, string(restaurantID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSubscriptions(rows)
}

func (r *PostgresRepository) DeleteByEndpoint(endpoint string) error {
	_, err := r.db.Exec(`DELETE FROM push_subscriptions WHERE endpoint = $1`, endpoint)
	return err
}

func scanSubscriptions(rows *sql.Rows) ([]*Subscription, error) {
	var out []*Subscription
	for rows.Next() {
		var s Subscription
		var restaurantID string
		if err := rows.Scan(&s.ID, &s.Role, &s.OrderNumber, &restaurantID, &s.Endpoint, &s.AuthKey, &s.P256DHKey); err != nil {
			return nil, err
		}
		s.RestaurantID = common.RestaurantID(restaurantID)
		out = append(out, &s)
	}
	return out, rows.Err()
}

func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
