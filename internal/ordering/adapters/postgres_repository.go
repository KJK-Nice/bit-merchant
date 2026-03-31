package adapters

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Save(o *order.Order) error {
	itemsJSON, err := json.Marshal(o.Items)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(
		`INSERT INTO orders (id, order_number, restaurant_id, session_id, items, total_amount, fiat_amount, payment_method, payment_status, fulfillment_status, created_at, updated_at, paid_at, preparing_at, ready_at, completed_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		 ON CONFLICT (id) DO UPDATE SET
		   order_number=EXCLUDED.order_number, items=EXCLUDED.items,
		   total_amount=EXCLUDED.total_amount, fiat_amount=EXCLUDED.fiat_amount,
		   payment_status=EXCLUDED.payment_status, fulfillment_status=EXCLUDED.fulfillment_status,
		   updated_at=EXCLUDED.updated_at, paid_at=EXCLUDED.paid_at,
		   preparing_at=EXCLUDED.preparing_at, ready_at=EXCLUDED.ready_at, completed_at=EXCLUDED.completed_at`,
		string(o.ID), string(o.OrderNumber), string(o.RestaurantID), o.SessionID,
		itemsJSON, o.TotalAmount, o.FiatAmount,
		string(o.PaymentMethod), string(o.PaymentStatus), string(o.FulfillmentStatus),
		o.CreatedAt, o.UpdatedAt, o.PaidAt, o.PreparingAt, o.ReadyAt, o.CompletedAt)
	return err
}

func (r *PostgresOrderRepository) FindByID(id common.OrderID) (*order.Order, error) {
	row := r.db.QueryRow(
		`SELECT id, order_number, restaurant_id, session_id, items, total_amount, fiat_amount, payment_method, payment_status, fulfillment_status, created_at, updated_at, paid_at, preparing_at, ready_at, completed_at
		 FROM orders WHERE id = $1`, string(id))
	return scanOrder(row)
}

func (r *PostgresOrderRepository) FindByOrderNumber(restaurantID common.RestaurantID, orderNumber string) (*order.Order, error) {
	row := r.db.QueryRow(
		`SELECT id, order_number, restaurant_id, session_id, items, total_amount, fiat_amount, payment_method, payment_status, fulfillment_status, created_at, updated_at, paid_at, preparing_at, ready_at, completed_at
		 FROM orders WHERE restaurant_id = $1 AND order_number = $2 ORDER BY created_at DESC LIMIT 1`,
		string(restaurantID), orderNumber)
	return scanOrder(row)
}

func (r *PostgresOrderRepository) FindByRestaurantID(restaurantID common.RestaurantID) ([]*order.Order, error) {
	return r.queryOrders(
		`SELECT id, order_number, restaurant_id, session_id, items, total_amount, fiat_amount, payment_method, payment_status, fulfillment_status, created_at, updated_at, paid_at, preparing_at, ready_at, completed_at
		 FROM orders WHERE restaurant_id = $1`, string(restaurantID))
}

func (r *PostgresOrderRepository) FindActiveByRestaurantID(restaurantID common.RestaurantID) ([]*order.Order, error) {
	return r.queryOrders(
		`SELECT id, order_number, restaurant_id, session_id, items, total_amount, fiat_amount, payment_method, payment_status, fulfillment_status, created_at, updated_at, paid_at, preparing_at, ready_at, completed_at
		 FROM orders WHERE restaurant_id = $1 AND fulfillment_status IN ('paid','preparing','ready')`,
		string(restaurantID))
}

func (r *PostgresOrderRepository) FindBySessionID(sessionID string) ([]*order.Order, error) {
	return r.queryOrders(
		`SELECT id, order_number, restaurant_id, session_id, items, total_amount, fiat_amount, payment_method, payment_status, fulfillment_status, created_at, updated_at, paid_at, preparing_at, ready_at, completed_at
		 FROM orders WHERE session_id = $1`, sessionID)
}

