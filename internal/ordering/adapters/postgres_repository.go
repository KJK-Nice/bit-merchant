package adapters

import (
	"database/sql"
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

const orderColumns = `id, order_number, restaurant_id, session_id, total_amount, fiat_amount,
	payment_method, payment_status, fulfillment_status,
	created_at, updated_at, paid_at, preparing_at, ready_at, completed_at`

func (r *PostgresOrderRepository) Save(o *order.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Exec(
		`INSERT INTO orders (id, order_number, restaurant_id, session_id, total_amount, fiat_amount,
			payment_method, payment_status, fulfillment_status,
			created_at, updated_at, paid_at, preparing_at, ready_at, completed_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		 ON CONFLICT (id) DO UPDATE SET
		   order_number=EXCLUDED.order_number,
		   total_amount=EXCLUDED.total_amount, fiat_amount=EXCLUDED.fiat_amount,
		   payment_status=EXCLUDED.payment_status, fulfillment_status=EXCLUDED.fulfillment_status,
		   updated_at=EXCLUDED.updated_at, paid_at=EXCLUDED.paid_at,
		   preparing_at=EXCLUDED.preparing_at, ready_at=EXCLUDED.ready_at, completed_at=EXCLUDED.completed_at`,
		string(o.ID), string(o.OrderNumber), string(o.RestaurantID), o.SessionID,
		o.TotalAmount, o.FiatAmount,
		string(o.PaymentMethod), string(o.PaymentStatus), string(o.FulfillmentStatus),
		o.CreatedAt, o.UpdatedAt, o.PaidAt, o.PreparingAt, o.ReadyAt, o.CompletedAt)
	if err != nil {
		return err
	}

	for _, item := range o.Items {
		_, err = tx.Exec(
			`INSERT INTO order_items (id, order_id, menu_item_id, name, quantity, unit_price, subtotal)
			 VALUES ($1,$2,$3,$4,$5,$6,$7)
			 ON CONFLICT (id) DO NOTHING`,
			string(item.ID), string(item.OrderID), string(item.MenuItemID),
			item.Name, item.Quantity, item.UnitPrice, item.Subtotal)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresOrderRepository) FindByID(id common.OrderID) (*order.Order, error) {
	row := r.db.QueryRow(
		`SELECT `+orderColumns+` FROM orders WHERE id = $1`, string(id))
	o, err := scanOrderRow(row)
	if err != nil {
		return nil, err
	}
	items, err := r.loadItems(string(o.ID))
	if err != nil {
		return nil, err
	}
	o.Items = items
	return o, nil
}

func (r *PostgresOrderRepository) FindByOrderNumber(restaurantID common.RestaurantID, orderNumber string) (*order.Order, error) {
	row := r.db.QueryRow(
		`SELECT `+orderColumns+` FROM orders WHERE restaurant_id = $1 AND order_number = $2 ORDER BY created_at DESC LIMIT 1`,
		string(restaurantID), orderNumber)
	o, err := scanOrderRow(row)
	if err != nil {
		return nil, err
	}
	items, err := r.loadItems(string(o.ID))
	if err != nil {
		return nil, err
	}
	o.Items = items
	return o, nil
}

func (r *PostgresOrderRepository) FindByRestaurantID(restaurantID common.RestaurantID) ([]*order.Order, error) {
	return r.queryOrders(
		`SELECT `+orderColumns+` FROM orders WHERE restaurant_id = $1`, string(restaurantID))
}

func (r *PostgresOrderRepository) FindActiveByRestaurantID(restaurantID common.RestaurantID) ([]*order.Order, error) {
	return r.queryOrders(
		`SELECT `+orderColumns+` FROM orders WHERE restaurant_id = $1 AND fulfillment_status IN ('paid','preparing','ready')`,
		string(restaurantID))
}

func (r *PostgresOrderRepository) FindBySessionID(sessionID string) ([]*order.Order, error) {
	return r.queryOrders(
		`SELECT `+orderColumns+` FROM orders WHERE session_id = $1`, sessionID)
}

func (r *PostgresOrderRepository) Update(o *order.Order) error {
	result, err := r.db.Exec(
		`UPDATE orders SET order_number=$2, total_amount=$3, fiat_amount=$4,
		   payment_method=$5, payment_status=$6, fulfillment_status=$7,
		   updated_at=$8, paid_at=$9, preparing_at=$10, ready_at=$11, completed_at=$12
		 WHERE id=$1`,
		string(o.ID), string(o.OrderNumber), o.TotalAmount, o.FiatAmount,
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
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, o := range result {
		items, err := r.loadItems(string(o.ID))
		if err != nil {
			return nil, err
		}
		o.Items = items
	}

	return result, nil
}

func (r *PostgresOrderRepository) loadItems(orderID string) ([]order.OrderItem, error) {
	rows, err := r.db.Query(
		`SELECT id, order_id, menu_item_id, name, quantity, unit_price, subtotal
		 FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []order.OrderItem
	for rows.Next() {
		var (
			id, oid, menuItemID, name string
			quantity                   int
			unitPrice, subtotal        float64
		)
		if err := rows.Scan(&id, &oid, &menuItemID, &name, &quantity, &unitPrice, &subtotal); err != nil {
			return nil, err
		}
		items = append(items, order.OrderItem{
			ID:         common.OrderItemID(id),
			OrderID:    common.OrderID(oid),
			MenuItemID: common.ItemID(menuItemID),
			Name:       name,
			Quantity:   quantity,
			UnitPrice:  unitPrice,
			Subtotal:   subtotal,
		})
	}
	return items, rows.Err()
}

func scanOrderRow(row *sql.Row) (*order.Order, error) {
	var (
		id, orderNum, restID, sessionID           string
		totalAmount                               int64
		fiatAmount                                float64
		payMethod, payStatus, fulStatus           string
		createdAt, updatedAt                      time.Time
		paidAt, preparingAt, readyAt, completedAt sql.NullTime
	)
	if err := row.Scan(&id, &orderNum, &restID, &sessionID, &totalAmount, &fiatAmount,
		&payMethod, &payStatus, &fulStatus,
		&createdAt, &updatedAt, &paidAt, &preparingAt, &readyAt, &completedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	return buildOrder(id, orderNum, restID, sessionID, totalAmount, fiatAmount,
		payMethod, payStatus, fulStatus, createdAt, updatedAt, paidAt, preparingAt, readyAt, completedAt), nil
}

func scanOrderRows(rows *sql.Rows) (*order.Order, error) {
	var (
		id, orderNum, restID, sessionID           string
		totalAmount                               int64
		fiatAmount                                float64
		payMethod, payStatus, fulStatus           string
		createdAt, updatedAt                      time.Time
		paidAt, preparingAt, readyAt, completedAt sql.NullTime
	)
	if err := rows.Scan(&id, &orderNum, &restID, &sessionID, &totalAmount, &fiatAmount,
		&payMethod, &payStatus, &fulStatus,
		&createdAt, &updatedAt, &paidAt, &preparingAt, &readyAt, &completedAt); err != nil {
		return nil, err
	}
	return buildOrder(id, orderNum, restID, sessionID, totalAmount, fiatAmount,
		payMethod, payStatus, fulStatus, createdAt, updatedAt, paidAt, preparingAt, readyAt, completedAt), nil
}

func buildOrder(id, orderNum, restID, sessionID string, totalAmount int64, fiatAmount float64,
	payMethod, payStatus, fulStatus string, createdAt, updatedAt time.Time,
	paidAt, preparingAt, readyAt, completedAt sql.NullTime) *order.Order {

	o := &order.Order{
		ID:                common.OrderID(id),
		OrderNumber:       common.OrderNumber(orderNum),
		RestaurantID:      common.RestaurantID(restID),
		SessionID:         sessionID,
		TotalAmount:       totalAmount,
		FiatAmount:        fiatAmount,
		PaymentMethod:     common.PaymentMethodType(payMethod),
		PaymentStatus:     common.PaymentStatus(payStatus),
		FulfillmentStatus: common.FulfillmentStatus(fulStatus),
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
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
	return o
}
