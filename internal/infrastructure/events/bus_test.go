package events_test

import (
	"context"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"

	"bitmerchant/internal/infrastructure/events"
)

// SubscriberForGroup must deliver every published message to BOTH the default
// subscriber AND each named-group subscriber. This is the in-memory contract
// that the wiring layer (service.go) relies on so SSE projections and
// push-notification handlers both react to the same domain event.
//
// On NATS this is enforced by giving each group a distinct QueueGroupPrefix —
// untestable here without testcontainers — but the gochannel backend has the
// same observable contract, so a memory-mode test is a sufficient guard
// against the wiring regression we're protecting against.
func TestSubscriberForGroup_FansOutToEveryGroup(t *testing.T) {
	bus, err := events.NewEventBusWithConfig(events.Config{Backend: "memory"})
	if err != nil {
		t.Fatalf("new event bus: %v", err)
	}
	t.Cleanup(func() { _ = bus.Close() })

	const topic = "test.fanout"
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	defaultCh, err := bus.Subscriber().Subscribe(ctx, topic)
	if err != nil {
		t.Fatalf("default subscribe: %v", err)
	}
	notifCh, err := bus.SubscriberForGroup("notif").Subscribe(ctx, topic)
	if err != nil {
		t.Fatalf("notif group subscribe: %v", err)
	}

	if err := bus.Publish(ctx, topic, map[string]string{"k": "v"}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	for _, sub := range []struct {
		name string
		ch   <-chan *message.Message
	}{
		{"default", defaultCh},
		{"notif", notifCh},
	} {
		select {
		case m := <-sub.ch:
			if m == nil {
				t.Fatalf("%s group: channel closed before message", sub.name)
			}
			m.Ack()
		case <-time.After(2 * time.Second):
			t.Fatalf("%s group: did not receive published message — fan-out broken", sub.name)
		}
	}
}
