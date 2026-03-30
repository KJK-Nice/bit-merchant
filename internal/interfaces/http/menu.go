package http

import (
	"log/slog"
	"net/http"

	"bitmerchant/internal/application/cart"
	"bitmerchant/internal/application/menu"
	"bitmerchant/internal/application/places"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/interfaces/templates"

	"github.com/labstack/echo/v4"
)

type MenuHandler struct {
	getMenuUseCase *menu.GetMenuUseCase
	cartService    *cart.CartService
	recordVisitUC  *places.RecordMenuVisitUseCase
}

func NewMenuHandler(getMenuUseCase *menu.GetMenuUseCase, cartService *cart.CartService, recordVisitUC *places.RecordMenuVisitUseCase) *MenuHandler {
	return &MenuHandler{
		getMenuUseCase: getMenuUseCase,
		cartService:    cartService,
		recordVisitUC:  recordVisitUC,
	}
}

func (h *MenuHandler) GetMenu(c echo.Context) error {
	restaurantID := c.QueryParam("restaurantID")
	if restaurantID == "" {
		restaurantID = "restaurant_1" // Default for MVP
	}

	// Get Menu Data
	menuData, err := h.getMenuUseCase.Execute(c.Request().Context(), domain.RestaurantID(restaurantID))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Get Cart Data
	sessionID := c.Get("sessionID").(string)
	cart := h.cartService.GetCart(sessionID)

	if h.recordVisitUC != nil {
		if err := h.recordVisitUC.Execute(c.Request().Context(), sessionID, domain.RestaurantID(restaurantID)); err != nil {
			slog.Warn("record menu visit failed", "error", err, "restaurantID", restaurantID)
		}
	}

	tableLabel := c.QueryParam("table")

	// Prevent caching so back button always fetches fresh state (updated cart)
	c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")

	return templates.MenuPage(menuData, cart, tableLabel).Render(c.Request().Context(), c.Response())
}
