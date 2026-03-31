package common

import (
	"context"
	"time"
)

const (
	EventOrderCreated   = "OrderCreated"
	EventOrderPaid      = "OrderPaid"
	EventOrderPreparing = "OrderPreparing"
	EventOrderReady     = "OrderReady"
	EventOrderCompleted = "OrderCompleted"
)

// DomainEvent represents a domain event interface.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// EventBus defines the interface for publishing domain events.
type EventBus interface {
	Publish(ctx context.Context, topic string, event interface{}) error
}
