package dashboard_test

import (
	"bitmerchant/internal/common"
	dashboard "bitmerchant/internal/dashboard/app/query"

	"bitmerchant/internal/infrastructure/repositories/memory"
	"bitmerchant/internal/ordering/domain/order"
	"context"

	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetOrderHistoryUseCase(t *testing.T) {
	orderRepo := memory.NewMemoryOrderRepository()
	uc := dashboard.NewGetOrderHistoryUseCase(orderRepo)
	restaurantID := common.RestaurantID("r1")

	// Seed orders
	items := []order.OrderItem{{MenuItemID: "i1", Name: "Item 1", Quantity: 1, UnitPrice: 10.0, Subtotal: 10.0}}

	o1, _ := order.NewOrder("o1", "1001", restaurantID, "session_1", items, 1000, common.PaymentMethodTypeCash)
	o1.PaymentStatus = common.PaymentStatusPaid
	o1.CreatedAt = time.Now().Add(-1 * time.Hour)
	_ = orderRepo.Save(o1)

	o2, _ := order.NewOrder("o2", "1002", restaurantID, "session_1", items, 2000, common.PaymentMethodTypeCash)
	o2.PaymentStatus = common.PaymentStatusPending
	o2.CreatedAt = time.Now()
	_ = orderRepo.Save(o2)

	o4, _ := order.NewOrder("o4", "1004", restaurantID, "session_1", items, 2500, common.PaymentMethodTypeCash)
	o4.PaymentStatus = common.PaymentStatusPaid
	o4.CreatedAt = time.Now().Add(10 * time.Minute)
	_ = orderRepo.Save(o4)

	// Order for another restaurant
	o3, _ := order.NewOrder("o3", "1003", "r2", "session_1", items, 3000, common.PaymentMethodTypeCash)
	_ = orderRepo.Save(o3)

	t.Run("Get Paid Orders Sorted By Newest First", func(t *testing.T) {
		orders, err := uc.Execute(context.Background(), restaurantID)
		assert.NoError(t, err)
		assert.Len(t, orders, 2)

		// Should be sorted by date desc (newest first)
		assert.Equal(t, "o4", string(orders[0].ID))
		assert.Equal(t, "o1", string(orders[1].ID))
	})
}
