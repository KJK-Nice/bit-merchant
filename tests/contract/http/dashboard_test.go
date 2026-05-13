package http_test

import (
	"bitmerchant/internal/common"
	dashboard "bitmerchant/internal/dashboard/app/query"

	httpMiddleware "bitmerchant/internal/common/http/middleware"
	dashboardhttp "bitmerchant/internal/dashboard/ports/http"
	"bitmerchant/internal/infrastructure/repositories/memory"
	"bitmerchant/internal/ordering/domain/order"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestDashboardHandler(t *testing.T) {
	// Setup
	orderRepo := memory.NewMemoryOrderRepository()
	restaurantRepo := memory.NewMemoryRestaurantRepository()

	// Seed restaurant
	r, _ := restaurant.NewRestaurant("restaurant_1", "Test Cafe")
	_ = restaurantRepo.Save(r)

	// Seed orders for stats
	items := []order.OrderItem{{MenuItemID: "i1", Name: "Item 1", Quantity: 1, UnitPrice: 10.0, Subtotal: 10.0}}
	o1, _ := order.NewOrder("o1", "1001", "restaurant_1", "session_1", items, 1000, common.PaymentMethodTypeCash)
	o1.PaymentStatus = common.PaymentStatusPaid
	o1.FiatAmount = 10.0
	_ = orderRepo.Save(o1)

	// Use Cases
	getStatsUC := dashboard.NewRestaurantDashboardStatsHandler(orderRepo, nil, nil)
	getHistoryUC := dashboard.NewPaidOrdersForRestaurantHandler(orderRepo, nil, nil)
	getTopItemsUC := dashboard.NewTopSellingMenuItemsHandler(orderRepo, nil, nil, nil)
	getStalledUC := dashboard.NewStalledOrdersHandler(orderRepo, nil, nil)
	getByHourUC := dashboard.NewOrdersByHourHandler(orderRepo, nil, nil)
	toggleOpenUC := restaurantCmd.NewToggleRestaurantOpenHandler(restaurantRepo, nil, nil)
	pauseUC := restaurantCmd.NewPauseRestaurantHandler(restaurantRepo, nil, nil)

	h := dashboardhttp.NewDashboardHandler(getStatsUC, getHistoryUC, getTopItemsUC, getStalledUC, getByHourUC, toggleOpenUC, pauseUC, restaurantRepo, orderRepo, nil, slog.Default())

	e := echo.New()

	t.Run("GET /dashboard", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(httpMiddleware.ContextRestaurantID, common.RestaurantID("restaurant_1"))

		err := h.Dashboard(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Sales Dashboard")
		assert.Contains(t, body, "10.00") // Total Sales
		assert.Contains(t, body, "Restaurant")
		assert.Contains(t, body, "Accepting orders")
	})

	t.Run("POST /dashboard/toggle-open", func(t *testing.T) {
		form := make(url.Values)
		req := httptest.NewRequest(http.MethodPost, "/dashboard/toggle-open", strings.NewReader(form.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(httpMiddleware.ContextRestaurantID, common.RestaurantID("restaurant_1"))

		err := h.ToggleOpen(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, rec.Code)
		loc := rec.Header().Get("Location")
		assert.Contains(t, loc, "/dashboard")

		updated, _ := restaurantRepo.FindByID("restaurant_1")
		assert.False(t, updated.IsOpen)
	})
}
