package dashboard_test

import (
	"context"
	"testing"
	"time"

	"bitmerchant/internal/application/dashboard"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
)

func TestGetOrderHistoryUseCase(t *testing.T) {
	orderRepo := memory.NewMemoryOrderRepository()
	uc := dashboard.NewGetOrderHistoryUseCase(orderRepo)
	restaurantID := domain.RestaurantID("r1")

	// Seed orders
	items := []domain.OrderItem{{MenuItemID: "i1", Name: "Item 1", Quantity: 1, UnitPrice: 10.0, Subtotal: 10.0}}
	
	o1, _ := domain.NewOrder("o1", "1001", restaurantID, "session_1", items, 1000, domain.PaymentMethodTypeCash)
	o1.CreatedAt = time.Now().Add(-1 * time.Hour)
	_ = orderRepo.Save(o1)

	o2, _ := domain.NewOrder("o2", "1002", restaurantID, "session_1", items, 2000, domain.PaymentMethodTypeCash)
	o2.CreatedAt = time.Now()
	_ = orderRepo.Save(o2)

	// Order for another restaurant
	o3, _ := domain.NewOrder("o3", "1003", "r2", "session_1", items, 3000, domain.PaymentMethodTypeCash)
	_ = orderRepo.Save(o3)

	t.Run("Get All Orders", func(t *testing.T) {
		orders, err := uc.Execute(context.Background(), restaurantID)
		assert.NoError(t, err)
		assert.Len(t, orders, 2)
		
		// Should be sorted by date desc (newest first)
		assert.Equal(t, "o2", string(orders[0].ID))
		assert.Equal(t, "o1", string(orders[1].ID))
	})
}

