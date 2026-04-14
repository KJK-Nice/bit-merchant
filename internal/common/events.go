package common

import (
	"context"
	"time"
)

const (
	EventOrderCreated   = "order.created"
	EventOrderPaid      = "order.paid"
	EventOrderPreparing = "order.preparing"
	EventOrderReady     = "order.ready"
	EventOrderCompleted = "order.completed"
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
