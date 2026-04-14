package kitchen_test

import (
	"bitmerchant/internal/common"
	kitchenCmd "bitmerchant/internal/ordering/app/command"
	"bitmerchant/internal/ordering/domain/order"

	"context"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMarkOrderReadyUseCase_Execute(t *testing.T) {
	t.Run("successfully marks order as ready", func(t *testing.T) {
		orderID := common.OrderID("order-123")
		existingOrder := createTestOrder("order-123", common.FulfillmentStatusPreparing, common.PaymentStatusPaid)

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

		uc := kitchenCmd.NewMarkOrderReadyUseCase(mockOrderRepo, mockEventBus)
		_, err := uc.Execute(context.Background(), orderID)

		assert.NoError(t, err)
		assert.Equal(t, common.FulfillmentStatusReady, savedOrder.FulfillmentStatus)
		assert.NotNil(t, savedOrder.ReadyAt)
		assert.Equal(t, "OrderReady", publishedEvent)
	})

	t.Run("fails if status transition is invalid", func(t *testing.T) {
		orderID := common.OrderID("order-123")
		// Paid -> Ready is invalid, must be Preparing first
		existingOrder := createTestOrder("order-123", common.FulfillmentStatusPaid, common.PaymentStatusPaid)

		mockOrderRepo := &mockOrderRepo{
			findByIDFn: func(id common.OrderID) (*order.Order, error) {
				return existingOrder, nil
			},
		}

		uc := kitchenCmd.NewMarkOrderReadyUseCase(mockOrderRepo, &mockEventBus{})
		_, err := uc.Execute(context.Background(), orderID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status transition")
	})
}
