package memory_test

import (
	"testing"

	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
)

func TestMemoryOrderRepository(t *testing.T) {
	repo := memory.NewMemoryOrderRepository()

	t.Run("Save and FindByID", func(t *testing.T) {
		item, _ := domain.NewOrderItem("oi1", "o1", "mi1", "Burger", 1, 10.0)
		order, _ := domain.NewOrder("o1", "101", "r1", "session_1", []domain.OrderItem{*item}, 100, domain.PaymentMethodTypeCash)

		err := repo.Save(order)
		assert.NoError(t, err)

		found, err := repo.FindByID("o1")
		assert.NoError(t, err)
		assert.Equal(t, order.ID, found.ID)
	})

	t.Run("FindByOrderNumber", func(t *testing.T) {
		item, _ := domain.NewOrderItem("oi2", "o2", "mi1", "Burger", 1, 10.0)
		order, _ := domain.NewOrder("o2", "102", "r1", "session_1", []domain.OrderItem{*item}, 100, domain.PaymentMethodTypeCash)
		repo.Save(order)

		found, err := repo.FindByOrderNumber("r1", "102")
		assert.NoError(t, err)
		assert.Equal(t, order.ID, found.ID)

		_, err = repo.FindByOrderNumber("r1", "999")
		assert.Error(t, err)
	})

	t.Run("FindActiveByRestaurantID", func(t *testing.T) {
		// Active
		item1, _ := domain.NewOrderItem("oi3", "o3", "mi1", "Burger", 1, 10.0)
		order1, _ := domain.NewOrder("o3", "103", "r2", "session_1", []domain.OrderItem{*item1}, 100, domain.PaymentMethodTypeCash)

		// Completed (not active)
		item2, _ := domain.NewOrderItem("oi4", "o4", "mi1", "Burger", 1, 10.0)
		order2, _ := domain.NewOrder("o4", "104", "r2", "session_1", []domain.OrderItem{*item2}, 100, domain.PaymentMethodTypeCash)
		order2.UpdateFulfillmentStatus(domain.FulfillmentStatusPreparing)
		order2.UpdateFulfillmentStatus(domain.FulfillmentStatusReady)
		order2.UpdateFulfillmentStatus(domain.FulfillmentStatusCompleted)

		repo.Save(order1)
		repo.Save(order2)

		active, err := repo.FindActiveByRestaurantID("r2")
		assert.NoError(t, err)
		assert.Len(t, active, 1)
		assert.Equal(t, "o3", string(active[0].ID))
	})
}
