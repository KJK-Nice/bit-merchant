package domain

import "time"

// DomainEvent represents a domain event interface
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// OrderPaid event is triggered when payment is confirmed
type OrderPaid struct {
	OrderID      OrderID
	RestaurantID RestaurantID
	OrderNumber  OrderNumber
	TotalAmount  int64
	CreatedAt    time.Time
}

// EventName returns the event name
func (e OrderPaid) EventName() string {
	return "OrderPaid"
}

// OccurredAt returns when the event occurred
func (e OrderPaid) OccurredAt() time.Time {
	return e.CreatedAt
}

// OrderStatusChanged event is triggered when order status changes
type OrderStatusChanged struct {
	OrderID        OrderID
	OrderNumber    OrderNumber
	PreviousStatus FulfillmentStatus
	NewStatus      FulfillmentStatus
	UpdatedAt      time.Time
}

// EventName returns the event name
func (e OrderStatusChanged) EventName() string {
	return "OrderStatusChanged"
}

// OccurredAt returns when the event occurred
func (e OrderStatusChanged) OccurredAt() time.Time {
	return e.UpdatedAt
}

// PaymentFailed event is triggered when payment fails
type PaymentFailed struct {
	PaymentID     PaymentID
	InvoiceID     string
	FailureReason string
	FailedAt      time.Time
}

// EventName returns the event name
func (e PaymentFailed) EventName() string {
	return "PaymentFailed"
}

// OccurredAt returns when the event occurred
func (e PaymentFailed) OccurredAt() time.Time {
	return e.FailedAt
}
