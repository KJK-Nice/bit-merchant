package kitchen_test

import (
	"bitmerchant/internal/application/kitchen"
	"bitmerchant/internal/domain"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkOrderPreparingUseCase_Execute(t *testing.T) {
	t.Run("successfully marks order as preparing", func(t *testing.T) {
		orderID := domain.OrderID("order-123")
		// Must be Paid to start preparing
		existingOrder := createTestOrder("order-123", domain.FulfillmentStatusPaid, domain.PaymentStatusPaid)
		
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

		uc := kitchen.NewMarkOrderPreparingUseCase(mockOrderRepo, mockEventBus)
		_, err := uc.Execute(context.Background(), orderID)

		assert.NoError(t, err)
		assert.Equal(t, domain.FulfillmentStatusPreparing, savedOrder.FulfillmentStatus)
		assert.NotNil(t, savedOrder.PreparingAt)
		assert.Equal(t, "OrderPreparing", publishedEvent)
	})

	t.Run("fails if order is not paid", func(t *testing.T) {
		orderID := domain.OrderID("order-123")
		existingOrder := createTestOrder("order-123", domain.FulfillmentStatusPaid, domain.PaymentStatusPending)

		mockOrderRepo := &mockOrderRepo{
			findByIDFn: func(id domain.OrderID) (*domain.Order, error) {
				return existingOrder, nil
			},
		}

		uc := kitchen.NewMarkOrderPreparingUseCase(mockOrderRepo, &mockEventBus{})
		_, err := uc.Execute(context.Background(), orderID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot prepare unpaid order")
	})

	t.Run("fails if status transition is invalid", func(t *testing.T) {
		orderID := domain.OrderID("order-123")
		// Already Ready, cannot go back to Preparing (unless allowed, but standard flow usually forward)
		// domain.go: FulfillmentStatusPreparing: {FulfillmentStatusReady}
		// domain.go: FulfillmentStatusReady: {FulfillmentStatusCompleted}
		// So Ready -> Preparing is NOT allowed in `isValidStatusTransition` map.
		existingOrder := createTestOrder("order-123", domain.FulfillmentStatusReady, domain.PaymentStatusPaid)

		mockOrderRepo := &mockOrderRepo{
			findByIDFn: func(id domain.OrderID) (*domain.Order, error) {
				return existingOrder, nil
			},
		}

		uc := kitchen.NewMarkOrderPreparingUseCase(mockOrderRepo, &mockEventBus{})
		_, err := uc.Execute(context.Background(), orderID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status transition")
	})
}

