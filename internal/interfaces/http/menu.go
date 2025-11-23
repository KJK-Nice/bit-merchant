package http

import (
	"net/http"

	"bitmerchant/internal/application/cart"
	"bitmerchant/internal/application/menu"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/interfaces/templates"

	"github.com/labstack/echo/v4"
)

type MenuHandler struct {
	getMenuUseCase *menu.GetMenuUseCase
	cartService    *cart.CartService
}

func NewMenuHandler(getMenuUseCase *menu.GetMenuUseCase, cartService *cart.CartService) *MenuHandler {
	return &MenuHandler{
		getMenuUseCase: getMenuUseCase,
		cartService:    cartService,
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

	return templates.MenuPage(menuData, cart).Render(c.Request().Context(), c.Response())
}
