package memory_test

import (
	"bitmerchant/internal/common"

	"bitmerchant/internal/infrastructure/repositories/memory"
	"bitmerchant/internal/ordering/domain/order"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMemoryOrderRepository(t *testing.T) {
	repo := memory.NewMemoryOrderRepository()

	t.Run("Save and FindByID", func(t *testing.T) {
		item, _ := order.NewOrderItem("oi1", "o1", "mi1", "Burger", 1, 10.0)
		order, _ := order.NewOrder("o1", "101", "r1", "session_1", []order.OrderItem{*item}, 100, common.PaymentMethodTypeCash)

		err := repo.Save(order)
		assert.NoError(t, err)

		found, err := repo.FindByID("o1")
		assert.NoError(t, err)
		assert.Equal(t, order.ID, found.ID)
	})

	t.Run("FindByOrderNumber", func(t *testing.T) {
		item, _ := order.NewOrderItem("oi2", "o2", "mi1", "Burger", 1, 10.0)
		order, _ := order.NewOrder("o2", "102", "r1", "session_1", []order.OrderItem{*item}, 100, common.PaymentMethodTypeCash)
		require.NoError(t, repo.Save(order))

		found, err := repo.FindByOrderNumber("r1", "102")
		assert.NoError(t, err)
		assert.Equal(t, order.ID, found.ID)

		_, err = repo.FindByOrderNumber("r1", "999")
		assert.Error(t, err)
	})

	t.Run("FindActiveByRestaurantID", func(t *testing.T) {
		// Active
		item1, _ := order.NewOrderItem("oi3", "o3", "mi1", "Burger", 1, 10.0)
		order1, _ := order.NewOrder("o3", "103", "r2", "session_1", []order.OrderItem{*item1}, 100, common.PaymentMethodTypeCash)

		// Completed (not active)
		item2, _ := order.NewOrderItem("oi4", "o4", "mi1", "Burger", 1, 10.0)
		order2, _ := order.NewOrder("o4", "104", "r2", "session_1", []order.OrderItem{*item2}, 100, common.PaymentMethodTypeCash)
		require.NoError(t, order2.UpdateFulfillmentStatus(common.FulfillmentStatusPreparing))
		require.NoError(t, order2.UpdateFulfillmentStatus(common.FulfillmentStatusReady))
		require.NoError(t, order2.UpdateFulfillmentStatus(common.FulfillmentStatusCompleted))

		require.NoError(t, repo.Save(order1))
		require.NoError(t, repo.Save(order2))

		active, err := repo.FindActiveByRestaurantID("r2")
		assert.NoError(t, err)
		assert.Len(t, active, 1)
		assert.Equal(t, "o3", string(active[0].ID))
	})
}
