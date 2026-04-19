package kitchen_test

import (
	"context"
	"testing"

	"bitmerchant/internal/common"
	kitchenCmd "bitmerchant/internal/ordering/app/command"
	"bitmerchant/internal/ordering/domain/order"

	"github.com/stretchr/testify/assert"
)

func TestMarkOrderCompletedHandler_Handle(t *testing.T) {
	t.Run("successfully marks order as completed", func(t *testing.T) {
		orderID := common.OrderID("order-123")
		existingOrder := createTestOrder("order-123", common.FulfillmentStatusReady, common.PaymentStatusPaid)

		var savedOrder *order.Order
		var publishedEvent string

		mockOrderRepo := &mockOrderRepo{
			findByIDFn: func(id common.OrderID) (*order.Order, error) {
				return existingOrder, nil
			},
			updateFn: func(order *order.Order) error {
				savedOrder = order
				return nil
			},
		}

		mockEventBus := &mockEventBus{
			publishFn: func(ctx context.Context, topic string, event interface{}) error {
				publishedEvent = topic
				return nil
			},
		}

		uc := kitchenCmd.NewMarkOrderCompletedHandler(mockOrderRepo, mockEventBus, nil, nil)
		_, err := uc.Handle(context.Background(), kitchenCmd.MarkOrderCompleted{OrderID: orderID})

		assert.NoError(t, err)
		assert.Equal(t, common.FulfillmentStatusCompleted, savedOrder.FulfillmentStatus)
		assert.NotNil(t, savedOrder.CompletedAt)
		assert.Equal(t, common.EventOrderCompleted, publishedEvent)
	})

	t.Run("fails if status transition is invalid", func(t *testing.T) {
		orderID := common.OrderID("order-123")
		// Preparing -> Completed is invalid, must become Ready first.
		existingOrder := createTestOrder("order-123", common.FulfillmentStatusPreparing, common.PaymentStatusPaid)

		mockOrderRepo := &mockOrderRepo{
			findByIDFn: func(id common.OrderID) (*order.Order, error) {
				return existingOrder, nil
			},
		}

		uc := kitchenCmd.NewMarkOrderCompletedHandler(mockOrderRepo, &mockEventBus{}, nil, nil)
		_, err := uc.Handle(context.Background(), kitchenCmd.MarkOrderCompleted{OrderID: orderID})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status transition")
	})
}
