package order_test

import (
	"bitmerchant/internal/common"

	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/infrastructure/payment/cash"
	"bitmerchant/internal/infrastructure/repositories/memory"
	"bitmerchant/internal/menu/domain/menu"
	"bitmerchant/internal/ordering/app/cart"
	orderCmd "bitmerchant/internal/ordering/app/command"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateOrderUseCase(t *testing.T) {
	orderRepo := memory.NewMemoryOrderRepository()
	paymentRepo := memory.NewMemoryPaymentRepository()
	restRepo := memory.NewMemoryRestaurantRepository()
	eventBus := events.NewEventBus()
	paymentMethod := cash.NewCashPaymentMethod()
	logger := logging.NewLogger()

	// Setup restaurant
	restID := common.RestaurantID("r1")
	restaurant, _ := restaurant.NewRestaurant(restID, "Test Rest")
	require.NoError(t, restRepo.Save(restaurant))

	_ = paymentRepo
	_ = paymentMethod
	uc := orderCmd.NewCreateOrderUseCase(
		orderRepo,
		restRepo,
		eventBus,
		logger,
	)

	t.Run("Execute Success", func(t *testing.T) {
		// Setup Cart
		cartSvc := cart.NewCartService()
		sessionID := "sess_1"
		item, _ := menu.NewMenuItem("i1", "c1", "r1", "Burger", 10.0)
		require.NoError(t, cartSvc.AddItem(sessionID, item, 2))

		userCart := cartSvc.GetCart(sessionID)

		req := orderCmd.CreateOrderRequest{
			RestaurantID:  "r1",
			SessionID:     sessionID,
			Cart:          userCart,
			PaymentMethod: common.PaymentMethodTypeCash,
		}

		resp, err := uc.Execute(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.OrderID)
		assert.NotEmpty(t, resp.OrderNumber)

		// Verify Order Saved
		savedOrder, _ := orderRepo.FindByID(resp.OrderID)
		assert.Equal(t, common.PaymentStatusPending, savedOrder.PaymentStatus)
		assert.Equal(t, int64(2000), savedOrder.TotalAmount) // 20.0 * 100
		assert.Equal(t, 20.0, savedOrder.FiatAmount)
	})
}
