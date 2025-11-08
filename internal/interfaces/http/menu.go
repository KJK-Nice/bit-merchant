package http

import (
	"net/http"

	"bitmerchant/internal/application/menu"
	"bitmerchant/internal/domain"

	"github.com/labstack/echo/v4"
)

// MenuHandler handles menu-related HTTP requests
type MenuHandler struct {
	getMenuUseCase *menu.GetMenuUseCase
}

// NewMenuHandler creates a new MenuHandler
func NewMenuHandler(getMenuUseCase *menu.GetMenuUseCase) *MenuHandler {
	return &MenuHandler{
		getMenuUseCase: getMenuUseCase,
	}
}

// GetMenu handles GET /menu
func (h *MenuHandler) GetMenu(c echo.Context) error {
	// Get restaurant ID from context or query param
	// For v1.0, single tenant - get from env or default
	restaurantID := domain.RestaurantID("rest_001") // TODO: Get from config

	result, err := h.getMenuUseCase.Execute(restaurantID)
	if err != nil {
		if err.Error() == "restaurant not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "restaurant not found"})
		}
		if err.Error() == "restaurant is closed" {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "restaurant is closed"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}
