package order

import (
	"errors"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/money"
)

// OrderItem represents an individual item within an order.
type OrderItem struct {
	ID         common.OrderItemID
	OrderID    common.OrderID
	MenuItemID common.ItemID
	Name       string
	Quantity   int
	UnitPrice  float64
	Subtotal   float64
	Currency   money.Currency
}

// NewOrderItem creates a new OrderItem. Currency defaults to USD.
func NewOrderItem(id common.OrderItemID, orderID common.OrderID, menuItemID common.ItemID, name string, quantity int, unitPrice float64) (*OrderItem, error) {
	return NewOrderItemWithCurrency(id, orderID, menuItemID, name, quantity, unitPrice, money.USD)
}

// NewOrderItemWithCurrency creates an OrderItem pinned to the order's currency.
func NewOrderItemWithCurrency(id common.OrderItemID, orderID common.OrderID, menuItemID common.ItemID, name string, quantity int, unitPrice float64, currency money.Currency) (*OrderItem, error) {
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}
	if unitPrice <= 0 {
		return nil, errors.New("unit price must be greater than 0")
	}
	if name == "" {
		return nil, errors.New("name must not be empty")
	}
	if currency.IsZero() {
		currency = money.USD
	}

	subtotal := float64(quantity) * unitPrice
	return &OrderItem{
		ID:         id,
		OrderID:    orderID,
		MenuItemID: menuItemID,
		Name:       name,
		Quantity:   quantity,
		UnitPrice:  unitPrice,
		Subtotal:   subtotal,
		Currency:   currency,
	}, nil
}
