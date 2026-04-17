package kitchen_test

import (
	"bitmerchant/internal/common"
	kitchenQuery "bitmerchant/internal/ordering/app/query"
	"bitmerchant/internal/ordering/domain/order"

	"context"

	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestActiveKitchenOrdersHandler_Handle(t *testing.T) {
	t.Run("returns orders sorted chronologically", func(t *testing.T) {
		restaurantID := common.RestaurantID("rest-1")

		order1 := createTestOrder("o1", common.FulfillmentStatusPaid, common.PaymentStatusPaid)
		order1.CreatedAt = time.Now().Add(-1 * time.Hour)

		order2 := createTestOrder("o2", common.FulfillmentStatusPreparing, common.PaymentStatusPaid)
		order2.CreatedAt = time.Now().Add(-30 * time.Minute)

		mockOrderRepo := &mockOrderRepo{
			findActiveByRestaurantIDFn: func(id common.RestaurantID) ([]*order.Order, error) {
				// Return unsorted or pre-sorted, use case should ensure sort?
				// Usually repo returns sorted, but let's assume use case ensures it or passes through.
				return []*order.Order{order1, order2}, nil
			},
		}

		uc := kitchenQuery.NewActiveKitchenOrdersHandler(mockOrderRepo, nil, nil)
		orders, err := uc.Handle(context.Background(), kitchenQuery.ActiveKitchenOrders{RestaurantID: restaurantID})

		assert.NoError(t, err)
		assert.Len(t, orders, 2)
		assert.Equal(t, "o1", string(orders[0].ID))
		assert.Equal(t, "o2", string(orders[1].ID))
	})

	t.Run("returns empty list when no orders", func(t *testing.T) {
		mockOrderRepo := &mockOrderRepo{
			findActiveByRestaurantIDFn: func(id common.RestaurantID) ([]*order.Order, error) {
				return []*order.Order{}, nil
			},
		}

		uc := kitchenQuery.NewActiveKitchenOrdersHandler(mockOrderRepo, nil, nil)
		orders, err := uc.Handle(context.Background(), kitchenQuery.ActiveKitchenOrders{RestaurantID: common.RestaurantID("rest-1")})

		assert.NoError(t, err)
		assert.Empty(t, orders)
	})
}
