package kitchen_test

import (
	"bitmerchant/internal/application/kitchen"
	"bitmerchant/internal/domain"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkOrderReadyUseCase_Execute(t *testing.T) {
	t.Run("successfully marks order as ready", func(t *testing.T) {
		orderID := domain.OrderID("order-123")
		existingOrder := createTestOrder("order-123", domain.FulfillmentStatusPreparing, domain.PaymentStatusPaid)
		
		var savedOrder *domain.Order
		var publishedEvent string

		mockOrderRepo := &mockOrderRepo{
			findByIDFn: func(id domain.OrderID) (*domain.Order, error) {
				return existingOrder, nil
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

		uc := kitchen.NewMarkOrderReadyUseCase(mockOrderRepo, mockEventBus)
		_, err := uc.Execute(context.Background(), orderID)

		assert.NoError(t, err)
		assert.Equal(t, domain.FulfillmentStatusReady, savedOrder.FulfillmentStatus)
		assert.NotNil(t, savedOrder.ReadyAt)
		assert.Equal(t, "OrderReady", publishedEvent)
	})

	t.Run("fails if status transition is invalid", func(t *testing.T) {
		orderID := domain.OrderID("order-123")
		// Paid -> Ready is invalid, must be Preparing first
		existingOrder := createTestOrder("order-123", domain.FulfillmentStatusPaid, domain.PaymentStatusPaid)

		mockOrderRepo := &mockOrderRepo{
			findByIDFn: func(id domain.OrderID) (*domain.Order, error) {
				return existingOrder, nil
			},
		}

		uc := kitchen.NewMarkOrderReadyUseCase(mockOrderRepo, &mockEventBus{})
		_, err := uc.Execute(context.Background(), orderID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status transition")
	})
}

