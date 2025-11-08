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
	ID                 OrderID
	OrderNumber        OrderNumber
	RestaurantID       RestaurantID
	Items              []OrderItem
	TotalAmount        int64
	FiatAmount         float64
	PaymentStatus      PaymentStatus
	FulfillmentStatus  FulfillmentStatus
	LightningInvoiceID string
	LightningInvoice   string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	CompletedAt        *time.Time
}

// OrderItem represents an individual item within an order
type OrderItem struct {
	ID                OrderItemID
	OrderID           OrderID
	MenuItemID        ItemID
	Quantity          int
	UnitPrice         float64
	UnitPriceSatoshis int64
	Subtotal          int64
}

// NewOrder creates a new Order with validation
func NewOrder(id OrderID, orderNumber OrderNumber, restaurantID RestaurantID, items []OrderItem, totalAmount int64, fiatAmount float64, invoiceID, invoice string) (*Order, error) {
	if len(items) == 0 {
		return nil, errors.New("order must have at least one item")
	}
	if totalAmount <= 0 {
		return nil, errors.New("total amount must be greater than 0")
	}
	if invoiceID == "" {
		return nil, errors.New("lightning invoice ID is required")
	}

	now := time.Now()
	return &Order{
		ID:                 id,
		OrderNumber:        orderNumber,
		RestaurantID:       restaurantID,
		Items:              items,
		TotalAmount:        totalAmount,
		FiatAmount:         fiatAmount,
		PaymentStatus:      PaymentStatusPaid,
		FulfillmentStatus:  FulfillmentStatusPaid,
		LightningInvoiceID: invoiceID,
		LightningInvoice:   invoice,
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil
}

// NewOrderItem creates a new OrderItem
func NewOrderItem(id OrderItemID, orderID OrderID, menuItemID ItemID, quantity int, unitPrice float64, unitPriceSatoshis int64) (*OrderItem, error) {
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}
	if unitPrice <= 0 {
		return nil, errors.New("unit price must be greater than 0")
	}
	if unitPriceSatoshis <= 0 {
		return nil, errors.New("unit price in satoshis must be greater than 0")
	}

	subtotal := int64(quantity) * unitPriceSatoshis
	return &OrderItem{
		ID:                id,
		OrderID:           orderID,
		MenuItemID:        menuItemID,
		Quantity:          quantity,
		UnitPrice:         unitPrice,
		UnitPriceSatoshis: unitPriceSatoshis,
		Subtotal:          subtotal,
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
