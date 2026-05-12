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

const orderColumns = `id, order_number, restaurant_id, session_id,
	COALESCE(subtotal_amount, total_amount), total_amount, COALESCE(tax_amount, 0), COALESCE(tip_amount, 0),
	fiat_amount, COALESCE(currency, 'USD'),
	COALESCE(customer_name, ''), COALESCE(table_label, ''),
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
		`INSERT INTO orders (id, order_number, restaurant_id, session_id,
			subtotal_amount, total_amount, tax_amount, tip_amount, fiat_amount, currency,
			customer_name, table_label,
			payment_method, payment_status, fulfillment_status,
			created_at, updated_at, paid_at, preparing_at, ready_at, completed_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
		 ON CONFLICT (id) DO UPDATE SET
		   order_number=EXCLUDED.order_number,
		   subtotal_amount=EXCLUDED.subtotal_amount,
		   total_amount=EXCLUDED.total_amount,
		   tax_amount=EXCLUDED.tax_amount, tip_amount=EXCLUDED.tip_amount,
		   fiat_amount=EXCLUDED.fiat_amount,
		   currency=EXCLUDED.currency,
		   customer_name=EXCLUDED.customer_name, table_label=EXCLUDED.table_label,
		   payment_status=EXCLUDED.payment_status, fulfillment_status=EXCLUDED.fulfillment_status,
		   updated_at=EXCLUDED.updated_at, paid_at=EXCLUDED.paid_at,
		   preparing_at=EXCLUDED.preparing_at, ready_at=EXCLUDED.ready_at, completed_at=EXCLUDED.completed_at`,
		string(o.ID), string(o.OrderNumber), string(o.RestaurantID), o.SessionID,
		o.Subtotal, o.TotalAmount, o.TaxAmount, o.TipAmount, o.FiatAmount, currency.Code,
		o.CustomerName, o.TableLabel,
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
			`INSERT INTO order_items (id, order_id, menu_item_id, name, quantity, unit_price, subtotal, currency, unit_price_minor, subtotal_minor, modifiers, special_instructions, prep_complete)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
			 ON CONFLICT (id) DO NOTHING`,
			string(item.ID), string(item.OrderID), string(item.MenuItemID),
			item.Name, item.Quantity, item.UnitPrice, item.Subtotal,
			itemCur.Code, unitPriceMinor, subtotalMinor,
			modifiersJSON, item.SpecialInstructions, item.PrepComplete)
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
		`UPDATE orders SET order_number=$2,
		   subtotal_amount=$3, total_amount=$4, tax_amount=$5, tip_amount=$6,
		   fiat_amount=$7, currency=$8,
		   customer_name=$9, table_label=$10,
		   payment_method=$11, payment_status=$12, fulfillment_status=$13,
		   updated_at=$14, paid_at=$15, preparing_at=$16, ready_at=$17, completed_at=$18
		 WHERE id=$1`,
		string(o.ID), string(o.OrderNumber),
		o.Subtotal, o.TotalAmount, o.TaxAmount, o.TipAmount,
		o.FiatAmount, currency.Code,
		o.CustomerName, o.TableLabel,
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

func (r *PostgresOrderRepository) UpdateItemPrepComplete(orderID common.OrderID, itemID common.OrderItemID, complete bool) error {
	result, err := r.db.Exec(
		`UPDATE order_items SET prep_complete = $3 WHERE id = $1 AND order_id = $2`,
		string(itemID), string(orderID), complete)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("order item not found")
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
		        COALESCE(modifiers, '[]'::jsonb), COALESCE(special_instructions, ''), COALESCE(prep_complete, false)
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
			prepComplete              bool
		)
		if err := rows.Scan(&id, &oid, &menuItemID, &name, &quantity, &unitPrice, &subtotal, &currencyCode, &modifiersJSON, &specialInstructions, &prepComplete); err != nil {
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
			PrepComplete:        prepComplete,
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

type orderRow struct {
	id, orderNum, restID, sessionID           string
	subtotal, totalAmount, taxAmount          int64
	tipAmount                                 int64
	fiatAmount                                float64
	currencyCode                              string
	customerName, tableLabel                  string
	payMethod, payStatus, fulStatus           string
	createdAt, updatedAt                      time.Time
	paidAt, preparingAt, readyAt, completedAt sql.NullTime
}

func (r *orderRow) targets() []any {
	return []any{
		&r.id, &r.orderNum, &r.restID, &r.sessionID,
		&r.subtotal, &r.totalAmount, &r.taxAmount, &r.tipAmount,
		&r.fiatAmount, &r.currencyCode,
		&r.customerName, &r.tableLabel,
		&r.payMethod, &r.payStatus, &r.fulStatus,
		&r.createdAt, &r.updatedAt, &r.paidAt, &r.preparingAt, &r.readyAt, &r.completedAt,
	}
}

func scanOrderRow(row *sql.Row) (*order.Order, error) {
	var r orderRow
	if err := row.Scan(r.targets()...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	return buildOrder(r), nil
}

func scanOrderRows(rows *sql.Rows) (*order.Order, error) {
	var r orderRow
	if err := rows.Scan(r.targets()...); err != nil {
		return nil, err
	}
	return buildOrder(r), nil
}

func buildOrder(r orderRow) *order.Order {
	currency, err := money.Parse(r.currencyCode)
	if err != nil {
		currency = money.USD
	}
	o := &order.Order{
		ID:                common.OrderID(r.id),
		OrderNumber:       common.OrderNumber(r.orderNum),
		RestaurantID:      common.RestaurantID(r.restID),
		SessionID:         r.sessionID,
		Subtotal:          r.subtotal,
		TaxAmount:         r.taxAmount,
		TipAmount:         r.tipAmount,
		TotalAmount:       r.totalAmount,
		FiatAmount:        r.fiatAmount,
		Currency:          currency,
		CustomerName:      r.customerName,
		TableLabel:        r.tableLabel,
		PaymentMethod:     common.PaymentMethodType(r.payMethod),
		PaymentStatus:     common.PaymentStatus(r.payStatus),
		FulfillmentStatus: common.FulfillmentStatus(r.fulStatus),
		CreatedAt:         r.createdAt,
		UpdatedAt:         r.updatedAt,
	}
	paidAt := r.paidAt
	preparingAt := r.preparingAt
	readyAt := r.readyAt
	completedAt := r.completedAt
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
