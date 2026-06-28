package http

import (
	"bitmerchant/internal/common"
	commonhttp "bitmerchant/internal/common/http"

	"bitmerchant/internal/interfaces/templates"
	"bitmerchant/internal/ordering/app/cart"

	// OrderHandler handles order-related HTTP requests
	orderCmd "bitmerchant/internal/ordering/app/command"
	orderQuery "bitmerchant/internal/ordering/app/query"
	"bitmerchant/internal/ordering/domain/order"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"fmt"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"net/http"
)

type OrderHandler struct {
	createOrder              orderCmd.CreateOrderHandler
	getCustomerOrderByLookup orderQuery.CustomerOrderByLookupHandler
	getCustomerOrders        orderQuery.CustomerOrdersForSessionHandler
	requestServer            orderCmd.RequestServerHandler
	requestBill              orderCmd.RequestBillHandler
	orderRepo                order.Repository
	restRepo                 restaurant.Repository
	cartService              *cart.CartService
	vapidPublicKey           string
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(
	createOrder orderCmd.CreateOrderHandler,
	getCustomerOrderByLookup orderQuery.CustomerOrderByLookupHandler,
	getCustomerOrders orderQuery.CustomerOrdersForSessionHandler,
	requestServer orderCmd.RequestServerHandler,
	requestBill orderCmd.RequestBillHandler,
	orderRepo order.Repository,
	restRepo restaurant.Repository,
	cartService *cart.CartService,
	vapidPublicKey string,
) *OrderHandler {
	return &OrderHandler{
		createOrder:              createOrder,
		getCustomerOrderByLookup: getCustomerOrderByLookup,
		getCustomerOrders:        getCustomerOrders,
		requestServer:            requestServer,
		requestBill:              requestBill,
		orderRepo:                orderRepo,
		restRepo:                 restRepo,
		cartService:              cartService,
		vapidPublicKey:           vapidPublicKey,
	}
}

// GetConfirmOrder handles GET /order/confirm
func (h *OrderHandler) GetConfirmOrder(c echo.Context) error {
	sessionID := c.Get("sessionID").(string)
	cart := h.cartService.GetCart(sessionID)

	if len(cart.Items) == 0 || cart.RestaurantID == "" {
		return c.Redirect(http.StatusFound, "/menu")
	}

	rest, err := h.restRepo.FindByID(cart.RestaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load restaurant: "+err.Error())
	}

	tableLabel := strings.TrimSpace(c.QueryParam("table"))

	return templates.OrderConfirmationPage(
		cart,
		string(cart.RestaurantID),
		rest.Name,
		tableLabel,
		rest.TaxRate,
		orderQuery.DefaultPrepTarget,
		commonhttp.CSRFToken(c),
		"",
	).Render(c.Request().Context(), c.Response())
}

// CreateOrder handles POST /order/create
func (h *OrderHandler) CreateOrder(c echo.Context) error {
	sessionID := c.Get("sessionID").(string)
	currentCart := h.cartService.GetCart(sessionID)

	if len(currentCart.Items) == 0 {
		return c.Redirect(http.StatusFound, "/menu")
	}

	req, ferr := parseCreateOrderForm(c, currentCart, sessionID)
	if ferr != nil {
		return h.handleCreateOrderFormError(c, currentCart, ferr)
	}

	resp, err := h.createOrder.Handle(c.Request().Context(), *req)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create order: "+err.Error())
	}

	h.cartService.ClearCart(sessionID)

	return c.Redirect(http.StatusFound, "/order/"+string(resp.OrderNumber))
}

// createOrderFormError carries a parse/validation failure plus the form values
// (restaurantID is set when we have enough context to re-render the page).
type createOrderFormError struct {
	httpStatus    int
	publicMessage string
	rerender      bool
	restaurantID  common.RestaurantID
}

func (e *createOrderFormError) Error() string { return e.publicMessage }

