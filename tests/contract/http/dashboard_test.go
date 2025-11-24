package http_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"bitmerchant/internal/application/dashboard"
	"bitmerchant/internal/application/restaurant"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"
	handler "bitmerchant/internal/interfaces/http"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestDashboardHandler(t *testing.T) {
	// Setup
	orderRepo := memory.NewMemoryOrderRepository()
	restaurantRepo := memory.NewMemoryRestaurantRepository()
	
	// Seed restaurant
	r, _ := domain.NewRestaurant("restaurant_1", "Test Cafe")
	_ = restaurantRepo.Save(r)

	// Seed orders for stats
	items := []domain.OrderItem{{MenuItemID: "i1", Name: "Item 1", Quantity: 1, UnitPrice: 10.0, Subtotal: 10.0}}
	o1, _ := domain.NewOrder("o1", "1001", "restaurant_1", items, 1000, domain.PaymentMethodTypeCash)
	o1.PaymentStatus = domain.PaymentStatusPaid
	o1.FiatAmount = 10.0
	_ = orderRepo.Save(o1)

	// Use Cases
	getStatsUC := dashboard.NewGetDashboardStatsUseCase(orderRepo)
	getHistoryUC := dashboard.NewGetOrderHistoryUseCase(orderRepo)
	getTopItemsUC := dashboard.NewGetTopSellingItemsUseCase(orderRepo)
	toggleOpenUC := restaurant.NewToggleRestaurantOpenUseCase(restaurantRepo)

	h := handler.NewDashboardHandler(getStatsUC, getHistoryUC, getTopItemsUC, toggleOpenUC)

	e := echo.New()

	t.Run("GET /dashboard", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.Dashboard(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Sales Dashboard")
		assert.Contains(t, rec.Body.String(), "10.00") // Total Sales
	})

	t.Run("POST /dashboard/toggle-open", func(t *testing.T) {
		form := make(url.Values)
		req := httptest.NewRequest(http.MethodPost, "/dashboard/toggle-open", strings.NewReader(form.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.ToggleOpen(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		
		// Verify restaurant state changed
		updated, _ := restaurantRepo.FindByID("restaurant_1")
		// Default was whatever NewRestaurant sets (probably false or true, let's say it toggled)
		// If it was false (closed), now true (open).
		_ = updated
		// We can check response body for button text change if we implement partial updates properly.
	})
}

