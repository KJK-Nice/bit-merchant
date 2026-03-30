package postgres

import (
	"database/sql"
	"errors"
	"time"

	"bitmerchant/internal/domain"
)

// OrderRepository implements domain.OrderRepository for PostgreSQL.
type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Save(order *domain.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := upsertOrder(tx, order); err != nil {
		return err
	}
	if err := replaceOrderItems(tx, order); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *OrderRepository) FindByID(id domain.OrderID) (*domain.Order, error) {
	row := r.db.QueryRow(
		`SELECT id, order_number, restaurant_id, session_id, total_amount, fiat_amount, payment_method, payment_status, fulfillment_status, created_at, updated_at, paid_at, preparing_at, ready_at, completed_at
		   FROM orders
		  WHERE id = $1`,
		string(id),
	)
	order, err := scanOrderRow(row)
	if err != nil {
		return nil, err
	}

	items, err := r.findItems(order.ID)
	if err != nil {
		return nil, err
	}
	order.Items = items
	return order, nil
}

func (r *OrderRepository) FindByOrderNumber(restaurantID domain.RestaurantID, orderNumber string) (*domain.Order, error) {
	row := r.db.QueryRow(
		`SELECT id, order_number, restaurant_id, session_id, total_amount, fiat_amount, payment_method, payment_status, fulfillment_status, created_at, updated_at, paid_at, preparing_at, ready_at, completed_at
		   FROM orders
		  WHERE restaurant_id = $1 AND order_number = $2
		  LIMIT 1`,
		string(restaurantID),
		orderNumber,
	)
	order, err := scanOrderRow(row)
	if err != nil {
		return nil, err
	}

	items, err := r.findItems(order.ID)
	if err != nil {
		return nil, err
	}
	order.Items = items
	return order, nil
}

