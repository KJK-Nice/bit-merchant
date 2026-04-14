package kitchen_test

import (
	"bitmerchant/internal/common"
	kitchenCmd "bitmerchant/internal/ordering/app/command"
	"bitmerchant/internal/ordering/domain/order"

	"context"
	"errors"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMarkOrderPaidUseCase_Execute(t *testing.T) {
	t.Run("successfully marks order as paid", func(t *testing.T) {
		orderID := common.OrderID("order-123")
		existingOrder := createTestOrder("order-123", common.FulfillmentStatusPaid, common.PaymentStatusPending)

		var savedOrder *order.Order
		var publishedEvent string

		mockOrderRepo := &mockOrderRepo{
			findByIDFn: func(id common.OrderID) (*order.Order, error) {
				if id == orderID {
					return existingOrder, nil
				}
				return nil, nil
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

		uc := kitchenCmd.NewMarkOrderPaidUseCase(mockOrderRepo, mockEventBus)
		_, err := uc.Execute(context.Background(), orderID)

		assert.NoError(t, err)
		assert.NotNil(t, savedOrder)
		assert.Equal(t, common.PaymentStatusPaid, savedOrder.PaymentStatus)
		assert.NotNil(t, savedOrder.PaidAt)
		assert.Equal(t, "OrderPaid", publishedEvent)
	})

	t.Run("returns error when order not found", func(t *testing.T) {
		mockOrderRepo := &mockOrderRepo{
			findByIDFn: func(id common.OrderID) (*order.Order, error) {
				return nil, nil
			},
		}

		uc := kitchenCmd.NewMarkOrderPaidUseCase(mockOrderRepo, &mockEventBus{})
		_, err := uc.Execute(context.Background(), common.OrderID("non-existent"))

		assert.Error(t, err)
		assert.Equal(t, "order not found", err.Error())
	})

	t.Run("returns error when update fails", func(t *testing.T) {
		orderID := common.OrderID("order-123")
		existingOrder := createTestOrder("order-123", common.FulfillmentStatusPaid, common.PaymentStatusPending)

		mockOrderRepo := &mockOrderRepo{
			findByIDFn: func(id common.OrderID) (*order.Order, error) {
				return existingOrder, nil
			},
			updateFn: func(order *order.Order) error {
				return errors.New("db error")
			},
		}

		uc := kitchenCmd.NewMarkOrderPaidUseCase(mockOrderRepo, &mockEventBus{})
		_, err := uc.Execute(context.Background(), orderID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}
