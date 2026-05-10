package adapters

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/money"
	"bitmerchant/internal/ordering/domain/order"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

const orderColumns = `id, order_number, restaurant_id, session_id, total_amount, fiat_amount, COALESCE(currency, 'USD'),
	payment_method, payment_status, fulfillment_status,
	created_at, updated_at, paid_at, preparing_at, ready_at, completed_at`

// NextOrderNumber atomically allocates the next order number for restaurantID.
// Race-free: the UPDATE in ON CONFLICT takes the row lock, so concurrent
// callers serialize on Postgres rather than racing in the application.
func (r *PostgresOrderRepository) NextOrderNumber(restaurantID common.RestaurantID) (int, error) {
	var n int
	err := r.db.QueryRow(`
		INSERT INTO restaurant_order_counters (restaurant_id, last_number)
		VALUES ($1, 1)
		ON CONFLICT (restaurant_id) DO UPDATE
			SET last_number = restaurant_order_counters.last_number + 1
		RETURNING last_number`,
		string(restaurantID),
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("allocate order number for restaurant %s: %w", restaurantID, err)
	}
	return n, nil
}

func (r *PostgresOrderRepository) Save(o *order.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	currency := o.Currency
	if currency.IsZero() {
		currency = money.USD
	}
	_, err = tx.Exec(
		`INSERT INTO orders (id, order_number, restaurant_id, session_id, total_amount, fiat_amount, currency,
			payment_method, payment_status, fulfillment_status,
			created_at, updated_at, paid_at, preparing_at, ready_at, completed_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		 ON CONFLICT (id) DO UPDATE SET
		   order_number=EXCLUDED.order_number,
		   total_amount=EXCLUDED.total_amount, fiat_amount=EXCLUDED.fiat_amount,
		   currency=EXCLUDED.currency,
		   payment_status=EXCLUDED.payment_status, fulfillment_status=EXCLUDED.fulfillment_status,
		   updated_at=EXCLUDED.updated_at, paid_at=EXCLUDED.paid_at,
		   preparing_at=EXCLUDED.preparing_at, ready_at=EXCLUDED.ready_at, completed_at=EXCLUDED.completed_at`,
		string(o.ID), string(o.OrderNumber), string(o.RestaurantID), o.SessionID,
		o.TotalAmount, o.FiatAmount, currency.Code,
		string(o.PaymentMethod), string(o.PaymentStatus), string(o.FulfillmentStatus),
		o.CreatedAt, o.UpdatedAt, o.PaidAt, o.PreparingAt, o.ReadyAt, o.CompletedAt)
	if err != nil {
		return err
	}

	for _, item := range o.Items {
		itemCur := item.Currency
		if itemCur.IsZero() {
			itemCur = currency
		}
		unitPriceMinor := money.FromMajor(item.UnitPrice, itemCur).Amount
		subtotalMinor := money.FromMajor(item.Subtotal, itemCur).Amount
		modifiersJSON, merr := marshalOrderModifiers(item.Modifiers)
		if merr != nil {
			return merr
		}
		_, err = tx.Exec(
			`INSERT INTO order_items (id, order_id, menu_item_id, name, quantity, unit_price, subtotal, currency, unit_price_minor, subtotal_minor, modifiers, special_instructions)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			 ON CONFLICT (id) DO NOTHING`,
			string(item.ID), string(item.OrderID), string(item.MenuItemID),
			item.Name, item.Quantity, item.UnitPrice, item.Subtotal,
			itemCur.Code, unitPriceMinor, subtotalMinor,
			modifiersJSON, item.SpecialInstructions)
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
	currency := o.Currency
	if currency.IsZero() {
		currency = money.USD
	}
	result, err := r.db.Exec(
		`UPDATE orders SET order_number=$2, total_amount=$3, fiat_amount=$4, currency=$5,
		   payment_method=$6, payment_status=$7, fulfillment_status=$8,
		   updated_at=$9, paid_at=$10, preparing_at=$11, ready_at=$12, completed_at=$13
		 WHERE id=$1`,
		string(o.ID), string(o.OrderNumber), o.TotalAmount, o.FiatAmount, currency.Code,
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
		`SELECT id, order_id, menu_item_id, name, quantity, unit_price, subtotal, COALESCE(currency, 'USD'),
		        COALESCE(modifiers, '[]'::jsonb), COALESCE(special_instructions, '')
		 FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []order.OrderItem
	for rows.Next() {
		var (
			id, oid, menuItemID, name string
			quantity                  int
			unitPrice, subtotal       float64
			currencyCode              string
			modifiersJSON             []byte
			specialInstructions       string
		)
		if err := rows.Scan(&id, &oid, &menuItemID, &name, &quantity, &unitPrice, &subtotal, &currencyCode, &modifiersJSON, &specialInstructions); err != nil {
			return nil, err
		}
		currency, err := money.Parse(currencyCode)
		if err != nil {
			currency = money.USD
		}
		items = append(items, order.OrderItem{
			ID:                  common.OrderItemID(id),
			OrderID:             common.OrderID(oid),
			MenuItemID:          common.ItemID(menuItemID),
			Name:                name,
			Quantity:            quantity,
			UnitPrice:           unitPrice,
			Subtotal:            subtotal,
			Currency:            currency,
			Modifiers:           unmarshalOrderModifiers(modifiersJSON),
			SpecialInstructions: specialInstructions,
		})
	}
	return items, rows.Err()
}

// jsonOrderModifier mirrors order.OrderItemModifier for JSON.
type jsonOrderModifier struct {
	GroupName  string  `json:"group_name"`
	OptionName string  `json:"option_name"`
	PriceDelta float64 `json:"price_delta"`
}

func marshalOrderModifiers(mods []order.OrderItemModifier) ([]byte, error) {
	if len(mods) == 0 {
		return []byte("[]"), nil
	}
	jms := make([]jsonOrderModifier, len(mods))
	for i, m := range mods {
		jms[i] = jsonOrderModifier{GroupName: m.GroupName, OptionName: m.OptionName, PriceDelta: m.PriceDelta}
	}
	return json.Marshal(jms)
}

func unmarshalOrderModifiers(data []byte) []order.OrderItemModifier {
	if len(data) == 0 {
		return nil
	}
	var jms []jsonOrderModifier
	if err := json.Unmarshal(data, &jms); err != nil {
		return nil
	}
	mods := make([]order.OrderItemModifier, len(jms))
	for i, jm := range jms {
		mods[i] = order.OrderItemModifier{GroupName: jm.GroupName, OptionName: jm.OptionName, PriceDelta: jm.PriceDelta}
	}
	return mods
}

func scanOrderRow(row *sql.Row) (*order.Order, error) {
	var (
		id, orderNum, restID, sessionID           string
		totalAmount                               int64
		fiatAmount                                float64
		currencyCode                              string
		payMethod, payStatus, fulStatus           string
		createdAt, updatedAt                      time.Time
		paidAt, preparingAt, readyAt, completedAt sql.NullTime
	)
	if err := row.Scan(&id, &orderNum, &restID, &sessionID, &totalAmount, &fiatAmount, &currencyCode,
		&payMethod, &payStatus, &fulStatus,
		&createdAt, &updatedAt, &paidAt, &preparingAt, &readyAt, &completedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	return buildOrder(id, orderNum, restID, sessionID, totalAmount, fiatAmount, currencyCode,
		payMethod, payStatus, fulStatus, createdAt, updatedAt, paidAt, preparingAt, readyAt, completedAt), nil
}

func scanOrderRows(rows *sql.Rows) (*order.Order, error) {
	var (
		id, orderNum, restID, sessionID           string
		totalAmount                               int64
		fiatAmount                                float64
		currencyCode                              string
		payMethod, payStatus, fulStatus           string
		createdAt, updatedAt                      time.Time
		paidAt, preparingAt, readyAt, completedAt sql.NullTime
	)
	if err := rows.Scan(&id, &orderNum, &restID, &sessionID, &totalAmount, &fiatAmount, &currencyCode,
		&payMethod, &payStatus, &fulStatus,
		&createdAt, &updatedAt, &paidAt, &preparingAt, &readyAt, &completedAt); err != nil {
		return nil, err
	}
	return buildOrder(id, orderNum, restID, sessionID, totalAmount, fiatAmount, currencyCode,
		payMethod, payStatus, fulStatus, createdAt, updatedAt, paidAt, preparingAt, readyAt, completedAt), nil
}

func buildOrder(id, orderNum, restID, sessionID string, totalAmount int64, fiatAmount float64, currencyCode string,
	payMethod, payStatus, fulStatus string, createdAt, updatedAt time.Time,
	paidAt, preparingAt, readyAt, completedAt sql.NullTime) *order.Order {

	currency, err := money.Parse(currencyCode)
	if err != nil {
		currency = money.USD
	}
	o := &order.Order{
		ID:                common.OrderID(id),
		OrderNumber:       common.OrderNumber(orderNum),
		RestaurantID:      common.RestaurantID(restID),
		SessionID:         sessionID,
		TotalAmount:       totalAmount,
		FiatAmount:        fiatAmount,
		Currency:          currency,
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