func (r *OrderRepository) FindByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.Order, error) {
	rows, err := r.db.Query(
		`SELECT id, order_number, restaurant_id, session_id, total_amount, fiat_amount, payment_method, payment_status, fulfillment_status, created_at, updated_at, paid_at, preparing_at, ready_at, completed_at
		   FROM orders
		  WHERE restaurant_id = $1
		  ORDER BY created_at DESC`,
		string(restaurantID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanOrdersWithItems(rows)
}

func (r *OrderRepository) FindActiveByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.Order, error) {
	rows, err := r.db.Query(
		`SELECT id, order_number, restaurant_id, session_id, total_amount, fiat_amount, payment_method, payment_status, fulfillment_status, created_at, updated_at, paid_at, preparing_at, ready_at, completed_at
		   FROM orders
		  WHERE restaurant_id = $1
		    AND fulfillment_status IN ($2, $3, $4)
		  ORDER BY created_at DESC`,
		string(restaurantID),
		string(domain.FulfillmentStatusPaid),
		string(domain.FulfillmentStatusPreparing),
		string(domain.FulfillmentStatusReady),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanOrdersWithItems(rows)
}

func (r *OrderRepository) FindBySessionID(sessionID string) ([]*domain.Order, error) {
	rows, err := r.db.Query(
		`SELECT id, order_number, restaurant_id, session_id, total_amount, fiat_amount, payment_method, payment_status, fulfillment_status, created_at, updated_at, paid_at, preparing_at, ready_at, completed_at
		   FROM orders
		  WHERE session_id = $1
		  ORDER BY created_at DESC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanOrdersWithItems(rows)
}

func (r *OrderRepository) Update(order *domain.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.Exec(
		`UPDATE orders
		    SET order_number = $2,
		        restaurant_id = $3,
		        session_id = $4,
		        total_amount = $5,
		        fiat_amount = $6,
		        payment_method = $7,
		        payment_status = $8,
		        fulfillment_status = $9,
		        updated_at = $10,
		        paid_at = $11,
		        preparing_at = $12,
		        ready_at = $13,
		        completed_at = $14
		  WHERE id = $1`,
		string(order.ID),
		string(order.OrderNumber),
		string(order.RestaurantID),
		order.SessionID,
		order.TotalAmount,
		order.FiatAmount,
		string(order.PaymentMethod),
		string(order.PaymentStatus),
		string(order.FulfillmentStatus),
		order.UpdatedAt,
		order.PaidAt,
		order.PreparingAt,
		order.ReadyAt,
		order.CompletedAt,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("order not found")
	}

	if err := replaceOrderItems(tx, order); err != nil {
		return err
	}

	return tx.Commit()
}

func upsertOrder(tx *sql.Tx, order *domain.Order) error {
	_, err := tx.Exec(
		`INSERT INTO orders (id, order_number, restaurant_id, session_id, total_amount, fiat_amount, payment_method, payment_status, fulfillment_status, created_at, updated_at, paid_at, preparing_at, ready_at, completed_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		 ON CONFLICT (id) DO UPDATE
		 SET order_number = EXCLUDED.order_number,
		     restaurant_id = EXCLUDED.restaurant_id,
		     session_id = EXCLUDED.session_id,
		     total_amount = EXCLUDED.total_amount,
		     fiat_amount = EXCLUDED.fiat_amount,
		     payment_method = EXCLUDED.payment_method,
		     payment_status = EXCLUDED.payment_status,
		     fulfillment_status = EXCLUDED.fulfillment_status,
		     updated_at = EXCLUDED.updated_at,
		     paid_at = EXCLUDED.paid_at,
		     preparing_at = EXCLUDED.preparing_at,
		     ready_at = EXCLUDED.ready_at,
		     completed_at = EXCLUDED.completed_at`,
		string(order.ID),
		string(order.OrderNumber),
		string(order.RestaurantID),
		order.SessionID,
		order.TotalAmount,
		order.FiatAmount,
		string(order.PaymentMethod),
		string(order.PaymentStatus),
		string(order.FulfillmentStatus),
		order.CreatedAt,
		order.UpdatedAt,
		order.PaidAt,
		order.PreparingAt,
		order.ReadyAt,
		order.CompletedAt,
	)
	return err
}

func replaceOrderItems(tx *sql.Tx, order *domain.Order) error {
	if _, err := tx.Exec(`DELETE FROM order_items WHERE order_id = $1`, string(order.ID)); err != nil {
		return err
	}

	for _, item := range order.Items {
		_, err := tx.Exec(
			`INSERT INTO order_items (id, order_id, menu_item_id, name, quantity, unit_price, subtotal)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			string(item.ID),
			string(order.ID),
			string(item.MenuItemID),
			item.Name,
			item.Quantity,
			item.UnitPrice,
			item.Subtotal,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *OrderRepository) findItems(orderID domain.OrderID) ([]domain.OrderItem, error) {
	rows, err := r.db.Query(
		`SELECT id, order_id, menu_item_id, name, quantity, unit_price, subtotal
		   FROM order_items
		  WHERE order_id = $1`,
		string(orderID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var (
			id         string
			oid        string
			menuItemID string
			name       string
			quantity   int
			unitPrice  float64
			subtotal   float64
		)
		if err := rows.Scan(&id, &oid, &menuItemID, &name, &quantity, &unitPrice, &subtotal); err != nil {
			return nil, err
		}
		items = append(items, domain.OrderItem{
			ID:         domain.OrderItemID(id),
			OrderID:    domain.OrderID(oid),
			MenuItemID: domain.ItemID(menuItemID),
			Name:       name,
			Quantity:   quantity,
			UnitPrice:  unitPrice,
			Subtotal:   subtotal,
		})
	}
	return items, rows.Err()
}

func (r *OrderRepository) scanOrdersWithItems(rows *sql.Rows) ([]*domain.Order, error) {
	var orders []*domain.Order
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		items, itemErr := r.findItems(order.ID)
		if itemErr != nil {
			return nil, itemErr
		}
		order.Items = items
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func scanOrder(rows *sql.Rows) (*domain.Order, error) {
	var (
		id                string
		orderNumber       string
		restaurantID      string
		sessionID         string
		totalAmount       int64
		fiatAmount        float64
		paymentMethod     string
		paymentStatus     string
		fulfillmentStatus string
		createdAt         sql.NullTime
		updatedAt         sql.NullTime
		paidAt            sql.NullTime
		preparingAt       sql.NullTime
		readyAt           sql.NullTime
		completedAt       sql.NullTime
	)
	if err := rows.Scan(
		&id,
		&orderNumber,
		&restaurantID,
		&sessionID,
		&totalAmount,
		&fiatAmount,
		&paymentMethod,
		&paymentStatus,
		&fulfillmentStatus,
		&createdAt,
		&updatedAt,
		&paidAt,
		&preparingAt,
		&readyAt,
		&completedAt,
	); err != nil {
		return nil, err
	}

	return &domain.Order{
		ID:                domain.OrderID(id),
		OrderNumber:       domain.OrderNumber(orderNumber),
		RestaurantID:      domain.RestaurantID(restaurantID),
		SessionID:         sessionID,
		TotalAmount:       totalAmount,
		FiatAmount:        fiatAmount,
		PaymentMethod:     domain.PaymentMethodType(paymentMethod),
		PaymentStatus:     domain.PaymentStatus(paymentStatus),
		FulfillmentStatus: domain.FulfillmentStatus(fulfillmentStatus),
		CreatedAt:         createdAt.Time,
		UpdatedAt:         updatedAt.Time,
		PaidAt:            nullTimePtr(paidAt),
		PreparingAt:       nullTimePtr(preparingAt),
		ReadyAt:           nullTimePtr(readyAt),
		CompletedAt:       nullTimePtr(completedAt),
	}, nil
}

func scanOrderRow(row *sql.Row) (*domain.Order, error) {
	var (
		id                string
		orderNumber       string
		restaurantID      string
		sessionID         string
		totalAmount       int64
		fiatAmount        float64
		paymentMethod     string
		paymentStatus     string
		fulfillmentStatus string
		createdAt         sql.NullTime
		updatedAt         sql.NullTime
		paidAt            sql.NullTime
		preparingAt       sql.NullTime
		readyAt           sql.NullTime
		completedAt       sql.NullTime
	)
	if err := row.Scan(
		&id,
		&orderNumber,
		&restaurantID,
		&sessionID,
		&totalAmount,
		&fiatAmount,
		&paymentMethod,
		&paymentStatus,
		&fulfillmentStatus,
		&createdAt,
		&updatedAt,
		&paidAt,
		&preparingAt,
		&readyAt,
		&completedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}

	return &domain.Order{
		ID:                domain.OrderID(id),
		OrderNumber:       domain.OrderNumber(orderNumber),
		RestaurantID:      domain.RestaurantID(restaurantID),
		SessionID:         sessionID,
		TotalAmount:       totalAmount,
		FiatAmount:        fiatAmount,
		PaymentMethod:     domain.PaymentMethodType(paymentMethod),
		PaymentStatus:     domain.PaymentStatus(paymentStatus),
		FulfillmentStatus: domain.FulfillmentStatus(fulfillmentStatus),
		CreatedAt:         createdAt.Time,
		UpdatedAt:         updatedAt.Time,
		PaidAt:            nullTimePtr(paidAt),
		PreparingAt:       nullTimePtr(preparingAt),
		ReadyAt:           nullTimePtr(readyAt),
		CompletedAt:       nullTimePtr(completedAt),
	}, nil
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}
