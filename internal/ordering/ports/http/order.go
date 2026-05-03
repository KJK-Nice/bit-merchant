package http

import (
	"bitmerchant/internal/common"
	commonhttp "bitmerchant/internal/common/http"

	"bitmerchant/internal/interfaces/templates"
	"bitmerchant/internal/ordering/app/cart"

	// OrderHandler handles order-related HTTP requests
	orderCmd "bitmerchant/internal/ordering/app/command"
	orderQuery "bitmerchant/internal/ordering/app/query"
	"fmt"

	"github.com/labstack/echo/v4"
	"net/http"
)

type OrderHandler struct {
	createOrder              orderCmd.CreateOrderHandler
	getCustomerOrderByLookup orderQuery.CustomerOrderByLookupHandler
	getCustomerOrders        orderQuery.CustomerOrdersForSessionHandler
	cartService              *cart.CartService
	vapidPublicKey           string
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(
	createOrder orderCmd.CreateOrderHandler,
	getCustomerOrderByLookup orderQuery.CustomerOrderByLookupHandler,
	getCustomerOrders orderQuery.CustomerOrdersForSessionHandler,
	cartService *cart.CartService,
	vapidPublicKey string,
) *OrderHandler {
	return &OrderHandler{
		createOrder:              createOrder,
		getCustomerOrderByLookup: getCustomerOrderByLookup,
		getCustomerOrders:        getCustomerOrders,
		cartService:              cartService,
		vapidPublicKey:           vapidPublicKey,
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
	if cart.RestaurantID == "" {
		return c.Redirect(http.StatusFound, "/menu")
	}

	return templates.OrderConfirmationPage(cart, string(cart.RestaurantID), commonhttp.CSRFToken(c)).Render(c.Request().Context(), c.Response())
}

// CreateOrder handles POST /order/create
func (h *OrderHandler) CreateOrder(c echo.Context) error {
	sessionID := c.Get("sessionID").(string)
	cart := h.cartService.GetCart(sessionID)

	if len(cart.Items) == 0 {
		return c.Redirect(http.StatusFound, "/menu")
	}

	restaurantID := common.RestaurantID(c.FormValue("restaurantID"))
	if restaurantID == "" || cart.RestaurantID != restaurantID {
		return c.String(http.StatusBadRequest, "Invalid restaurant for this order")
	}

	// Get payment method from form
	paymentMethodVal := c.FormValue("paymentMethod")
	var paymentMethod common.PaymentMethodType
	if paymentMethodVal == "cash" {
		paymentMethod = common.PaymentMethodTypeCash
	} else {
		// Default or Error
		paymentMethod = common.PaymentMethodTypeCash
	}

	req := orderCmd.CreateOrder{
		RestaurantID:  restaurantID,
		SessionID:     sessionID,
		Cart:          cart,
		PaymentMethod: paymentMethod,
	}

	resp, err := h.createOrder.Handle(c.Request().Context(), req)
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

	sessionID, _ := c.Get("sessionID").(string)
	result, cerr := h.getCustomerOrderByLookup.Handle(c.Request().Context(), orderQuery.CustomerOrderByLookup{
		SessionID:   sessionID,
		OrderNumber: orderNumber,
	})
	if cerr != nil {
		if cerr.Error() == "order not found" {
			return c.String(http.StatusNotFound, "Order not found")
		}
		return c.String(http.StatusInternalServerError, cerr.Error())
	}

	return templates.OrderStatusPage(result, h.vapidPublicKey).Render(c.Request().Context(), c.Response())
}

// GetLookup renders the lookup/history page (REPLACED functionality)
func (h *OrderHandler) GetLookup(c echo.Context) error {
	sessionID := c.Get("sessionID").(string)
	orders, err := h.getCustomerOrders.Handle(c.Request().Context(), orderQuery.CustomerOrdersForSession{SessionID: sessionID})
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to retrieve orders: "+err.Error())
	}

	return templates.OrderHistoryPage(orders).Render(c.Request().Context(), c.Response())
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
