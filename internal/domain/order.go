package domain

import (
	"errors"
	"time"
)

// OrderID represents a unique order identifier
type OrderID string

// OrderNumber represents a human-readable order number
type OrderNumber string

// OrderItemID represents a unique order item identifier
type OrderItemID string

// PaymentStatus represents payment status
type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusFailed  PaymentStatus = "failed"
	PaymentStatusExpired PaymentStatus = "expired"
)

// FulfillmentStatus represents order fulfillment status
type FulfillmentStatus string

const (
	FulfillmentStatusPaid      FulfillmentStatus = "paid"
	FulfillmentStatusPreparing FulfillmentStatus = "preparing"
	FulfillmentStatusReady     FulfillmentStatus = "ready"
	FulfillmentStatusCompleted FulfillmentStatus = "completed"
)

// Order represents a customer purchase record
type Order struct {
	ID                OrderID
	OrderNumber       OrderNumber
	RestaurantID      RestaurantID
	Items             []OrderItem
	TotalAmount       int64
	FiatAmount        float64
	PaymentMethod     PaymentMethodType
	PaymentStatus     PaymentStatus
	FulfillmentStatus FulfillmentStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
	PaidAt            *time.Time
	PreparingAt       *time.Time
	ReadyAt           *time.Time
	CompletedAt       *time.Time
}

// NewOrder creates a new Order with validation
func NewOrder(id OrderID, orderNumber OrderNumber, restaurantID RestaurantID, items []OrderItem, totalAmount int64, paymentMethod PaymentMethodType) (*Order, error) {
	if len(items) == 0 {
		return nil, errors.New("order must have at least one item")
	}
	if totalAmount <= 0 {
		return nil, errors.New("total amount must be greater than 0")
	}

	now := time.Now()
	return &Order{
		ID:                id,
		OrderNumber:       orderNumber,
		RestaurantID:      restaurantID,
		Items:             items,
		TotalAmount:       totalAmount,
		PaymentMethod:     paymentMethod,
		PaymentStatus:     PaymentStatusPending,
		FulfillmentStatus: FulfillmentStatusPaid, // This seems wrong in original code too, should probably be derived
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

// OrderItem represents an individual item within an order
type OrderItem struct {
	ID         OrderItemID
	OrderID    OrderID
	MenuItemID ItemID
	Name       string
	Quantity   int
	UnitPrice  float64
	Subtotal   float64
}

// NewOrderItem creates a new OrderItem
func NewOrderItem(id OrderItemID, orderID OrderID, menuItemID ItemID, name string, quantity int, unitPrice float64) (*OrderItem, error) {
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}
	if unitPrice <= 0 {
		return nil, errors.New("unit price must be greater than 0")
	}
	if name == "" {
		return nil, errors.New("name must not be empty")
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
	}, nil
}

// UpdateFulfillmentStatus updates order fulfillment status with validation
func (o *Order) UpdateFulfillmentStatus(newStatus FulfillmentStatus) error {
	if !isValidStatusTransition(o.FulfillmentStatus, newStatus) {
		return errors.New("invalid status transition")
	}

	o.FulfillmentStatus = newStatus
	o.UpdatedAt = time.Now()

	if newStatus == FulfillmentStatusCompleted {
		now := time.Now()
		o.CompletedAt = &now
	}

	return nil
}

// isValidStatusTransition validates status transitions
func isValidStatusTransition(current, new FulfillmentStatus) bool {
	validTransitions := map[FulfillmentStatus][]FulfillmentStatus{
		FulfillmentStatusPaid:      {FulfillmentStatusPreparing},
		FulfillmentStatusPreparing: {FulfillmentStatusReady},
		FulfillmentStatusReady:     {FulfillmentStatusCompleted},
		FulfillmentStatusCompleted: {},
	}

	allowed, exists := validTransitions[current]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == new {
			return true
		}
	}

	return false
}
