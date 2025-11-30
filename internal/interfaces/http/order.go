package http

import (
	"fmt"
	"net/http"

	"bitmerchant/internal/application/cart"
	"bitmerchant/internal/application/order"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/interfaces/templates"

	"github.com/labstack/echo/v4"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	createOrderUseCase      *order.CreateOrderUseCase
	getOrderByNumberUseCase *order.GetOrderByNumberUseCase
	cartService             *cart.CartService
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(
	createOrderUseCase *order.CreateOrderUseCase,
	getOrderByNumberUseCase *order.GetOrderByNumberUseCase,
	cartService *cart.CartService,
) *OrderHandler {
	return &OrderHandler{
		createOrderUseCase:      createOrderUseCase,
		getOrderByNumberUseCase: getOrderByNumberUseCase,
		cartService:             cartService,
	}
}

// GetConfirmOrder handles GET /order/confirm
func (h *OrderHandler) GetConfirmOrder(c echo.Context) error {
	sessionID := c.Get("sessionID").(string)
	cart := h.cartService.GetCart(sessionID)
	
	// Validate cart not empty
	if len(cart.Items) == 0 {
		return c.Redirect(http.StatusFound, "/menu")
	}

	return templates.OrderConfirmationPage(cart).Render(c.Request().Context(), c.Response())
}

// CreateOrder handles POST /order/create
func (h *OrderHandler) CreateOrder(c echo.Context) error {
	sessionID := c.Get("sessionID").(string)
	cart := h.cartService.GetCart(sessionID)

	if len(cart.Items) == 0 {
		return c.Redirect(http.StatusFound, "/menu")
	}

	restaurantID := domain.RestaurantID("restaurant_1") // Default for MVP
	
	// Get payment method from form
	paymentMethodVal := c.FormValue("paymentMethod")
	var paymentMethod domain.PaymentMethodType
	if paymentMethodVal == "cash" {
		paymentMethod = domain.PaymentMethodTypeCash
	} else {
		// Default or Error
		paymentMethod = domain.PaymentMethodTypeCash
	}

	req := order.CreateOrderRequest{
		RestaurantID:  restaurantID,
		SessionID:     sessionID,
		Cart:          cart,
		PaymentMethod: paymentMethod,
	}

	resp, err := h.createOrderUseCase.Execute(c.Request().Context(), req)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create order: "+err.Error())
	}

	// Clear cart
	h.cartService.ClearCart(sessionID)

	// Redirect to status page
	return c.Redirect(http.StatusFound, "/order/"+string(resp.OrderNumber))
}

// GetOrder handles GET /order/:orderNumber
func (h *OrderHandler) GetOrder(c echo.Context) error {
	orderNumber := c.Param("orderNumber")
	if orderNumber == "" {
		return c.String(http.StatusBadRequest, "Order number required")
	}

	restaurantID := domain.RestaurantID("restaurant_1") // Default for MVP

	result, err := h.getOrderByNumberUseCase.Execute(c.Request().Context(), restaurantID, orderNumber)
	if err != nil {
		if err.Error() == "order not found" {
			return c.String(http.StatusNotFound, "Order not found")
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return templates.OrderStatusPage(result).Render(c.Request().Context(), c.Response())
}

// GetLookup renders the lookup page
func (h *OrderHandler) GetLookup(c echo.Context) error {
	return templates.OrderLookupPage().Render(c.Request().Context(), c.Response())
}

// PostLookup handles the lookup form submission
func (h *OrderHandler) PostLookup(c echo.Context) error {
	orderNumber := c.FormValue("orderNumber")
	if orderNumber == "" {
		return c.Redirect(http.StatusSeeOther, "/order/lookup")
	}
	// Redirect to status page
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/order/%s", orderNumber))
}
