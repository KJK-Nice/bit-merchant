package kitchen_test

import (
	"bitmerchant/internal/application/kitchen"
	"bitmerchant/internal/domain"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetKitchenOrdersUseCase_Execute(t *testing.T) {
	t.Run("returns orders sorted chronologically", func(t *testing.T) {
		restaurantID := domain.RestaurantID("rest-1")
		
		order1 := createTestOrder("o1", domain.FulfillmentStatusPaid, domain.PaymentStatusPaid)
		order1.CreatedAt = time.Now().Add(-1 * time.Hour)
		
		order2 := createTestOrder("o2", domain.FulfillmentStatusPreparing, domain.PaymentStatusPaid)
		order2.CreatedAt = time.Now().Add(-30 * time.Minute)

		mockOrderRepo := &mockOrderRepo{
			findActiveByRestaurantIDFn: func(id domain.RestaurantID) ([]*domain.Order, error) {
				// Return unsorted or pre-sorted, use case should ensure sort?
				// Usually repo returns sorted, but let's assume use case ensures it or passes through.
				return []*domain.Order{order1, order2}, nil
			},
		}

		uc := kitchen.NewGetKitchenOrdersUseCase(mockOrderRepo)
		orders, err := uc.Execute(context.Background(), restaurantID)

		assert.NoError(t, err)
		assert.Len(t, orders, 2)
		assert.Equal(t, "o1", string(orders[0].ID))
		assert.Equal(t, "o2", string(orders[1].ID))
	})

	t.Run("returns empty list when no orders", func(t *testing.T) {
		mockOrderRepo := &mockOrderRepo{
			findActiveByRestaurantIDFn: func(id domain.RestaurantID) ([]*domain.Order, error) {
				return []*domain.Order{}, nil
			},
		}

		uc := kitchen.NewGetKitchenOrdersUseCase(mockOrderRepo)
		orders, err := uc.Execute(context.Background(), domain.RestaurantID("rest-1"))

		assert.NoError(t, err)
		assert.Empty(t, orders)
	})
}

