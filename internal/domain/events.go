package domain

import (
	"context"
	"time"
)

// Event constants
const (
	EventOrderCreated   = "OrderCreated"
	EventOrderPaid      = "OrderPaid"
	EventOrderPreparing = "OrderPreparing"
	EventOrderReady     = "OrderReady"
	EventOrderCompleted = "OrderCompleted"
)

// DomainEvent represents a domain event interface
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// EventBus defines the interface for publishing domain events
type EventBus interface {
	Publish(ctx context.Context, topic string, event interface{}) error
}

// OrderCreated event is triggered when order is created
type OrderCreated struct {
	OrderID      OrderID
	RestaurantID RestaurantID
	OrderNumber  OrderNumber
	TotalAmount  int64
	CreatedAt    time.Time
}

// EventName returns the event name
func (e OrderCreated) EventName() string {
	return EventOrderCreated
}

// OccurredAt returns when the event occurred
func (e OrderCreated) OccurredAt() time.Time {
	return e.CreatedAt
}

// OrderPaid event is triggered when payment is confirmed
type OrderPaid struct {
	OrderID      OrderID
	RestaurantID RestaurantID
	OrderNumber  OrderNumber
	TotalAmount  int64
	PaidAt       time.Time
}

// EventName returns the event name
func (e OrderPaid) EventName() string {
	return EventOrderPaid
}

// OccurredAt returns when the event occurred
func (e OrderPaid) OccurredAt() time.Time {
	return e.PaidAt
}

// OrderPreparing event is triggered when order starts preparing
type OrderPreparing struct {
	OrderID      OrderID
	RestaurantID RestaurantID
	OrderNumber  OrderNumber
	PreparingAt  time.Time
}

// EventName returns the event name
func (e OrderPreparing) EventName() string {
	return EventOrderPreparing
}

// OccurredAt returns when the event occurred
func (e OrderPreparing) OccurredAt() time.Time {
	return e.PreparingAt
}

// OrderReady event is triggered when order is ready
type OrderReady struct {
	OrderID      OrderID
	RestaurantID RestaurantID
	OrderNumber  OrderNumber
	ReadyAt      time.Time
}

// EventName returns the event name
func (e OrderReady) EventName() string {
	return EventOrderReady
}

// OccurredAt returns when the event occurred
func (e OrderReady) OccurredAt() time.Time {
	return e.ReadyAt
}

// OrderCompleted event is triggered when order is completed
type OrderCompleted struct {
	OrderID      OrderID
	RestaurantID RestaurantID
	OrderNumber  OrderNumber
	CompletedAt  time.Time
}

// EventName returns the event name
func (e OrderCompleted) EventName() string {
	return EventOrderCompleted
}

// OccurredAt returns when the event occurred
func (e OrderCompleted) OccurredAt() time.Time {
	return e.CompletedAt
}
