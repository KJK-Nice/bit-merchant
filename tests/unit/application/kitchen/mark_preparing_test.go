package kitchen_test

import (
	"bitmerchant/internal/common"
	kitchenCmd "bitmerchant/internal/ordering/app/command"
	"bitmerchant/internal/ordering/domain/order"

	"context"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMarkOrderPreparingUseCase_Execute(t *testing.T) {
	t.Run("successfully marks order as preparing", func(t *testing.T) {
		orderID := common.OrderID("order-123")
		// Must be Paid to start preparing
		existingOrder := createTestOrder("order-123", common.FulfillmentStatusPaid, common.PaymentStatusPaid)

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

		uc := kitchenCmd.NewMarkOrderPreparingUseCase(mockOrderRepo, mockEventBus)
		_, err := uc.Execute(context.Background(), orderID)

		assert.NoError(t, err)
		assert.Equal(t, common.FulfillmentStatusPreparing, savedOrder.FulfillmentStatus)
		assert.NotNil(t, savedOrder.PreparingAt)
		assert.Equal(t, "order.preparing", publishedEvent)
	})

	t.Run("fails if order is not paid", func(t *testing.T) {
		orderID := common.OrderID("order-123")
		existingOrder := createTestOrder("order-123", common.FulfillmentStatusPaid, common.PaymentStatusPending)

		mockOrderRepo := &mockOrderRepo{
			findByIDFn: func(id common.OrderID) (*order.Order, error) {
				return existingOrder, nil
			},
		}

		uc := kitchenCmd.NewMarkOrderPreparingUseCase(mockOrderRepo, &mockEventBus{})
		_, err := uc.Execute(context.Background(), orderID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot prepare unpaid order")
	})

	t.Run("fails if status transition is invalid", func(t *testing.T) {
		orderID := common.OrderID("order-123")
		// Already Ready, cannot go back to Preparing (unless allowed, but standard flow usually forward)
		// domain.go: FulfillmentStatusPreparing: {FulfillmentStatusReady}
		// domain.go: FulfillmentStatusReady: {FulfillmentStatusCompleted}
		// So Ready -> Preparing is NOT allowed in `isValidStatusTransition` map.
		existingOrder := createTestOrder("order-123", common.FulfillmentStatusReady, common.PaymentStatusPaid)

		mockOrderRepo := &mockOrderRepo{
			findByIDFn: func(id common.OrderID) (*order.Order, error) {
				return existingOrder, nil
			},
		}

		uc := kitchenCmd.NewMarkOrderPreparingUseCase(mockOrderRepo, &mockEventBus{})
		_, err := uc.Execute(context.Background(), orderID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status transition")
	})
}