func (r *PostgresOrderRepository) Update(o *order.Order) error {
	itemsJSON, err := json.Marshal(o.Items)
	if err != nil {
		return err
	}
	result, err := r.db.Exec(
		`UPDATE orders SET order_number=$2, items=$3, total_amount=$4, fiat_amount=$5,
		   payment_method=$6, payment_status=$7, fulfillment_status=$8,
		   updated_at=$9, paid_at=$10, preparing_at=$11, ready_at=$12, completed_at=$13
		 WHERE id=$1`,
		string(o.ID), string(o.OrderNumber), itemsJSON, o.TotalAmount, o.FiatAmount,
		string(o.PaymentMethod), string(o.PaymentStatus), string(o.FulfillmentStatus),
		o.UpdatedAt, o.PaidAt, o.PreparingAt, o.ReadyAt, o.CompletedAt)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("order not found")
	}
	return nil
}

func (r *PostgresOrderRepository) queryOrders(query string, args ...interface{}) ([]*order.Order, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*order.Order
	for rows.Next() {
		o, err := scanOrderRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, o)
	}
	return result, rows.Err()
}

func scanOrder(row *sql.Row) (*order.Order, error) {
	var (
		id, orderNum, restID, sessionID            string
		itemsJSON                                  []byte
		totalAmount                                int64
		fiatAmount                                 float64
		payMethod, payStatus, fulStatus            string
		createdAt, updatedAt                       time.Time
		paidAt, preparingAt, readyAt, completedAt sql.NullTime
	)
	if err := row.Scan(&id, &orderNum, &restID, &sessionID, &itemsJSON, &totalAmount, &fiatAmount, &payMethod, &payStatus, &fulStatus, &createdAt, &updatedAt, &paidAt, &preparingAt, &readyAt, &completedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	return buildOrder(id, orderNum, restID, sessionID, itemsJSON, totalAmount, fiatAmount, payMethod, payStatus, fulStatus, createdAt, updatedAt, paidAt, preparingAt, readyAt, completedAt)
}

func scanOrderRows(rows *sql.Rows) (*order.Order, error) {
	var (
		id, orderNum, restID, sessionID            string
		itemsJSON                                  []byte
		totalAmount                                int64
		fiatAmount                                 float64
		payMethod, payStatus, fulStatus            string
		createdAt, updatedAt                       time.Time
		paidAt, preparingAt, readyAt, completedAt sql.NullTime
	)
	if err := rows.Scan(&id, &orderNum, &restID, &sessionID, &itemsJSON, &totalAmount, &fiatAmount, &payMethod, &payStatus, &fulStatus, &createdAt, &updatedAt, &paidAt, &preparingAt, &readyAt, &completedAt); err != nil {
		return nil, err
	}
	return buildOrder(id, orderNum, restID, sessionID, itemsJSON, totalAmount, fiatAmount, payMethod, payStatus, fulStatus, createdAt, updatedAt, paidAt, preparingAt, readyAt, completedAt)
}

func buildOrder(id, orderNum, restID, sessionID string, itemsJSON []byte, totalAmount int64, fiatAmount float64, payMethod, payStatus, fulStatus string, createdAt, updatedAt time.Time, paidAt, preparingAt, readyAt, completedAt sql.NullTime) (*order.Order, error) {
	var items []order.OrderItem
	if len(itemsJSON) > 0 {
		_ = json.Unmarshal(itemsJSON, &items)
	}
	o := &order.Order{
		ID: common.OrderID(id), OrderNumber: common.OrderNumber(orderNum),
		RestaurantID: common.RestaurantID(restID), SessionID: sessionID,
		Items: items, TotalAmount: totalAmount, FiatAmount: fiatAmount,
		PaymentMethod: common.PaymentMethodType(payMethod),
		PaymentStatus: common.PaymentStatus(payStatus),
		FulfillmentStatus: common.FulfillmentStatus(fulStatus),
		CreatedAt: createdAt, UpdatedAt: updatedAt,
	}
	if paidAt.Valid {
		t := paidAt.Time
		o.PaidAt = &t
	}
	if preparingAt.Valid {
		t := preparingAt.Time
		o.PreparingAt = &t
	}
	if readyAt.Valid {
		t := readyAt.Time
		o.ReadyAt = &t
	}
	if completedAt.Valid {
		t := completedAt.Time
		o.CompletedAt = &t
	}
	return o, nil
}
