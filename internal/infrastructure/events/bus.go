package events

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

// EventBus wraps Watermill pub/sub for domain events
type EventBus struct {
	publisher  message.Publisher
	subscriber message.Subscriber
}

// NewEventBus creates a new in-memory event bus
func NewEventBus() *EventBus {
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{},
		watermill.NewStdLogger(false, false),
	)

	return &EventBus{
		publisher:  pubSub,
		subscriber: pubSub,
	}
}

// Publish publishes a domain event
func (b *EventBus) Publish(ctx context.Context, topic string, event interface{}) error {
	// Convert domain event to Watermill message
	// This is a simplified implementation
	// In production, you'd serialize the event properly
	msg := message.NewMessage(watermill.NewUUID(), []byte(topic))
	return b.publisher.Publish(topic, msg)
}

// Subscribe subscribes to domain events
func (b *EventBus) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	return b.subscriber.Subscribe(ctx, topic)
}

// Close closes the event bus
func (b *EventBus) Close() error {
	return b.publisher.Close()
}
