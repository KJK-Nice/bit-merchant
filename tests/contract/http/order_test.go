package http_test

import (
	"bitmerchant/internal/common"

	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/infrastructure/payment/cash"
	"bitmerchant/internal/infrastructure/repositories/memory"
	"bitmerchant/internal/ordering/app/cart"
	orderCmd "bitmerchant/internal/ordering/app/command"
	orderQuery "bitmerchant/internal/ordering/app/query"
	"bitmerchant/internal/ordering/domain/order"
	orderinghttp "bitmerchant/internal/ordering/ports/http"

	// Setup Dependencies
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrderEndpoints(t *testing.T) {

	orderRepo := memory.NewMemoryOrderRepository()
	paymentRepo := memory.NewMemoryPaymentRepository()
	restRepo := memory.NewMemoryRestaurantRepository()
	eventBus := events.NewEventBus()
	paymentMethod := cash.NewCashPaymentMethod()
	logger := logging.NewLogger()

	// Seed restaurant
	rest, _ := restaurant.NewRestaurant("restaurant_1", "Test Restaurant")
	require.NoError(t, restRepo.Save(rest))

	_ = paymentRepo
	_ = paymentMethod
	createUC := orderCmd.NewCreateOrderHandler(orderRepo, restRepo, eventBus, logger.Logger, nil)
	getCustomerOrderUC := orderQuery.NewCustomerOrderByLookupHandler(orderRepo, nil, nil)
	getCustomerOrdersUC := orderQuery.NewCustomerOrdersForSessionHandler(orderRepo, nil, nil)
	cartService := cart.NewCartService()
	requestServerUC := orderCmd.NewRequestServerHandler(orderRepo, eventBus, logger.Logger, nil)
	requestBillUC := orderCmd.NewRequestBillHandler(orderRepo, eventBus, logger.Logger, nil)

	h := orderinghttp.NewOrderHandler(createUC, getCustomerOrderUC, getCustomerOrdersUC, requestServerUC, requestBillUC, orderRepo, restRepo, cartService, "")

	e := echo.New()

	t.Run("Get Order", func(t *testing.T) {
		// Setup existing order
		item, _ := order.NewOrderItem("oi1", "o1", "mi1", "Burger", 1, 10.0)
		existingOrder, _ := order.NewOrder(
			"o1",
			"1234",
			"restaurant_1",
			"session_1",
			[]order.OrderItem{*item},
			1000,
			common.PaymentMethodTypeCash,
		)
		existingOrder.FiatAmount = 10.0
		require.NoError(t, orderRepo.Save(existingOrder))

		req := httptest.NewRequest(http.MethodGet, "/order/1234", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("sessionID", "session_1")
		c.SetPath("/order/:orderNumber")
		c.SetParamNames("orderNumber")
		c.SetParamValues("1234")

		err := h.GetOrder(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Order #1234")
	})

	t.Run("Order History EmptyStateLinksToMyPlaces", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/order/lookup", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("sessionID", "session-empty")

		err := h.GetLookup(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "href=\"/my-places\"")
	})
}
