package order_test

import (
	"bitmerchant/internal/common"
	"bitmerchant/internal/common/money"

	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/infrastructure/payment/cash"
	"bitmerchant/internal/infrastructure/repositories/memory"
	"bitmerchant/internal/menu/domain/menu"
	"bitmerchant/internal/ordering/app/cart"
	orderCmd "bitmerchant/internal/ordering/app/command"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"context"
	"sync"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateOrderHandler(t *testing.T) {
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
	uc := orderCmd.NewCreateOrderHandler(
		orderRepo,
		restRepo,
		eventBus,
		logger.Logger,
		nil,
	)

	t.Run("Execute Success", func(t *testing.T) {
		// Setup Cart
		cartSvc := cart.NewCartService()
		sessionID := "sess_1"
		item, _ := menu.NewMenuItem("i1", "c1", "r1", "Burger", 10.0)
		require.NoError(t, cartSvc.AddItem(sessionID, item, 2))

		userCart := cartSvc.GetCart(sessionID)

		req := orderCmd.CreateOrder{
			RestaurantID:  "r1",
			SessionID:     sessionID,
			Cart:          userCart,
			PaymentMethod: common.PaymentMethodTypeCash,
		}

		resp, err := uc.Handle(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.OrderID)
		assert.NotEmpty(t, resp.OrderNumber)

		// Verify Order Saved
		savedOrder, _ := orderRepo.FindByID(resp.OrderID)
		assert.Equal(t, common.PaymentStatusPending, savedOrder.PaymentStatus)
		assert.Equal(t, int64(2000), savedOrder.TotalAmount) // 20.0 * 100 (USD cents)
		assert.Equal(t, 20.0, savedOrder.FiatAmount)
		assert.Equal(t, money.USD, savedOrder.Currency)
		assert.Equal(t, "$20.00", savedOrder.Total().Format())
	})
}

// TestCreateOrderHandler_SatoshiRestaurant exercises the niche-defining SAT
// path: a Lightning-priced restaurant with whole-sat amounts. Verifies that
// totals stay integer-clean (no float drift) and format with thousands
// separators and the "sats" suffix.
func TestCreateOrderHandler_SatoshiRestaurant(t *testing.T) {
	orderRepo := memory.NewMemoryOrderRepository()
	restRepo := memory.NewMemoryRestaurantRepository()
	eventBus := events.NewEventBus()
	logger := logging.NewLogger()

	restID := common.RestaurantID("r_sat")
	rest, err := restaurant.NewRestaurantWithCurrency(restID, "Lightning Cafe", money.SAT)
	require.NoError(t, err)
	require.NoError(t, restRepo.Save(rest))

	uc := orderCmd.NewCreateOrderHandler(orderRepo, restRepo, eventBus, logger.Logger, nil)

	cartSvc := cart.NewCartService()
	sessionID := "sess_sat"
	item, err := menu.NewMenuItemWithCurrency("i1", "c1", restID, "Espresso", 5_000, money.SAT)
	require.NoError(t, err)
	require.NoError(t, cartSvc.AddItem(sessionID, item, 3))

	resp, err := uc.Handle(context.Background(), orderCmd.CreateOrder{
		RestaurantID:  restID,
		SessionID:     sessionID,
		Cart:          cartSvc.GetCart(sessionID),
		PaymentMethod: common.PaymentMethodTypeCash,
	})
	require.NoError(t, err)

	saved, err := orderRepo.FindByID(resp.OrderID)
	require.NoError(t, err)
	assert.Equal(t, money.SAT, saved.Currency)
	assert.Equal(t, int64(15_000), saved.TotalAmount, "3 × 5,000 sats = 15,000 sats (whole units, no scale)")
	assert.Equal(t, "15,000 sats", saved.Total().Format())
}

// Regression: replaces a previously-random rand.Intn(10000) generator that
// hit the (restaurant_id, order_number) UNIQUE constraint under concurrent
// load (birthday paradox). With per-restaurant atomic counters, N concurrent
// CreateOrder calls for the same restaurant must yield N distinct numbers.
func TestCreateOrderHandler_ConcurrentNumbersAreUnique(t *testing.T) {
	orderRepo := memory.NewMemoryOrderRepository()
	restRepo := memory.NewMemoryRestaurantRepository()
	eventBus := events.NewEventBus()
	logger := logging.NewLogger()

	restID := common.RestaurantID("r1")
	rest, _ := restaurant.NewRestaurant(restID, "Test Rest")
	require.NoError(t, restRepo.Save(rest))

	uc := orderCmd.NewCreateOrderHandler(orderRepo, restRepo, eventBus, logger.Logger, nil)

	const concurrency = 25
	results := make([]string, concurrency)
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cartSvc := cart.NewCartService()
			sessionID := "sess_" + string(rune('A'+idx))
			item, _ := menu.NewMenuItem("i1", "c1", "r1", "Burger", 10.0)
			require.NoError(t, cartSvc.AddItem(sessionID, item, 1))
			req := orderCmd.CreateOrder{
				RestaurantID:  restID,
				SessionID:     sessionID,
				Cart:          cartSvc.GetCart(sessionID),
				PaymentMethod: common.PaymentMethodTypeCash,
			}
			resp, err := uc.Handle(context.Background(), req)
			require.NoError(t, err)
			results[idx] = string(resp.OrderNumber)
		}(i)
	}
	wg.Wait()

	seen := make(map[string]struct{}, concurrency)
	for _, n := range results {
		assert.NotEmpty(t, n)
		_, dup := seen[n]
		assert.Falsef(t, dup, "duplicate order number %q", n)
		seen[n] = struct{}{}
	}
	assert.Len(t, seen, concurrency)
}
