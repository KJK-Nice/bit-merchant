package dashboard_test

import (
	"context"
	"testing"

	"bitmerchant/internal/application/dashboard"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
)

func TestGetTopSellingItemsUseCase(t *testing.T) {
	orderRepo := memory.NewMemoryOrderRepository()
	uc := dashboard.NewGetTopSellingItemsUseCase(orderRepo)
	restaurantID := domain.RestaurantID("r1")

	// Order 1: 2 Burgers, 1 Soda
	items1 := []domain.OrderItem{
		{MenuItemID: "burger", Name: "Burger", Quantity: 2, UnitPrice: 10.0, Subtotal: 20.0},
		{MenuItemID: "soda", Name: "Soda", Quantity: 1, UnitPrice: 3.0, Subtotal: 3.0},
	}
	o1, _ := domain.NewOrder("o1", "1001", restaurantID, "session_1", items1, 2300, domain.PaymentMethodTypeCash)
	o1.PaymentStatus = domain.PaymentStatusPaid
	_ = orderRepo.Save(o1)

	// Order 2: 1 Burger
	items2 := []domain.OrderItem{
		{MenuItemID: "burger", Name: "Burger", Quantity: 1, UnitPrice: 10.0, Subtotal: 10.0},
	}
	o2, _ := domain.NewOrder("o2", "1002", restaurantID, "session_1", items2, 1000, domain.PaymentMethodTypeCash)
	o2.PaymentStatus = domain.PaymentStatusPaid
	_ = orderRepo.Save(o2)

	// Order 3 (Unpaid): 10 Steaks (Should be excluded)
	items3 := []domain.OrderItem{
		{MenuItemID: "steak", Name: "Steak", Quantity: 10, UnitPrice: 50.0, Subtotal: 500.0},
	}
	o3, _ := domain.NewOrder("o3", "1003", restaurantID, "session_1", items3, 50000, domain.PaymentMethodTypeCash)
	o3.PaymentStatus = domain.PaymentStatusPending
	_ = orderRepo.Save(o3)

	t.Run("Get Top Items", func(t *testing.T) {
		items, err := uc.Execute(context.Background(), restaurantID)
		assert.NoError(t, err)
		
		// Expected: Burger (3), Soda (1)
		assert.Len(t, items, 2)
		
		assert.Equal(t, "Burger", items[0].Name)
		assert.Equal(t, 3, items[0].Quantity)
		assert.Equal(t, 30.0, items[0].Revenue)

		assert.Equal(t, "Soda", items[1].Name)
		assert.Equal(t, 1, items[1].Quantity)
		assert.Equal(t, 3.0, items[1].Revenue)
	})
}

