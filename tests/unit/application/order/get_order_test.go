package order_test

import (
	"bitmerchant/internal/common"

	"bitmerchant/internal/infrastructure/repositories/memory"
	orderQuery "bitmerchant/internal/ordering/app/query"
	"bitmerchant/internal/ordering/domain/order"
	"context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOrderByNumberForRestaurantHandler(t *testing.T) {
	repo := memory.NewMemoryOrderRepository()
	uc := orderQuery.NewOrderByNumberForRestaurantHandler(repo, nil, nil)

	t.Run("Handle Success", func(t *testing.T) {
		// Setup order
		item, _ := order.NewOrderItem("oi1", "o1", "mi1", "Burger", 1, 10.0)
		existingOrder, _ := order.NewOrder(
			"o1",
			"1234",
			"r1",
			"session_1",
			[]order.OrderItem{*item},
			1000,
			common.PaymentMethodTypeCash,
		)
		require.NoError(t, repo.Save(existingOrder))

		result, err := uc.Handle(context.Background(), orderQuery.OrderByNumberForRestaurant{
			RestaurantID: common.RestaurantID("r1"),
			OrderNumber:  "1234",
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, existingOrder.ID, result.ID)
	})

	t.Run("Handle Not Found", func(t *testing.T) {
		_, err := uc.Handle(context.Background(), orderQuery.OrderByNumberForRestaurant{
			RestaurantID: common.RestaurantID("r1"),
			OrderNumber:  "9999",
		})
		assert.Error(t, err)
	})
}
