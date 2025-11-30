package order_test

import (
	"context"
	"testing"

	"bitmerchant/internal/application/cart"
	"bitmerchant/internal/application/order"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/infrastructure/payment/cash"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
)

func TestCreateOrderUseCase(t *testing.T) {
	orderRepo := memory.NewMemoryOrderRepository()
	paymentRepo := memory.NewMemoryPaymentRepository()
	restRepo := memory.NewMemoryRestaurantRepository()
	eventBus := events.NewEventBus()
	paymentMethod := cash.NewCashPaymentMethod()
	logger := logging.NewLogger()

	// Setup restaurant
	restID := domain.RestaurantID("r1")
	restaurant, _ := domain.NewRestaurant(restID, "Test Rest")
	restRepo.Save(restaurant)
	
	uc := order.NewCreateOrderUseCase(
		orderRepo,
		paymentRepo,
		restRepo,
		eventBus,
		paymentMethod,
		logger,
	)

	t.Run("Execute Success", func(t *testing.T) {
		// Setup Cart
		cartSvc := cart.NewCartService()
		sessionID := "sess_1"
		item, _ := domain.NewMenuItem("i1", "c1", "r1", "Burger", 10.0)
		cartSvc.AddItem(sessionID, item, 2)
		
		userCart := cartSvc.GetCart(sessionID)
		
		req := order.CreateOrderRequest{
			RestaurantID: "r1",
			SessionID:    sessionID,
			Cart:         userCart,
			PaymentMethod: domain.PaymentMethodTypeCash,
		}

		resp, err := uc.Execute(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.OrderID)
		assert.NotEmpty(t, resp.OrderNumber)
		
		// Verify Order Saved
		savedOrder, _ := orderRepo.FindByID(resp.OrderID)
		assert.Equal(t, domain.PaymentStatusPending, savedOrder.PaymentStatus)
		assert.Equal(t, int64(2000), savedOrder.TotalAmount) // 20.0 * 100
		assert.Equal(t, 20.0, savedOrder.FiatAmount)
	})
}

