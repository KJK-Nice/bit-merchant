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

func TestRestaurantDashboardStatsHandler(t *testing.T) {
	orderRepo := memory.NewMemoryOrderRepository()
	h := dashboard.NewRestaurantDashboardStatsHandler(orderRepo, nil, nil)
	restaurantID := common.RestaurantID("r1")
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	prevMonth := startOfMonth.AddDate(0, -1, 0).Add(12 * time.Hour)

	// Seed orders
	items := []order.OrderItem{{MenuItemID: "i1", Name: "Item 1", Quantity: 1, UnitPrice: 10.0, Subtotal: 10.0}}

	// Order 1: Today, Paid, $10
	o1, _ := order.NewOrder("o1", "1001", restaurantID, "session_1", items, 1000, common.PaymentMethodTypeCash)
	o1.FiatAmount = 10.0
	o1.PaymentStatus = common.PaymentStatusPaid
	o1.CreatedAt = startOfToday.Add(1 * time.Hour)
	_ = orderRepo.Save(o1)

	// Order 2: 2 days ago, Paid, $20 (in week and month)
	o2, _ := order.NewOrder("o2", "1002", restaurantID, "session_1", items, 2000, common.PaymentMethodTypeCash)
	o2.FiatAmount = 20.0
	o2.PaymentStatus = common.PaymentStatusPaid
	o2.CreatedAt = startOfToday.AddDate(0, 0, -2).Add(2 * time.Hour)
	_ = orderRepo.Save(o2)

	// Order 3: 10 days ago, Paid, $50 (excluded from week, included in month only if still same month)
	o3, _ := order.NewOrder("o3", "1003", restaurantID, "session_1", items, 5000, common.PaymentMethodTypeCash)
	o3.FiatAmount = 50.0
	o3.PaymentStatus = common.PaymentStatusPaid
	o3.CreatedAt = startOfToday.AddDate(0, 0, -10).Add(3 * time.Hour)
	_ = orderRepo.Save(o3)

	// Order 4: Today, Pending (always excluded)
	o4, _ := order.NewOrder("o4", "1004", restaurantID, "session_1", items, 500, common.PaymentMethodTypeCash)
	o4.FiatAmount = 5.0
	o4.PaymentStatus = common.PaymentStatusPending
	o4.CreatedAt = startOfToday.Add(4 * time.Hour)
	_ = orderRepo.Save(o4)

	// Order 5: Previous month, Paid, $80 (excluded from month and week/today)
	o5, _ := order.NewOrder("o5", "1005", restaurantID, "session_1", items, 8000, common.PaymentMethodTypeCash)
	o5.FiatAmount = 80.0
	o5.PaymentStatus = common.PaymentStatusPaid
	o5.CreatedAt = prevMonth
	_ = orderRepo.Save(o5)

	t.Run("Get Today's Stats", func(t *testing.T) {
		stats, err := h.Handle(context.Background(), dashboard.RestaurantDashboardStats{
			RestaurantID: restaurantID,
			Range:        dashboard.DateRangeToday,
		})
		assert.NoError(t, err)

		assert.Equal(t, 1, stats.OrderCount)
		assert.Equal(t, 10.0, stats.TotalSales)
		assert.Equal(t, 10.0, stats.AverageOrderValue)
	})

	t.Run("Get Weekly Stats", func(t *testing.T) {
		stats, err := h.Handle(context.Background(), dashboard.RestaurantDashboardStats{
			RestaurantID: restaurantID,
			Range:        dashboard.DateRangeWeek,
		})
		assert.NoError(t, err)

		assert.Equal(t, 2, stats.OrderCount)
		assert.Equal(t, 30.0, stats.TotalSales)
		assert.Equal(t, 15.0, stats.AverageOrderValue)
	})

	t.Run("Get Monthly Stats", func(t *testing.T) {
		stats, err := h.Handle(context.Background(), dashboard.RestaurantDashboardStats{
			RestaurantID: restaurantID,
			Range:        dashboard.DateRangeMonth,
		})
		assert.NoError(t, err)

		expectedCount := 2 // o1 + o2
		expectedSales := 30.0
		if !o3.CreatedAt.Before(startOfMonth) {
			expectedCount++
			expectedSales += 50.0
		}

		assert.Equal(t, expectedCount, stats.OrderCount)
		assert.Equal(t, expectedSales, stats.TotalSales)
		assert.Equal(t, expectedSales/float64(expectedCount), stats.AverageOrderValue)
	})

	t.Run("Invalid Range Defaults To Today", func(t *testing.T) {
		stats, err := h.Handle(context.Background(), dashboard.RestaurantDashboardStats{
			RestaurantID: restaurantID,
			Range:        dashboard.DateRange("invalid"),
		})
		assert.NoError(t, err)

		assert.Equal(t, 1, stats.OrderCount)
		assert.Equal(t, 10.0, stats.TotalSales)
		assert.Equal(t, 10.0, stats.AverageOrderValue)
	})
}
