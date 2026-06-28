package order_test

import (
	"context"
	"sync"
	"testing"

	"bitmerchant/internal/common"
	"bitmerchant/internal/infrastructure/repositories/memory"
	orderCmd "bitmerchant/internal/ordering/app/command"
	"bitmerchant/internal/ordering/domain/order"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingBus counts events published per topic so we can assert idempotency.
type recordingBus struct {
	mu     sync.Mutex
	topics []string
}

func (b *recordingBus) Publish(_ context.Context, topic string, _ interface{}) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.topics = append(b.topics, topic)
	return nil
}

func (b *recordingBus) count(topic string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := 0
	for _, t := range b.topics {
		if t == topic {
			n++
		}
	}
	return n
}

func seedActiveOrder(t *testing.T, repo order.Repository) *order.Order {
	t.Helper()
	item, _ := order.NewOrderItem("oi1", "o1", "mi1", "Bao", 1, 9.0)
	o, err := order.NewOrder("o1", "A29F", "r1", "sess1", []order.OrderItem{*item}, 900, common.PaymentMethodTypeCash)
	require.NoError(t, err)
	o.CustomerName = "Maya"
	o.TableLabel = "7"
	require.NoError(t, repo.Save(o))
	return o
}

func TestRequestServerHandler_PublishesOnceThenThrottles(t *testing.T) {
	repo := memory.NewMemoryOrderRepository()
	bus := &recordingBus{}
	seedActiveOrder(t, repo)

	h := orderCmd.NewRequestServerHandler(repo, bus, nil, nil)

	_, err := h.Handle(context.Background(), orderCmd.RequestServer{OrderID: "o1"})
	require.NoError(t, err)
	assert.Equal(t, 1, bus.count(common.EventServerCalled), "first call publishes")

	saved, _ := repo.FindByID("o1")
	assert.NotNil(t, saved.ServerCalledAt, "request persisted on the order")

	// Immediate repeat is within the throttle window: idempotent success, no event.
	_, err = h.Handle(context.Background(), orderCmd.RequestServer{OrderID: "o1"})
	require.NoError(t, err)
	assert.Equal(t, 1, bus.count(common.EventServerCalled), "repeat within window does not publish")
}

func TestRequestBillHandler_PublishesOnceThenThrottles(t *testing.T) {
	repo := memory.NewMemoryOrderRepository()
	bus := &recordingBus{}
	seedActiveOrder(t, repo)

	h := orderCmd.NewRequestBillHandler(repo, bus, nil, nil)

	_, err := h.Handle(context.Background(), orderCmd.RequestBill{OrderID: "o1"})
	require.NoError(t, err)
	assert.Equal(t, 1, bus.count(common.EventBillRequested))

	_, err = h.Handle(context.Background(), orderCmd.RequestBill{OrderID: "o1"})
	require.NoError(t, err)
	assert.Equal(t, 1, bus.count(common.EventBillRequested))
}