// parseCreateOrderForm validates the confirm-page form and returns the command
// payload. Returns a structured *createOrderFormError so the handler can decide
// between rendering an inline page error and a plain status response.
func parseCreateOrderForm(c echo.Context, currentCart *cart.Cart, sessionID string) (*orderCmd.CreateOrder, *createOrderFormError) {
	restaurantID := common.RestaurantID(c.FormValue("restaurantID"))
	if restaurantID == "" || currentCart.RestaurantID != restaurantID {
		return nil, &createOrderFormError{httpStatus: http.StatusBadRequest, publicMessage: "Invalid restaurant for this order"}
	}

	customerName := strings.TrimSpace(c.FormValue("customerName"))
	if customerName == "" {
		return nil, &createOrderFormError{httpStatus: http.StatusBadRequest, publicMessage: "Name for pickup is required.", rerender: true, restaurantID: restaurantID}
	}

	tipPercent, terr := parseTipPercent(c.FormValue("tipPercent"))
	if terr != nil {
		return nil, &createOrderFormError{httpStatus: http.StatusBadRequest, publicMessage: terr.Error()}
	}

	if v := c.FormValue("paymentMethod"); v != "" && v != "cash" {
		return nil, &createOrderFormError{httpStatus: http.StatusBadRequest, publicMessage: "Unsupported payment method"}
	}

	return &orderCmd.CreateOrder{
		RestaurantID:  restaurantID,
		SessionID:     sessionID,
		Cart:          currentCart,
		PaymentMethod: common.PaymentMethodTypeCash,
		CustomerName:  customerName,
		TableLabel:    strings.TrimSpace(c.FormValue("table")),
		TipPercent:    tipPercent,
	}, nil
}

func parseTipPercent(raw string) (int, error) {
	if raw == "" {
		return cart.DefaultTipPercent, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || !cart.IsAllowedTipPercent(n) {
		return 0, fmt.Errorf("invalid tip percent")
	}
	return n, nil
}

func (h *OrderHandler) handleCreateOrderFormError(c echo.Context, currentCart *cart.Cart, ferr *createOrderFormError) error {
	if ferr.rerender {
		return h.rerenderConfirmWithError(c, currentCart, ferr.restaurantID, ferr.publicMessage)
	}
	return c.String(ferr.httpStatus, ferr.publicMessage)
}

// rerenderConfirmWithError re-renders the confirm page with an inline error message and a 400 status.
func (h *OrderHandler) rerenderConfirmWithError(c echo.Context, currentCart *cart.Cart, restaurantID common.RestaurantID, errMsg string) error {
	rest, err := h.restRepo.FindByID(restaurantID)
	if err != nil {
		return c.String(http.StatusBadRequest, errMsg)
	}
	tableLabel := strings.TrimSpace(c.FormValue("table"))
	c.Response().WriteHeader(http.StatusBadRequest)
	return templates.OrderConfirmationPage(
		currentCart,
		string(restaurantID),
		rest.Name,
		tableLabel,
		rest.TaxRate,
		orderQuery.DefaultPrepTarget,
		commonhttp.CSRFToken(c),
		errMsg,
	).Render(c.Request().Context(), c.Response())
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

	view, verr := orderQuery.BuildOrderStatusView(h.orderRepo, result, orderQuery.DefaultPrepTarget)
	if verr != nil {
		return c.String(http.StatusInternalServerError, verr.Error())
	}

	return templates.OrderStatusPage(view, h.vapidPublicKey).Render(c.Request().Context(), c.Response())
}

// resolveCustomerOrder loads the order for the requesting session by order number.
// Scoping by sessionID ensures only the customer who placed the order can act on it.
func (h *OrderHandler) resolveCustomerOrder(c echo.Context) (*order.Order, error) {
	orderNumber := c.Param("orderNumber")
	if orderNumber == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Order number required")
	}
	sessionID, _ := c.Get("sessionID").(string)
	result, err := h.getCustomerOrderByLookup.Handle(c.Request().Context(), orderQuery.CustomerOrderByLookup{
		SessionID:   sessionID,
		OrderNumber: orderNumber,
	})
	if err != nil {
		if err.Error() == "order not found" {
			return nil, echo.NewHTTPError(http.StatusNotFound, "Order not found")
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return result, nil
}

// CallServer handles POST /order/:orderNumber/call-server. Idempotent — repeated
// taps within the throttle window are no-ops at the command layer.
func (h *OrderHandler) CallServer(c echo.Context) error {
	o, err := h.resolveCustomerOrder(c)
	if err != nil {
		return err
	}
	if _, cerr := h.requestServer.Handle(c.Request().Context(), orderCmd.RequestServer{OrderID: o.ID}); cerr != nil {
		return c.String(http.StatusInternalServerError, cerr.Error())
	}
	return c.NoContent(http.StatusOK)
}

// RequestBill handles POST /order/:orderNumber/request-bill.
func (h *OrderHandler) RequestBill(c echo.Context) error {
	o, err := h.resolveCustomerOrder(c)
	if err != nil {
		return err
	}
	if _, cerr := h.requestBill.Handle(c.Request().Context(), orderCmd.RequestBill{OrderID: o.ID}); cerr != nil {
		return c.String(http.StatusInternalServerError, cerr.Error())
	}
	return c.NoContent(http.StatusOK)
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
