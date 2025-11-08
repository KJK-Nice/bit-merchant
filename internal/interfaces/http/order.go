package http

import (
	"net/http"

	"bitmerchant/internal/application/order"
	"bitmerchant/internal/domain"

	"github.com/labstack/echo/v4"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	getOrderUseCase *order.GetOrderByNumberUseCase
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(getOrderUseCase *order.GetOrderByNumberUseCase) *OrderHandler {
	return &OrderHandler{
		getOrderUseCase: getOrderUseCase,
	}
}

// GetOrder handles GET /order/:orderNumber
func (h *OrderHandler) GetOrder(c echo.Context) error {
	orderNumber := c.Param("orderNumber")
	if orderNumber == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "order number is required"})
	}

	// Get restaurant ID (v1.0 single tenant)
	restaurantID := domain.RestaurantID("rest_001") // TODO: Get from config

	result, err := h.getOrderUseCase.Execute(restaurantID, domain.OrderNumber(orderNumber))
	if err != nil {
		if err.Error() == "order not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}
