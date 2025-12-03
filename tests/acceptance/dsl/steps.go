package dsl

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bitmerchant/internal/domain"

	"github.com/stretchr/testify/require"
)

// Step represents an executable step in a test scenario
type Step interface {
	Execute(t *testing.T, app *TestApplication)
}

// AddToCartStep represents adding an item to cart
type AddToCartStep struct {
	SessionID string
	ItemID    string
	Quantity  int
}

func (s *AddToCartStep) Execute(t *testing.T, app *TestApplication) {
	item, err := app.menuItemRepo.FindByID(domain.ItemID(s.ItemID))
	require.NoError(t, err)
	require.NotNil(t, item)

	err = app.cartService.AddItem(s.SessionID, item, s.Quantity)
	require.NoError(t, err)
}

// CreateOrderStep represents creating an order
type CreateOrderStep struct {
	SessionID string
}

func (s *CreateOrderStep) Execute(t *testing.T, app *TestApplication) {
	req := httptest.NewRequest(http.MethodPost, "/order/create", nil)
	req.AddCookie(&http.Cookie{Name: "bitmerchant_session", Value: s.SessionID})

	rec := httptest.NewRecorder()
	c := app.echo.NewContext(req, rec)
	c.Set("sessionID", s.SessionID)

	err := app.orderHandler.CreateOrder(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusFound, rec.Code)

	// Extract order ID and number from repository
	restaurantID := domain.RestaurantID("restaurant_1")
	orders, err := app.orderRepo.FindByRestaurantID(restaurantID)
	require.NoError(t, err)
	require.NotEmpty(t, orders, "Order should be created")

	// Store the most recent order (last one)
	latestOrder := orders[len(orders)-1]
	if app.context != nil {
		app.context.SetCreatedOrder(latestOrder.ID, latestOrder.OrderNumber)
	}

	// Give time for event processing
	time.Sleep(100 * time.Millisecond)
}

// ViewOrderStep represents viewing an order
type ViewOrderStep struct {
	SessionID   string
	OrderNumber string
}

