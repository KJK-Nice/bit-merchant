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

func TestGetDashboardStatsUseCase(t *testing.T) {
	orderRepo := memory.NewMemoryOrderRepository()
	uc := dashboard.NewGetDashboardStatsUseCase(orderRepo)
	restaurantID := domain.RestaurantID("r1")

	// Seed orders
	items := []domain.OrderItem{{MenuItemID: "i1", Name: "Item 1", Quantity: 1, UnitPrice: 10.0, Subtotal: 10.0}}

	// Order 1: Today, Paid, $10
	o1, _ := domain.NewOrder("o1", "1001", restaurantID, "session_1", items, 1000, domain.PaymentMethodTypeCash)
	o1.FiatAmount = 10.0
	o1.PaymentStatus = domain.PaymentStatusPaid
	o1.CreatedAt = time.Now()
	_ = orderRepo.Save(o1)

	// Order 2: Today, Paid, $20
	o2, _ := domain.NewOrder("o2", "1002", restaurantID, "session_1", items, 2000, domain.PaymentMethodTypeCash)
	o2.FiatAmount = 20.0
	o2.PaymentStatus = domain.PaymentStatusPaid
	o2.CreatedAt = time.Now()
	_ = orderRepo.Save(o2)

	// Order 3: Yesterday, Paid, $50 (Should be excluded from "today" stats unless range specified)
	o3, _ := domain.NewOrder("o3", "1003", restaurantID, "session_1", items, 5000, domain.PaymentMethodTypeCash)
	o3.FiatAmount = 50.0
	o3.PaymentStatus = domain.PaymentStatusPaid
	o3.CreatedAt = time.Now().AddDate(0, 0, -1)
	_ = orderRepo.Save(o3)

	// Order 4: Today, Pending (Should be excluded from sales stats?) -> Usually sales stats count confirmed/paid orders
	o4, _ := domain.NewOrder("o4", "1004", restaurantID, "session_1", items, 500, domain.PaymentMethodTypeCash)
	o4.FiatAmount = 5.0
	o4.PaymentStatus = domain.PaymentStatusPending
	o4.CreatedAt = time.Now()
	_ = orderRepo.Save(o4)

	t.Run("Get Today's Stats", func(t *testing.T) {
		stats, err := uc.Execute(context.Background(), restaurantID, dashboard.DateRangeToday)
		assert.NoError(t, err)
		
		// Should include o1, o2. o3 is yesterday. o4 is pending (assume we only count paid for sales).
		// Note: Requirement says "view daily sales". Usually means "sales confirmed".
		
		assert.Equal(t, 2, stats.OrderCount)
		assert.Equal(t, 30.0, stats.TotalSales)
		assert.Equal(t, 15.0, stats.AverageOrderValue)
	})
}

