package kitchen_test

import (
	"bitmerchant/internal/application/kitchen"
	"bitmerchant/internal/domain"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkOrderPaidUseCase_Execute(t *testing.T) {
	t.Run("successfully marks order as paid", func(t *testing.T) {
		orderID := domain.OrderID("order-123")
		existingOrder := createTestOrder("order-123", domain.FulfillmentStatusPaid, domain.PaymentStatusPending)
		
		var savedOrder *domain.Order
		var publishedEvent string

		mockOrderRepo := &mockOrderRepo{
			findByIDFn: func(id domain.OrderID) (*domain.Order, error) {
				if id == orderID {
					return existingOrder, nil
				}
				return nil, nil
			},
			updateFn: func(order *domain.Order) error {
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

		uc := kitchen.NewMarkOrderPaidUseCase(mockOrderRepo, mockEventBus)
		_, err := uc.Execute(context.Background(), orderID)

		assert.NoError(t, err)
		assert.NotNil(t, savedOrder)
		assert.Equal(t, domain.PaymentStatusPaid, savedOrder.PaymentStatus)
		assert.NotNil(t, savedOrder.PaidAt)
		assert.Equal(t, "OrderPaid", publishedEvent)
	})

	t.Run("returns error when order not found", func(t *testing.T) {
		mockOrderRepo := &mockOrderRepo{
			findByIDFn: func(id domain.OrderID) (*domain.Order, error) {
				return nil, nil
			},
		}
		
		uc := kitchen.NewMarkOrderPaidUseCase(mockOrderRepo, &mockEventBus{})
		_, err := uc.Execute(context.Background(), domain.OrderID("non-existent"))

		assert.Error(t, err)
		assert.Equal(t, "order not found", err.Error())
	})

	t.Run("returns error when update fails", func(t *testing.T) {
		orderID := domain.OrderID("order-123")
		existingOrder := createTestOrder("order-123", domain.FulfillmentStatusPaid, domain.PaymentStatusPending)

		mockOrderRepo := &mockOrderRepo{
			findByIDFn: func(id domain.OrderID) (*domain.Order, error) {
				return existingOrder, nil
			},
			updateFn: func(order *domain.Order) error {
				return errors.New("db error")
			},
		}

		uc := kitchen.NewMarkOrderPaidUseCase(mockOrderRepo, &mockEventBus{})
		_, err := uc.Execute(context.Background(), orderID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}