func (s *ViewOrderStep) Execute(t *testing.T, app *TestApplication) {
	req := httptest.NewRequest(http.MethodGet, "/order/"+s.OrderNumber, nil)
	req.AddCookie(&http.Cookie{Name: "bitmerchant_session", Value: s.SessionID})

	rec := httptest.NewRecorder()
	c := app.echo.NewContext(req, rec)
	c.Set("sessionID", s.SessionID)
	c.SetPath("/order/:orderNumber")
	c.SetParamNames("orderNumber")
	c.SetParamValues(s.OrderNumber)

	err := app.orderHandler.GetOrder(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
}

// ViewKitchenDashboardStep represents viewing kitchen dashboard
type ViewKitchenDashboardStep struct{}

func (s *ViewKitchenDashboardStep) Execute(t *testing.T, app *TestApplication) {
	req := httptest.NewRequest(http.MethodGet, "/kitchen", nil)
	rec := httptest.NewRecorder()
	c := app.echo.NewContext(req, rec)

	err := app.kitchenHandler.GetKitchen(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
}

// MarkOrderPaidStep represents marking an order as paid
type MarkOrderPaidStep struct {
	OrderID string // If empty, uses context's created order ID
}

func (s *MarkOrderPaidStep) Execute(t *testing.T, app *TestApplication) {
	orderID := s.OrderID
	if orderID == "" && app.context != nil {
		orderID = string(app.context.GetCreatedOrderID())
	}
	require.NotEmpty(t, orderID, "OrderID must be provided or order must be created first")

	req := httptest.NewRequest(http.MethodPost, "/kitchen/order/"+orderID+"/mark-paid", nil)
	rec := httptest.NewRecorder()
	c := app.echo.NewContext(req, rec)
	c.SetPath("/kitchen/order/:id/mark-paid")
	c.SetParamNames("id")
	c.SetParamValues(orderID)

	err := app.kitchenHandler.MarkPaid(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	time.Sleep(100 * time.Millisecond)
}

// MarkOrderPreparingStep represents marking an order as preparing
type MarkOrderPreparingStep struct {
	OrderID string // If empty, uses context's created order ID
}

func (s *MarkOrderPreparingStep) Execute(t *testing.T, app *TestApplication) {
	orderID := s.OrderID
	if orderID == "" && app.context != nil {
		orderID = string(app.context.GetCreatedOrderID())
	}
	require.NotEmpty(t, orderID, "OrderID must be provided or order must be created first")

	req := httptest.NewRequest(http.MethodPost, "/kitchen/order/"+orderID+"/mark-preparing", nil)
	rec := httptest.NewRecorder()
	c := app.echo.NewContext(req, rec)
	c.SetPath("/kitchen/order/:id/mark-preparing")
	c.SetParamNames("id")
	c.SetParamValues(orderID)

	err := app.kitchenHandler.MarkPreparing(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	// Give more time for event processing and state updates
	time.Sleep(200 * time.Millisecond)
}

// MarkOrderReadyStep represents marking an order as ready
type MarkOrderReadyStep struct {
	OrderID string // If empty, uses context's created order ID
}

func (s *MarkOrderReadyStep) Execute(t *testing.T, app *TestApplication) {
	orderID := s.OrderID
	if orderID == "" && app.context != nil {
		orderID = string(app.context.GetCreatedOrderID())
	}
	require.NotEmpty(t, orderID, "OrderID must be provided or order must be created first")

	req := httptest.NewRequest(http.MethodPost, "/kitchen/order/"+orderID+"/mark-ready", nil)
	rec := httptest.NewRecorder()
	c := app.echo.NewContext(req, rec)
	c.SetPath("/kitchen/order/:id/mark-ready")
	c.SetParamNames("id")
	c.SetParamValues(orderID)

	err := app.kitchenHandler.MarkReady(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	time.Sleep(100 * time.Millisecond)

	// Capture SSE events
	captureSSEEvents(t, app)
}

// captureSSEEvents captures SSE events that were broadcast
func captureSSEEvents(t *testing.T, app *TestApplication) {
	if app.testSSEHandler == nil || app.context == nil {
		return
	}

	// Capture events for kitchen stream
	kitchenClient := app.context.GetSSEClient("/kitchen/stream")
	if kitchenClient != nil {
		broadcasts := app.testSSEHandler.GetCapturedBroadcasts("kitchen")
		for _, msg := range broadcasts {
			kitchenClient.CaptureEvent(msg)
		}
	}

	// Capture events for order stream
	if app.context.GetCreatedOrderNumber() != "" {
		orderTopic := fmt.Sprintf("order:%s", app.context.GetCreatedOrderNumber())
		orderPath := fmt.Sprintf("/order/%s/stream", app.context.GetCreatedOrderNumber())
		orderClient := app.context.GetSSEClient(orderPath)
		if orderClient != nil {
			broadcasts := app.testSSEHandler.GetCapturedBroadcasts(orderTopic)
			for _, msg := range broadcasts {
				orderClient.CaptureEvent(msg)
			}
		}
	}
}

// PublishEventStep represents publishing a domain event
type PublishEventStep struct {
	Event interface{}
}

func (s *PublishEventStep) Execute(t *testing.T, app *TestApplication) {
	ctx := context.Background()
	var topic string

	switch s.Event.(type) {
	case domain.OrderCreated:
		topic = domain.EventOrderCreated
	case domain.OrderPaid:
		topic = domain.EventOrderPaid
	case domain.OrderPreparing:
		topic = domain.EventOrderPreparing
	case domain.OrderReady:
		topic = domain.EventOrderReady
	default:
		t.Fatalf("Unknown event type: %T", s.Event)
	}

	err := app.eventBus.Publish(ctx, topic, s.Event)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Capture SSE events
	captureSSEEvents(t, app)
}
