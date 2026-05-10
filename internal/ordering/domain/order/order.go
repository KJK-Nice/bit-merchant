package order

import (
	"errors"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/money"
)

// Order represents a customer purchase record.
type Order struct {
	ID                common.OrderID
	OrderNumber       common.OrderNumber
	RestaurantID      common.RestaurantID
	SessionID         string
	Items             []OrderItem
	TotalAmount       int64
	FiatAmount        float64
	Currency          money.Currency
	PaymentMethod     common.PaymentMethodType
	PaymentStatus     common.PaymentStatus
	FulfillmentStatus common.FulfillmentStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
	PaidAt            *time.Time
	PreparingAt       *time.Time
	ReadyAt           *time.Time
	CompletedAt       *time.Time
}

// Total returns the order total as money.Money. Falls back to USD when the
// order was loaded from a row that predates currency support.
func (o *Order) Total() money.Money {
	c := o.Currency
	if c.IsZero() {
		c = money.USD
	}
	return money.New(o.TotalAmount, c)
}

// NewOrder creates a new Order with validation. Currency defaults to USD.
func NewOrder(id common.OrderID, orderNumber common.OrderNumber, restaurantID common.RestaurantID, sessionID string, items []OrderItem, totalAmount int64, paymentMethod common.PaymentMethodType) (*Order, error) {
	return NewOrderWithCurrency(id, orderNumber, restaurantID, sessionID, items, totalAmount, paymentMethod, money.USD)
}

// NewOrderWithCurrency creates an Order pinned to the restaurant's base currency.
func NewOrderWithCurrency(id common.OrderID, orderNumber common.OrderNumber, restaurantID common.RestaurantID, sessionID string, items []OrderItem, totalAmount int64, paymentMethod common.PaymentMethodType, currency money.Currency) (*Order, error) {
	if len(items) == 0 {
		return nil, errors.New("order must have at least one item")
	}
	if totalAmount <= 0 {
		return nil, errors.New("total amount must be greater than 0")
	}
	if sessionID == "" {
		return nil, errors.New("session ID is required")
	}
	if currency.IsZero() {
		currency = money.USD
	}

	now := time.Now()
	return &Order{
		ID:                id,
		OrderNumber:       orderNumber,
		RestaurantID:      restaurantID,
		SessionID:         sessionID,
		Items:             items,
		TotalAmount:       totalAmount,
		Currency:          currency,
		PaymentMethod:     paymentMethod,
		PaymentStatus:     common.PaymentStatusPending,
		FulfillmentStatus: common.FulfillmentStatusPaid,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

// AllItemsPrepComplete reports whether every line item is marked prep complete.
// An order with no items returns false (defensive — should not occur in practice).
func (o *Order) AllItemsPrepComplete() bool {
	if len(o.Items) == 0 {
		return false
	}
	for _, item := range o.Items {
		if !item.PrepComplete {
			return false
		}
	}
	return true
}

// ItemPrepComplete reports the current prep_complete flag for a line item.
// Returns ok=false if the item ID is not part of this order.
func (o *Order) ItemPrepComplete(itemID common.OrderItemID) (bool, bool) {
	for i := range o.Items {
		if o.Items[i].ID == itemID {
			return o.Items[i].PrepComplete, true
		}
	}
	return false, false
}

// SetItemPrepComplete sets the prep_complete flag for a single line item.
// Returns ok=false if the item ID is not part of this order.
func (o *Order) SetItemPrepComplete(itemID common.OrderItemID, complete bool) bool {
	for i := range o.Items {
		if o.Items[i].ID == itemID {
			o.Items[i].PrepComplete = complete
			o.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// MarkPaid transitions payment status to paid.
func (o *Order) MarkPaid() {
	o.PaymentStatus = common.PaymentStatusPaid
	now := time.Now()
	o.PaidAt = &now
	o.UpdatedAt = now
}

// StartPreparing transitions fulfillment to preparing.
func (o *Order) StartPreparing() error {
	if o.PaymentStatus != common.PaymentStatusPaid {
		return errors.New("cannot prepare unpaid order")
	}
	if err := o.updateFulfillmentStatus(common.FulfillmentStatusPreparing); err != nil {
		return err
	}
	now := time.Now()
	o.PreparingAt = &now
	return nil
}

// MarkReady transitions fulfillment to ready.
func (o *Order) MarkReady() error {
	if err := o.updateFulfillmentStatus(common.FulfillmentStatusReady); err != nil {
		return err
	}
	now := time.Now()
	o.ReadyAt = &now
	return nil
}

// Complete transitions fulfillment to completed.
func (o *Order) Complete() error {
	if err := o.updateFulfillmentStatus(common.FulfillmentStatusCompleted); err != nil {
		return err
	}
	now := time.Now()
	o.CompletedAt = &now
	return nil
}

// UpdateFulfillmentStatus updates order fulfillment status with validation (kept for backward compat).
func (o *Order) UpdateFulfillmentStatus(newStatus common.FulfillmentStatus) error {
	return o.updateFulfillmentStatus(newStatus)
}

func (o *Order) updateFulfillmentStatus(newStatus common.FulfillmentStatus) error {
	if !isValidStatusTransition(o.FulfillmentStatus, newStatus) {
		return errors.New("invalid status transition")
	}
	o.FulfillmentStatus = newStatus
	o.UpdatedAt = time.Now()
	if newStatus == common.FulfillmentStatusCompleted {
		now := time.Now()
		o.CompletedAt = &now
	}
	return nil
}

func isValidStatusTransition(current, next common.FulfillmentStatus) bool {
	validTransitions := map[common.FulfillmentStatus][]common.FulfillmentStatus{
		common.FulfillmentStatusPaid:      {common.FulfillmentStatusPreparing},
		common.FulfillmentStatusPreparing: {common.FulfillmentStatusReady},
		common.FulfillmentStatusReady:     {common.FulfillmentStatusCompleted},
		common.FulfillmentStatusCompleted: {},
	}

	allowed, exists := validTransitions[current]
	if !exists {
		return false
	}
	for _, status := range allowed {
		if status == next {
			return true
		}
	}
	return false
}
