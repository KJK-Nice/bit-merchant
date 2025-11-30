package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bitmerchant/internal/application/cart"
	"bitmerchant/internal/application/order"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/infrastructure/payment/cash"
	"bitmerchant/internal/infrastructure/repositories/memory"
	handler "bitmerchant/internal/interfaces/http"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestOrderEndpoints(t *testing.T) {
	// Setup Dependencies
	orderRepo := memory.NewMemoryOrderRepository()
	paymentRepo := memory.NewMemoryPaymentRepository()
	restRepo := memory.NewMemoryRestaurantRepository()
	eventBus := events.NewEventBus()
	paymentMethod := cash.NewCashPaymentMethod()
	logger := logging.NewLogger()

	// Seed restaurant
	rest, _ := domain.NewRestaurant("restaurant_1", "Test Restaurant")
	restRepo.Save(rest)
	
	createUC := order.NewCreateOrderUseCase(orderRepo, paymentRepo, restRepo, eventBus, paymentMethod, logger)
	getUC := order.NewGetOrderByNumberUseCase(orderRepo)
	cartService := cart.NewCartService()
	
	h := handler.NewOrderHandler(createUC, getUC, cartService)
	
	e := echo.New()

	t.Run("Get Order", func(t *testing.T) {
		// Setup existing order
		item, _ := domain.NewOrderItem("oi1", "o1", "mi1", "Burger", 1, 10.0)
		existingOrder, _ := domain.NewOrder(
			"o1", 
			"1234", 
			"restaurant_1", 
			[]domain.OrderItem{*item}, 
			1000, 
			domain.PaymentMethodTypeCash,
		)
		existingOrder.FiatAmount = 10.0
		orderRepo.Save(existingOrder)

		req := httptest.NewRequest(http.MethodGet, "/order/1234", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/order/:orderNumber")
		c.SetParamNames("orderNumber")
		c.SetParamValues("1234")

		err := h.GetOrder(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Order #1234")
	})
}

