package dashboard_test

import (
	"context"
	"testing"
	"time"

	"bitmerchant/internal/application/cart"
	"bitmerchant/internal/application/dashboard"
	"bitmerchant/internal/application/order"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/infrastructure/payment/cash"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
)

func TestDashboardIntegration(t *testing.T) {
	// Infrastructure
	orderRepo := memory.NewMemoryOrderRepository()
	paymentRepo := memory.NewMemoryPaymentRepository()
	restRepo := memory.NewMemoryRestaurantRepository()
	eventBus := events.NewEventBus()
	paymentMethod := cash.NewCashPaymentMethod()
	logger := logging.NewLogger()

	// Setup restaurant for order creation check
	restaurant, _ := domain.NewRestaurant("restaurant_1", "Integration Cafe")
	restRepo.Save(restaurant)

	// Use Cases
	createOrderUC := order.NewCreateOrderUseCase(orderRepo, paymentRepo, restRepo, eventBus, paymentMethod, logger)
	getStatsUC := dashboard.NewGetDashboardStatsUseCase(orderRepo)

	t.Run("Order Creation Reflected in Stats", func(t *testing.T) {
		// 1. Create an Order
		cartSvc := cart.NewCartService()
		sessionID := "sess_integration"
		item, _ := domain.NewMenuItem("i1", "c1", "r1", "Burger", 15.0)
		cartSvc.AddItem(sessionID, item, 2) // 2 Burgers = $30
		userCart := cartSvc.GetCart(sessionID)

		req := order.CreateOrderRequest{
			RestaurantID:  "restaurant_1", // Must match dashboard default
			SessionID:     sessionID,
			Cart:          userCart,
			PaymentMethod: domain.PaymentMethodTypeCash,
		}

		resp, err := createOrderUC.Execute(context.Background(), req)
		assert.NoError(t, err)

		// 2. Mark Order as Paid (since stats only count paid orders)
		// We can use repo directly or kitchen use case. Using repo for simplicity here.
		savedOrder, _ := orderRepo.FindByID(resp.OrderID)
		savedOrder.PaymentStatus = domain.PaymentStatusPaid
		savedOrder.CreatedAt = time.Now() // Ensure it's today
		_ = orderRepo.Save(savedOrder)

		// 3. Check Stats
		stats, err := getStatsUC.Execute(context.Background(), "restaurant_1", dashboard.DateRangeToday)
		assert.NoError(t, err)

		assert.Equal(t, 1, stats.OrderCount)
		assert.Equal(t, 30.0, stats.TotalSales)
		assert.Equal(t, 30.0, stats.AverageOrderValue)
	})
}

